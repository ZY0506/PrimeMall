package oss

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

// OssConfig 配置结构（增加可选字段）
type OssConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	BucketName      string
	Host            string   // 前端上传的 endpoint，例如 https://bucket.oss-cn-hangzhou.aliyuncs.com
	UploadDir       string   // 目录前缀，例如 "uploads/" （内部会自动补斜杠）
	ExpireTime      int64    // 凭证有效期（秒）
	CallbackUrl     string   // 回调地址，为空表示不使用回调
	MaxFileSize     int64    // 最大文件大小（字节），0或小于0则使用默认 10MB
	AllowedExts     []string // 允许的扩展名（不带点，如 []string{"jpg","png"}），为空跳过校验
}

// PolicyToken 返回给前端的上传凭证
type PolicyToken struct {
	AccessKeyId     string `json:"access_key_id"`
	Host            string `json:"host"`
	Policy          string `json:"policy"`
	Signature       string `json:"signature"`
	Expire          string `json:"expire"`
	ExpireTimestamp int64  `json:"expire_timestamp"` // Unix 时间戳，便于前端计算剩余时间
	Dir             string `json:"dir"`
	ObjectKey       string `json:"object_key"` // 前端必须使用的完整对象路径
	Callback        string `json:"callback"`
}

// 默认文件大小 10MB
const defaultMaxFileSize = 10 << 20

// GeneratePolicyToken 生成前端直传 OSS 所需的 Policy 和签名（支持回调）
func GeneratePolicyToken(cfg OssConfig, originalFilename string, userId uint64) (*PolicyToken, error) {
	// 1. 校验并规范化上传目录
	dir := normalizeUploadDir(cfg.UploadDir)

	// 2. 校验扩展名白名单
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if len(cfg.AllowedExts) > 0 {
		if !isExtAllowed(ext, cfg.AllowedExts) {
			return nil, errors.New("unsupported file type")
		}
	}

	// 3. 生成唯一文件名
	uniqueFilename := uuid.New().String() + ext
	objectKey := fmt.Sprintf("%suser_%d/%s", dir, userId, uniqueFilename)

	// 4. 有效期
	expireTime := time.Now().Unix() + cfg.ExpireTime
	expireIsoTime := time.Unix(expireTime, 0).UTC().Format("2006-01-02T15:04:05Z")

	// 5. 处理回调（若配置了回调地址）
	var callbackBase64 string
	if cfg.CallbackUrl != "" {
		// 使用 x-www-form-urlencoded 格式，避免 JSON 解析问题
		callbackBody := "bucket=${bucket}&object=${object}&etag=${etag}&size=${size}"
		callbackMap := map[string]string{
			"callbackUrl":      cfg.CallbackUrl,
			"callbackBody":     callbackBody,
			"callbackBodyType": "application/x-www-form-urlencoded",
		}
		callbackJSON, err := json.Marshal(callbackMap)
		if err != nil {
			logx.Errorf("marshal callback failed: %v", err)
			return nil, err
		}
		logx.Infof("callbackJSON=%s", string(callbackJSON))
		callbackBase64 = base64.StdEncoding.EncodeToString(callbackJSON)
	}

	// 6. 构造 Policy conditions（不再包含 $callback 条件）
	maxSize := cfg.MaxFileSize
	if maxSize <= 0 {
		maxSize = defaultMaxFileSize
	}

	conditions := []interface{}{
		[]interface{}{"content-length-range", 0, maxSize},
		[]interface{}{"eq", "$key", objectKey},
		[]interface{}{"eq", "$success_action_status", "200"},
	}

	policyDoc := map[string]interface{}{
		"expiration": expireIsoTime,
		"conditions": conditions,
	}
	policyBytes, err := json.Marshal(policyDoc)
	if err != nil {
		logx.Errorf("marshal policy failed: %v", err)
		return nil, err
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyBytes)

	// 7. 签名
	mac := hmac.New(sha1.New, []byte(cfg.AccessKeySecret))
	mac.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 8. 返回凭证
	return &PolicyToken{
		AccessKeyId:     cfg.AccessKeyId,
		Host:            cfg.Host,
		Policy:          policyBase64,
		Signature:       signature,
		Expire:          expireIsoTime,
		ExpireTimestamp: expireTime,
		Dir:             dir,
		ObjectKey:       objectKey,
		Callback:        callbackBase64, // 仍然返回给前端，但 Policy 不再强制校验
	}, nil
}

// normalizeUploadDir 确保目录以 / 结尾，且不以 / 开头（OSS 对象 key 不应以 / 开头）
func normalizeUploadDir(dir string) string {
	if dir == "" {
		return ""
	}
	// 去掉开头的斜杠
	dir = strings.TrimPrefix(dir, "/")
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	return dir
}

// isExtAllowed 检查扩展名是否在白名单内（白名单存储不带点的格式，如 "jpg"）
func isExtAllowed(ext string, allowedExts []string) bool {
	if ext == "" {
		return false
	}
	// ext 以 "." 开头，例如 ".jpg"，去掉点后比较
	extNoDot := ext[1:]
	for _, allowed := range allowedExts {
		if strings.EqualFold(extNoDot, allowed) {
			return true
		}
	}
	return false
}

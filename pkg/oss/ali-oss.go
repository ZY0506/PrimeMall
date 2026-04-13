package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type OssConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	BucketName      string
	Host            string
	UploadDir       string
	ExpireTime      int64
	CallbackUrl     string
}

// PolicyToken 返回给前端的上传凭证
type PolicyToken struct {
	AccessKeyId string `json:"access_key_id"`
	Host        string `json:"host"`
	Policy      string `json:"policy"`
	Signature   string `json:"signature"`
	Expire      string `json:"expire"`
	Dir         string `json:"dir"`
	ObjectKey   string `json:"object_key"` // 前端必须使用的完整对象路径
}

// GeneratePolicyToken 生成前端直传 OSS 所需的 Policy 和签名
func GeneratePolicyToken(cfg OssConfig, originalFilename string) (*PolicyToken, error) {
	// 1. 生成唯一文件名
	ext := filepath.Ext(originalFilename)
	uniqueFilename := uuid.New().String() + ext
	objectKey := cfg.UploadDir + uniqueFilename

	// 2. 计算过期时间
	expireTime := time.Now().Unix() + cfg.ExpireTime
	expireIsoTime := time.Unix(expireTime, 0).UTC().Format("2006-01-02T15:04:05Z")

	// 3. 构造回调参数 (使用 JSON 格式)
	//callbackBody := fmt.Sprintf(
	//	`{"bucket":"${bucket}","object":"${object}","etag":"${etag}","size":${size},"mimeType":"${mimeType}"}`,
	//)
	//callbackMap := map[string]string{
	//	"callbackUrl":      cfg.CallbackUrl,
	//	"callbackBody":     callbackBody,
	//	"callbackBodyType": "application/json",
	//}
	//callbackJSON, _ := json.Marshal(callbackMap)
	//callbackBase64 := base64.StdEncoding.EncodeToString(callbackJSON)

	// 4. 构造 Policy 文档
	policyDoc := map[string]interface{}{
		"expiration": expireIsoTime,
		"conditions": []interface{}{
			[]interface{}{"content-length-range", 0, 10485760}, // 限制文件最大 10MB
			[]interface{}{"eq", "$key", objectKey},             // 强制使用后端生成的文件名
			[]interface{}{"eq", "$success_action_status", "200"},
			//[]interface{}{"eq", "$callback", callbackBase64}, // 限制回调参数，暂时先不做回调
		},
	}
	policyBytes, err := json.Marshal(policyDoc)
	if err != nil {
		logx.Errorf("marshal policy failed: %v", err)
		return nil, err
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyBytes)

	// 5. 计算签名 (HMAC-SHA1)
	mac := hmac.New(sha1.New, []byte(cfg.AccessKeySecret))
	mac.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 6. 返回凭证
	return &PolicyToken{
		AccessKeyId: cfg.AccessKeyId,
		Host:        cfg.Host,
		Policy:      policyBase64,
		Signature:   signature,
		Expire:      expireIsoTime,
		Dir:         cfg.UploadDir,
		ObjectKey:   objectKey,
	}, nil
}

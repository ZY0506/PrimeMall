package ctxdata

import (
	"context"
	"encoding/json"
	"github.com/ZY0506/PrimeMall/common/response"
	"google.golang.org/grpc/metadata"
	"strconv"
)

const (
	ContextKeyClientInfo      = "client_info"
	ContextKeyUserId          = "user_id"
	ContextKeyAdminId         = "admin_id"
	ContextKeyClientIp        = "client_ip"
	ContextKeyDeviceID        = "client_device_id"
	ContextKeyClientUserAgent = "client_user_agent"
)

type ClientInfo struct {
	IP        string
	UserAgent string
	DeviceID  string
}

// GetUserIdFromCtx 从 context 中获取登录用户的 ID
func GetUserIdFromCtx(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	uids := md.Get(ContextKeyUserId)
	if len(uids) == 0 {
		return 0, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	uid, err := strconv.ParseUint(uids[0], 10, 64)
	if err != nil {
		return 0, response.NewBizError(response.ErrCodeInvalidParam, "参数错误")
	}
	return uid, nil
}

// GetClientInfoFromCtx 从 context 中获取 ClientInfo
func GetClientInfoFromCtx(ctx context.Context) (*ClientInfo, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	ip := md.Get(ContextKeyClientIp)
	userAgent := md.Get(ContextKeyClientUserAgent)
	deviceID := md.Get(ContextKeyDeviceID)
	if len(ip) == 0 || len(userAgent) == 0 || len(deviceID) == 0 {
		return nil, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	return &ClientInfo{
		IP:        ip[0],
		UserAgent: userAgent[0],
		DeviceID:  deviceID[0],
	}, nil
}

// PutUserIdToCtx 将用户 ID 放入 context 中
func PutUserIdToCtx(ctx context.Context, uid uint64) (context.Context, error) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy() // 避免修改原始 map
	}
	md.Set(ContextKeyUserId, strconv.FormatUint(uid, 10))
	return metadata.NewOutgoingContext(ctx, md), nil
}

func ParseAndPutUserIdToCtx(ctx context.Context) (context.Context, error) {
	uid, ok := ctx.Value(ContextKeyUserId).(json.Number)
	if !ok {
		return ctx, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	userId, _ := uid.Int64()
	return PutUserIdToCtx(ctx, uint64(userId))
}

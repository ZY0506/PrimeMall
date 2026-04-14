package ctxdata

import (
	"context"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
	"strconv"
)

// GetUserIdFromCtx 从 context 中获取登录用户的 ID
func GetUserIdFromCtx(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	uids := md.Get("user_id")
	if len(uids) == 0 {
		return 0, response.NewBizError(response.ErrCodeMissingParam, "缺少必要参数")
	}
	uid, err := strconv.ParseUint(uids[0], 10, 64)
	if err != nil {
		logx.WithContext(ctx).Errorf("参数转换失败，error:%v", err)
		return 0, response.NewBizError(response.ErrCodeInvalidParam, "参数错误")
	}
	return uid, nil
}

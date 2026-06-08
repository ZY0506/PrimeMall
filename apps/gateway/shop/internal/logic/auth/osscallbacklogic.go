// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OssCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewOssCallbackLogic OSS上传回调（服务端自行幂等）
func NewOssCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OssCallbackLogic {
	return &OssCallbackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OssCallbackLogic) OssCallback(req *types.OssCallbackReq) (err error) {
	l.Logger.Infof("接收到 OSS 回调，内容: %v", req)

	// 从object中解析出user_id
	parts := strings.Split(req.Object, "/")
	if len(parts) < 2 {
		l.Logger.Errorf("object格式错误,object=%s", req.Object)
		return errorx.NewBizError(response.ErrCodeInvalidObject, "object格式错误")
	}

	// parts[1] = "user_12345"
	userIDStr := strings.TrimPrefix(parts[1], "user_")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		l.Logger.Errorf("user_id解析失败,err=%v", err)
		return err
	}

	// 注入userID
	l.ctx, err = ctxdata.PutUserIdToCtx(l.ctx, uint64(userID))
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return err
	}

	// 调用rpc更新头像
	_, err = l.svcCtx.UserRpc.UpdateUserInfo(l.ctx, &user.UpdateUserInfoReq{
		Avatar: fmt.Sprintf("%s/%s", l.svcCtx.Config.OSSConfig.Host, req.Object),
	})
	if err != nil {
		l.Logger.Errorf("调用RPC更新用户信息失败,err=%v", err)
		return err
	}

	return nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package profile

import (
	"context"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOssTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetOssTokenLogic 获取OSS上传凭证
func NewGetOssTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOssTokenLogic {
	return &GetOssTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOssTokenLogic) GetOssToken(req *types.OssTokenReq) (resp *types.OssTokenResp, err error) {
	// 注入user_id
	l.ctx, err = ctxdata.ParseAndPutUserIdToCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("注入UserID失败，error=%v", err)
		return nil, err
	}

	ossToken, err := l.svcCtx.UserRpc.GetOssToken(l.ctx, &user.OssTokenReq{
		FileName: req.FileName,
	})
	if err != nil {
		l.Logger.Errorf("调用rpc获取OSS上传凭证失败，error=%v", err)
		return nil, err
	}

	return &types.OssTokenResp{
		AccessKeyId: ossToken.AccessId,
		Host:        ossToken.Host,
		Policy:      ossToken.Policy,
		Signature:   ossToken.Signature,
		Expire:      ossToken.Expire.AsTime().Unix(),
		Dir:         ossToken.Dir,
		ObjectKey:   ossToken.ObjectKey,
		Callback:    ossToken.Callback,
	}, nil
}

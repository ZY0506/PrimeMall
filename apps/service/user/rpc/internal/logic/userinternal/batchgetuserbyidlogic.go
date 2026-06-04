package userinternallogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetUserByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetUserByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetUserByIdLogic {
	return &BatchGetUserByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量获取用户基本信息
func (l *BatchGetUserByIdLogic) BatchGetUserById(in *user.BatchGetUserByIdReq) (*user.BatchUserBasicInfoResp, error) {
	// todo: add your logic here and delete this line

	return &user.BatchUserBasicInfoResp{}, nil
}

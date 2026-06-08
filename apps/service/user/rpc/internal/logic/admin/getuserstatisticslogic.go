package adminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatisticsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserStatisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatisticsLogic {
	return &GetUserStatisticsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserStatistics 获取用户统计数据
func (l *GetUserStatisticsLogic) GetUserStatistics(in *user.GetUserStatisticsReq) (*user.GetUserStatisticsResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetUserStatisticsResp{}, nil
}

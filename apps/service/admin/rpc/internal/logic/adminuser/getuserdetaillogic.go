package adminuserlogic

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	userrpc "github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/admin"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDetailLogic {
	return &GetUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserDetailLogic) GetUserDetail(in *admin.AdminGetUserDetailReq) (*admin.AdminGetUserDetailResp, error) {
	resp, err := l.svcCtx.UserRpc.GetUserDetail(l.ctx, &userrpc.GetUserDetailReq{
		UserId: in.UserId,
	})
	if err != nil {
		l.Logger.Errorf("获取用户详情失败,error=%v", err)
		return nil, err
	}

	result := &admin.AdminGetUserDetailResp{
		Id:         resp.UserInfo.Id,
		Phone:      resp.UserInfo.Phone,
		Nickname:   resp.UserInfo.Nickname,
		Avatar:     resp.UserInfo.Avatar,
		Gender:     resp.UserInfo.Gender,
		Status:     int64(resp.UserInfo.Status),
		StatusDesc: getUserStatusDesc(int64(resp.UserInfo.Status)),
	}

	// 处理生日
	if resp.UserInfo.Birthday != nil {
		result.Birthday = resp.UserInfo.Birthday.AsTime().Format(time.RFC3339)
	}

	// 处理创建时间
	if resp.UserInfo.CreatedAt != nil {
		result.CreatedAt = timestamppb.New(resp.UserInfo.CreatedAt.AsTime())
	}

	// 处理最后登录时间
	if resp.UserInfo.LastLoginTime != nil {
		result.LastLoginTime = timestamppb.New(resp.UserInfo.LastLoginTime.AsTime())
	}

	return result, nil
}

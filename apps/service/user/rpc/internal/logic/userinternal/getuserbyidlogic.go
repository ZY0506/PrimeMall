package userinternallogic

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByIdLogic {
	return &GetUserByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserById 根据用户ID获取用户基本信息（用于订单、购物车等）
func (l *GetUserByIdLogic) GetUserById(in *user.GetUserByIdReq) (*user.UserBasicInfo, error) {
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("用户不存在,userId=%d", in.UserId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}

	return &user.UserBasicInfo{
		Id:       userInfo.Id,
		Phone:    userInfo.Phone,
		Nickname: userInfo.Nickname,
		Avatar:   userInfo.Avatar,
		Status:   user.UserStatus(userInfo.Status),
	}, nil
}

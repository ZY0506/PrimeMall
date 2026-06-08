package adminlogic

import (
	"context"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListUsers 获取用户列表（分页、筛选、搜索）
func (l *ListUsersLogic) ListUsers(in *user.ListUsersReq) (*user.ListUsersResp, error) {

	sortFieldMap := map[user.SortField]string{
		user.SortField_SORT_FIELD_UNKNOWN:         "id",
		user.SortField_SORT_FIELD_ID:              "id",
		user.SortField_SORT_FIELD_CREATED_AT:      "created_at",
		user.SortField_SORT_FIELD_LAST_LOGIN_TIME: "last_login_time",
	}
	sortOrder := "desc"
	if in.SortOrder == user.SortOrder_SORT_ORDER_ASC {
		sortOrder = "asc"
	}
	// 查询列表
	listResp, total, err := l.svcCtx.UserModel.FindListByPage(l.ctx, &model.UserListFilter{
		Page:      in.Page,
		PageSize:  in.Size,
		Phone:     in.Phone,
		Nickname:  in.Nickname,
		Status:    int64(in.Status),
		StartTime: in.StartTime.AsTime(),
		EndTime:   in.EndTime.AsTime(),
		SortField: sortFieldMap[in.SortField],
		SortOrder: sortOrder,
	})
	if err != nil {
		l.Logger.Errorf("获取用户列表失败：%v", err)
		return nil, err
	}

	// 封装返回响应
	list := make([]*user.UserListItem, 0)
	for _, userInfo := range listResp {
		list = append(list, &user.UserListItem{
			Id:            userInfo.Id,
			Phone:         userInfo.Phone,
			Nickname:      userInfo.Nickname,
			Avatar:        userInfo.Avatar,
			Gender:        userInfo.Gender,
			Birthday:      timestamppb.New(userInfo.Birthday.Time),
			Status:        user.UserStatus(userInfo.Status),
			CreatedAt:     timestamppb.New(userInfo.CreatedAt),
			LastLoginTime: timestamppb.New(userInfo.LastLoginTime.Time),
			LastLoginIp:   userInfo.LastLoginIp,
		})
	}

	l.Logger.Info("获取用户列表成功")

	return &user.ListUsersResp{
		List:  list,
		Total: total,
	}, nil
}

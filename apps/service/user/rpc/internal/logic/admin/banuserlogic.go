package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type BanUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBanUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BanUserLogic {
	return &BanUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// BanUser 封禁用户（限制下单/封禁账号）
func (l *BanUserLogic) BanUser(in *user.BanUserReq) (*user.EmptyResp, error) {
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("用户不存在,userId=%d", in.UserId)
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("获取用户信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}
	if userInfo.DeletedAt.Valid {
		l.Logger.Errorf("账号已注销,userId=%d", in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "账号已注销")
	}
	if userInfo.Status == constants.USER_STATUS_BANNED {
		l.Logger.Errorf("用户已封禁,userId=%d", in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeUserDisabled, "用户已封禁")
	}
	if userInfo.Status == constants.USER_STATUS_RESTRICTED && in.ActionType == user.BanType_BAN_TYPE_LOGIN {
		l.Logger.Errorf("用户已限制下单,userId=%d", in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeUserDisabled, "用户已限制下单")
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {

		actionType := constants.BAN_TYPE_UNKNOWN
		if in.ActionType == user.BanType_BAN_TYPE_ORDER {
			userInfo.Status = constants.USER_STATUS_RESTRICTED
			actionType = constants.BAN_TYPE_ORDER
		} else {
			userInfo.Status = constants.USER_STATUS_BANNED
			actionType = constants.BAN_TYPE_LOGIN
		}
		// 更新状态
		err = l.svcCtx.UserModel.UpdateTx(ctx, session, userInfo)
		if err != nil {
			l.Logger.Errorf("更新用户状态失败,userId=%d,err=%v", in.UserId, err)
			return err
		}

		// 记录日志
		_, err = l.svcCtx.PunishLogModel.InsertTx(ctx, session, &model.UserPunishLog{
			UserId:     in.UserId,
			Phone:      userInfo.Phone,
			ActionType: int64(actionType),
			Reason:     in.Reason,
			BannedBy:   in.Operator,
			StartTime:  sql.NullTime{Time: time.Now(), Valid: true},
			EndTime:    sql.NullTime{Time: in.EndTime.AsTime(), Valid: true},
			CreatedAt:  time.Now(),
		})
		if err != nil {
			l.Logger.Errorf("记录用户处罚日志失败,userId=%d,err=%v", in.UserId, err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Info("封禁用户成功")
	return &user.EmptyResp{}, nil
}

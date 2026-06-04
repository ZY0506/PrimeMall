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

type UnbanUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnbanUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbanUserLogic {
	return &UnbanUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UnbanUser 解封用户（恢复为正常状态）
func (l *UnbanUserLogic) UnbanUser(in *user.UnbanUserReq) (*user.EmptyResp, error) {

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
	if userInfo.Status == constants.USER_STATUS_NORMAL {
		l.Logger.Errorf("用户已正常,userId=%d", in.UserId)
		return nil, errorx.NewBizError(response.ErrCodeUserUnbanned, "账号已解封")
	}

	// 查找最后一条记录
	log, err := l.svcCtx.PunishLogModel.FindLastLogByUserId(l.ctx, userInfo.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("风控日志不存在,userId=%d", in.UserId)
			return nil, errorx.NewBizError(response.ErrCodeUserPunishLogNotFound, "风控日志不存在")
		}
		l.Logger.Errorf("获取用户处罚信息失败,userId=%d,err=%v", in.UserId, err)
		return nil, err
	}

	// 事务操作解封用户
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		userInfo.Status = constants.USER_STATUS_NORMAL
		err = l.svcCtx.UserModel.UpdateTx(ctx, session, userInfo)
		if err != nil {
			l.Logger.Errorf("更新用户信息失败,userId=%d,err=%v", in.UserId, err)
			return err
		}

		log.EndTime = sql.NullTime{Time: time.Now(), Valid: true}
		log.UnbannedBy = sql.NullString{String: in.Operator, Valid: true}
		err = l.svcCtx.PunishLogModel.UpdateTx(ctx, session, log)
		if err != nil {
			l.Logger.Errorf("更新用户处罚信息失败,userId=%d,err=%v", in.UserId, err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Info("解封用户成功")

	return &user.EmptyResp{}, nil
}

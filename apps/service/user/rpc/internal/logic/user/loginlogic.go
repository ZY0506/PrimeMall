package userlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Login 手机号密码登录
func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {

	// 提前获取客户端信息
	clientInfo, err := ctxdata.GetClientInfoFromCtx(l.ctx)
	if err != nil {
		l.Logger.Infof("获取客户端信息失败,使用默认值,error=%v", err)
		clientInfo = &ctxdata.ClientInfo{IP: "0.0.0.0", UserAgent: ""}
	}

	// 1. 检查用户是否存在
	userInfo, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, in.Phone)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Infof("用户不存在,phone=%s", in.Phone)
			return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "账号或密码错误")
		}
		l.Logger.Errorf("数据库查询失败，error=%v", err)
		return nil, err
	}

	// 2. 检查用户状态（提前检查，避免无效密码比较）
	if userInfo.Status == constants.USER_STATUS_BANNED {
		l.Logger.Errorf("用户被禁用,phone=%v", in.Phone)
		// 获取封禁原因
		failReason := ""
		userPunish, err := l.svcCtx.PunishLogModel.FindLastLogByUserId(l.ctx, userInfo.Id)
		if err != nil && !errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("查询封禁原因失败，error=%v", err)
		} else if userPunish != nil {
			failReason = userPunish.Reason
		}
		l.recordLoginLog(userInfo.Id, clientInfo.IP, clientInfo.UserAgent, constants.LOGIN_STATUS_FAIL, failReason)
		return nil, errorx.NewBizError(response.ErrCodeUserDisabled, "用户被禁用")
	}
	if userInfo.DeletedAt.Valid {
		l.Logger.Errorf("用户已注销,phone=%v", in.Phone)
		l.recordLoginLog(userInfo.Id, clientInfo.IP, clientInfo.UserAgent, constants.LOGIN_STATUS_FAIL, "用户已注销")
		return nil, errorx.NewBizError(response.ErrCodeUserDeleted, "用户已注销")
	}

	// 3. 检查密码
	if !pwd.CompareHashAndPassword(userInfo.Password, in.Password) {
		l.Logger.Infof("密码错误,phone=%v", in.Phone)
		l.recordLoginLog(userInfo.Id, clientInfo.IP, clientInfo.UserAgent, constants.LOGIN_STATUS_FAIL, "密码错误")
		return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "账号或密码错误")
	}

	// 4. 颁发 token
	aToken, err := jwt.GenToken(
		jwt.Claims{UserId: userInfo.Id, JTI: uuid.New().String(), Role: "user"}, []byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.AccessTokenExpire, l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("生成access token失败,error=%v", err)
		return nil, err
	}
	rToken, err := jwt.GenToken(
		jwt.Claims{UserId: userInfo.Id, JTI: uuid.New().String(), Role: "user"}, []byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.RefreshTokenExpire, l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("创建refresh token失败,error=%v", err)
		return nil, err
	}

	l.Logger.Infof("用户登录成功，id:%d", userInfo.Id)

	// 更新最后登录信息（仅成功时）
	go l.updateLastLoginInfo(*userInfo, clientInfo.IP)

	// 记录成功日志
	l.recordLoginLog(userInfo.Id, clientInfo.IP, clientInfo.UserAgent, constants.LOGIN_STATUS_SUCCESS, "")

	return &user.LoginResp{
		AccessToken:   aToken,
		AccessExpire:  timestamppb.New(time.Now().Add(time.Duration(l.svcCtx.Config.JWT.AccessTokenExpire) * time.Second)),
		RefreshToken:  rToken,
		RefreshExpire: timestamppb.New(time.Now().Add(time.Duration(l.svcCtx.Config.JWT.RefreshTokenExpire) * time.Second)),
	}, nil
}

// updateLastLoginInfo 更新最后登录信息
func (l *LoginLogic) updateLastLoginInfo(userInfo model.User, ip string) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("更新登录信息协程panic: %v", r)
		}
	}()
	// 使用带超时的 context，避免后台操作阻塞过久
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userInfo.LastLoginTime = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	userInfo.LastLoginIp = ip

	if err := l.svcCtx.UserModel.Update(ctx, &userInfo); err != nil {
		logx.Errorf("更新用户登录信息失败,error=%v", err)
	}
}

// recordLoginLog 记录登录日志
func (l *LoginLogic) recordLoginLog(userId uint64, ip string, userAgent string, status int64, failReason string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("记录登录日志协程panic: %v", r)
			}
		}()
		// 使用带超时的 context，避免后台操作阻塞过久
		logCtx, logCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer logCancel()
		_, err := l.svcCtx.LoginLogModel.Insert(
			logCtx,
			&model.UserLoginLog{
				UserId:     userId,
				LoginType:  constants.LOGIN_BY_PASSWORD,
				LoginIp:    ip,
				UserAgent:  userAgent,
				Status:     status,
				FailReason: failReason,
				CreatedAt:  time.Now(),
			},
		)
		if err != nil {
			logx.Errorf("插入登录日志失败,error=%v", err)
		}
	}()
}

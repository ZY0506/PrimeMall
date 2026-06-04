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
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type MobileLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMobileLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MobileLoginLogic {
	return &MobileLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MobileLogin 手机号验证码登录
func (l *MobileLoginLogic) MobileLogin(in *user.MobileLoginReq) (*user.LoginResp, error) {

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
			return nil, errorx.NewBizError(response.ErrCodeUserNotFound, "用户不存在")
		}
		l.Logger.Errorf("数据库查询失败，error=%v", err)
		return nil, err
	}

	// 2. 检查用户状态
	if userInfo.Status == constants.USER_STATUS_BANNED {
		l.Logger.Errorf("用户被禁用,phone=%v", in.Phone)

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

	// 3. 校验验证码
	ok, err := l.svcCtx.CaptchaSvc.Verify(l.ctx, in.Phone, constants.SCENE_LOGIN, in.Code)
	if err != nil {
		l.Logger.Errorf("验证码校验失败，error=%v", err)
		return nil, err
	}

	if !ok {
		l.Logger.Infof("验证码错误,phone=%v", in.Phone)
		l.recordLoginLog(userInfo.Id, clientInfo.IP, clientInfo.UserAgent, constants.LOGIN_STATUS_FAIL, "验证码错误")
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码错误")
	}

	// 4. 颁发token
	aToken, err := jwt.GenToken(
		jwt.Claims{
			UserId: userInfo.Id,
			JTI:    uuid.New().String(),
			Role:   "user",
		},
		[]byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.AccessTokenExpire,
		l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("生成access token失败,error=%v", err)
		return nil, err
	}

	rToken, err := jwt.GenToken(
		jwt.Claims{
			UserId: userInfo.Id,
			JTI:    uuid.New().String(),
			Role:   "user",
		},
		[]byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.RefreshTokenExpire,
		l.svcCtx.Config.JWT.Issuer,
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
func (l *MobileLoginLogic) updateLastLoginInfo(userInfo model.User, ip string) {
	defer func() {
		if r := recover(); r != nil {
			logx.Errorf("更新登录信息协程panic: %v", r)
		}
	}()
	ctx := context.Background()

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
func (l *MobileLoginLogic) recordLoginLog(userId uint64, ip string, userAgent string, status int64, failReason string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("记录登录日志协程panic: %v", r)
			}
		}()
		_, err := l.svcCtx.LoginLogModel.Insert(
			context.Background(),
			&model.UserLoginLog{
				UserId:     userId,
				LoginType:  constants.LOGIN_BY_CAPTCHA,
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

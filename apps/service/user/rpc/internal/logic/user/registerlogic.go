package userlogic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/common/base62"
	"github.com/ZY0506/PrimeMall/common/constants"
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

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Register 用户注册
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.LoginResp, error) {
	var ok bool
	var err error

	// 校验两次密码是否一致
	if in.Password != in.ConfirmPassword {
		return nil, errorx.NewBizError(response.ErrCodePasswordNotMatch, "两次输入密码不一致")
	}

	// 校验手机号是否已经存在（排除已注销账户）
	existingUser, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, in.Phone)
	if err == nil {
		// 如果用户存在且未注销，则手机号已注册
		if !existingUser.DeletedAt.Valid {
			l.Logger.Infof("手机号已存在,phone=%s", in.Phone)
			return nil, errorx.NewBizError(response.ErrCodePhoneRegistered, "手机号已注册")
		}
		// 已注销用户允许重新注册
		l.Logger.Infof("手机号已注销,phone=%s,允许重新注册", in.Phone)
	} else if !errors.Is(err, model.ErrNotFound) {
		l.Logger.Errorf("查询用户失败,error=%v", err)
		return nil, err
	}

	// 验证码校验
	ok, err = l.svcCtx.CaptchaSvc.Verify(l.ctx, in.Phone, constants.SCENE_REGISTER, in.Code)
	if err != nil {
		l.Logger.Errorf("验证码校验失败,error=%v", err)
		return nil, err
	}
	if !ok {
		return nil, errorx.NewBizError(response.ErrCodeCaptchaWrong, "验证码错误")
	}

	// 3、幂等校验
	idempotencyKey := fmt.Sprintf("%s%s", constants.IDEMPOTENCY_KEY+constants.USER_SERVICE, in.IdempotencyKey)
	ok, err = l.svcCtx.Client.SetNX(l.ctx, idempotencyKey, "1", constants.IDEMPOTENCY_EXIRE).Result()
	if err != nil {
		l.Logger.Errorf("幂等校验失败,error=%v", err)
		return nil, err
	}
	if !ok {
		return nil, errorx.NewBizError(response.ErrCodeTooFrequent, "请勿重复操作")
	}

	// TODO: 若后续失败，删除键（需确保删除操作的原子性，可配合 Lua 脚本）
	defer func() {
		if err != nil {
			l.svcCtx.Client.Del(l.ctx, idempotencyKey)
		}
	}()

	// 生成token要用的用户ID（新注册用新ID，注销复活用旧ID）
	userId := uint64(0)
	isRevived := existingUser != nil && existingUser.DeletedAt.Valid
	if isRevived {
		userId = existingUser.Id
		l.Logger.Infof("复用已注销账号, id=%d", userId)
	} else {
		userId, err = l.svcCtx.IDGenerator.NextID()
		if err != nil {
			l.Logger.Errorf("生成用户ID失败,error=%v", err)
			return nil, err
		}
	}

	// 生成token
	aToken, err := jwt.GenToken(
		jwt.Claims{UserId: userId, JTI: uuid.New().String(), Role: "user"}, []byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.AccessTokenExpire, l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("生成access token失败,error=%v", err)
		return nil, err
	}
	rToken, err := jwt.GenToken(
		jwt.Claims{UserId: userId, JTI: uuid.New().String(), Role: "user"}, []byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.RefreshTokenExpire, l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("创建refresh token失败,error=%v", err)
		return nil, err
	}

	// 4、 创建/复活用户
	if in.Nickname == "" {
		in.Nickname = base62.GenerateNickname("prime_mall_", userId)
	}
	encodePassword, err := pwd.GenerateFromPassword(in.Password)
	if err != nil {
		l.Logger.Errorf("密码加密失败,error=%v", err)
		return nil, err
	}

	if isRevived {
		// 已注销账号复活：UPDATE 现有记录，将 deleted_at 置 NULL
		existingUser.Phone = in.Phone
		existingUser.Password = encodePassword
		existingUser.Nickname = in.Nickname
		existingUser.Avatar = "https://avatars.githubusercontent.com/u/583231?v=4"
		existingUser.Gender = constants.USER_GENDER_UNKNOWN
		existingUser.Birthday = sql.NullTime{}
		existingUser.Status = constants.USER_STATUS_NORMAL
		existingUser.LastLoginTime = sql.NullTime{}
		existingUser.LastLoginIp = ""
		existingUser.ExtInfo = sql.NullString{}
		existingUser.DeletedAt = sql.NullTime{Valid: false} // 复活：清除删除标记
		existingUser.CreatedAt = time.Now()
		existingUser.UpdatedAt = time.Now()
		err = l.svcCtx.UserModel.Update(l.ctx, existingUser)
	} else {
		// 全新注册：INSERT 新记录
		_, err = l.svcCtx.UserModel.Insert(l.ctx, &model.User{
			Id:            userId,
			Phone:         in.Phone,
			Password:      encodePassword,
			Nickname:      in.Nickname,
			Avatar:        "https://avatars.githubusercontent.com/u/583231?v=4",
			Gender:        constants.USER_GENDER_UNKNOWN,
			Birthday:      sql.NullTime{},
			Status:        constants.USER_STATUS_NORMAL,
			LastLoginTime: sql.NullTime{},
			LastLoginIp:   "",
			ExtInfo:       sql.NullString{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
	}
	if err != nil {
		l.Logger.Errorf("创建用户失败,error=%v", err)
		return nil, err
	}

	l.Logger.Infof("用户注册成功，id:%d", userId)

	return &user.LoginResp{
		AccessToken:   aToken,
		AccessExpire:  timestamppb.New(time.Now().Add(time.Duration(l.svcCtx.Config.JWT.AccessTokenExpire) * time.Second)),
		RefreshToken:  rToken,
		RefreshExpire: timestamppb.New(time.Now().Add(time.Duration(l.svcCtx.Config.JWT.RefreshTokenExpire) * time.Second)),
	}, nil
}

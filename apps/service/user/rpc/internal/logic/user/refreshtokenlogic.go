package userlogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RefreshToken token刷新
func (l *RefreshTokenLogic) RefreshToken(in *user.RefreshTokenReq) (*user.LoginResp, error) {
	// 1、从ctx中获取当前用户id
	uid, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户id失败")
		return nil, err
	}

	// 2、检查token是否在黑名单中
	if l.svcCtx.Client.Exists(l.ctx, constants.RTOKEN_BLACKLIST_KEY+in.RefreshJti).Val() == 1 {
		l.Logger.Errorf("token已失效")
		return nil, errorx.NewBizError(response.ErrCodeTokenInvalid, "Token 无效")
	}

	// 3、将旧的token加入黑名单
	refreshExpire := l.svcCtx.Config.JWT.RefreshTokenExpire
	err = l.svcCtx.Client.SetNX(l.ctx, constants.RTOKEN_BLACKLIST_KEY+in.RefreshJti, 1,
		time.Duration(refreshExpire)*time.Second,
	).Err()
	if err != nil {
		l.Logger.Errorf("将旧的token加入黑名单失败")
		return nil, err
	}

	// 4、生成新的 token
	aToken, err := jwt.GenToken(
		jwt.Claims{UserId: uid, JTI: uuid.New().String(), Role: "user"}, []byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.AccessTokenExpire, l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("生成access token失败")
		return nil, err
	}
	rToken, err := jwt.GenToken(
		jwt.Claims{UserId: uid, JTI: uuid.New().String(), Role: "user"}, []byte(l.svcCtx.Config.JWT.Secret),
		l.svcCtx.Config.JWT.RefreshTokenExpire, l.svcCtx.Config.JWT.Issuer,
	)
	if err != nil {
		l.Logger.Errorf("生成refresh token失败")
		return nil, err
	}

	l.Logger.Info("刷新token成功")

	// 5、返回新的token
	return &user.LoginResp{
		AccessToken:   aToken,
		AccessExpire:  timestamppb.New(time.Now().Add(time.Duration(l.svcCtx.Config.JWT.AccessTokenExpire) * time.Second)),
		RefreshToken:  rToken,
		RefreshExpire: timestamppb.New(time.Now().Add(time.Duration(l.svcCtx.Config.JWT.RefreshTokenExpire) * time.Second)),
	}, nil
}

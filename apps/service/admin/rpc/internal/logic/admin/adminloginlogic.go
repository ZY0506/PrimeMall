package adminlogic

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/ZY0506/PrimeMall/common/jwt"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/ZY0506/PrimeMall/pkg/pwd"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ---- 管理员自身管理 ----
func (l *AdminLoginLogic) AdminLogin(in *admin.AdminLoginReq) (*admin.AdminLoginResp, error) {
	// 获取客户端信息
	clientInfo, err := ctxdata.GetClientInfoFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取客户端信息失败,error=%v", err)
		return nil, err
	}
	// 1. 查询管理员
	adminUser, err := l.svcCtx.AdminModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "账号或密码错误")
		}
		l.Logger.Errorf("查询管理员信息失败，error: %v", err)
		return nil, err
	}

	// 2. 检查状态
	if adminUser.Status == constants.ADMIN_STATUS_DISABLED {
		return nil, errorx.NewBizError(response.ErrCodeAdminDisabled, "管理员已被禁用")
	}

	// 3. 验证密码
	if !pwd.CompareHashAndPassword(adminUser.Password, in.Password) {
		return nil, errorx.NewBizError(response.ErrCodePasswordWrong, "账号或密码错误")
	}

	// 4. 生成JWT token
	token, err := jwt.GenToken(jwt.Claims{
		UserId: adminUser.Id,
		JTI:    uuid.New().String(),
		Role:   "admin",
	}, []byte(l.svcCtx.Config.JWT.Secret), l.svcCtx.Config.JWT.AccessTokenExpire, l.svcCtx.Config.JWT.Issuer)
	if err != nil {
		l.Logger.Errorf("生成token失败,error=%v", err)
		return nil, errorx.NewBizError(response.ErrCodeTokenInvalid, "生成token失败")
	}

	// 5. 异步更新登录信息
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l.Logger.Errorf("更新管理员登录信息协程panic: %v", r)
			}
		}()
		// 使用带超时的 context，避免后台操作阻塞过久
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = l.svcCtx.AdminModel.UpdateLoginInfo(ctx, adminUser.Id, clientInfo.IP)
	}()

	return &admin.AdminLoginResp{
		Token:      token,
		ExpireTime: time.Now().Add(time.Duration(l.svcCtx.Config.JWT.AccessTokenExpire) * time.Second).Unix(),
	}, nil
}

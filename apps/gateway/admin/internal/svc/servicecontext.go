package svc

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/client/admin"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/client/adminaftersale"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/client/admincoupon"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/client/adminorder"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/client/adminproduct"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/client/adminuser"
	"github.com/ZY0506/PrimeMall/common/interceptor"
	"github.com/ZY0506/PrimeMall/common/middleware"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config               config.Config
	AdminRpc             admin.Admin
	AdminUserRpc         adminuser.AdminUser
	AdminProductRpc      adminproduct.AdminProduct
	AdminOrderRpc        adminorder.AdminOrder
	AdminAfterSaleRpc    adminaftersale.AdminAfterSale
	AdminCouponRpc       admincoupon.AdminCoupon
	ClientInfoMiddleware *middleware.ClientInfoMiddleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := zrpc.MustNewClient(c.AdminRpc,
		zrpc.WithUnaryClientInterceptor(interceptor.ClientInfoUnaryInterceptor()))
	return &ServiceContext{
		Config:               c,
		AdminRpc:             admin.NewAdmin(conn),
		AdminUserRpc:         adminuser.NewAdminUser(conn),
		AdminProductRpc:      adminproduct.NewAdminProduct(conn),
		AdminOrderRpc:        adminorder.NewAdminOrder(conn),
		AdminAfterSaleRpc:    adminaftersale.NewAdminAfterSale(conn),
		AdminCouponRpc:       admincoupon.NewAdminCoupon(conn),
		ClientInfoMiddleware: middleware.NewClientInfoMiddleware(),
	}
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package svc

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/config"
	"github.com/ZY0506/PrimeMall/apps/service/marketing/rpc/client/marketing"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/cart"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/client/order"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/client/payment"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/client/paymentinternal"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/client/product"
	"github.com/ZY0506/PrimeMall/apps/service/search/rpc/client/search"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/client/user"
	"github.com/ZY0506/PrimeMall/common/interceptor"
	"github.com/ZY0506/PrimeMall/common/middleware"
	"github.com/ZY0506/PrimeMall/pkg/rdb"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"net/http"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config                   config.Config
	Client                   *redis.Client
	UserRpc                  user.User
	ProductRpc               product.Product
	OrderRpc                 order.Order
	CartRpc                  cart.Cart
	PaymentRpc               payment.Payment
	PaymentInternalRpc       paymentinternal.PaymentInternal
	MarketingRpc             marketing.Marketing
	SearchRpc                search.Search
	CorsMiddleware           *middleware.CorsMiddleware
	ClientInfoMiddleware     *middleware.ClientInfoMiddleware
	TokenBlacklistMiddleware rest.Middleware
	CasbinEnforcerMiddleware rest.Middleware
	TokenBucketMiddleware    rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	userClient := zrpc.MustNewClient(c.UserRpc,
		zrpc.WithUnaryClientInterceptor(interceptor.ClientInfoUnaryInterceptor()))
	// 初始化redis
	client, err := rdb.InitRedis(c.RDB.RedisAddr, c.RDB.RedisPassword, c.RDB.RedisDB, c.RDB.RedisPoolSize, c.RDB.RedisMinIdleConns)
	if err != nil {
		logx.Errorf("redis init failed,error:%v", err.Error())
		panic(err)
	}
	tokenBlacklistMiddleware := middleware.NewTokenBlacklistMiddleware(client)

	// 初始化 Casbin 权限中间件
	casbinMiddleware, err := middleware.NewCasbinEnforcerMiddleware(c.CasbinConf.ModelPath, c.CasbinConf.PolicyPath)
	if err != nil {
		logx.Errorf("Casbin init failed, casbin permission middleware disabled, error:%v", err.Error())
		// Casbin 初始化失败时，使用一个放行中间件（不阻塞服务启动）
		casbinMiddleware = nil
	}
	var casbinHandle rest.Middleware
	if casbinMiddleware != nil {
		casbinHandle = casbinMiddleware.Handle
	} else {
		casbinHandle = func(next http.HandlerFunc) http.HandlerFunc {
			return next
		}
	}

	tokenBucketMiddleware := middleware.NewTokenBucketMiddleware(c.RateLimit.PerIPSpeed, c.RateLimit.PerIPBurst)

	return &ServiceContext{
		Config:                   c,
		UserRpc:                  user.NewUser(userClient),
		ProductRpc:               product.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		OrderRpc:                 order.NewOrder(zrpc.MustNewClient(c.OrderRpc)),
		CartRpc:                  cart.NewCart(zrpc.MustNewClient(c.OrderRpc)),
		PaymentRpc:               payment.NewPayment(zrpc.MustNewClient(c.PaymentRpc)),
		PaymentInternalRpc:       paymentinternal.NewPaymentInternal(zrpc.MustNewClient(c.PaymentInternalRpc)),
		MarketingRpc:             marketing.NewMarketing(zrpc.MustNewClient(c.MarketingRpc)),
		SearchRpc:                search.NewSearch(zrpc.MustNewClient(c.SearchRpc)),
		Client:                   client,
		CorsMiddleware:           middleware.NewCorsMiddleware(),
		ClientInfoMiddleware:     middleware.NewClientInfoMiddleware(),
		TokenBlacklistMiddleware: tokenBlacklistMiddleware.Handle,
		CasbinEnforcerMiddleware: casbinHandle,
		TokenBucketMiddleware:    tokenBucketMiddleware.Handle,
	}
}

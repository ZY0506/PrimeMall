package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"path/filepath"

	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/config"
	cartServer "github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/server/cart"
	orderServer "github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/server/order"
	orderadminServer "github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/server/orderadmin"
	orderinternalServer "github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/server/orderinternal"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/order/rpc/types/order"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "apps/service/order/rpc/etc/order.yaml", "the config file")

func main() {
	flag.Parse()

	projectRoot, _ := filepath.Abs("./")
	envPath := filepath.Join(projectRoot, ".env")

	if err := godotenv.Load(envPath); err != nil {
		fmt.Println("Load .env file failed")
		panic(err)
	}

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	defer ctx.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		order.RegisterOrderServer(grpcServer, orderServer.NewOrderServer(ctx))
		order.RegisterOrderAdminServer(grpcServer, orderadminServer.NewOrderAdminServer(ctx))
		order.RegisterOrderInternalServer(grpcServer, orderinternalServer.NewOrderInternalServer(ctx))
		order.RegisterCartServer(grpcServer, cartServer.NewCartServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 启动定时关单任务（兜底方案，防止MQ延迟消息丢失导致过期订单未关闭）
	go ctx.CloseExpiredOrders(ctx.Ctx)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

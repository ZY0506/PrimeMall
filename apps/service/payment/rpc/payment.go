package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"path/filepath"

	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/config"
	paymentServer "github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/server/payment"
	paymentadminServer "github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/server/paymentadmin"
	paymentinternalServer "github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/server/paymentinternal"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/payment/rpc/types/payment"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/payment.yaml", "the config file")

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

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		payment.RegisterPaymentServer(grpcServer, paymentServer.NewPaymentServer(ctx))
		payment.RegisterPaymentAdminServer(grpcServer, paymentadminServer.NewPaymentAdminServer(ctx))
		payment.RegisterPaymentInternalServer(grpcServer, paymentinternalServer.NewPaymentInternalServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"path/filepath"

	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/config"
	productServer "github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/server/product"
	productadminServer "github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/server/productadmin"
	productinternalServer "github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/server/productinternal"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/product.yaml", "the config file")

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
		product.RegisterProductServer(grpcServer, productServer.NewProductServer(ctx))
		product.RegisterProductAdminServer(grpcServer, productadminServer.NewProductAdminServer(ctx))
		product.RegisterProductInternalServer(grpcServer, productinternalServer.NewProductInternalServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

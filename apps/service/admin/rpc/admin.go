package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/config"
	adminserver "github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/server/admin"
	adminaftersaleserver "github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/server/adminaftersale"
	adminorderserver "github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/server/adminorder"
	adminproductserver "github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/server/adminproduct"
	adminuserserver "github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/server/adminuser"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/types/admin"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "apps/service/admin/rpc/etc/admin.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载 .env 文件
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
		// 注册所有5个服务
		admin.RegisterAdminServer(grpcServer, adminserver.NewAdminServer(ctx))
		admin.RegisterAdminUserServer(grpcServer, adminuserserver.NewAdminUserServer(ctx))
		admin.RegisterAdminProductServer(grpcServer, adminproductserver.NewAdminProductServer(ctx))
		admin.RegisterAdminOrderServer(grpcServer, adminorderserver.NewAdminOrderServer(ctx))
		admin.RegisterAdminAfterSaleServer(grpcServer, adminaftersaleserver.NewAdminAfterSaleServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"path/filepath"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/config"
	userServer "github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/server/user"
	userinternalServer "github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/server/userinternal"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "apps/service/user/rpc/etc/user.yaml", "the config file")

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
		user.RegisterUserServer(grpcServer, userServer.NewUserServer(ctx))
		user.RegisterUserInternalServer(grpcServer, userinternalServer.NewUserInternalServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

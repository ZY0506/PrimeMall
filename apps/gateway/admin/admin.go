package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/config"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/handler"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/admin.yaml", "the config file")

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

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	// 注册全局中间件
	server.Use(ctx.ClientInfoMiddleware.Handle) // 请求头信息处理中间件

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting admin gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

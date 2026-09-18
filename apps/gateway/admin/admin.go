package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/config"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/handler"
	"github.com/ZY0506/PrimeMall/apps/gateway/admin/internal/svc"
	"github.com/ZY0506/PrimeMall/common/validate"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/admin.yaml", "the config file")

func main() {
	flag.Parse()
	logx.SetLevel(logx.ErrorLevel)

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

	// 注册全局请求参数校验器。admin 的请求类型目前没有任何 validate tag，
	// 这一步是为后续补 tag 铺路，今天不产生行为变化。
	httpx.SetValidator(validate.New())

	ctx := svc.NewServiceContext(c)
	// 注册全局中间件
	server.Use(ctx.ClientInfoMiddleware.Handle) // 请求头信息处理中间件

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting admin gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

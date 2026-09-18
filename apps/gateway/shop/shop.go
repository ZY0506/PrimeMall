// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/config"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/handler"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/common/validate"
	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/shop-api.yaml", "the config file")

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

	// 注册全局请求参数校验器。不注册的话 httpx.Parse 会跳过全部 validate tag 直接放行。
	httpx.SetValidator(validate.New())

	ctx := svc.NewServiceContext(c)
	// 注册全局中间件
	server.Use(ctx.CorsMiddleware.ClientHandle) // 跨域处理中间件
	server.Use(ctx.ClientInfoMiddleware.Handle) // 请求头信息处理中间件
	server.Use(ctx.TokenBucketMiddleware)       // 令牌桶限流中间件

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

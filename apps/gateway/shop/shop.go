// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/config"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/handler"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// defaultValidator implements httpx.Validator using reflect-based struct tag parsing.
type defaultValidator struct{}

func (v *defaultValidator) Validate(r *http.Request, data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}
	return v.validateStruct(val)
}

func (v *defaultValidator) validateStruct(val reflect.Value) error {
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" || tag == "-" {
			continue
		}
		fieldVal := val.Field(i)
		if err := v.validateField(field, fieldVal, tag); err != nil {
			return err
		}
	}
	return nil
}

func (v *defaultValidator) validateField(field reflect.StructField, val reflect.Value, tag string) error {
	// Split by comma (rules like "oneof=..." may contain spaces)
	rules := strings.Split(tag, ",")
	omitempty := false
	validations := make([]string, 0, len(rules))
	for _, r := range rules {
		r = strings.TrimSpace(r)
		if r == "omitempty" {
			omitempty = true
		} else {
			validations = append(validations, r)
		}
	}

	if omitempty && val.IsZero() {
		return nil
	}

	for _, rule := range validations {
		if err := v.applyRule(field, val, rule); err != nil {
			return err
		}
	}
	return nil
}

func (v *defaultValidator) applyRule(field reflect.StructField, val reflect.Value, rule string) error {
	switch {
	case rule == "required":
		if val.IsZero() {
			return fmt.Errorf("'%s' 为必填字段", field.Name)
		}
	case strings.HasPrefix(rule, "min="):
		minStr := rule[4:]
		return v.validateMin(field.Name, val, minStr)
	case strings.HasPrefix(rule, "max="):
		maxStr := rule[4:]
		return v.validateMax(field.Name, val, maxStr)
	case strings.HasPrefix(rule, "oneof="):
		options := strings.Split(rule[6:], " ")
		return v.validateOneof(field.Name, val, options)
	case strings.HasPrefix(rule, "eqfield="):
		return nil
	case strings.HasPrefix(rule, "nefield="):
		return nil
	case strings.HasPrefix(rule, "datetime="):
		return nil
	}
	return nil
}

func (v *defaultValidator) validateMin(name string, val reflect.Value, minStr string) error {
	switch val.Kind() {
	case reflect.String:
		min, _ := strconv.Atoi(minStr)
		if len(val.String()) < min {
			return fmt.Errorf("'%s' 长度不能少于 %d 个字符", name, min)
		}
	case reflect.Int, reflect.Int64:
		min, _ := strconv.ParseInt(minStr, 10, 64)
		if val.Int() < min {
			return fmt.Errorf("'%s' 不能小于 %d", name, min)
		}
	}
	return nil
}

func (v *defaultValidator) validateMax(name string, val reflect.Value, maxStr string) error {
	switch val.Kind() {
	case reflect.String:
		max, _ := strconv.Atoi(maxStr)
		if len(val.String()) > max {
			return fmt.Errorf("'%s' 长度不能超过 %d 个字符", name, max)
		}
	case reflect.Int, reflect.Int64:
		max, _ := strconv.ParseInt(maxStr, 10, 64)
		if val.Int() > max {
			return fmt.Errorf("'%s' 不能大于 %d", name, max)
		}
	}
	return nil
}

func (v *defaultValidator) validateLen(name string, val reflect.Value, lenStr string) error {
	switch val.Kind() {
	case reflect.String:
		expectedLen, _ := strconv.Atoi(lenStr)
		if len(val.String()) != expectedLen {
			return fmt.Errorf("'%s' 长度必须为 %d 个字符", name, expectedLen)
		}
	}
	return nil
}

func (v *defaultValidator) validateOneof(name string, val reflect.Value, options []string) error {
	var strVal string
	switch val.Kind() {
	case reflect.String:
		strVal = val.String()
	case reflect.Int, reflect.Int64:
		strVal = strconv.FormatInt(val.Int(), 10)
	default:
		return nil
	}
	for _, opt := range options {
		if strVal == opt {
			return nil
		}
	}
	return fmt.Errorf("'%s' 的值必须在 [%s] 范围内", name, strings.Join(options, ", "))
}

var configFile = flag.String("f", "etc/shop-api.yaml", "the config file")

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

	// 注册全局请求参数校验器
	httpx.SetValidator(&defaultValidator{})

	ctx := svc.NewServiceContext(c)
	// 注册全局中间件
	server.Use(ctx.CorsMiddleware.ClientHandle) // 跨域处理中间件
	server.Use(ctx.ClientInfoMiddleware.Handle) // 请求头信息处理中间件
	server.Use(ctx.TokenBucketMiddleware)       // 令牌桶限流中间件

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

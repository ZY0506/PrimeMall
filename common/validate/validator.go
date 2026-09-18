// Package validate 提供基于 go-playground/validator 的 HTTP 请求参数校验器。
package validate

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Validator 实现 go-zero 的 httpx.Validator 接口，通过 httpx.SetValidator 注册。
//
// 背景：go-zero 只提供 httpx.Validator 接口和 httpx.SetValidator 这个注册点，
// 未注册时 httpx.Parse 会跳过全部 tag 校验直接放行——框架本身并不含内置的 tag 校验器
// （httpx.getValidator() 返回的包级变量默认就是 nil）。所以这里不是"退回框架默认实现"，
// 而是把校验能力真正接上。
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// New 构造校验器并注册中文错误文案。
// 在 main 里随进程启动调用一次；配置失败直接 panic，与 conf.MustLoad 的失败策略一致。
func New() *Validator {
	v := validator.New()

	// 错误文案里用请求字段名（json/form/path tag）而不是 Go 字段名，因为这些错误会经
	// response.ClientError 原样透给客户端。注意这不影响 eqfield/nefield/gtefield 的参数解析，
	// 那些参数是按 Go 字段名查找的（validator 内部走 reflect.FieldByName）。
	v.RegisterTagNameFunc(fieldDisplayName)

	locale := zh.New()
	trans, ok := ut.New(locale, locale).GetTranslator("zh")
	if !ok {
		panic("validate: 获取 zh 翻译器失败")
	}
	if err := zhtranslations.RegisterDefaultTranslations(v, trans); err != nil {
		panic(fmt.Sprintf("validate: 注册中文校验文案失败: %v", err))
	}

	return &Validator{validate: v, trans: trans}
}

// Validate 实现 httpx.Validator。校验失败时返回的 error 已是中文文案。
func (v *Validator) Validate(_ *http.Request, data any) error {
	if err := v.validate.Struct(data); err != nil {
		return &FieldError{err: err, trans: v.trans}
	}
	return nil
}

// fieldDisplayName 依次尝试 json/form/path tag，取不到再退回 Go 字段名。
func fieldDisplayName(field reflect.StructField) string {
	for _, tag := range []string{"json", "form", "path"} {
		name := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return field.Name
}

// FieldError 把 validator 的多字段错误渲染成一条客户端可读的中文消息。
type FieldError struct {
	err   error
	trans ut.Translator
}

func (e *FieldError) Error() string {
	var fieldErrs validator.ValidationErrors
	if !errors.As(e.err, &fieldErrs) {
		return e.err.Error()
	}

	// 缺失翻译时 FieldError.Translate 会自动退回英文原文，这里不再兜底。
	msgs := make([]string, 0, len(fieldErrs))
	for _, fe := range fieldErrs {
		msgs = append(msgs, fe.Translate(e.trans))
	}
	return strings.Join(msgs, "; ")
}

// Unwrap 保留原始错误，便于调用方用 errors.Is/As 追溯。
func (e *FieldError) Unwrap() error { return e.err }

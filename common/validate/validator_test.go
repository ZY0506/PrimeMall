package validate

import (
	"strings"
	"testing"
)

// 包级构造一次即可：New() 会注册翻译器，没必要每个用例重建。
var testValidator = New()

// 用例结构体镜像 apps/gateway/shop/internal/types/types.go 中的真实请求类型。

type eqFieldReq struct {
	Password        string `json:"password" validate:"required,min=8,max=20"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type neFieldReq struct {
	OldPassword string `json:"old_password" validate:"required,min=8,max=20"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=20,nefield=OldPassword"`
}

type datetimeReq struct {
	Birthday string `json:"birthday,optional" validate:"omitempty,datetime=2006-01-02"`
}

type gteFieldReq struct {
	MinPrice int64 `form:"min_price,optional" validate:"omitempty,min=0"`
	MaxPrice int64 `form:"max_price,optional" validate:"omitempty,min=0,gtefield=MinPrice"`
}

type requiredIfReq struct {
	PayType string `json:"pay_type" validate:"required,oneof=APP H5 JSAPI QR"`
	OpenId  string `json:"openid,optional" validate:"required_if=PayType JSAPI"`
}

type diveReq struct {
	Images []string `json:"images,optional" validate:"omitempty,max=9,dive,min=1,max=255"`
}

type requiredSliceReq struct {
	Items []int `json:"items" validate:"required,min=1,max=50"`
}

type cjkReq struct {
	Nickname string `json:"nickname" validate:"required,min=1,max=4"`
}

type phoneReq struct {
	Phone string `json:"phone" validate:"required,len=11"`
}

func validateErr(t *testing.T, data any) string {
	t.Helper()
	err := testValidator.Validate(nil, data)
	if err == nil {
		t.Fatalf("期望校验失败，实际通过: %+v", data)
	}
	return err.Error()
}

func validateOK(t *testing.T, data any) {
	t.Helper()
	if err := testValidator.Validate(nil, data); err != nil {
		t.Fatalf("期望校验通过，实际失败: %v", err)
	}
}

// 错误文案必须是中文且用请求字段名——这些消息经 response.ClientError 直接透给客户端。
func TestErrorIsChineseAndUsesRequestFieldName(t *testing.T) {
	msg := validateErr(t, &eqFieldReq{Password: "abcd1234"})
	if !strings.Contains(msg, "必填") {
		t.Fatalf("错误文案应为中文，实际: %s", msg)
	}
	if !strings.Contains(msg, "confirm_password") {
		t.Fatalf("错误文案应使用 json 字段名，实际: %s", msg)
	}
}

func TestEqField(t *testing.T) {
	validateOK(t, &eqFieldReq{Password: "abcd1234", ConfirmPassword: "abcd1234"})
	msg := validateErr(t, &eqFieldReq{Password: "abcd1234", ConfirmPassword: "abcd12345"})
	if !strings.Contains(msg, "confirm_password") {
		t.Fatalf("eqfield 未命中 confirm_password，实际: %s", msg)
	}
}

func TestNeField(t *testing.T) {
	validateOK(t, &neFieldReq{OldPassword: "abcd1234", NewPassword: "abcd12345"})
	validateErr(t, &neFieldReq{OldPassword: "abcd1234", NewPassword: "abcd1234"})
}

func TestDatetime(t *testing.T) {
	validateOK(t, &datetimeReq{Birthday: "2000-01-02"})
	validateOK(t, &datetimeReq{Birthday: ""}) // omitempty
	validateErr(t, &datetimeReq{Birthday: "2000/01/02"})
}

func TestGteField(t *testing.T) {
	validateOK(t, &gteFieldReq{MinPrice: 100, MaxPrice: 200})
	validateOK(t, &gteFieldReq{MinPrice: 100, MaxPrice: 100})
	validateErr(t, &gteFieldReq{MinPrice: 200, MaxPrice: 100})

	// MinPrice 为 0 而 MaxPrice 非 0：两侧同 Kind（int64），不应 panic。
	validateOK(t, &gteFieldReq{MinPrice: 0, MaxPrice: 5})
}

func TestRequiredIf(t *testing.T) {
	validateOK(t, &requiredIfReq{PayType: "APP"})
	validateOK(t, &requiredIfReq{PayType: "JSAPI", OpenId: "ooxx"})
	validateErr(t, &requiredIfReq{PayType: "JSAPI"})
}

func TestDiveAndSliceMax(t *testing.T) {
	validateOK(t, &diveReq{Images: nil})
	validateOK(t, &diveReq{Images: []string{"a.png", "b.png"}})

	// 切片上的 max 在旧实现里是空操作，现在按元素个数校验——这是可见的行为变化。
	validateErr(t, &diveReq{Images: make([]string, 10)})

	// dive 之后的 min/max 作用于每个元素
	validateErr(t, &diveReq{Images: []string{"a.png", ""}})
}

func TestRequiredAndLengthOnSlice(t *testing.T) {
	validateOK(t, &requiredSliceReq{Items: []int{1}})

	// required 对切片用的是 !IsNil()，与旧实现的 IsZero() 等价：nil 被拒。
	validateErr(t, &requiredSliceReq{Items: nil})

	// 非 nil 空切片能过 required，但会被 min=1 拦下。
	// 旧实现对切片上的 min/max 是空操作，所以 "items": [] 以前是放行的，现在不放行。
	msg := validateErr(t, &requiredSliceReq{Items: []int{}})
	if !strings.Contains(msg, "至少") {
		t.Fatalf("空切片应由切片长度规则 min=1 拦截，实际: %s", msg)
	}
}

// 字符串长度按符文计：旧实现的 min=/max= 用的是 len()（字节），len= 用的是符文，自身不一致。
// 4 个汉字 = 12 字节，旧实现会误判为超长；v10 统一按符文计。
func TestStringLengthCountsRunesNotBytes(t *testing.T) {
	validateOK(t, &cjkReq{Nickname: "四个字符"})
	validateErr(t, &cjkReq{Nickname: "五个字符啊"})
}

// len= 在 ASCII 上的语义与旧实现一致，回归保护。
func TestLenUnchangedForASCII(t *testing.T) {
	validateOK(t, &phoneReq{Phone: "13800138000"})
	validateErr(t, &phoneReq{Phone: "138"})
}

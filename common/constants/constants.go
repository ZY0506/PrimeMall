package constants

import "time"

// 场景
const (
	SCENE_LOGIN        = "login"
	SCENE_REGISTER     = "register"
	SCENE_RESET_PWD    = "reset_pwd"
	SCENE_UPDATE_PWD   = "update_pwd"
	SCENE_UPDATE_PHONE = "update_phone"

	LOGIN_BY_PASSWORD = "password"
	LOGIN_BY_CAPTCHA  = "captcha"
)

// 验证码
const (
	CAPTCHA_KEY       = "captcha:"
	CAPTCHA_COOL_KEY  = "captcha_cool:"
	CAPTCHA_EXPIRE    = 5 * 60 * time.Second // 单位：秒
	CAPTCHA_COOL_DOWN = 60 * time.Second     // 单位：秒
)

// 通用幂等键
const (
	IDEMPOTENCY_KEY   = "idempotency:"
	IDEMPOTENCY_EXIRE = 30 * time.Second
)

// 用户状态
const (
	USER_STATUS_UNKNOWN    = iota // 未知
	USER_STATUS_NORMAL            // 正常
	USER_STATUS_RESTRICTED        // 限制下单
	USER_STATUS_BANNED            // 封禁
)

// 用户性别
const (
	USER_GENDER_UNKNOWN = iota
	USER_GENDER_MALE    // 男
	USER_GENDER_FEMALE
)

// token黑名单
const (
	RTOKEN_BLACKLIST_KEY = "jwt:blacklist:refresh:"
	ATOKEN_BLACKLIST_KEY = "jwt:blacklist:access:"
)

// 登录状态
const (
	LOGIN_STATUS_SUCCESS = int64(1)
	LOGIN_STATUS_FAIL    = int64(2)
)

// 地址tag
const (
	TAG_HOME   = "HOME"
	TAG_OFFICE = "OFFICE"
	TAG_SCHOOL = "SCHOOL"
)

const (
	PRODUCT_STATUS_UNKNOWN = iota
	PRODUCT_STATUS_ON_SALE
	PRODUCT_STATUS_OFF_SALE
)

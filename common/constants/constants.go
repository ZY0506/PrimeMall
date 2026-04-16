package constants

import "time"

// 场景
const (
	SCENE_LOGIN     = "login"
	SCENE_REGISTER  = "register"
	SCENE_RESET_PWD = "reset_pwd"

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
	USER_GENDER_MALE
	USER_GENDER_FEMALE
)

// token黑名单
const (
	RTOKEN_BLACKLIST_KEY = "rToken_blacklist:"
	ATOKEN_BLACKLIST_KEY = "aToken_blacklist:"
)

const (
	LOGIN_STATUS_SUCCESS = int64(1)
	LOGIN_STATUS_FAIL    = int64(1)
)

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
	USER_SERVICE      = "user_service:"
	PRODUCT_SERVICE   = "product_service:"
	ORDER_SERVICE     = "order_service:"
	PAYMENT_SERVICE   = "payment_service:"
	SEARCH_SERVICE    = "search_service:"
	MARKETING_SERVICE = "marketing_service:"
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

// MAX_RETRY_COUNT 重试机制次数
const (
	MAX_RETRY_COUNT = 3
)

const (
	STOCK_CHANGE_TYPE_LOCKED        = iota + 1 // 锁定库存
	STOCK_CHANGE_TYPE_UNLOCKED                 // 解锁库存
	STOCK_CHANGE_TYPE_DEDUCT                   // 扣减库存
	STOCK_CHANGE_TYPE_ROLLBACK                 // 回滚库存(订单取消/退款)
	STOCK_CHANGE_TYPE_REVERT_DEDUCT            // 恢复库存
)

// 前缀（订单部分）
const (
	PREFIX_ORDER_SN         = "PM"
	PREFIX_SETTLEMENT_TOKEN = "SToken"
	SETTLEMENT_TOKEN_KEY    = "order:settlement_token:"
	SETTLEMENT_TOKEN_EXPIRE = 60 * time.Second
	ORDER_CACHE_KEY         = "order:info:"
	ORDER_CACHE_EXPIRE      = 15 * 60 * time.Second
)

// 订单类型
const (
	ORDER_TYPE_NORMAL = iota + 1
	ORDER_TYPE_SECKILL
	ORDER_TYPE_GROUPON
)

// 订单状态
const (
	ORDER_STATUS_PENDING_PAY = 10 // 待支付
	ORDER_STATUS_PAID        = 20 // 已支付
	ORDER_STATUS_SHIPPED     = 30 // 已发货
	ORDER_STATUS_COMPLETED   = 40 // 已完成
	ORDER_STATUS_CANCELED    = 50 // 已取消
	ORDER_STATUS_AFTER_SALE  = 60 // 售后中
)

// 存储订单信息的key
const (
	ORDER_SN      = "order_sn"
	ORDER_USER_ID = "userId"
	ORDER_INFO    = "order_info"
	ORDER_ITEMS   = "order_items"
)

// ORDER_EXPIRE 订单过期时间
const (
	ORDER_EXPIRE = 15 * 60 * time.Second
)

// 售后
const (
	PREFIX_AFTER_SALE_SN        = "AS"
	AFTER_SALE_STATUS_PENDING   = 10 // 待审核
	AFTER_SALE_STATUS_RETURNING = 20 // 待退货
	AFTER_SALE_STATUS_REFUNDING = 30 // 退款中
	AFTER_SALE_STATUS_COMPLETED = 40 // 已完成
	AFTER_SALE_STATUS_REFUSED   = 50 // 拒绝
	AFTER_SALE_STATUS_CANCELED  = 60 // 已取消
)

// 退款状态
const (
	REFUND_STATUS_PENDING_REFUND = 0  // 未退款
	REFUND_STATUS_REFUNDING      = 10 // 退款中
	REFUND_STATUS_REFUNDED       = 20 // 退款成功
	REFUND_STATUS_REFUND_FAILED  = 30 // 退款失败
)

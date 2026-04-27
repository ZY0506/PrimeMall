package response

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
	"net/http"
)

const (
	InternalError = 500
	RequestError  = 400
)

// SuccessCode 成功码
const (
	SuccessCode = 0 // 成功
)

// 参数与请求错误 (11xxx)
const (
	ErrCodeInvalidParam     = 11001 // 请求参数无效
	ErrCodeMissingParam     = 11002 // 缺少必要参数
	ErrCodeParamType        = 11003 // 参数类型错误
	ErrCodeParamRange       = 11004 // 参数超出范围
	ErrCodeInvalidPage      = 11005 // 分页参数无效
	ErrCodeInvalidTimestamp = 11006 // 时间格式无效
)

// 认证与授权错误 (12xxx)
const (
	ErrCodeUnauthorized      = 12001 // 未认证，请先登录
	ErrCodeTokenInvalid      = 12002 // Token 无效
	ErrCodeTokenExpired      = 12003 // Token 已过期
	ErrCodeTokenMissing      = 12004 // 缺少 Token
	ErrCodePermissionDenied  = 12005 // 权限不足
	ErrCodePasswordWrong     = 12006 // 账号或密码错误
	ErrCodeCaptchaWrong      = 12007 // 验证码错误
	ErrCodeCaptchaExpired    = 12008 // 验证码已过期
	ErrCodePhoneNotExist     = 12009 // 手机号未注册
	ErrCodePhoneRegistered   = 12010 // 手机号已注册
	ErrCodeTooFrequent       = 12011 // 访问过于频繁
	ErrCodePasswordNotMatch  = 12012 // 两次输入密码不一致
	ErrCodeSameAsOldPassword = 12013 // 新密码与旧密码一致
)

// 用户模块 (2xxxx)
const (
	ErrCodeUserNotFound     = 20001 // 用户不存在
	ErrCodeUserDisabled     = 20002 // 账号已被禁用
	ErrCodeUserLocked       = 20003 // 账号已被锁定
	ErrCodeUserDeleted      = 20004 // 账号已注销
	ErrCodeUserRestricted   = 20005 // 账号受限（限制下单等）
	ErrCodeNicknameExists   = 20006 // 昵称已被占用
	ErrCodeAvatarUploadFail = 20007 // 头像上传失败
	ErrCodeOssCallbackFail  = 20008 // OSS 回调处理失败
)

// 商品模块 (3xxxx)
const (
	ErrCodeProductNotFound      = 30001 // 商品不存在
	ErrCodeProductOffline       = 30002 // 商品已下架
	ErrCodeProductDeleted       = 30003 // 商品已删除
	ErrCodeSkuNotFound          = 30004 // SKU 不存在
	ErrCodeSkuStockInsufficient = 30005 // SKU 库存不足
	ErrCodeSkuSoldOut           = 30006 // SKU 已售罄
	ErrCodeCategoryNotFound     = 30007 // 分类不存在
	ErrCodeCategoryHasChildren  = 30008 // 分类含有子分类，无法删除
	ErrCodeProductInvalidStatus = 30009 // 商品状态非法
	ErrCodePictureNotFound      = 30009 // 图片不存在
	ErrCodeCategoryDisabled     = 30010 // 分类被禁用
)

// 订单模块 (4xxxx)
const (
	ErrCodeOrderNotFound            = 40001 // 订单不存在
	ErrCodeOrderStatusInvalid       = 40002 // 订单状态不允许当前操作
	ErrCodeOrderCancelFailed        = 40003 // 订单取消失败（非待支付状态）
	ErrCodeOrderConfirmFailed       = 40004 // 确认收货失败
	ErrCodeOrderExpired             = 40005 // 订单已过期
	ErrCodeOrderItemMismatch        = 40006 // 订单商品信息不匹配
	ErrCodeSettlementTokenInvalid   = 40007 // 结算令牌无效或已过期
	ErrCodeIdempotentConflict       = 40008 // 重复提交订单（幂等冲突）
	ErrCodeOrderAmountMismatch      = 40009 // 订单金额校验失败
	ErrCodeAddressNotBelongUser     = 40010 // 地址不属于当前用户
	ErrCodeFreightTemplateNotFound  = 40011 // 运费模板不存在
	ErrCodeFreightCalculationFailed = 40012 // 运费计算失败

)

// 支付模块 (5xxxx)
const (
	ErrCodePaymentNotFound       = 50001 // 支付单不存在
	ErrCodePaymentAlreadyPaid    = 50002 // 支付单已支付
	ErrCodePaymentExpired        = 50003 // 支付单已过期
	ErrCodePaymentChannelError   = 50004 // 支付渠道错误
	ErrCodePaymentAmountMismatch = 50005 // 支付金额与订单金额不一致
	ErrCodePaymentCloseFailed    = 50006 // 关闭支付单失败
	ErrCodeRefundNotFound        = 50007 // 退款单不存在
	ErrCodeRefundStatusInvalid   = 50008 // 退款状态不允许操作
	ErrCodeRefundAmountExceed    = 50009 // 退款金额超过可退金额
	ErrCodeCallbackVerifyFailed  = 50010 // 支付回调验签失败
)

// 营销/优惠券模块 (6xxxx)
const (
	ErrCodeCouponNotFound       = 60001 // 优惠券不存在
	ErrCodeCouponExpired        = 60002 // 优惠券已过期
	ErrCodeCouponNotStarted     = 60003 // 优惠券未开始
	ErrCodeCouponAlreadyClaimed = 60004 // 优惠券已领取过
	ErrCodeCouponStockExhausted = 60005 // 优惠券库存不足
	ErrCodeCouponNotAvailable   = 60006 // 优惠券不可用（不满足门槛等）
	ErrCodeUserCouponNotFound   = 60007 // 用户优惠券记录不存在
	ErrCodeUserCouponUsed       = 60008 // 优惠券已被使用
	ErrCodeCouponLockFailed     = 60009 // 优惠券锁定失败
	ErrCodeCouponUnlockFailed   = 60010 // 优惠券解锁失败
)

// 地址模块 (7xxxx)
const (
	ErrCodeAddressNotFound      = 70001 // 地址不存在
	ErrCodeAddressLimitExceeded = 70002 // 地址数量超出限制（最多20条）
	ErrCodeDefaultAddressDelete = 70003 // 默认地址不可直接删除，请先设置其他默认地址
	ErrCodeAddressMissingFields = 70004 // 地址缺少必填字段
)

// 购物车模块 (8xxxx)
const (
	ErrCodeCartItemNotFound        = 80001 // 购物车商品不存在
	ErrCodeCartItemQuantityInvalid = 80002 // 购物车数量无效
	ErrCodeCartSkuMismatch         = 80003 // 购物车 SKU 信息已变更
	ErrCodeCartClearFailed         = 80004 // 清空购物车失败
)

// 售后模块 (9xxxx)
const (
	ErrCodeAfterSaleNotFound      = 90001 // 售后单不存在
	ErrCodeAfterSaleStatusInvalid = 90002 // 售后单状态不允许操作
	ErrCodeAfterSaleAlreadyExist  = 90003 // 该商品已申请过售后
	ErrCodeAfterSaleTimeExceed    = 90004 // 超出售后申请时限
	ErrCodeAfterSaleReasonInvalid = 90005 // 售后原因无效
	ErrCodeAfterSaleImageRequired = 90006 // 请上传凭证图片
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string {
	return e.Msg
}

func NewBizError(code int, msg string) error {
	return &BizError{
		Code: code,
		Msg:  msg,
	}
}

func Success(w http.ResponseWriter, r *http.Request, data interface{}) {
	httpx.WriteJsonCtx(r.Context(), w, http.StatusOK, Response{
		Code: SuccessCode,
		Msg:  "success",
		Data: data,
	})
}

func ClientError(ctx context.Context, w http.ResponseWriter, bizCode int, errMsg string) {
	logx.WithContext(ctx).Errorf("ClientError: code=%d, msg=%v", bizCode, errMsg)
	httpx.WriteJsonCtx(ctx, w, http.StatusBadRequest, Response{
		Code: bizCode,
		Msg:  errMsg,
		Data: nil,
	})
}

func LogicError(ctx context.Context, w http.ResponseWriter, err error) {
	var bizError *BizError
	if errors.As(err, &bizError) {
		logx.WithContext(ctx).Errorf("Business Warning: code=%d, msg=%v", bizError.Code, bizError.Msg)
		httpx.WriteJsonCtx(ctx, w, http.StatusBadRequest, Response{
			Code: bizError.Code,
			Msg:  bizError.Msg,
			Data: nil,
		})
		return
	} else {
		st, ok := status.FromError(err)
		if ok {
			for _, detail := range st.Details() {
				switch e := detail.(type) {
				case *user.ErrDetail:
					// 解析到rpc业务错误
					httpx.WriteJsonCtx(ctx, w, http.StatusOK, Response{
						Code: int(e.Code),
						Msg:  e.Msg,
						Data: nil,
					})
					return
				}
			}
		}
	}
	logx.WithContext(ctx).Errorf("InternalError: error=%v", err)
	httpx.WriteJsonCtx(ctx, w, http.StatusInternalServerError, Response{
		Code: InternalError,
		Msg:  "Internal Error",
		Data: nil,
	})
	return
}

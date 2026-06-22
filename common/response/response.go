package response

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

const (
	InternalError = 500
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
	ErrCodeInvalidObject    = 11007 // object格式错误
)

// 认证与授权错误 (12xxx)
const (
	ErrCodeUnauthorized      = 12001 // 未认证，请先登录
	ErrCodeTokenInvalid      = 12002 // Token 无效
	ErrCodePermissionDenied  = 12005 // 权限不足
	ErrCodePasswordWrong     = 12006 // 账号或密码错误
	ErrCodeCaptchaWrong      = 12007 // 验证码错误
	ErrCodePhoneRegistered   = 12010 // 手机号已注册
	ErrCodeTooFrequent       = 12011 // 访问过于频繁
	ErrCodePasswordNotMatch  = 12012 // 两次输入密码不一致
	ErrCodeSameAsOldPassword = 12013 // 新密码与旧密码一致
)

// 用户模块 (2xxxx)
const (
	ErrCodeUserNotFound          = 20001 // 用户不存在
	ErrCodeUserDisabled          = 20002 // 账号已被禁用
	ErrCodeUserDeleted           = 20004 // 账号已注销
	ErrCodeUserRestricted        = 20005 // 账号受限（限制下单等）
	ErrCodeUserUnbanned          = 20009 // 账号已解封
	ErrCodeUserPunishLogNotFound = 20010 // 风控日志不存在
)

// 商品模块 (3xxxx)
const (
	ErrCodeProductNotFound     = 30001 // 商品不存在
	ErrCodeProductOffline      = 30002 // 商品已下架
	ErrCodeSkuNotFound         = 30004 // SKU 不存在
	ErrCodeCategoryNotFound    = 30007 // 分类不存在
	ErrCodeCategoryHasChildren = 30008 // 分类含有子分类，无法删除
	ErrCodeCategoryDisabled    = 30011 // 分类被禁用
	ErrCodeInvalidQuantity     = 30012 // 数量非法
)

// 订单模块 (4xxxx)
const (
	ErrCodeOrderNotFound            = 40001 // 订单不存在
	ErrCodeOrderStatusInvalid       = 40002 // 订单状态不允许当前操作
	ErrCodeOrderCancelFailed        = 40003 // 订单取消失败（非待支付状态）
	ErrCodeOrderExpired             = 40005 // 订单已过期
	ErrCodeIdempotentConflict       = 40008 // 重复提交订单（幂等冲突）
	ErrCodeAddressNotBelongUser     = 40010 // 地址不属于当前用户
	ErrCodeFreightTemplateNotFound  = 40011 // 运费模板不存在
	ErrCodeFreightCalculationFailed = 40012 // 运费计算失败
	ErrCodePreOrderFailed           = 40013 // 预下单失败
)

// 支付模块 (5xxxx)
const (
	ErrCodePaymentNotFound     = 50001 // 支付单不存在
	ErrCodeRefundNotFound      = 50007 // 退款单不存在
	ErrCodeRefundStatusInvalid = 50008 // 退款状态不允许操作
	ErrCodeRefundAmountExceed  = 50009 // 退款金额超过可退金额
)

// 营销/优惠券模块 (6xxxx)
const (
	ErrCodeCouponNotFound       = 60001 // 优惠券不存在
	ErrCodeCouponExpired        = 60002 // 优惠券已过期
	ErrCodeCouponNotStarted     = 60003 // 优惠券未开始
	ErrCodeCouponAlreadyClaimed = 60004 // 优惠券已领取过
	ErrCodeCouponStockExhausted = 60005 // 优惠券库存不足
	ErrCodeCouponNotAvailable   = 60006 // 优惠券不可用（不满足门槛等）
)

// 地址模块 (7xxxx)
const (
	ErrCodeAddressNotFound      = 70001 // 地址不存在
	ErrCodeAddressLimitExceeded = 70002 // 地址数量超出限制（最多20条）
	ErrCodeDefaultAddressDelete = 70003 // 默认地址不可直接删除，请先设置其他默认地址
)

// 管理员模块 (10xxx)
const (
	ErrCodeAdminNotFound       = 10001 // 管理员不存在
	ErrCodeAdminDisabled       = 10002 // 管理员已被禁用
	ErrCodeAdminUsernameExists = 10003 // 管理员用户名已存在
	ErrCodeRoleNotFound        = 10004 // 角色不存在
)

// 售后模块 (9xxxx)
const (
	ErrCodeAfterSaleNotFound      = 90001 // 售后单不存在
	ErrCodeAfterSaleStatusInvalid = 90002 // 售后单状态不允许操作
	ErrCodeAfterSaleAlreadyExist  = 90003 // 该商品已申请过售后
	ErrCodeAfterSaleReasonInvalid = 90005 // 售后原因无效
	ErrCodeAmountInvalid          = 90007 // 金额无效
	ErrCodeQuantityInvalid        = 90008 // 数量无效
	ErrCodeAfterSaleTypeInvalid   = 90009 // 仅退货退款订单可提交物流信息
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
		httpx.WriteJsonCtx(ctx, w, http.StatusOK, Response{
			Code: bizError.Code,
			Msg:  bizError.Msg,
			Data: nil,
		})
		return
	} else {
		if code, msg, ok := errorx.ParseBizError(err); ok {
			httpx.WriteJsonCtx(ctx, w, http.StatusOK, Response{
				Code: int(code),
				Msg:  msg,
				Data: nil,
			})
			return
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

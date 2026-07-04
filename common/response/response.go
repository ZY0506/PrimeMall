package response

import (
	"context"
	"errors"
	"github.com/ZY0506/PrimeMall/common/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
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
	httpx.WriteJsonCtx(ctx, w, http.StatusOK, Response{ // 统一返回 200，前端通过 code 区分
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

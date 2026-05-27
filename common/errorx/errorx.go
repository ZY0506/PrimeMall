package errorx

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bizCodeSep = "||"

// NewBizError 创建带业务错误码的 gRPC 错误
// 业务码和消息编码在 status.Message 中，通过分隔符解析，避免 common 包依赖具体 proto 类型
func NewBizError(code int32, msg string) error {
	st := status.New(codes.InvalidArgument, fmt.Sprintf("%d%s%s", code, bizCodeSep, msg))
	return st.Err()
}

// ParseBizError 从 error 中解析业务错误码和消息
func ParseBizError(err error) (code int32, msg string, ok bool) {
	st, ok := status.FromError(err)
	if !ok {
		return 0, "", false
	}
	if st.Code() != codes.InvalidArgument {
		return 0, "", false
	}
	parts := strings.SplitN(st.Message(), bizCodeSep, 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	c, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return 0, "", false
	}
	return int32(c), parts[1], true
}

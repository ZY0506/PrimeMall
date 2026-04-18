package errorx

import (
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewBizError(code int32, msg string) error {
	st := status.New(codes.InvalidArgument, msg)

	detail := &user.ErrDetail{
		Code: code,
		Msg:  msg,
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		return st.Err()
	}

	return stWithDetails.Err()
}

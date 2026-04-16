package interceptor

import (
	"context"
	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientInfoUnaryInterceptor gRPC 客户端一元拦截器，自动注入 ClientInfo 到 metadata
func ClientInfoUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 从 HTTP Context 中获取 ClientInfo
		if clientInfo, ok := ctx.Value(ctxdata.ContextKeyClientInfo).(*ctxdata.ClientInfo); ok && clientInfo != nil {
			// 将 ClientInfo 注入到 gRPC metadata
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.New(nil)
			} else {
				md = md.Copy()
			}

			md.Set(ctxdata.ContextKeyClientIp, clientInfo.IP)
			md.Set(ctxdata.ContextKeyDeviceID, clientInfo.UserAgent)
			md.Set(ctxdata.ContextKeyClientUserAgent, clientInfo.DeviceID)

			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

package libgrpc

import (
	"context"

	"hp-backend-lib/log"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func Dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			if imd, ok := metadata.FromIncomingContext(ctx); ok {
				for k, v := range imd {
					ctx = metadata.AppendToOutgoingContext(ctx, k, v[0])
				}
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}))
	if err != nil {
		liblog.Fatalf("failed to dial %s: %v", addr, err)
	}
	return conn
}

func UnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	return grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))
}

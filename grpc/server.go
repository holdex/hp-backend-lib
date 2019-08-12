package libgrpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/holdex/hp-backend-lib/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func NewServer(opt ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(append(opt,
		UnaryInterceptors(
			MakeLoggingUnaryServerInterceptor(),
			ValidatorUnaryServerInterceptor,
			grpc_auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
				return ctx, status.Error(codes.Unimplemented, "authentication not implemented")
			}),
		))...)
}

func Serve(grpcServer *grpc.Server, listenAddr string) {
	reflection.Register(grpcServer)

	errCh := make(chan error)
	go func() {
		netListener, err := net.Listen("tcp", listenAddr)
		if err != nil {
			errCh <- fmt.Errorf("listen on %s: %v", listenAddr, err)
			return
		}
		liblog.Infof("serving on %s", listenAddr)
		if err := grpcServer.Serve(netListener); err != nil {
			errCh <- fmt.Errorf("grpc serve: %v", err)
		}
		close(errCh)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-c
		close(errCh)
	}()

	if err := <-errCh; err != nil {
		liblog.Fatalln(err)
	}
}

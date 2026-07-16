package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"qrcode-generation-service/pkg/logger"
)

type Server struct {
	GrpcServer *grpc.Server
}

func NewServer() *Server {
	return &Server{
		GrpcServer: grpc.NewServer(grpc.UnaryInterceptor(recoveryInterceptor)),
	}
}

// recoveryInterceptor turns a panic in a handler into an Internal error rather
// than letting it crash the server process.
func recoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("recovered panic in %s: %v", info.FullMethod, r)
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(ctx, req)
}

// Run runs gRPC server.
func (s *Server) Run(l net.Listener) error {
	return s.GrpcServer.Serve(l)
}

// Stop stops gRPC server.
func (s *Server) Stop() {
	s.GrpcServer.GracefulStop()
}

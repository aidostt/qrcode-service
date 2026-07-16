package server

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"qrcode-generation-service/pkg/logger"
)

type Server struct {
	GrpcServer *grpc.Server
}

func NewServer() *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(loggingInterceptor, recoveryInterceptor),
	)
	registerHealth(grpcServer)
	return &Server{GrpcServer: grpcServer}
}

// registerHealth exposes the standard gRPC health service so orchestrators can
// probe the server.
func registerHealth(s *grpc.Server) {
	h := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, h)
	h.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
}

// loggingInterceptor emits one structured access-log line per call, carrying the
// request id propagated by the gateway so a request can be correlated across
// services.
func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	dur := time.Since(start)
	code := status.Code(err)
	logger.L().Info("grpc request",
		"method", info.FullMethod,
		"request_id", requestIDFromContext(ctx),
		"duration_ms", dur.Milliseconds(),
		"code", code.String(),
	)
	recordMetrics(info.FullMethod, code.String(), dur)
	return resp, err
}

func requestIDFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if v := md.Get("x-request-id"); len(v) > 0 {
			return v[0]
		}
	}
	return ""
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

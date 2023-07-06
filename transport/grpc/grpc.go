package grpc_transport

import (
	"fmt"
	"net"
	"time"

	"github.com/Moranilt/jwt-http2/jwt_gRPC"
	"github.com/Moranilt/jwt-http2/middleware"
	service "github.com/Moranilt/jwt-http2/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Transport struct {
	*grpc.Server
}

func New(service *service.Server, mw *middleware.Middleware) *Transport {
	server := &Transport{
		Server: grpc.NewServer(grpc.ConnectionTimeout(10*time.Second), grpc.UnaryInterceptor(mw.UnaryInterceptor)),
	}
	jwt_gRPC.RegisterAuthenticationServer(server, service)
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	reflection.Register(server.Server)

	return server
}

func (s *Transport) MakeListener(port string) (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	return lis, nil
}

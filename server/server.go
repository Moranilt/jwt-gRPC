package server

import (
	"context"

	jwt_http2 "github.com/Moranilt/jwt-http2/auth"
	"github.com/Moranilt/jwt-http2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Service struct {
	jwt_http2.UnimplementedAuthenticationServer
	log         *logger.Logger
	redis       *redis.Client
	publicCert  []byte
	privateCert []byte
}

func New(log *logger.Logger, r *redis.Client, public, private []byte) *Service {
	return &Service{
		log:         log,
		redis:       r,
		publicCert:  public,
		privateCert: private,
	}
}

func (s *Service) CreateTokens(ctx context.Context, req *jwt_http2.CreateTokensRequest) (*jwt_http2.CreateTokensResponse, error) {
	s.log.WithRequestInfo(ctx).WithFields(logrus.Fields{
		"req": req,
	}).Info()

	return &jwt_http2.CreateTokensResponse{
		AccessToken:  "token",
		RefreshToken: "rtoken",
	}, nil
}

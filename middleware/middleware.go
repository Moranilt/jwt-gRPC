package middleware

import (
	"context"
	"time"

	"github.com/Moranilt/jwt-http2/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Middleware struct {
	log *logger.Logger
}

func New(log *logger.Logger) *Middleware {
	return &Middleware{
		log: log,
	}
}

func (m *Middleware) UnaryInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	reqID := uuid.NewString()
	start := time.Now()
	newCtx := context.WithValue(ctx, logger.CtxRequestId, reqID)

	h, err := handler(newCtx, req)

	if err != nil {
		m.log.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": time.Since(start),
			"error":    err.Error(),
			"req":      req,
			"id":       reqID,
		}).Error()
	} else {
		m.log.WithFields(logrus.Fields{
			"method":   info.FullMethod,
			"duration": time.Since(start),
			"req":      req,
			"id":       reqID,
		}).Info()
	}

	return h, err
}

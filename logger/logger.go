package logger

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type ContextKey string

const (
	CtxRequestId ContextKey = "request_id"
)

type Logger struct {
	logrus.Logger
}

func New() *Logger {
	log := Logger{}
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout

	return &log
}

func (l *Logger) WithRequestInfo(ctx context.Context) *logrus.Entry {
	requestId := ctx.Value(CtxRequestId)

	return l.WithFields(logrus.Fields{
		"id": requestId,
	})
}

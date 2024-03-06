package logger

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

import (
	"context"
)

type ctxLoggerKey struct{}
type ctxLoggerValue struct {
	logger *zap.Logger
}

var loggerKey ctxLoggerKey

const (
	traceIDField    = "trace.id"
	spanIDField     = "span.id"
	traceFlagsField = "trace.flags"
)

// SetTraceInfoInterceptor ...
func SetTraceInfoInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ctx = context.WithValue(ctx, loggerKey, ctxLoggerValue{logger: logger})
		return handler(ctx, req)
	}
}

// Extract ...
func Extract(ctx context.Context) *zap.Logger {
	val, ok := ctx.Value(loggerKey).(ctxLoggerValue)
	if !ok {
		return zap.NewNop()
	}
	return val.logger
}

// WrapError ...
func WrapError(ctx context.Context, err error) {
	Extract(ctx).WithOptions(zap.AddCallerSkip(2)).
		Error("WrapError", zap.Error(err))
}

// ToContext ...
func ToContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, ctxLoggerValue{logger: l})
}

// GetRawLogger ...
func GetRawLogger(ctx context.Context) *zap.Logger {
	val, ok := ctx.Value(loggerKey).(ctxLoggerValue)
	if !ok {
		return zap.NewNop()
	}
	return val.logger
}

// Debug ...
func Debug(ctx context.Context, msg string) {
	Extract(ctx).Sugar().Debug(msg)
}

// Debugf ...
func Debugf(ctx context.Context, format string, args ...interface{}) {
	Extract(ctx).Sugar().Debugf(format, args)
}

// Info ...
func Info(ctx context.Context, msg string) {
	Extract(ctx).Sugar().Info(msg)
}

// Infof ...
func Infof(ctx context.Context, format string, args ...interface{}) {
	Extract(ctx).Sugar().Infof(format, args...)
}

// Warn ...
func Warn(ctx context.Context, msg string) {
	Extract(ctx).Sugar().Info(msg)
}

// Warnf ...
func Warnf(ctx context.Context, format string, args ...interface{}) {
	Extract(ctx).Sugar().Warnf(format, args)
}

// Error ...
func Error(ctx context.Context, msg string) {
	Extract(ctx).Sugar().Error(msg)
}

// Errorf ...
func Errorf(ctx context.Context, format string, args ...interface{}) {
	Extract(ctx).Sugar().Errorf(format, args)
}

// Fatal ...
func Fatal(ctx context.Context, msg string) {
	Extract(ctx).Sugar().Fatal(msg)
}

// Fatalf ...
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	Extract(ctx).Sugar().Fatalf(format, args)
}

// Panic ...
func Panic(ctx context.Context, msg string) {
	Extract(ctx).Sugar().Panic(msg)
}

// Panicf ...
func Panicf(ctx context.Context, format string, args ...interface{}) {
	Extract(ctx).Sugar().Panicf(format, args)
}

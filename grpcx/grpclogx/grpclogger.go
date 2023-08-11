package grpclogx

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"log/slog"
)

type GrpcLogger struct {
	v        int
	delegate slog.Handler
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (gl *GrpcLogger) Info(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (gl *GrpcLogger) Infoln(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (gl *GrpcLogger) Infof(format string, args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, fmt.Sprintf(format, args...), 0)
	gl.delegate.Handle(ctx, record)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (gl *GrpcLogger) Warning(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelWarn, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (gl *GrpcLogger) Warningln(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelWarn, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (gl *GrpcLogger) Warningf(format string, args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelWarn, fmt.Sprintf(format, args...), 0)
	gl.delegate.Handle(ctx, record)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (gl *GrpcLogger) Error(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelError, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (gl *GrpcLogger) Errorln(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelError, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (gl *GrpcLogger) Errorf(format string, args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), 0)
	gl.delegate.Handle(ctx, record)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (gl *GrpcLogger) Fatal(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelError, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
	os.Exit(1)
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (gl *GrpcLogger) Fatalln(args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelError, "", 0)
	record.Add(args...)
	gl.delegate.Handle(ctx, record)
	os.Exit(1)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (gl *GrpcLogger) Fatalf(format string, args ...interface{}) {
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelError, fmt.Sprintf(format, args...), 0)
	gl.delegate.Handle(ctx, record)
	os.Exit(1)
}

// V reports whether verbosity level l is at least the requested verbose level.
// 返回false，则不打印 指定 verbose 等级的日志
func (gl *GrpcLogger) V(l int) bool {
	return l <= gl.v
}

func NewGrpcLog(l slog.Handler) *GrpcLogger {
	var v int
	vLevel := os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")
	if vl, err := strconv.Atoi(vLevel); err == nil {
		v = vl
	}
	return &GrpcLogger{
		v:        v,
		delegate: l,
	}
}

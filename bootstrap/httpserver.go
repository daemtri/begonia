package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

type HttpServerRunOption struct {
	Addr string `flag:"addr" default:"0.0.0.0:80" usage:"http服务监听地址"`
}

type HttpServer struct {
	addr    string
	handler http.Handler
	server  http.Server
}

func NewHttpServer(ctx context.Context, opt *HttpServerRunOption, handler http.Handler) (*HttpServer, error) {
	return &HttpServer{
		addr:    opt.Addr,
		handler: handler,
		server: http.Server{
			Addr:    opt.Addr,
			Handler: handler,
			BaseContext: func(net.Listener) context.Context {
				return ctx
			},
		},
	}, nil
}

func (gs *HttpServer) Enabled() bool {
	if r, ok := gs.handler.(interface{ Enabled() bool }); ok {
		return r.Enabled()
	}
	return true
}

func (gs *HttpServer) BroadCastAddr() string {
	_, port, err := net.SplitHostPort(gs.addr)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("http://:%s", port)
}

func (gs *HttpServer) Run(ctx context.Context) error {
	logger.Info("http server listening", "addr", gs.addr)
	lis, err := net.Listen("tcp", gs.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	return gs.server.Serve(lis)
}

func (gs *HttpServer) GracefulStop() {
	logger.Info("http server shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := gs.server.Shutdown(ctx); err != nil {
		logger.Error("http server shutdown error: %v", err)
	}
}

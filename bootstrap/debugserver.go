package bootstrap

import (
	"context"
	"net/http"
	"time"

	"github.com/arl/statsviz"
	"github.com/maruel/panicparse/v2/stack/webstack"
)

type DebugServer struct {
	Enable bool   `flag:"enable" default:"false" usage:"是否开启debug http服务"`
	Addr   string `flag:"addr" default:":8078" usage:"debug http服务监听地址"`

	server *http.Server
}

func (s *DebugServer) Enabled() bool {
	return s.Enable
}

func (s *DebugServer) BroadCastAddr() string {
	return ""
}

// Run 启动服务
func (s *DebugServer) Run(_ context.Context) error {
	// 通过 http://127.0.0.1:8078/debug/statsviz/ 访问debug信息
	if err := statsviz.Register(http.DefaultServeMux); err != nil {
		return err
	}
	// 漂亮的Goroutine打印
	http.HandleFunc("/debug/panicparse", webstack.SnapshotHandler)
	s.server = &http.Server{
		Handler: http.DefaultServeMux,
		Addr:    s.Addr,
	}
	logger.Info("debug server已启动", "addr", s.Addr)
	return s.server.ListenAndServe()
}

func (s *DebugServer) GracefulStop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}

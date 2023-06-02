package bootstrap

import (
	"context"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/container"
	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
	"golang.org/x/sync/errgroup"
)

var (
	logger = logx.GetLogger("bootstrap")
)

type Server interface {
	Enabled() bool
	BroadCastAddr() string
	Run(ctx context.Context) error
	GracefulStop()
}

type Engine interface {
	Run(ctx context.Context) error
}

func NewEngine(ctx context.Context) (Engine, error) {
	return &EngineImpl{
		servers: container.Invoke[container.Set[Server]](ctx),
	}, nil
}

type EngineImpl struct {
	servers []Server
}

func (engine *EngineImpl) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	for _, server := range engine.servers {
		s := server
		if s.Enabled() {
			group.Go(func() error { return s.Run(ctx) })
		}
	}

	group.Go(func() error {
		defer logx.Recover(logger)
		<-ctx.Done()
		for _, server := range engine.servers {
			if server.Enabled() {
				server.GracefulStop()
			}
		}
		return nil
	})

	return group.Wait()
}

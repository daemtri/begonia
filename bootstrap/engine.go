package bootstrap

import (
	"context"

	"github.com/daemtri/begonia/di/container"
	"github.com/daemtri/begonia/logx"
	"golang.org/x/sync/errgroup"
)

var (
	logger = logx.GetLogger("bootstrap")
)

type Runable interface {
	Enabled() bool
	Run(ctx context.Context) error
	GracefulStop()
}

type Server interface {
	Runable
	BroadCastAddr() string
}

type Engine interface {
	Run(ctx context.Context) error
}

func NewEngine(ctx context.Context) (Engine, error) {
	return &EngineImpl{
		runables: container.Invoke[container.Set[Runable]](ctx),
	}, nil
}

type EngineImpl struct {
	runables []Runable
}

func (engine *EngineImpl) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	for _, runable := range engine.runables {
		s := runable
		if s.Enabled() {
			group.Go(func() error { return s.Run(ctx) })
		}
	}

	group.Go(func() error {
		defer logx.Recover(logger)
		<-ctx.Done()
		for _, runable := range engine.runables {
			if runable.Enabled() {
				runable.GracefulStop()
			}
		}
		return nil
	})

	return group.Wait()
}

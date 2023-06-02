package service

import (
	"net/http"

	"git.bianfeng.com/stars/wegame/wan/wanx/example/box_example/contract"
	"golang.org/x/exp/slog"
)

type ConsumerService struct {
	logger *slog.Logger
}

func NewConsumerService(repo contract.UserRepository, logger *slog.Logger) (*ConsumerService, error) {
	return &ConsumerService{logger: logger}, nil
}

func (c *ConsumerService) AddRoute(mux *http.ServeMux) {
	c.logger.Info("ConsumerService add Route")
}

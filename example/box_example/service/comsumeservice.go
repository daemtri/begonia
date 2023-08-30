package service

import (
	"net/http"

	"log/slog"

	"github.com/daemtri/begonia/example/box_example/contract"
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

package bootstrap

import (
	"context"
	"fmt"

	"github.com/daemtri/begonia/api/transmit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BusinessService struct {
	transmit.UnimplementedBusinessServiceServer

	rr *RouteRegistrar
}

func NewBusinessService(rr *RouteRegistrar) (*BusinessService, error) {
	b := &BusinessService{
		rr: rr,
	}
	return b, nil
}

func (bs *BusinessService) Dispatch(ctx context.Context, req *transmit.DispatchRequest) (*transmit.DispatchReply, error) {
	h, ok := bs.rr.routes[req.Msgid]
	if !ok {
		return nil, status.Error(codes.Unimplemented, fmt.Sprintf("unknown msgid %d", req.Msgid))
	}

	if err := h(ctx, req.Data); err != nil {
		return nil, status.Convert(err).Err()
	}
	return &transmit.DispatchReply{}, nil
}

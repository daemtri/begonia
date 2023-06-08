package bootstrap

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type GrpcServer struct {
	addr   string
	server *grpc.Server
}

func (gs *GrpcServer) Init(addr string, server *grpc.Server) {
	gs.addr = addr
	gs.server = server
}

func (gs *GrpcServer) Enabled() bool {
	return len(gs.server.GetServiceInfo()) > 0
}

func (gs *GrpcServer) BroadCastAddr() string {
	_, port, err := net.SplitHostPort(gs.addr)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("grpc://:%s", port)
}

func (gs *GrpcServer) Run(ctx context.Context) error {
	logger.Info("grpc server listening on %s", gs.addr)
	lis, err := net.Listen("tcp", gs.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	return gs.server.Serve(lis)
}

func (gs *GrpcServer) GracefulStop() {
	logger.Info("grpc server shutdown")
	gs.server.GracefulStop()
}

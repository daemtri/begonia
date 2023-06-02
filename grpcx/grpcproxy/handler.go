package grpcproxy

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
)

var (
	clientStreamDescForProxying = &grpc.StreamDesc{
		ServerStreams: true,
		ClientStreams: true,
	}
)

// RegisterService sets up a proxy handler for a particular gRPC service and method.
// The behaviour is the same as if you were registering a handler method, e.g. from a generated pb.go file.
func RegisterService(server *grpc.Server, director StreamDirector, serviceName string, methodNames ...string) {
	streamer := &handler{director}
	fakeDesc := &grpc.ServiceDesc{
		ServiceName: serviceName,
		HandlerType: (*interface{})(nil),
	}
	for _, m := range methodNames {
		streamDesc := grpc.StreamDesc{
			StreamName:    m,
			Handler:       streamer.handle,
			ServerStreams: true,
			ClientStreams: true,
		}
		fakeDesc.Streams = append(fakeDesc.Streams, streamDesc)
	}
	server.RegisterService(fakeDesc, streamer)
}

// TransparentHandler returns a handler that attempts to proxy all requests that are not registered in the server.
// The indented use here is as a transparent proxy, where the server doesn't know about the services implemented by the
// backends. It should be used as a `grpc.UnknownServiceHandler`.
func TransparentHandler(director StreamDirector) grpc.StreamHandler {
	streamer := &handler{director: director}
	return streamer.handle
}

type handler struct {
	director StreamDirector
}

// handler is where the real magic of proxying happens.
// It is invoked like any gRPC server stream and uses the emptypb.Empty type server
// to proxy calls between the input and output streams.
func (s *handler) handle(_ interface{}, serverStream grpc.ServerStream) error {
	// 获取请求流的目的接口名称
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}
	// We require that the director's returned context inherits from the serverStream.Context().
	outgoingCtx, backendConn, err := s.director(serverStream.Context(), fullMethodName)
	if err != nil {
		return err
	}

	clientCtx, clientCancel := context.WithCancel(outgoingCtx)
	defer clientCancel()

	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, backendConn, fullMethodName)
	if err != nil {
		return err
	}

	s2cErrChan := make(chan error, 1)
	defer close(s2cErrChan)
	go func(s2cErrChan chan<- error) {
		// 启动流控 请求方(serverStream)->服务方(clientStream)
		s2cErrChan <- s.forwardServerToClient(serverStream, clientStream)
	}(s2cErrChan)

	// 启动流控，服务方(clientStream)->请求方(serverStream)
	c2sErr := s.forwardClientToServer(clientStream, serverStream)
	s2cErr := <-s2cErrChan
	if s2cErr != nil {
		return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
	}
	if c2sErr != nil {
		//return status.Errorf(codes.Internal, "failed proxying c2s: %v", c2sErr)
		return c2sErr
	}
	return nil
}

func (s *handler) forwardServerToClient(ss grpc.ServerStream, cs grpc.ClientStream) error {
	f := &emptypb.Empty{}
	for {
		if err := ss.RecvMsg(f); err != nil {
			// 正常情况，应该是这里返回io.EOF
			if err == io.EOF {
				cs.CloseSend()
				return nil
			}
			return err
		}
		if err := cs.SendMsg(f); err != nil {
			// 如果发送消息出错,cs.RecvMsg也会出错,所以我们不用管
			return err
		}
	}
}

func (s *handler) forwardClientToServer(cs grpc.ClientStream, ss grpc.ServerStream) error {
	defer func() {
		// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
		// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
		// will be nil.
		ss.SetTrailer(cs.Trailer())
	}()
	f := &emptypb.Empty{}
	if err := cs.RecvMsg(f); err != nil {
		return err
	}
	// grpc中客户端到服务器的header只能在第一个客户端消息后才可以读取到，
	// 同时又必须在flush第一个msg之前写入到流中。
	md, err := cs.Header()
	if err != nil {
		return err
	}
	if err := ss.SendHeader(md); err != nil {
		return err
	}
	if err := ss.SendMsg(f); err != nil {
		return err
	}
	for {
		if err := cs.RecvMsg(f); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := ss.SendMsg(f); err != nil {
			return err
		}
	}
}

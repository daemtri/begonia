package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/daemtri/begonia/bootstrap"
	"github.com/daemtri/begonia/contract"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type Integrator struct {
	Grpc   GrpcServiceRegistrar
	Http   chi.Router
	PubSub contract.PubSubConsumerRegistrar
	Task   contract.TaskProcessorRegistrar
}

func newIntegrator(
	lsr GrpcServiceRegistrar,
	psr contract.PubSubConsumerRegistrar,
	tpr contract.TaskProcessorRegistrar,
	mux chi.Router,
) (*Integrator, error) {
	reg := &Integrator{
		Grpc:   lsr,
		PubSub: psr,
		Task:   tpr,
		Http:   mux,
	}
	return reg, nil
}

func (it *Integrator) integrate(mr *moduleRuntime) {
	currentModule = mr
	it.Http.With(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r2 := r.WithContext(withObjectContainer(r.Context(), currentModule))
			h.ServeHTTP(w, r2)
		})
	}).Route("/"+mr.moduleName, func(r chi.Router) {
		mr.module.Integrate(Integrator{
			Grpc:   it.Grpc,
			PubSub: it.PubSub,
			Task:   it.Task,
			Http:   r,
		})
	})
	currentModule = nil
}

type GrpcServiceRegistrar interface {
	contract.RouteRegistrar
	grpc.ServiceRegistrar
}

type grpcServiceRegistrarImpl struct {
	ci      *bootstrap.ContextInjector
	route   contract.RouteRegistrar
	service grpc.ServiceRegistrar
}

func newGrpcServiceRegistrarImpl(
	rr *bootstrap.RouteRegistrar,
	sr *bootstrap.ServiceRegistrar,
	ci *bootstrap.ContextInjector,
) (*grpcServiceRegistrarImpl, error) {
	lsri := &grpcServiceRegistrarImpl{
		route:   rr,
		service: sr,
		ci:      ci,
	}
	return lsri, nil
}

func (gr *grpcServiceRegistrarImpl) RegisterService(desc *grpc.ServiceDesc, impl any) {
	mr := currentModule
	gr.ci.Bind(desc.ServiceName, func(ctx context.Context) context.Context {
		return withObjectContainer(ctx, mr)
	})
	gr.service.RegisterService(desc, impl)
}

func Route[K ~int32, T proto.Message](msgID K, handleFunc func(ctx context.Context, req T) error) contract.RouteCell {
	mr := currentModule
	return contract.RouteCell{
		MsgID: int32(msgID),
		HandleFunc: func(ctx context.Context, req []byte) error {
			var x T
			v := x.ProtoReflect().New().Interface()
			if req != nil {
				if err := proto.Unmarshal(req, v); err != nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("message id %d does not match the message type, error %s", msgID, err))
				}
			}
			return handleFunc(withObjectContainer(ctx, mr), v.(T))
		},
	}
}

func (gr *grpcServiceRegistrarImpl) RegisterRoute(routes ...contract.RouteCell) {
	gr.route.RegisterRoute(routes...)
}

type httpServerMux struct {
	chi.Router
}

func newHttpServerMux() (*httpServerMux, error) {
	return &httpServerMux{
		Router: chi.NewRouter(),
	}, nil
}

func (mux *httpServerMux) Enabled() bool {
	logger.Info("http server enable check, routes", "routes", len(mux.Routes()))
	return len(mux.Routes()) > 0
}

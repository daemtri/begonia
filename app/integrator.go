package app

import (
	"net/http"

	"git.bianfeng.com/stars/wegame/wan/wanx/contract"
	"github.com/go-chi/chi/v5"
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
			r2 := r.WithContext(withModuleRuntime(r.Context(), currentModule))
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

package specify

import (
	"google.golang.org/grpc/balancer"
)

type BalancerBuilder struct{}

func (s BalancerBuilder) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	return &Balancer{
		cc:   cc,
		opts: opts,
	}
}

func (s BalancerBuilder) Name() string {
	return "specify"
}

type Balancer struct {
	cc   balancer.ClientConn
	opts balancer.BuildOptions

	state balancer.ClientConnState
}

func (b *Balancer) UpdateClientConnState(state balancer.ClientConnState) error {
	b.state = state
	return nil
}

func (b *Balancer) ResolverError(err error) {

}

func (b *Balancer) UpdateSubConnState(conn balancer.SubConn, state balancer.SubConnState) {

}

func (b *Balancer) Close() {

}

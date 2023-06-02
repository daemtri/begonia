package grpcresolver

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"google.golang.org/grpc/resolver"
)

// RegisterInCluster registers the solver builder to grpc with sgr schema
func RegisterInCluster(name string, reg component.Discovery) {
	resolver.Register(NewBuilder(name, reg))
}

// mdnsBuilder implements the builder interface of grpc resolver
type sgrBuilder struct {
	schema    string
	registrar component.Discovery
}

// Build creates a new resolver for the given target.
// gRPC dial calls Build synchronously, and fails if the returned error is
// not nil.
func (b *sgrBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	ti, err := parseResolverTarget(target)
	if err != nil {
		return nil, err
	}

	// DNS address (non-IP).
	d := &sgrResolver{
		target:               ti,
		clientConn:           cc,
		disableServiceConfig: opts.DisableServiceConfig,
		discovery:            b.registrar,
	}

	return d, d.Init()
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *sgrBuilder) Scheme() string {
	return b.schema
}

func NewBuilder(name string, reg component.Discovery) resolver.Builder {
	return &sgrBuilder{schema: name, registrar: reg}
}

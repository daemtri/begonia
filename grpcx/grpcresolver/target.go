package grpcresolver

import (
	"errors"

	"google.golang.org/grpc/resolver"
)

type targetInfo struct {
	instanceName string
	serviceName  string
	domain       string
	target       resolver.Target
}

func parseResolverTarget(target resolver.Target) (*targetInfo, error) {
	// relay://serviceName/endpointPrefix
	endpoint := target.URL.Path
	authority := target.URL.Host

	if authority == "" {
		return nil, errors.New("could not parse target. Invalid Authority")
	}
	if endpoint == "" {
		return nil, errors.New("could not parse target. Invalid Endpoint")
	}

	ti := &targetInfo{
		target:      target,
		serviceName: authority,
	}
	return ti, nil
}

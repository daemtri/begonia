package specify

import (
	"strings"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// PolicyMetadataKey 是grpc metadata key，
	// 对应的value形式为 KEY=VALUE, KEY和VALUE从SubConnInfo的Address的Metadata从获取
	// example:
	//		Specify-Policy: app-id=123
	PolicyMetadataKey = "Specify-Policy"
)

func init() {
	balancer.Register(base.NewBalancerBuilder("specify", PickBuilder{}, base.Config{HealthCheck: true}))
}

type PickBuilder struct{}

func (p PickBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	return &Picker{SCS: info.ReadySCs}
}

type Picker struct {
	SCS map[balancer.SubConn]base.SubConnInfo
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	result := balancer.PickResult{}
	md, ok := metadata.FromOutgoingContext(info.Ctx)
	if !ok {
		return result, status.Error(codes.InvalidArgument, "balancer specify: 无法获取metadata")
	}
	policy := firstOrEmpty(md.Get(PolicyMetadataKey))
	if policy == "" {
		return result, status.Errorf(codes.InvalidArgument, "balancer specify: policy为空")
	}
	policyArr := strings.SplitN(policy, "=", 2)
	if len(policyArr) != 2 {
		return result, status.Errorf(codes.InvalidArgument, "balancer specify: policy格式错误")
	}
	for k, v := range p.SCS {
		val := v.Address.BalancerAttributes.Value(policyArr[0])
		valStr, ok := val.(string)
		if !ok {
			continue
		}
		if valStr == policyArr[1] {
			result.SubConn = k
			return result, nil
		}
	}
	return result, balancer.ErrNoSubConnAvailable
}

func firstOrEmpty(x []string) string {
	if len(x) == 0 {
		return ""
	}
	return x[0]
}

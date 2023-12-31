package chash

import (
	"context"
	"log"

	"github.com/serialx/hashring"
	"google.golang.org/grpc/balancer"
)

// TODO add lock for maps or use sync.map
type consistentHashPicker struct {
	subConns   map[string]balancer.SubConn // address string -> balancer.SubConn
	hashRing   *hashring.HashRing
	needReport bool
	reportChan chan<- PickResult
}

type PickResult struct {
	Ctx context.Context
	SC  balancer.SubConn
}

func NewConsistentHashPicker(subConns map[string]balancer.SubConn) *consistentHashPicker {
	addrs := make([]string, 0)
	for addr := range subConns {
		addrs = append(addrs, addr)
	}
	log.Printf("consistent hash picker built with addresses %v\n", addrs)
	return &consistentHashPicker{
		subConns:   subConns,
		hashRing:   hashring.New(addrs),
		needReport: false,
	}
}

func NewConsistentHashPickerWithReportChan(subConns map[string]balancer.SubConn, reportChan chan<- PickResult) *consistentHashPicker {
	addrs := make([]string, 0)
	for addr := range subConns {
		addrs = append(addrs, addr)
	}
	log.Printf("consistent hash picker built with addresses %v\n", addrs)
	return &consistentHashPicker{
		subConns:   subConns,
		hashRing:   hashring.New(addrs),
		needReport: true,
		reportChan: reportChan,
	}
}

func (p *consistentHashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var ret balancer.PickResult
	key, ok := info.Ctx.Value(Key).(string)
	if !ok {
		// TODO: 修改为其他错误
		return ret, balancer.ErrNoSubConnAvailable
	}
	log.Printf("pick for %s\n", key)
	if targetAddr, ok := p.hashRing.GetNode(key); ok {
		ret.SubConn = p.subConns[targetAddr]
		if p.needReport {
			p.reportChan <- PickResult{Ctx: info.Ctx, SC: ret.SubConn}
		}
	}
	//ret.SubConn = p.subConns["localhost:50000"]
	if ret.SubConn == nil {
		return ret, balancer.ErrNoSubConnAvailable
	}
	return ret, nil
}

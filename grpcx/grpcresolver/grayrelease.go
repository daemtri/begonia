package grpcresolver

import (
	"git.bianfeng.com/stars/wegame/wan/wanx/pkg/slicemap"
	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"
	"github.com/coreos/go-semver/semver"
)

// grayReleaseHandle 处理同时存在多个版本的情况
func grayReleaseHandle(sis []component.ServiceEntry, grc string) []component.ServiceEntry {
	switch grc {
	case "old_version":
		return filterServiceEntry(sis, 0)
	case "latest_version":
		return filterServiceEntry(sis, 1)
	}
	return sis
}

func filterServiceEntry(sis []component.ServiceEntry, policy int) []component.ServiceEntry {
	if len(sis) == 0 {
		return sis
	}
	slicemap.Sort(sis, func(left, right component.ServiceEntry) bool {
		// TODO: 优化性能
		if policy <= 0 {
			return semver.New(left.Version).LessThan(*semver.New(right.Version))
		}
		return semver.New(right.Version).LessThan(*semver.New(left.Version))
	})
	ret := make([]component.ServiceEntry, 0, len(sis))
	ret = append(ret, sis[0])
	for i := 1; i < len(sis); i++ {
		if sis[i].Version == ret[0].Version {
			ret = append(ret, sis[i])
		} else {
			break
		}
	}
	return ret
}

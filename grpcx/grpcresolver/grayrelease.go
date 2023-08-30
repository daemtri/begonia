package grpcresolver

import (
	"slices"

	"github.com/coreos/go-semver/semver"
	"github.com/daemtri/begonia/runtime/component"
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
	slices.SortFunc(sis, func(left, right component.ServiceEntry) int {
		if policy <= 0 {
			return semver.New(left.Version).Compare(*semver.New(right.Version))
		}
		return semver.New(right.Version).Compare(*semver.New(left.Version))
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

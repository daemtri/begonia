package component

import (
	"context"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Service struct {
	// 服务列表
	Entries []ServiceEntry
	// 服务配置(调度策略)
	Configs []ConfigItem
}

// Discovery 服务注册发现
type Discovery interface {
	Interface

	// Register 注册service，Register只能被调用一次
	Register(ctx context.Context, service ServiceEntry) error
	// Lookup 查询指定id和name的ServiceEntry
	Lookup(ctx context.Context, name, id string) (*ServiceEntry, error)
	// Browse 查询指定name的所有ServiceEntry
	Browse(ctx context.Context, name string) (*Service, error)
	// Watch 监听服务变化
	Watch(ctx context.Context, name string) Iterator[*Service]
}

// ServiceEntry 表示一个APP(Service)在服务发现系统中的一个实例(节点)
type ServiceEntry struct {
	// ID 是注册到服务发现系统中的全局唯一的ID，建议使用UUID
	ID string `json:"id"`
	// Name 为注册到服务发现系统中的名称
	Name string `json:"name"`
	// Alias 服务模块名
	Alias string `json:"alias"`
	// Version 应用版本,形如：v1.0.0
	Version string `json:"version"`
	// Endpoints 地址，如:127.0.0.1:3001
	Endpoints []string `json:"endpoints"`
	// Metadata is the kv pair metadata associated with the service instance.
	// 注意： key 和 value 中不能包含符号`=`
	Metadata map[string]string `json:"metadata"`
}

func (se *ServiceEntry) Equal(se2 *ServiceEntry) bool {
	if se.ID != se2.ID || se.Name != se2.Name {
		return false
	}

	if !slices.Equal(se.Endpoints, se2.Endpoints) {
		return false
	}

	return maps.Equal(se.Metadata, se2.Metadata)
}

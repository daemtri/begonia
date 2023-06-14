package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

type ServiceType string

const (
	// ServiceTypeCluster 服务类型为有状态服务
	ServiceTypeCluster ServiceType = "cluster"
	// ServiceTypeService 服务类型为无状态服务
	ServiceTypeService ServiceType = "service"
)

var ParseServiceType = func(serviceName string) ServiceType {
	ids := strings.TrimPrefix(serviceName, "app")
	id, err := strconv.ParseUint(ids, 16, 32)
	if err != nil {
		panic(fmt.Errorf("invalid service name: %s", serviceName))
	}
	// id是一个十六进制数字子字符串，例如：10A，解析为数字
	idu32 := uint32(id)
	if idu32&0x100 == 0x100 {
		return ServiceTypeCluster
	}
	return ServiceTypeService
}

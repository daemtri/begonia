package servicemesh

import (
	"fmt"
	"strconv"
	"strings"
)

type ServiceType string

const (
	ServiceTypeCluster ServiceType = "cluster"
	ServiceTypeService ServiceType = "service"
)

func ParseServiceTypeFromServiceName(name string) ServiceType {
	ids := strings.TrimPrefix(name, "app")
	id, err := strconv.Atoi(ids)
	if err != nil {
		panic(fmt.Errorf("invalid service name: %s", name))
	}
	idu32 := uint(id)
	// 如果最前面4位为0001则为cluster，如果为0000则为service
	if idu32>>28 == 1 {
		return ServiceTypeCluster
	} else {
		return ServiceTypeService
	}
}

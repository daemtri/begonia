package runtime

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/runtime/component"

	"github.com/segmentio/ksuid"
)

const (
	hostNameKey = "hostname"
	versionKey  = "version"
)

var (
	namespace    string = ""
	serviceEntry        = component.ServiceEntry{
		Metadata: map[string]string{},
	}
)

func init() {
	SetNamespace("default")
	SetServiceName(execname())
	SetServiceAlias(GetServiceName())
	SetServiceID(ksuid.New().String())
	SetServiceMetadata("uptime", time.Now().Format(time.RFC3339))
	SetServiceMetadata("go", runtime.Version())
	info, ok := debug.ReadBuildInfo()
	if ok {
		SetServiceMetadata("vcs.revision", getSettingFromDebugInfo(info, "vcs.revision"))
		SetServiceMetadata("vcs.time", getVcsTimeFromDebugInfo(info))
		SetServiceMetadata("vcs.modified", getSettingFromDebugInfo(info, "vcs.modified"))
	}
	if hostname, err := os.Hostname(); err == nil {
		SetServiceMetadata(hostNameKey, hostname)
	}
	SetServiceMetadata("process", strconv.Itoa(os.Getpid()))
}

func SetNamespace(ns string) {
	if namespace == "" {
		namespace = ns
	}
}

func GetNamespace() string {
	if namespace == "" {
		return "default"
	}
	return namespace
}

func GetName() string {
	return GetServiceAlias()
}

func GetServiceEntry() component.ServiceEntry {
	return serviceEntry
}

func AddServiceEndpoint(endpoint string) {
	serviceEntry.Endpoints = append(serviceEntry.Endpoints, endpoint)
}

func GetServiceEndpoints() []string {
	return serviceEntry.Endpoints
}

func SetServiceID(id string) {
	serviceEntry.ID = id
}

func GetServiceID() string {
	return serviceEntry.ID
}

func SetServiceName(name string) {
	serviceEntry.Name = name
}

func GetServiceName() string {
	return serviceEntry.Name
}

func SetServiceAlias(alias string) {
	serviceEntry.Alias = alias
}

func GetServiceAlias() string {
	return serviceEntry.Alias
}

func SetServiceVersion(ver string) {
	serviceEntry.Version = ver
}

func GetServiceVersion() string {
	return serviceEntry.Version
}

func GetHostName() string {
	val, ok := serviceEntry.Metadata[hostNameKey]
	if !ok {
		return ""
	}
	return val
}

func SetServiceMetadata(k, v string) {
	serviceEntry.Metadata[k] = v
}

func GetServiceMetadata() map[string]string {
	return serviceEntry.Metadata
}

func getVcsTimeFromDebugInfo(info *debug.BuildInfo) string {
	vcsTime := getSettingFromDebugInfo(info, "vcs.time")
	if vcsTime == "" {
		return vcsTime
	}
	t, err := time.ParseInLocation(time.RFC3339, vcsTime, time.UTC)
	if err != nil {
		return vcsTime
	}
	return t.Local().Format(time.RFC3339)
}

func getSettingFromDebugInfo(info *debug.BuildInfo, key string) string {
	for i := range info.Settings {
		if info.Settings[i].Key == key {
			return info.Settings[i].Value
		}
	}
	return ""
}

func execname() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeName := strings.ToLower(filepath.Base(exePath))
	return strings.TrimSuffix(exeName, ".exe")
}

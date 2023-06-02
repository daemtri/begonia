package netx

import (
	"net"
	"strings"
	"time"
)

// DialTimeout 支持连接 tcp或者unix domain socket地址
func DialTimeout(addr string, d time.Duration) (net.Conn, error) {
	if strings.HasPrefix(addr, "unix:") || strings.HasSuffix(addr, ".sock") {
		return net.DialTimeout("unix", strings.TrimPrefix(addr, "unix:"), d)
	}
	return net.DialTimeout("tcp", strings.TrimPrefix(addr, "tcp:"), d)
}

// Dial 支持连接 tcp或者unix domain socket地址
func Dial(addr string) (net.Conn, error) {
	if strings.HasPrefix(addr, "unix:") || strings.HasSuffix(addr, ".sock") {
		return net.Dial("unix", strings.TrimPrefix(addr, "unix:"))
	}
	return net.Dial("tcp", strings.TrimPrefix(addr, "tcp:"))
}

// Listen 支持监听tcp或者unix domain socket地址
func Listen(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, "unix:") || strings.HasSuffix(addr, ".sock") {
		return net.Listen("unix", strings.TrimPrefix(addr, "unix:"))
	}
	return net.Listen("tcp", strings.TrimPrefix(addr, "tcp:"))
}

package netx

import "fmt"

// FindLocalIP 查找局域网IP，如果找到多个，则使用找到的第一个
func FindLocalIP() (string, error) {
	ifaces := ListMulticastInterfaces()
	if len(ifaces) == 0 {
		return "", fmt.Errorf("无法找到支持多播的地址")
	}
	host := "localhost"
	for i := range ifaces {
		v4, v6 := AddrsForInterface(&ifaces[i])
		if len(v4) > 0 {
			host = v4[0].String()
			break
		}
		if len(v6) > 0 {
			host = v6[0].String()
		}
	}
	return host, nil
}

// ListLocalIP 查找所有本地IP
func ListLocalIP() []string {
	ifaces := ListMulticastInterfaces()
	if len(ifaces) == 0 {
		return nil
	}
	hosts := make([]string, 0, len(ifaces))
	for i := range ifaces {
		v4, v6 := AddrsForInterface(&ifaces[i])
		if len(v4) > 0 {
			hosts = append(hosts, v4[0].String())
			break
		}
		if len(v6) > 0 {
			hosts = append(hosts, v6[0].String())
		}
	}
	return hosts
}

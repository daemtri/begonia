package netx

import (
	"net"
	"reflect"
	"testing"
)

func Test_ListMulticastInterfaces(t *testing.T) {
	ifaces := ListMulticastInterfaces()
	for i := range ifaces {
		v4, v6 := AddrsForInterface(&ifaces[i])
		t.Log("v4", v4)
		t.Log("v6", v6)
	}
	t.Log("iface", ifaces)
	tests := []struct {
		name string
		want []net.Interface
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListMulticastInterfaces(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListMulticastInterfaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_AddrsForInterface(t *testing.T) {
	type args struct {
		iface *net.Interface
	}
	tests := []struct {
		name  string
		args  args
		want  []net.IP
		want1 []net.IP
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := AddrsForInterface(tt.args.iface)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddrsForInterface() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("AddrsForInterface() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

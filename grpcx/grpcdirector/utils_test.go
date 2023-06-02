package grpcdirector

import (
	"testing"

	"google.golang.org/grpc/metadata"
)

func Test_resolveAppName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "/mtx.sample.v1.Sample/TestUnary",
			args: args{
				name: "/mtx.sample.v1.Sample/TestUnary",
			},
			want: "mtx.sample.v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := resolveAppName(tt.args.name, metadata.MD{}); got != tt.want {
				t.Errorf("resolveAppName() = %v, want %v", got, tt.want)
			}
		})
	}
}

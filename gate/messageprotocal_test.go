package gate

import (
	"testing"
)

func TestGetAppId(t *testing.T) {
	type args struct {
		messageId int32
	}
	tests := []struct {
		name string
		args args
		want uint8
	}{
		{
			name: "gate",
			args: args{
				messageId: 0x10001,
			},
			want: 1,
		},
		{
			name: "lobby",
			args: args{
				messageId: 0x20001,
			},
			want: 2,
		},
		{
			name: "global",
			args: args{
				messageId: 0x30001,
			},
			want: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAppId(tt.args.messageId); got != tt.want {
				t.Errorf("GetAppId() = %v, want %v", got, tt.want)
			}
		})
	}
}

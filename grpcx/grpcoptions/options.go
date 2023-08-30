package grpcoptions

import (
	"flag"

	"github.com/daemtri/begonia/di/box/validate"
)

type Options struct {
	// MaxSendSize GRPC服务参数
	MaxSendMsgSize        int
	MaxRecvMsgSize        int
	InitialWindowSize     int
	InitialConnWindowSize int
	MaxConcurrentStreams  uint
}

func NewOptions() Options {
	// create default server run options
	s := Options{
		MaxSendMsgSize:        4 * 1024 * 1024,
		MaxRecvMsgSize:        4 * 1024 * 1024,
		InitialWindowSize:     1 * 1024 * 1024,
		InitialConnWindowSize: 1 * 1024 * 1024,
		MaxConcurrentStreams:  10000,
	}
	return s
}

func (o *Options) ValidateFlags() error {
	// TODO: 添加自定义规则
	return validate.Struct(o)
}

func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.IntVar(&o.MaxSendMsgSize, "max-send-msg-size", o.MaxSendMsgSize, "GRPC MaxSendMsgSize")
	fs.IntVar(&o.MaxRecvMsgSize, "max-recv-msg-size", o.MaxRecvMsgSize, "GRPC MaxRecvMsgSize")
	fs.IntVar(&o.InitialWindowSize, "init-window-size", o.InitialWindowSize, "GRPC InitialWindowSize")
	fs.IntVar(&o.InitialConnWindowSize, "init-conn-window-size", o.InitialConnWindowSize, "GRPC InitialConnWindowSize")
	fs.UintVar(&o.MaxConcurrentStreams, "max-concurrent-streams", o.MaxConcurrentStreams, "GRPC MaxConcurrentStreams")
}

package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Client struct {
	*clientv3.Client
}

type Options struct {
	Endpoints            []string      `flag:"endpoints"`
	AutoSyncInterval     time.Duration `flag:"auto_sync_interval"`
	DialTimeout          time.Duration `flag:"dial_timeout"`
	DialKeepAliveTime    time.Duration `flag:"dial_keep_alive_time"`
	DialKeepAliveTimeout time.Duration `flag:"dial_keep_alive_timeout"`
	MaxCallSendMsgSize   int           `flag:"max_call_send_msg_size"`
	MaxCallRecvMsgSize   int           `flag:"max_call_recv_msg_size"`
	Username             string        `flag:"username"`
	Password             string        `flag:"password"`
	RejectOldCluster     bool          `flag:"reject_old_cluster"`
	PermitWithoutStream  bool          `flag:"permit_without_stream"`
}

func (opt *Options) Build(ctx context.Context) (*Client, error) {
	base, err := clientv3.New(clientv3.Config{
		Endpoints:            opt.Endpoints,
		AutoSyncInterval:     opt.AutoSyncInterval,
		DialTimeout:          opt.DialTimeout,
		DialKeepAliveTime:    opt.DialKeepAliveTime,
		DialKeepAliveTimeout: opt.DialKeepAliveTimeout,
		MaxCallSendMsgSize:   opt.MaxCallSendMsgSize,
		MaxCallRecvMsgSize:   opt.MaxCallRecvMsgSize,
		Username:             opt.Username,
		Password:             opt.Password,
		RejectOldCluster:     opt.RejectOldCluster,
		Context:              ctx,
		PermitWithoutStream:  opt.PermitWithoutStream,
	})
	return &Client{Client: base}, err
}

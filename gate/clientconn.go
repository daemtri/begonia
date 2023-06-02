package gate

import (
	"context"
	"errors"
	"net"
	"time"

	"git.bianfeng.com/stars/wegame/wan/wanx/logx"
)

var (
	ErrorNoUser = errors.New("no user")
)

type ClientConnNotifier interface {
	OnStart()
	OnMessage(ctx context.Context, userId int64, data []byte) error
	OnClose()
}

type ClientConn interface {
	Start() error
	GracefulClose()
	Send(userId int64, data []byte) error
	SendLeaveCluster(userId int64) (err error)
}

func NewClientConn(appId uint8,
	clusterId uint8,
	gateId uint8,
	endpoint string,
	sendChainSize int,
	maxFrameSize uint32,
	handler ClientConnNotifier) ClientConn {
	return &clientConn{
		ready:         false,
		appId:         appId,
		clusterId:     clusterId,
		gateClusterId: gateId,
		endpoint:      endpoint,
		parser:        NewRouterFrameParser(maxFrameSize),
		sendChan:      make(chan []byte, sendChainSize),
		handler:       handler,
	}
}

type clientConn struct {
	conn  net.Conn
	ready bool

	sendChan   chan []byte
	cancelFunc context.CancelFunc

	appId         uint8
	clusterId     uint8
	gateClusterId uint8
	endpoint      string
	parser        RouterFrameParser
	handler       ClientConnNotifier
}

func (client *clientConn) Start() error {
	conn, err := net.DialTimeout("tcp", client.endpoint, time.Second*5)
	if err != nil {
		return err
	}

	client.conn = conn
	childCtx, cancel := context.WithCancel(context.Background())
	client.cancelFunc = cancel

	go func(ctx context.Context, client *clientConn) {
		defer logx.Recover(logger)
		defer logger.Info("connection to gate lost",
			"clusterId", client.gateClusterId)
		defer client.cancelFunc()
		defer client.OnClose()

		for {
			select {
			case <-childCtx.Done():
				return
			default:
				frame, ok, err := client.parser.ReadOneFrame(client.conn)
				if !ok {
					logger.Warn("read failed", "error", err)
					return
				}
				if err = client.OnFrame(ctx, frame); err != nil {
					logger.Warn("handle failed", "error", err)
					return
				}
			}
		}
	}(childCtx, client)
	_, err = conn.Write(
		client.parser.CreateAuthorFrame(client.appId, client.clusterId))
	if err != nil {
		cancel()
		_ = conn.Close()
		client.conn = nil
		client.cancelFunc = nil
		return err
	}

	client.handler.OnStart()
	return nil
}

func (client *clientConn) GracefulClose() {
	if client.cancelFunc != nil {
		client.cancelFunc()
	}
	if client.conn != nil {
		_ = client.conn.Close()
	}
	client.ready = false
}

func (client *clientConn) Send(userId int64, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("failed enqueue")
		}
	}()
	client.sendChan <- client.parser.Wrap(userId, data)
	return nil
}

func (client *clientConn) SendLeaveCluster(userId int64) (err error) {
	client.sendNoUser(context.TODO(), userId)
	return nil
}

func (client *clientConn) OnFrame(ctx context.Context, frame []byte) error {
	if !client.ready {
		appId, instanceId, ok := client.parser.DecodeAuthorFrame(frame)
		if !ok || appId != ServerTypeGate ||
			instanceId != client.gateClusterId {
			return errors.New("wrong auth")
		}
		client.ready = true
		go client.sendWorker(ctx)
		return nil
	}
	userId, data, ok := client.parser.Parse(frame)
	if !ok {
		return errors.New("wrong frame")
	}

	if client.handler != nil {
		err := client.handler.OnMessage(ctx, userId, data)
		if err == ErrorNoUser {
			client.sendNoUser(ctx, userId)
		}
		return err
	}
	return nil
}

func (client *clientConn) OnClose() {
	if client.handler != nil {
		client.handler.OnClose()
	}
}

func (client *clientConn) sendWorker(ctx context.Context) {
	defer logx.Recover(logger)
	defer logger.Info("send worker quit", "app", client.appId, "instance", client.clusterId)
	defer close(client.sendChan)

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-client.sendChan:
			if !ok {
				ctx.Done()
				return
			}
			_, err := client.conn.Write(data)
			if err != nil {
				ctx.Done()
				return
			}
		}
	}
}

func (client *clientConn) sendNoUser(_ context.Context, userId int64) {
	client.sendChan <- client.parser.Wrap(userId, nil)
}

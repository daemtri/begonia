// Package gate 定义Gate与各服务器间通信的接口
package gate

import (
	"github.com/daemtri/begonia/logx"
)

type ServerType = uint8

const (
	ServerTypeGate            ServerType = 1
	SentryServerType          ServerType = 2
	ServerNameGate                       = "gate"
	ServerNameSentry                     = "sentry"
	ClientAuthorMsgId         int32      = 0x10001
	ToClientAuthorResultMsgId int32      = 0x10002
	ClientPingMsgId           int32      = 0x10003
	ToClientPongMsgId         int32      = 0x10004
	ErrorMsgId                int32      = 0x18000

	ToClientErrorCodeInternalError int32 = 1
	ToClientErrorCodeNotInCluster  int32 = 2
	ToClientErrorCodeNoCluster     int32 = 3
	ToClientNoHandler              int32 = 4
	ToClientKick                   int32 = 5
	ToClientInvokeDeadlineExceeded int32 = 6

	SessionDomain = "session"
)

func IsGateMessage(messageId int32) bool {
	return GetAppId(messageId) == ServerTypeGate
}

func GetAppId(messageId int32) ServerType {
	return ServerType((messageId >> 16) & 0xFF)
}

func IsCluster(appId ServerType) bool {
	return appId == ServerTypeGate || (appId&0x80) != 0
}

var logger = logx.GetLogger("/gf/gate")

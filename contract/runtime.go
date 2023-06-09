package contract

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type Task struct {
	typename   string
	payload    []byte
	ScheduleAt time.Time
}

func (t *Task) Type() string    { return t.typename }
func (t *Task) Payload() []byte { return t.payload }

type Scheduler interface {
	AddTask(task *Task) error
}

type PubSubInterface interface {
	Publish()
}

// ConfigInterface 配置接口
type ConfigInterface[T any] interface {
	// Instance 用于获取配置项的值，如果配置项不存在则返回默认值，
	// 如果配置类型是指针，则返回nil。
	Instance() T
	// Watch 用于监听配置项的变化。
	SpanWatch(ctx context.Context, fn func(T) error)
}

type DistrubutedLocker interface {
	Lock(key string, fn func()) error
	TryLock(key string, fn func()) error
}

type Notifier interface {
}

type ClusterInterface interface {
}

type ServiceInterface interface {
	ClientConn() grpc.ClientConnInterface
}

type UserInfoInterface interface {
	GetUserID() uint32
	GetTenantID() uint32
	GetSource() string
	GetGameID() uint32
	GetVersion() uint32
}

package contract

import (
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

type ConfigInterface interface {
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

type AuthInterface interface {
	GetUserID() int32
	GetTenantID() int32
	GetGameID() int32
	GetClientVersion() int32
}

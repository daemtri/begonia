package schedule

import (
	"context"
	"time"
)

type Task struct {
	typename   string
	payload    []byte
	ScheduleAt time.Time
}

func (t *Task) Type() string    { return t.typename }
func (t *Task) Payload() []byte { return t.payload }

type Scheduler interface {
	AddTask(cxt context.Context, typename string, task any, opts ...TaskOption) error
}

type taskOptions struct {
	scheduleAt time.Time
	key        string
}

type TaskOption interface {
	apply(*taskOptions)
}

type funcTaskOption func(*taskOptions)

func (f funcTaskOption) apply(opts *taskOptions) {
	f(opts)
}

func WithScheduleAt(scheduleAt time.Time) TaskOption {
	return funcTaskOption(func(opts *taskOptions) {
		opts.scheduleAt = scheduleAt
	})
}

func WithKey(key string) TaskOption {
	return funcTaskOption(func(opts *taskOptions) {
		opts.key = key
	})
}

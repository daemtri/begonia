package component

import (
	"context"
	"sync"
	"time"
)

type Mutex interface {
	// TryLock locks the mutex if not already locked by another session.
	// If lock is held by another session, return immediately after attempting necessary cleanup
	// The ctx argument is used for the sending/receiving Txn RPC.
	TryLock(ctx context.Context) error

	// Lock locks the mutex with a cancelable context. If the context is canceled
	// while trying to acquire the lock, the mutex tries to clean its stale lock entry.
	Lock(ctx context.Context) error

	// Unlock 解锁
	Unlock(ctx context.Context) error
}

// Session represents a lease kept alive for the lifetime of a client.
// Fault-tolerant applications may use sessions to reason about liveness.
type Session interface {
	NewMutex(pfx string) Mutex
	NewLocker(pfx string) sync.Locker
}

type Concurrency interface {
	Interface

	// NewSession gets the leased session.
	NewSession(ctx context.Context, ttl time.Duration) (Session, error)
}

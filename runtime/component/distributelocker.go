package component

import (
	"context"
)

// Locker 互斥锁接口,使用传递context实现可重入锁的功能
type Locker interface {
	// TryLock 尝试对互斥锁进行加锁，如果该锁已经被其他会话占用，则返回一个error
	// ctx 参数用于发送/接收 Txn RPC。一旦获取锁成功，锁不会因为ctx的取消而自动释放
	// context.Contex 返回值可以用于其他操作，如果其他操作内部尝试获取同一个锁，则会直接返成功。
	TryLock() (context.Context, error)
	// Lock 对互斥锁进行加锁，并使用可取消的上下文。如果在尝试获取锁的过程中上下文被取消，则会不再尝试获取锁。
	// 一旦锁获取成功，锁不会因为ctx的取消而自动释放
	// context.Contex 返回值可以用于其他操作，如果其他操作内部尝试获取同一个锁，则会直接返成功。
	Lock() (context.Context, error)
	// Unlock 对互斥锁进行解锁，并清理与互斥锁相关联的锁条目。
	// ctx 参数用于发送/接收 Txn RPC。
	Unlock()
}

type DistrubutedLocker interface {
	GetLock(ctx context.Context, key string) Locker
}

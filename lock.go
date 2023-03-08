package locker

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	redsync "github.com/go-redsync/redsync/v4"
	goredis "github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/google/wire"
)

// LockerProvider
var LockerProvider = wire.NewSet(NewLocker, wire.Bind(new(ILocker), new(*Locker)))

// ILocker
type ILocker interface {
	Lock(ctx context.Context, resource string, timeout int) (*redsync.Mutex, error)
	TryLock(ctx context.Context, resource string, expire time.Duration, tries int) (*redsync.Mutex, error)
	Unlock(ctx context.Context, mutex *redsync.Mutex) (bool, error)
}

// Locker
type Locker struct {
	client  redis.UniversalClient
	redsync *redsync.Redsync
}

// NewLocker
func NewLocker(resource, token string,
	client redis.UniversalClient, timeout int) *Locker {
	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &Locker{
		client:  client,
		redsync: rs,
	}
}

// Lock
func (lock *Locker) Lock(ctx context.Context, resource string) (*redsync.Mutex, error) {
	if resource == "" {
		return nil, fmt.Errorf("resource undefined")
	}

	mutex := lock.redsync.NewMutex(resource)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	return mutex, nil
}

// TryLock
func (lock *Locker) TryLock(ctx context.Context, resource string, expire time.Duration, tries int) (*redsync.Mutex, error) {
	options := []redsync.Option{
		redsync.WithExpiry(expire),
		redsync.WithTries(tries),
	}
	mutex := lock.redsync.NewMutex(resource, options...)
	err := mutex.LockContext(ctx)
	if err != nil {
		return nil, err
	}
	return mutex, nil
}

// Unlock
func (lock *Locker) Unlock(ctx context.Context, mutex *redsync.Mutex) (bool, error) {
	return mutex.UnlockContext(ctx)
}

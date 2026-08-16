package service

import (
	"context"
	"sync"
)

type firstTokenObserverContextKey struct{}

type firstTokenObserver struct {
	once   sync.Once
	notify func(int)
}

func WithFirstTokenObserver(ctx context.Context, notify func(int)) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if notify == nil {
		return ctx
	}
	return context.WithValue(ctx, firstTokenObserverContextKey{}, &firstTokenObserver{notify: notify})
}

func NotifyFirstToken(ctx context.Context, firstTokenMs int) {
	if ctx == nil || firstTokenMs < 0 {
		return
	}
	observer, _ := ctx.Value(firstTokenObserverContextKey{}).(*firstTokenObserver)
	if observer == nil || observer.notify == nil {
		return
	}
	observer.once.Do(func() {
		observer.notify(firstTokenMs)
	})
}

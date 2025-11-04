package eventbus

import "context"

// 函数型处理器
type EventFunc func(ctx context.Context, evt Event) error

// 接口型处理器
type EventHandle interface {
	Handle(ctx context.Context, evt Event) error
}

type DropPolicy int

const (
	DropBlock  DropPolicy = iota // 缓冲满时阻塞
	DropNewest                   // 丢弃最新（这次要写入的）
	DropOldest                   // 丢弃最旧（队首）
)

type EventOptions struct {
	Async      bool
	Buffer     int
	DropPolicy DropPolicy
	Priority   int // 值越大优先处理（异步时生效）
	Once       bool
}

type EventBus interface {
	SubscribeFunc(pattern string, h EventFunc, opt EventOptions) (subID string)
	Subscribe(pattern string, h EventHandle, opt EventOptions) (subID string)
	Unsubscribe(subID string) bool
	Publish(ctx context.Context, evt Event) error
	Close(ctx context.Context) error
}

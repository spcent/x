package eventbus

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 测试用 DTO
type shippedDTO struct {
	OrderID int `json:"order_id"`
	UserID  int `json:"user_id"`
}

// 辅助：等待直到条件成立或超时
func waitUntil(t *testing.T, d time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout after %s waiting for condition", d)
}

// 辅助：新建一个 JSON 事件
func mustNewEvt(t *testing.T, key string, o, u int) Event {
	t.Helper()
	return NewJSONEvent(key, shippedDTO{OrderID: o, UserID: u})
}

func TestSyncPriorityOrder(t *testing.T) {
	b := New()
	defer b.Close(context.Background())

	var seq []int
	var mu sync.Mutex

	// 优先级高的先执行
	SubscribeFunc := b.SubscribeFunc
	SubscribeFunc("order.shipped", func(ctx context.Context, evt Event) error {
		mu.Lock()
		seq = append(seq, 1)
		mu.Unlock()
		return nil
	}, EventOptions{Async: false, Priority: 1})

	SubscribeFunc("order.shipped", func(ctx context.Context, evt Event) error {
		mu.Lock()
		seq = append(seq, 2)
		mu.Unlock()
		return nil
	}, EventOptions{Async: false, Priority: 10})

	err := b.Publish(context.Background(), mustNewEvt(t, "order.shipped", 1001, 42))
	if err != nil {
		t.Fatalf("publish error: %v", err)
	}

	// 立即可见（同步）
	if len(seq) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(seq))
	}
	if seq[0] != 2 || seq[1] != 1 {
		t.Fatalf("priority order wrong, got %v (expect [2,1])", seq)
	}
}

func TestAsyncFanoutAndDecode(t *testing.T) {
	b := New(WithAsyncWorkers(2))
	defer b.Close(context.Background())

	var c1, c2 int32

	// 通配 & 精确同时收到
	b.SubscribeFunc("order.*", func(ctx context.Context, evt Event) error {
		var dto shippedDTO
		var err error
		if dto, err = DecodeJSON[shippedDTO](evt); err != nil {
			t.Errorf("decode error: %v", err)
			return nil
		}
		if dto.OrderID != 2001 || dto.UserID != 7 {
			t.Errorf("unexpected dto %+v", dto)
		}
		atomic.AddInt32(&c1, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 64, Priority: 5})

	b.SubscribeFunc("order.shipped", func(ctx context.Context, evt Event) error {
		_, err := DecodeJSON[shippedDTO](evt)
		if err != nil {
			t.Errorf("decode error: %v", err)
			return nil
		}
		atomic.AddInt32(&c2, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 64, Priority: 5})

	_ = b.Publish(context.Background(), mustNewEvt(t, "order.shipped", 2001, 7))

	waitUntil(t, 500*time.Millisecond, func() bool {
		return atomic.LoadInt32(&c1) == 1 && atomic.LoadInt32(&c2) == 1
	})
}

func TestOnceOnlyOnce(t *testing.T) {
	b := New(WithAsyncWorkers(1))
	defer b.Close(context.Background())

	var cnt int32
	b.SubscribeFunc("user.created", func(ctx context.Context, evt Event) error {
		atomic.AddInt32(&cnt, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 16, Once: true})

	_ = b.Publish(context.Background(), mustNewEvt(t, "user.created", 1, 1))
	_ = b.Publish(context.Background(), mustNewEvt(t, "user.created", 1, 1))
	_ = b.Publish(context.Background(), mustNewEvt(t, "user.created", 1, 1))

	waitUntil(t, 400*time.Millisecond, func() bool {
		return atomic.LoadInt32(&cnt) == 1
	})
}

func TestDropNewestPolicy(t *testing.T) {
	b := New(WithAsyncWorkers(1))
	defer b.Close(context.Background())

	// 异步消费者刻意变慢
	var processed int32
	b.SubscribeFunc("slow.*", func(ctx context.Context, evt Event) error {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&processed, 1)
		return nil
	}, EventOptions{
		Async:      true,
		Buffer:     1,          // 很小的缓冲
		DropPolicy: DropNewest, // 溢出丢最新
		Priority:   5,
	})

	// 快速发布 10 条
	for i := 0; i < 10; i++ {
		_ = b.Publish(context.Background(), mustNewEvt(t, "slow.event", i, i))
	}

	// 等最多 1 秒
	time.Sleep(1 * time.Second)
	got := atomic.LoadInt32(&processed)

	// 因为消费者处理慢 & Buffer 很小 & DropNewest，大部分会被丢弃
	if got >= 10 {
		t.Fatalf("DropNewest expected to drop some, processed=%d", got)
	}
	if got == 0 {
		t.Fatalf("expected at least one processed")
	}
}

func TestDropOldestKeepsRecent(t *testing.T) {
	b := New(WithAsyncWorkers(1))
	defer b.Close(context.Background())

	var lastOrder atomic.Int64
	// 处理慢，模拟拥塞
	b.SubscribeFunc("recent.*", func(ctx context.Context, evt Event) error {
		time.Sleep(30 * time.Millisecond)
		dto, err := DecodeJSON[shippedDTO](evt)
		if err == nil {
			lastOrder.Store(int64(dto.OrderID))
		}
		return nil
	}, EventOptions{
		Async:      true,
		Buffer:     1,
		DropPolicy: DropOldest, // 溢出丢最旧
		Priority:   5,
	})

	// 迅速推入 8 条，最后一条的 OrderID=999
	for i := 0; i < 7; i++ {
		_ = b.Publish(context.Background(), mustNewEvt(t, "recent.event", i, i))
	}
	_ = b.Publish(context.Background(), mustNewEvt(t, "recent.event", 999, 0))

	time.Sleep(600 * time.Millisecond) // 给足时间

	lo := lastOrder.Load()
	if lo != 999 {
		t.Fatalf("DropOldest should keep most recent; got last=%d want=999 (non-deterministic but expected)", lo)
	}
}

func TestWildcardAndExact(t *testing.T) {
	b := New(WithAsyncWorkers(1))
	defer b.Close(context.Background())

	var w, e int32

	b.SubscribeFunc("order.*", func(ctx context.Context, evt Event) error {
		atomic.AddInt32(&w, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 16})

	b.SubscribeFunc("order.shipped", func(ctx context.Context, evt Event) error {
		atomic.AddInt32(&e, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 16})

	_ = b.Publish(context.Background(), mustNewEvt(t, "order.shipped", 1, 1))

	waitUntil(t, 400*time.Millisecond, func() bool {
		return atomic.LoadInt32(&w) == 1 && atomic.LoadInt32(&e) == 1
	})
}

func TestUnsubscribe(t *testing.T) {
	b := New(WithAsyncWorkers(1))
	defer b.Close(context.Background())

	var cnt int32

	id := b.SubscribeFunc("ping", func(ctx context.Context, evt Event) error {
		atomic.AddInt32(&cnt, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 8})

	_ = b.Publish(context.Background(), mustNewEvt(t, "ping", 0, 0))
	waitUntil(t, 300*time.Millisecond, func() bool { return atomic.LoadInt32(&cnt) == 1 })

	ok := b.Unsubscribe(id)
	if !ok {
		t.Fatalf("unsubscribe returned false")
	}

	_ = b.Publish(context.Background(), mustNewEvt(t, "ping", 0, 0))
	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&cnt) != 1 {
		t.Fatalf("handler invoked after unsubscribe")
	}
}

func TestCloseGracefully(t *testing.T) {
	b := New(WithAsyncWorkers(2))

	var cnt int32
	b.SubscribeFunc("close.*", func(ctx context.Context, evt Event) error {
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&cnt, 1)
		return nil
	}, EventOptions{Async: true, Buffer: 64})

	// 发几条
	for i := 0; i < 5; i++ {
		_ = b.Publish(context.Background(), mustNewEvt(t, "close.event", i, i))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := b.Close(ctx); err != nil {
		t.Fatalf("close error: %v", err)
	}

	// 已关闭后再次 Publish 不应 panic；（当前实现可能无效果，这里只是验证不会崩）
	_ = b.Publish(context.Background(), mustNewEvt(t, "close.event", 99, 0))
}

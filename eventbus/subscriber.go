package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PanicHandler func(subID, pattern string, evt Event, recovered any, stack []byte)

type subscriber struct {
	id      string
	pattern string
	fn      EventFunc
	opt     EventOptions
	ch      chan job
	closed  chan struct{}
}

type bus struct {
	mu      sync.RWMutex
	subs    map[string]*subscriber
	pattern map[string]map[string]bool

	aw      *asyncWorker
	workers int
	wg      sync.WaitGroup

	ctx    context.Context
	cancel context.CancelFunc

	panicHandler PanicHandler
}

type BusOption func(*bus)

func WithAsyncWorkers(n int) BusOption {
	return func(b *bus) { b.workers = n }
}

func WithPanicHandler(h PanicHandler) BusOption {
	return func(b *bus) { b.panicHandler = h }
}

func New(opts ...BusOption) EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	b := &bus{
		subs:    make(map[string]*subscriber),
		pattern: make(map[string]map[string]bool),
		aw:      newAsyncWorker(),
		workers: 4,
		ctx:     ctx,
		cancel:  cancel,
	}

	for _, o := range opts {
		o(b)
	}

	b.startWorkers()
	return b
}

func (b *bus) startWorkers() {
	for i := 0; i < b.workers; i++ {
		b.wg.Add(1)

		go func() {
			defer b.wg.Done()
			for {
				j, ok := b.aw.pop()
				if !ok {
					return
				}
				b.dispatchAsyncJob(j)
			}
		}()
	}
}

func (b *bus) safeInvoke(s *subscriber, ctx context.Context, evt Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			st := debug.Stack()
			if b.panicHandler != nil {
				b.panicHandler(s.id, s.pattern, evt, r, st)
			} else {
				log.Printf("[eventbus] handler panic sub=%s pattern=%s key=%s id=%s err=%v\n%s",
					s.id, s.pattern, evt.Key(), evt.ID(), r, string(st))
			}
			err = fmt.Errorf("handler panic: %v", r)
		}
	}()

	return s.fn(ctx, evt)
}

func (b *bus) SubscribeFunc(pattern string, h EventFunc, opt EventOptions) (subID string) {
	return b.subscribe(pattern, h, opt)
}

func (b *bus) Subscribe(pattern string, h EventHandle, opt EventOptions) (subID string) {
	fn := func(ctx context.Context, evt Event) error {
		return h.Handle(ctx, evt)
	}
	return b.subscribe(pattern, fn, opt)
}

func (b *bus) subscribe(pattern string, fn EventFunc, opt EventOptions) (subID string) {
	if opt.Buffer <= 0 {
		opt.Buffer = 256
	}

	sub := &subscriber{
		id:      uuid.NewString(),
		pattern: pattern,
		fn:      fn,
		opt:     opt,
		closed:  make(chan struct{}),
	}
	if opt.Async {
		sub.ch = make(chan job, opt.Buffer)
		b.wg.Add(1)
		go func(s *subscriber) {
			defer b.wg.Done()
			for {
				select {
				case <-b.ctx.Done():
					return
				case <-s.closed:
					return
				case j := <-s.ch:
					err := b.safeInvoke(s, context.Background(), j.evt)
					if err != nil {
						log.Printf("[eventbus] handler error sub=%s pattern=%s key=%s id=%s err=%v",
							s.id, s.pattern, j.evt.Key(), j.evt.ID(), err)
					}
					if s.opt.Once {
						b.Unsubscribe(s.id)
						return
					}
				}
			}
		}(sub)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.subs[sub.id] = sub
	if _, ok := b.pattern[pattern]; !ok {
		b.pattern[pattern] = map[string]bool{}
	}
	b.pattern[pattern][sub.id] = true
	return sub.id
}

func (b *bus) Unsubscribe(subID string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub, ok := b.subs[subID]
	if !ok {
		return false
	}

	delete(b.subs, subID)
	if set, ok := b.pattern[sub.pattern]; ok {
		delete(set, subID)
		if len(set) == 0 {
			delete(b.pattern, sub.pattern)
		}
	}

	close(sub.closed)
	return true
}

func (b *bus) Publish(ctx context.Context, evt Event) error {
	if evt == nil {
		return nil
	}

	subs := b.snapshot(evt.Key())
	if len(subs) == 0 {
		return nil
	}

	hasAsync := false
	// 同步（按优先级）
	for _, s := range subs {
		if !s.opt.Async {
			if err := b.safeInvoke(s, ctx, evt); err != nil {
				// 如需 StopOnError，可改为 return err
				log.Printf("[eventbus] handler error sub=%s pattern=%s key=%s id=%s err=%v",
					s.id, s.pattern, evt.Key(), evt.ID(), err)
				return err
			}
			if s.opt.Once {
				b.Unsubscribe(s.id)
			}
		} else {
			hasAsync = true
		}
	}

	if hasAsync {
		// 异步（全局优先队列 + 各订阅者缓冲回压）
		now := time.Now()
		// 压一次，由 worker fan-out
		b.aw.push(job{evt: evt, priority: maxPriority(subs), ts: now})
	}

	return nil
}

func (b *bus) Close(ctx context.Context) error {
	b.cancel()
	b.aw.close()

	done := make(chan struct{})
	go func() { b.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return errors.New("event bus close timeout")
	}
}

func (b *bus) snapshot(key string) []*subscriber {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var res []*subscriber

	// 精确匹配
	for sid := range b.pattern[key] {
		res = append(res, b.subs[sid])
	}

	// 前缀通配："order.*" 命中 "order.created"
	for pat, set := range b.pattern {
		if strings.HasSuffix(pat, "*") {
			prefix := strings.TrimSuffix(pat, "*")
			if strings.HasPrefix(key, prefix) {
				for sid := range set {
					res = append(res, b.subs[sid])
				}
			}
		}
	}

	// 同步执行顺序：优先级高在前
	if len(res) > 1 {
		for i := 1; i < len(res); i++ {
			j := i
			for j > 0 && res[j-1].opt.Priority < res[j].opt.Priority {
				res[j], res[j-1] = res[j-1], res[j]
				j--
			}
		}
	}

	return res
}

func (b *bus) dispatchAsyncJob(j job) {
	subs := b.snapshot(j.evt.Key())
	for _, s := range subs {
		if !s.opt.Async {
			continue
		}

		select {
		case s.ch <- j:
		default:
			switch s.opt.DropPolicy {
			case DropBlock:
				s.ch <- j
			case DropNewest:
				// 丢这条
			case DropOldest:
				select {
				case <-s.ch:
				default:
				}
				select {
				case s.ch <- j:
				default:
				}
			}
		}
	}
}

func maxPriority(subs []*subscriber) int {
	maxp := 0
	for _, s := range subs {
		if s.opt.Priority > maxp {
			maxp = s.opt.Priority
		}
	}

	return maxp
}

package runner

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// TestGroupRun_NormalExit 测试所有actor在超时前正常退出
func TestGroupRun_NormalExit(t *testing.T) {
	g := &Group{ShutdownTimeout: 2 * time.Second}

	// 第一个actor立即返回错误
	firstErr := fmt.Errorf("first actor error")
	g.Add(func() error {
		return firstErr
	}, func(err error) {})

	// 第二个actor在interrupt时立即退出
	secondDone := make(chan struct{})
	g.Add(func() error {
		<-secondDone // 等待interrupt触发
		return nil
	}, func(err error) {
		close(secondDone) // 触发退出
	})

	// 执行Run并验证结果
	err := g.Run()
	if err != firstErr {
		t.Errorf("expected error %v, got %v", firstErr, err)
	}
}

// TestGroupRun_TimeoutExit 测试存在actor未及时退出时触发超时
func TestGroupRun_TimeoutExit(t *testing.T) {
	g := &Group{ShutdownTimeout: 1 * time.Second} // 缩短超时时间便于测试

	// 第一个actor立即返回错误
	firstErr := fmt.Errorf("first actor error")
	g.Add(func() error {
		return firstErr
	}, func(err error) {})

	// 第二个actor故意延迟退出（超过超时时间）
	secondDone := make(chan struct{})
	g.Add(func() error {
		<-secondDone // 永远不关闭，模拟阻塞
		return nil
	}, func(err error) {
		// 不触发退出，模拟未及时响应中断
	})

	// 执行Run并验证超时
	err := g.Run()
	if err != firstErr {
		t.Errorf("expected error %v, got %v", firstErr, err)
	}
}

func TestZero(t *testing.T) {
	var g Group
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if err != nil {
			t.Errorf("%v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestOne(t *testing.T) {
	myError := errors.New("foobar")
	var g Group
	g.Add(func() error { return myError }, func(error) {})
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := myError, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout")
	}
}

func TestMany(t *testing.T) {
	interrupt := errors.New("interrupt")
	var g Group
	g.Add(func() error { return interrupt }, func(error) {})
	cancel := make(chan struct{})
	g.Add(func() error { <-cancel; return nil }, func(error) { close(cancel) })
	res := make(chan error)
	go func() { res <- g.Run() }()
	select {
	case err := <-res:
		if want, have := interrupt, err; want != have {
			t.Errorf("want %v, have %v", want, have)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("timeout")
	}
}

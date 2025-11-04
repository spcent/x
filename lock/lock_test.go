package lock

import (
	"context"
	"fmt"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// _redisClient redis 客户端
var _redisClient *redis.Client

// InitTestRedis 实例化一个可以用于单元测试的redis
func InitTestRedis() {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	// 打开下面命令可以测试链接关闭的情况
	// defer mr.Close()

	_redisClient = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	fmt.Println("mini redis addr:", mr.Addr())
}

func TestLockWithDefaultTimeout(t *testing.T) {
	InitTestRedis()
	expiration := 2 * time.Second

	lock := NewRedisLock(_redisClient, "lock1", expiration)
	ok, err := lock.Lock(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Fatal("lock is not ok")
	}

	ok, err = lock.Unlock(context.Background())
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Unlock is not ok")
	}

	t.Log(ok)
}

func TestLockWithTimeout(t *testing.T) {
	InitTestRedis()
	expiration := 2 * time.Second

	t.Run("should lock/unlock success", func(t *testing.T) {
		ctx := context.Background()
		lock1 := NewRedisLock(_redisClient, "lock2", expiration)
		ok, err := lock1.Lock(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)

		ok, err = lock1.Unlock(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)
	})

	t.Run("should unlock failed", func(t *testing.T) {
		ctx := context.Background()
		lock2 := NewRedisLock(_redisClient, "lock3", expiration)
		ok, err := lock2.Lock(ctx)
		assert.Nil(t, err)
		assert.True(t, ok)

		time.Sleep(3 * time.Second)

		ok, err = lock2.Unlock(ctx)
		fmt.Println("===*****************", ok, err)
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

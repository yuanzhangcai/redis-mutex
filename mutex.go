package mutex

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

var (
	gOpts  *InitOptions
	client *redis.Client

	// ErrLockFailed 上锁失败
	ErrLockFailed = errors.New("lock failed")
	// ErrUnlockInvalid 无效解锁
	ErrUnlockInvalid = errors.New("unlock invalid")
	// ErrRefreshTTLFailed 刷新锁的存活时间失败
	ErrRefreshTTLFailed = errors.New("refresh ttl failed")

	luaRefresh = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)

	luaDel = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)

	luaPTTL = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pttl", KEYS[1])
		else
			return -1
		end`)
)

// Init 初始化分布式锁
func Init(opt ...InitOption) error {
	gOpts = newInitOptions(opt...)

	if client != nil {
		return nil
	}

	client = redis.NewClient(&redis.Options{
		Addr:     gOpts.RedisServer,
		Password: gOpts.Password,
	})

	_, err := client.Ping().Result()
	if err != nil {
		client = nil
		return err
	}

	return nil
}

// Locker 分布式锁接口
type Locker interface {
	Lock() error
	RefreshTTL() error
	Unlock() error
	TTL() (time.Duration, error)
}

// redisMutex redis分布式锁
type redisMutex struct {
	key   string
	value string
	opts  *Options
	stop  chan bool
	m     sync.Mutex
}

// NewMutex 创建分布式锁
func NewMutex(key string, opt ...Option) Locker {
	opts := newOptions(opt...)
	return &redisMutex{
		key:   gOpts.Prefix + key,
		value: strconv.FormatInt(time.Now().UnixNano(), 10),
		opts:  opts,
	}
}

// Lock 上锁
func (c *redisMutex) Lock() error {
	var timer *time.Timer
	for deadline := time.Now().Add(c.opts.Timeout); time.Now().Before(deadline); {
		ok, err := client.SetNX(c.key, c.value, c.opts.TTL).Result()
		if err != nil {
			return err
		} else if ok {
			if c.opts.AutoRefreshTTL {
				c.autoRefresh()
			}

			return nil
		}

		if c.opts.Retry == nil {
			return ErrLockFailed
		}

		after := c.opts.Retry.After()
		if after == 0 {
			return ErrLockFailed
		}

		if timer == nil {
			timer = time.NewTimer(after)
			defer timer.Stop()
		} else {
			timer.Reset(after)
		}

		select {
		case <-c.opts.Ctx.Done():
			return c.opts.Ctx.Err()
		case <-timer.C:
		}
	}

	return ErrLockFailed
}

// RefreshTTL 刷新锁的存活时间
func (c *redisMutex) RefreshTTL() error {
	ttl := strconv.FormatInt(int64(c.opts.TTL/time.Millisecond), 10)
	status, err := luaRefresh.Run(client, []string{c.key}, c.value, ttl).Result()
	if err != nil {
		return err
	} else if status == int64(1) {
		return nil
	}
	return ErrRefreshTTLFailed
}

// Unlock 解锁
func (c *redisMutex) Unlock() error {
	if c.opts.AutoRefreshTTL {
		c.m.Lock()
		close(c.stop)
		c.stop = nil
		c.m.Unlock()
	}
	res, err := luaDel.Run(client, []string{c.key}, c.value).Result()
	if err == redis.Nil {
		return ErrUnlockInvalid
	} else if err != nil {
		return err
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return ErrUnlockInvalid
	}

	return nil
}

// TTL 返回锁的存活时间
func (c *redisMutex) TTL() (time.Duration, error) {
	res, err := luaPTTL.Run(client, []string{c.key}, c.value).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	if num := res.(int64); num > 0 {
		return time.Duration(num) * time.Millisecond, nil
	}
	return 0, nil
}

//
func (c *redisMutex) autoRefresh() {
	c.m.Lock()
	c.stop = make(chan bool)
	stop := c.stop
	c.m.Unlock()
	go func() {
		var err error
		timer := time.NewTicker(c.opts.TTL / 2)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				err = c.RefreshTTL()
				if err != nil {
					return
				}
			case <-stop:
				return
			}
		}
	}()
}

// RetryStrategy 重试策略
type RetryStrategy interface {
	// After 返回重试时间，当前时之后多长时间重试
	After() time.Duration
}

// defaultRetryStrategy 默认重试策略
type defaultRetryStrategy struct {
	cnt uint
}

func (c *defaultRetryStrategy) After() time.Duration {
	c.cnt++

	after := c.cnt * 10
	if after > 200 {
		after = 200
	}
	return time.Duration(after) * time.Millisecond
}

package mutex

import (
	"context"
	"time"
)

var (
	defaultPrefix         = "mutex_"                // 默认redis前缀
	defaultRedisServer    = "127.0.0.1:6379"        // 默认redis地址
	defaultTimeout        = 3000 * time.Millisecond // 默认超时时间
	defaultTTL            = 5000 * time.Millisecond // 默认存活时间
	defaultAutoRefreshTTL = false                   // 默认自动刷新锁的存活时间
)

// InitOptions 分布式锁参数
type InitOptions struct {
	Prefix      string // redis键值前缀
	RedisServer string // redis服务器信息
	Password    string // redis密码
}

// InitOption 设置初始化分布锁参数参数
type InitOption func(*InitOptions)

// newInitOptions 创建初始化分布式锁参数对象
func newInitOptions(opts ...InitOption) *InitOptions {
	opt := &InitOptions{
		Prefix:      defaultPrefix,
		RedisServer: defaultRedisServer,
	}

	for _, one := range opts {
		one(opt)
	}
	return opt
}

// RedisServer 设置redis地址
func RedisServer(server string) InitOption {
	return func(opt *InitOptions) {
		opt.RedisServer = server
	}
}

// Password 设置redis密码
func Password(password string) InitOption {
	return func(opt *InitOptions) {
		opt.Password = password
	}
}

// Prefix 设置redis前缀
func Prefix(prefix string) InitOption {
	if prefix != "" && prefix[len(prefix)-1] != '_' {
		prefix += "_"
	}

	return func(opt *InitOptions) {
		opt.Prefix = prefix
	}
}

// Options 分布式锁参数
type Options struct {
	TTL            time.Duration // 锁的存活时间
	Timeout        time.Duration // 上锁超时时间
	AutoRefreshTTL bool          // 是否自动刷新锁的存活时间
	Retry          RetryStrategy // 重试策略
	Ctx            context.Context
}

// Option 设置分布锁参数
type Option func(*Options)

// newOptions 创建分布式锁参数对象
func newOptions(opts ...Option) *Options {
	opt := &Options{
		Timeout:        defaultTimeout,
		TTL:            defaultTTL,
		AutoRefreshTTL: defaultAutoRefreshTTL,
		Retry:          &defaultRetryStrategy{},
		Ctx:            context.Background(),
	}

	for _, one := range opts {
		one(opt)
	}
	return opt
}

// TTL 设置分布式的存活时间
func TTL(ttl time.Duration) Option {
	return func(opt *Options) {
		opt.TTL = ttl
	}
}

// Timeout 设置分布式的存活时间
func Timeout(timeout time.Duration) Option {
	return func(opt *Options) {
		opt.Timeout = timeout
	}
}

// AutoRefresh 设置分布式锁过期时间是否自动刷新
func AutoRefresh(autoRefresh bool) Option {
	return func(opt *Options) {
		opt.AutoRefreshTTL = autoRefresh
	}
}

// Context 设置上下文
func Context(ctx context.Context) Option {
	return func(opt *Options) {
		opt.Ctx = ctx
	}
}

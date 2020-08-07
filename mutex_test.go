package mutex

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Run("Init failed", func(t *testing.T) {
		err := Init(RedisServer("127.0.0.1:63790"), Password("12345678"), Prefix("lock_demo"))
		assert.NotNil(t, err)
		assert.Nil(t, client)
	})

	t.Run("Init success", func(t *testing.T) {
		err := Init(RedisServer("127.0.0.1:6379"), Password("12345678"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})

	t.Run("Init success again", func(t *testing.T) {
		err := Init(RedisServer("127.0.0.1:6379"), Password("12345678"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
	})
}

func TestMutex(t *testing.T) {
	_ = Init(RedisServer("127.0.0.1:6379"), Password("12345678"))

	var m Locker
	t.Run("create locker", func(t *testing.T) {
		m = NewMutex("demo_ock", TTL(10*time.Second), Timeout(5*time.Second), AutoRefresh(false), Context(context.Background()))
		assert.NotNil(t, m)
	})

	t.Run("Lock success", func(t *testing.T) {
		err := m.Lock()
		assert.Nil(t, err)
	})

	t.Run("TTL success", func(t *testing.T) {
		ttl, err := m.TTL()
		assert.Nil(t, err)
		assert.Less(t, int64(0), int64(ttl))
		assert.Less(t, int64(ttl), int64(10*time.Second))
	})

	t.Run("RefreshTTL success", func(t *testing.T) {
		time.Sleep(5 * time.Second)
		err := m.RefreshTTL()
		assert.Nil(t, err)

		ttl, err := m.TTL()
		assert.Nil(t, err)
		assert.Less(t, int64(7*time.Second), int64(ttl))
		assert.Less(t, int64(ttl), int64(10*time.Second))
	})

	t.Run("Lock again", func(t *testing.T) {
		err := m.Lock()
		assert.NotNil(t, err)
	})

	t.Run("Unlock", func(t *testing.T) {
		err := m.Unlock()
		assert.Nil(t, err)
	})

	t.Run("Unlock again", func(t *testing.T) {
		err := m.Unlock()
		assert.NotNil(t, err)
	})
}

func TestAutoRefreshTTL(t *testing.T) {
	_ = Init(RedisServer("127.0.0.1:6379"), Password("12345678"))

	m := NewMutex("demo_lock_2", TTL(500*time.Millisecond))
	_ = m.Lock()
	time.Sleep(2 * time.Second)
	ret := client.Get("mutex_demo_lock_2")
	assert.NotNil(t, ret)
	assert.NotEmpty(t, ret.Val())
	_ = m.Unlock()
}

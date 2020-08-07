package main

import (
	"fmt"
	"time"

	mutex "github.com/yuanzhangcai/redis-mutex"
)

func main() {

	err := mutex.Init(mutex.RedisServer("127.0.0.1:6379"), mutex.Password("12345678"), mutex.Prefix("lock_demo"))
	if err != nil {
		fmt.Println(1, err)
		return
	}

	m := mutex.NewMutex("Lock_key", mutex.TTL(300*time.Second))
	err = m.Lock()
	if err != nil {
		fmt.Println(2, err)
		return
	}

	fmt.Println("do something now.")
	time.Sleep(25 * time.Second)
	m.Unlock()
	fmt.Println("unlock")
	time.Sleep(25 * time.Second)
}

package main

import (
	"fmt"
	"time"

	"github.com/y-matsuwitter/mcond"
	redis "gopkg.in/redis.v2"
)

func main() {
	client := redis.NewTCPClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	mc := mcond.NewMCond(client)
	key := "test"
	mc.AddCond(key)
	mc.AddHost("localhost:8080")
	mc.AddProcessing(key)
	go func() {
		fmt.Println("doing heavy task.")
		time.Sleep(5 * time.Second)
		mc.AddCompleted(key)
		mc.Broadcast(key)
	}()
	fmt.Println(mc)
	mc.WaitForAvailable(key)
	fmt.Println("completed!!")
	time.Sleep(time.Second)
}

package main

import (
	"fmt"
	"time"

	"github.com/y-matsuwitter/mcond"
)

func main() {
	mc := mcond.NewMCond(mcond.MCondOption{})
	key := "test"
	mc.AddCond(key)
	mc.AddHost("localhost:9012")
	mc.Clear()

	mc.AddProcessing(key)
	go func() {
		fmt.Println("doing heavy task.")
		time.Sleep(5 * time.Second)
		mc.AddCompleted(key)
		mc.Broadcast(key)
	}()

	mc.WaitForAvailable(key)
	fmt.Println("completed!!")
	time.Sleep(time.Second)
}

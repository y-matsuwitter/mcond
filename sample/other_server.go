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
	mc.Start()
	go func() {
		mc.WaitForAvailable(key)
		fmt.Println("this key available!! :" + key)
	}()
	time.Sleep(time.Minute)
}

package main

import (
	"fmt"
	"log"
	"net/http"

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

	http.HandleFunc("/cache/broadcast/"+key, func(w http.ResponseWriter, r *http.Request) {
		mc.Broadcast(key)
	})
	go func() {
		mc.WaitForAvailable(key)
		fmt.Println("this key available!! :" + key)
	}()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

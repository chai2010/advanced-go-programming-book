package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

func doClientWork(client *rpc.Client) {
	//watch success
	go func() {
		var keyChanged string
		err := client.Call("KVStoreService.Watch", 6, &keyChanged)
		//err := client.Call("KVStoreService.Watch", 1, &keyChanged)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("watch:", keyChanged)
	} ()

	err := client.Call(
		"KVStoreService.Set", [2]string{"abc", "abc-value"},
		new(struct{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second*3)

	err = client.Call(
		"KVStoreService.Set", [2]string{"abc", "abc-value-change"},
		new(struct{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second*2)

	//timeout
	go func() {
		var keyChanged string
		err := client.Call("KVStoreService.Watch", 1, &keyChanged)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("watch:", keyChanged)
	} ()

	err = client.Call(
		"KVStoreService.Set", [2]string{"abc", "abc-value"},
		new(struct{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second*3)

	err = client.Call(
		"KVStoreService.Set", [2]string{"abc", "abc-value-change"},
		new(struct{}),
	)
	if err != nil {
		log.Fatal(err)
	}

}


func main(){
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	doClientWork(client)
}
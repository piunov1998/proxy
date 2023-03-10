package main

import (
	"log"
	"ws_rest_wrapper/internal/client"
)

func main() {
	config := client.Config{
		WsHost:   "localhost:7070",
		WsPath:   "/ws",
		SelfHost: "pa.orbismap.com",
	}
	proxyClient := client.New(&config)
	err := proxyClient.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

package main

import (
	"log"
	"ws_rest_wrapper/internal/server"
)

func main() {
	config := server.Config{
		WsPath: "/ws",
		Host:   "localhost:7070",
	}
	proxyServer := server.New(&config)
	err := proxyServer.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

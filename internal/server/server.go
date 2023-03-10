package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"ws_rest_wrapper/internal/handlers"
)

type Config struct {
	WsPath string // /ws
	Host   string // localhost:7070
}

type ProxyServer struct {
	config *Config
}

func New(config *Config) ProxyServer {
	return ProxyServer{config: config}
}

func (s *ProxyServer) Run() error {

	mux := http.NewServeMux()

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	ws := handlers.NewWsHandler(ctx)
	api := handlers.NewHttpHandler(&ws)

	mux.Handle("/", &api)
	mux.HandleFunc("/ws", ws.Connect)

	return http.ListenAndServe(s.config.Host, mux)
}

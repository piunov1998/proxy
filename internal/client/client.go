package client

import (
	"context"
	"os"
	"os/signal"
	"ws_rest_wrapper/internal/worker"
)

type Config struct {
	WsHost   string
	WsPath   string
	SelfHost string
}

type ProxyClient struct {
	config *Config
}

func New(config *Config) ProxyClient {
	return ProxyClient{config: config}
}

func (a *ProxyClient) Run() error {
	w := worker.Worker{
		WsHost:   a.config.WsHost,   // 149.154.67.79:5678
		WsPath:   a.config.WsPath,   // /
		SelfHost: a.config.SelfHost, // localhost:7777
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	return w.Run(ctx)
}

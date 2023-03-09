package app

import (
	"net/http"
	"ws_rest_wrapper/internal/handlers"
)

type App struct {
}

func (a *App) Run() error {
	mux := http.NewServeMux()
	mux.Handle("/", handlers.Handler{})
	return http.ListenAndServe("0.0.0.0:7777", mux)
}

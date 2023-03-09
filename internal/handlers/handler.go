package handlers

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
	request2 "ws_rest_wrapper/internal/request"
)

type Handler struct {
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	wsUrl := url.URL{Scheme: "ws", Host: "149.154.67.79:5678", Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl.String(), nil)

	defer conn.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("error during message consumption -> %s", err)
				return
			}
			log.Printf("got message -> %s\n", message)
		}
	}()
	body := map[string]any{
		"a": 1,
		"b": 2,
	}
	bytes, _ := json.Marshal(body)
	wsRequest := request2.ProxyRequest{
		Method: "GET",
		Path:   "/",
		Body:   bytes,
	}
	err = conn.WriteJSON(wsRequest)
	if err != nil {
		log.Printf("error during message production -> %s", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 10):
	}
	writer.Write([]byte("Closed"))
}

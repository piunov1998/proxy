package worker

import (
	"bytes"
	"context"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
	"ws_rest_wrapper/internal/request"
	"ws_rest_wrapper/internal/response"
)

type Worker struct {
	WsHost string
	WsPath string
}

func (w *Worker) Run(ctx context.Context) error {
	inbox := make(chan request.ProxyRequest)
	outbox := make(chan response.ProxyResponse)

	u := url.URL{Scheme: "ws", Host: w.WsHost, Path: w.WsPath}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer close(outbox)

	go w.messageConsumer(conn, inbox)

	running := true
	for running {
		select {
		case <-ctx.Done():
			running = false
		case req := <-inbox:
			go w.messageHandler(req, outbox)
		case resp := <-outbox:
			go w.produceMessage(conn, resp)
		}
	}

	return nil
}

func (w *Worker) messageConsumer(conn *websocket.Conn, inbox chan request.ProxyRequest) {
	defer close(inbox)

	var req request.ProxyRequest
	for {
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Printf("error during message consumption -> %s\n", err)
			continue
		}

		inbox <- req
	}
}

func (w *Worker) produceMessage(conn *websocket.Conn, resp response.ProxyResponse) {
	err := conn.WriteJSON(resp)
	if err != nil {
		log.Printf("error during producing message -> %s\n", err)
	}
}

func (w *Worker) messageHandler(request request.ProxyRequest, outbox chan response.ProxyResponse) {
	headers := http.Header{}
	for key, value := range request.Headers {
		headers.Set(key, value)
	}
	body := bytes.NewBuffer(request.Body)
	u := url.URL{Scheme: "http", Host: "localhost", Path: request.Path}
	prepared, err := http.NewRequest(request.Method, u.String(), body)
	if err != nil {
		log.Printf("error during preparing request -> %s\n", err)
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(prepared)
	if err != nil {
		log.Printf("error during making request -> %s\n", err)
	}

	respHeaders := make(map[string][]string)
	for key, value := range resp.Header {
		respHeaders[key] = value
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	wsResponse := response.ProxyResponse{
		Id:      request.Id,
		Headers: headers,
		Body:    respBody,
		Status:  resp.StatusCode,
	}
	outbox <- wsResponse
}

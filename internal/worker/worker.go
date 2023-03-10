package worker

import (
	"bufio"
	"bytes"
	"context"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
	"ws_rest_wrapper/package/request"
	"ws_rest_wrapper/package/response"
)

type Worker struct {
	WsHost   string
	WsPath   string
	SelfHost string
}

func (w *Worker) Run(ctx context.Context) error {
	inbox := make(chan request.ProxyRequest)
	outbox := make(chan response.ProxyResponse)

	u := url.URL{Scheme: "ws", Host: w.WsHost, Path: w.WsPath}
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
	}
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()
	defer close(outbox)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go w.consumeMessages(ctx, conn, inbox, &wg)

	running := true
	for running {
		select {
		case <-ctx.Done():
			running = false
		case req := <-inbox:
			go w.handleMessage(req, outbox)
		case resp := <-outbox:
			go w.produceMessage(conn, resp)
		}
	}

	wg.Wait()

	return nil
}

func (w *Worker) consumeMessages(ctx context.Context, conn *websocket.Conn, inbox chan request.ProxyRequest, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(inbox)

	var req request.ProxyRequest
	for {
		select {
		default:
		case <-ctx.Done():
			return
		}
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

func (w *Worker) handleMessage(request request.ProxyRequest, outbox chan response.ProxyResponse) {

	reader := bufio.NewReader(bytes.NewBuffer(request.RawHttp))
	prepared, err := http.ReadRequest(reader)
	if err != nil {
		log.Printf("error during preparing request -> %s\n", err)
		return
	}
	path := prepared.URL.Path
	prepared.RequestURI = ""
	u := url.URL{
		Scheme: "http",
		Host:   w.SelfHost,
		Path:   path,
	}
	prepared.URL = &u

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(prepared)
	if err != nil {
		log.Printf("error during making request -> %s\n", err)
		return
	}

	rawHttp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("error during parsing response -> %s\n", err)
		return
	}

	wsResponse := response.ProxyResponse{
		Id:      request.Id,
		Status:  resp.StatusCode,
		RawHttp: rawHttp,
	}
	outbox <- wsResponse
}

package handlers

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"ws_rest_wrapper/package/request"
	"ws_rest_wrapper/package/response"
)

type WsHandler struct {
	conn     *websocket.Conn
	channels map[int]chan response.ProxyResponse
	ctx      context.Context
}

func NewWsHandler(ctx context.Context) WsHandler {
	channels := make(map[int]chan response.ProxyResponse)
	return WsHandler{conn: nil, channels: channels, ctx: ctx}
}

func (ws *WsHandler) IsConnected() bool {
	return ws.conn != nil
}

func (ws *WsHandler) Connect(w http.ResponseWriter, r *http.Request) {
	if ws.IsConnected() {
		_, _ = w.Write([]byte("connection already exists"))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	upgrade := websocket.Upgrader{HandshakeTimeout: 5 * time.Second}
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error during upgrading connection -> %s", err)
		return
	}

	ws.conn = conn
	err = ws.StartPolling()
	if err != nil {
		log.Printf("unable to start polling, closing connection")
		_ = ws.conn.Close()
		ws.conn = nil
	}
}

func (ws *WsHandler) Disconnect() error {
	err := ws.conn.Close()
	ws.conn = nil
	return err
}

func (ws *WsHandler) StartPolling() error {
	if !ws.IsConnected() {
		return errors.New("impossible to start polling -> no connection")
	}

	for {
		select {
		default:
		case <-ws.ctx.Done():
			return nil
		}

		var resp response.ProxyResponse
		err := ws.conn.ReadJSON(&resp)
		if err != nil {
			log.Printf("error during recieving message -> %s", err)
		}

		id := resp.Id
		if output, exists := ws.channels[id]; exists {
			output <- resp
			delete(ws.channels, id)
		} else {
			log.Printf("undefined response id -> %d", id)
		}
	}
}

func (ws *WsHandler) Pass(request request.ProxyRequest, output chan response.ProxyResponse) error {
	if !ws.IsConnected() {
		return errors.New("no connection exists")
	}

	err := ws.conn.WriteJSON(request)
	if err != nil {
		log.Printf("error during message producing -> %s", err)
	}

	ws.channels[request.Id] = output

	return nil
}

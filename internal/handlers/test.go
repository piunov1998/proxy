package handlers

import "net/http"

type WSHandler struct {
}

func (h WSHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Ok!"))
	writer.WriteHeader(200)
}

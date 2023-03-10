package handlers

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"ws_rest_wrapper/package/request"
	"ws_rest_wrapper/package/response"
)

type Proxy interface {
	Pass(request request.ProxyRequest, output chan response.ProxyResponse) error
}

type HttpHandler struct {
	proxy Proxy
}

func NewHttpHandler(proxy Proxy) HttpHandler {
	return HttpHandler{proxy: proxy}
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawReq, _ := httputil.DumpRequest(r, true)
	req := request.ProxyRequest{
		Id:      0,
		RawHttp: rawReq,
	}

	output := make(chan response.ProxyResponse)
	err := h.proxy.Pass(req, output)
	if err != nil {
		w.Write([]byte("server unavailable"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	proxyResp := <-output

	reader := bufio.NewReader(bytes.NewBuffer(proxyResp.RawHttp))
	resp, _ := http.ReadResponse(reader, r)

	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
	}
	_, _ = io.Copy(w, resp.Body)
	_ = resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
}

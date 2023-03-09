package response

type ProxyResponse struct {
	Id      int                 `json:"id"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
	Status  int                 `json:"status"`
}

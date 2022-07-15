package jsonrpcv2

type Request struct {
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      *int           `json:"id"`
}

type Response struct {
	Result  map[string]any `json:"result"`
	Error   map[string]any `json:"error"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      *int           `json:"id"`
}

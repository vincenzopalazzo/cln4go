package plugin

type RPCCallback[T any] = func(*Plugin[T], map[string]any) (map[string]any, error)
type voidRPCCallback[T any] = func(*Plugin[T], map[string]any)

type rpcOption struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Deprecated  bool   `json:"deprecated"`
}

type request struct {
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      *int           `json:"idm,omitempty"`
}

type response struct {
	Result  map[string]any `json:"result,omitempty"`
	Error   map[string]any `json:"error,omitempty"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      *int           `json:"id,omitempty"`
}

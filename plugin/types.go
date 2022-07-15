package plugin

// https://github.com/golang/go/issues/46477#issuecomment-1134888278
// type RPCCallback[T any] = func(*Plugin[T], map[string]any) (map[string]any, error)
// type voidRPCCallback[T any] = func(*Plugin[T], map[string]any)

// RPCCommand
// FIXME: override the command pattern with the generic type alias when implemented
type RPCCommand[T any] interface {
	Call(*Plugin[T], map[string]any) (map[string]any, error)
}

type RPCEvent[T any] interface {
	Call(*Plugin[T], map[string]any)
}

type rpcOption struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Deprecated  bool   `json:"deprecated"`
	Value       any    `json:"-"`
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

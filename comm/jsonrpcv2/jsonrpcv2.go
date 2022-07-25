package jsonrpcv2

type Request struct {
	Method  string `json:"method"`
	Params  any    `json:"params"`
	Jsonrpc string `json:"jsonrpc"`
	Id      *int   `json:"id"`
}

//TODO: core lightning should be consistent with the type that return
// in the params
func (instance *Request) GetParams() map[string]any {
	unboxed, ok := instance.Params.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return unboxed
}

type Response struct {
	Result  map[string]any `json:"result"`
	Error   map[string]any `json:"error"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      *int           `json:"id"`
}

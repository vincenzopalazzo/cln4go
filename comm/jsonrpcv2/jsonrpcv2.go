package jsonrpcv2

// RequestId in the JSON RPC 2.0 Protocol there is the possibility to have
// an id as int or a string, so this type should make the encoder smarter
// and able to decode both types
type RequestId interface {
	// FIXME: remove the any when cln remove the id type
	*string | *int | any
}

type Request[I RequestId] struct {
	Method  string `json:"method"`
	Params  any    `json:"params"`
	Jsonrpc string `json:"jsonrpc"`
	Id      I      `json:"id,omitempty"`
}

// TODO: core lightning should be consistent with the type that return
// in the params
func (instance *Request[I]) GetParams() map[string]any {
	unboxed, ok := instance.Params.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return unboxed
}

type Response[I RequestId] struct {
	Result  map[string]any `json:"result"`
	Error   map[string]any `json:"error"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      I              `json:"id, omitempty"`
}

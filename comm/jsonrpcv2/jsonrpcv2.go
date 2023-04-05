package jsonrpcv2

// Id in the JSON RPC 2.0 Protocol there is the possibility to have
// an id as int or a string, so this type should make the encoder smarter
// and able to decode both types
type Id interface {
	// FIXME: remove the any when cln remove the id type
	*string | *int | any
}

type Request[I Id, P any] struct {
	Method  string `json:"method"`
	Params  P      `json:"params"`
	Jsonrpc string `json:"jsonrpc"`
	Id      I      `json:"id,omitempty"`
}

// TODO: core lightning should be consistent with the type that return
// in the params
func (instance *Request[I, P]) GetParams() P {
	return instance.Params
}

type Response[I Id, R any] struct {
	Result  R              `json:"result"`
	Error   map[string]any `json:"error"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      I              `json:"id,omitempty"`
}

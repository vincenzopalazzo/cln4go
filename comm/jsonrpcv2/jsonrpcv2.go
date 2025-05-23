package jsonrpcv2

import "fmt"

// Id in the JSON RPC 2.0 Protocol there is the possibility to have
// an id as int or a string, so this type should make the encoder smarter
// and able to decode both types
type Id interface {
	// FIXME: remove the any when cln remove the id type
	*string | *int | any
}

type Params interface {
	map[string]any | []any
}

type Request struct {
	Method  string `json:"method"`
	Params  any    `json:"params"`
	Jsonrpc string `json:"jsonrpc"`
	Id      Id     `json:"id,omitempty"`
}

// TODO: core lightning should be consistent with the type that return
// in the params
func (instance *Request) GetParams() map[string]any {
	switch params := instance.Params.(type) {
	case map[string]any:
		return params
	case []any:
		if len(params) != 0 {
			panic(fmt.Sprintf("%s", params))
		}
		return map[string]any{}
	default:
		panic("Params has a different type")
	}
}

type Response[R any] struct {
	Result  R             `json:"result"`
	Error   *JSONRPCError `json:"error"`
	Jsonrpc string        `json:"jsonrpc"`
	Id      Id            `json:"id,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error object.
type JSONRPCError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

func (self *JSONRPCError) Error() string {
	return fmt.Sprintf("JSONRPCError: code=%d, message=%s, data=%v", self.Code, self.Message, self.Data)
}

func MakeRPCError(code int, message string, data map[string]any) error {
	return &JSONRPCError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

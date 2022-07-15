package client

import (
	"encoding/json"
	"reflect"
)

// Client - Common interface for all the client supported by the library.
type Client interface {
	Call(method string, payload map[string]any) (map[string]any, error)
}

// Call - Generic call for perform a RPC call.
func Call[C Client, Req any, Resp any](client *C, method string, payload Req) (*Resp, error) {
	result, err := (*client).Call(method, fromTypeToMap(payload))
	if err != nil {
		return nil, err
	}

	// FIXME: this has performance issue!
	byteResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var typ Resp
	// FIXME: make the Unmarshall independent from the library used
	if err := json.Unmarshal(byteResult, &typ); err != nil {
		return nil, err
	}
	return &typ, nil
}

// PlainCall - Generic call for perform a RPC call and return an hash map as response.
func PlainCall[C Client](client C, method string, payload map[string]any) (map[string]any, error) {
	result, err := client.Call(method, payload)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FIXME: move in https://github.com/LNOpenMetrics/lnmetrics.utils
func fromTypeToMap(typeInstance any) map[string]any {
	if reflect.ValueOf(typeInstance).Kind() != reflect.Struct {
		return typeInstance.(map[string]any)
	}
	mapp := make(map[string]any)
	elem := reflect.ValueOf(&typeInstance).Elem()
	relType := elem.Type()
	for i := 0; i < relType.NumField(); i++ {
		mapp[relType.Field(i).Name] = elem.Field(i).Interface()
	}
	return mapp
}

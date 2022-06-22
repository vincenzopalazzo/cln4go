package client

import (
	"encoding/json"
)

// Client - Common interface for all the client supported by the library.
type Client interface {
	Call(method string, payload map[string]any) (map[string]any, error)
}

// Call - Generic call for perform a RPC call.
func Call[C Client, R any](client *C, method string, payload map[string]any) (*R, error) {
	result, err := (*client).Call(method, payload)
	if err != nil {
		return nil, err
	}

	// FIXME: this has performance issue!
	byteResult, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var typ R
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

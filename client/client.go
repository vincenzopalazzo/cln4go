package client

import (
	"encoding/json"
)

type Client interface {
	Call(method string, payload map[string]any) ([]byte, error)
}

func Call[C Client, R any](client C, method string, payload map[string]any) (*R, error) {
	jsonResp, err := client.Call(method, payload)
	if err != nil {
		return nil, err
	}
	var typ R
	if err := json.Unmarshal(jsonResp, &typ); err != nil {
		return nil, err
	}
	return &typ, nil
}

package client

type Client interface {
	call(method string, payload map[string]interface{}) (interface{}, error)
}

type UNIXRpc struct {
}

func New() Client {
	return &UNIXRpc{}
}

func (instance *UNIXRpc) call(method string, payload map[string]interface{}) (interface{}, error) {
	return nil, nil
}

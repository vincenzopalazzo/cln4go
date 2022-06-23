package plugin

type rpcMethod[T any] struct {
	name            string
	usage           string
	description     string
	longDescription string
	callback        RPCCallback[T]
}

func (instance *rpcMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	return instance.callback(plugin, request)
}

type rpcNotification[T any] struct {
	onEvent  string
	callback voidRPCCallback[T]
}

func (instance *rpcNotification[T]) Call(plugin *Plugin[T], request map[string]any) {
	instance.callback(plugin, request)
}

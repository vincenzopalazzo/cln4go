package plugin

type rpcMethod[T any] struct {
	Name            string                                                                  `json:"name"`
	Usage           string                                                                  `json:"usage"`
	Description     string                                                                  `json:"description"`
	LongDescription string                                                                  `json:"long_description"`
	callback        func(plugin *Plugin[T], request map[string]any) (map[string]any, error) `json:"-"`
}

func (instance *rpcMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	return instance.callback(plugin, request)
}

type rpcNotification[T any] struct {
	onEvent  string
	callback func(plugin *Plugin[T], request map[string]any)
}

func (instance *rpcNotification[T]) Call(plugin *Plugin[T], request map[string]any) {
	instance.callback(plugin, request)
}

type rpcHook[T any] struct {
	name     string
	before   []string
	after    []string
	callback func(plugin *Plugin[T], request map[string]any) (map[string]any, error)
}

func (instance *rpcHook[T]) Call(plugin *Plugin[T], request map[string]any) {
	if _, err := instance.callback(plugin, request); err != nil {
		plugin.tracer.Infof("%s", err)
	}
}

package plugin

func getManifest[T any](plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	return map[string]any{"hello": "hello from go 1.18"}, nil
}

func initMethod[T any](plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	return map[string]any{"hello": "hello from go 1.18"}, nil
}

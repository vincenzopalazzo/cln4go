package comm

func GenerateArray[V any](mapData map[string]V) []V {
	v := make([]V, 0, len(mapData))

	for _, value := range mapData {
		v = append(v, value)
	}
	return v
}

func GenerateKeyArray[V any](mapData map[string]V) []string {
	k := make([]string, 0, len(mapData))

	for key, _ := range mapData {
		k = append(k, key)
	}

	return k
}

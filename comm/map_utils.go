package comm

func GenerateArray[V any](mapData map[string]V, skipIfExist map[string]bool) []V {
	v := make([]V, 0, len(mapData))

	for key, value := range mapData {
		_, found := skipIfExist[key]
		if found {
			continue
		}
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

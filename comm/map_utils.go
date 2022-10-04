package comm

// GenerateArray take a map with any data and return an array of values, if a
// value with the same key is specified inside the `skipIfExist` map the value will
// be not insert inside the list.
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

// GenerateKeyArray return an array with all the key value of the `mapData`
func GenerateKeyArray[V any](mapData map[string]V) []string {
	k := make([]string, 0, len(mapData))

	for key, _ := range mapData {
		k = append(k, key)
	}

	return k
}

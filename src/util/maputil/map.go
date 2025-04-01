package maputil

func Map[K comparable, T any, U any](m map[K]T, mapper func(T) U) map[K]U {
	result := make(map[K]U, len(m))
	for k, v := range m {
		result[k] = mapper(v)
	}
	return result
}

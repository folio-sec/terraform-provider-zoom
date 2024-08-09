package util

func ArrayMap[T, V any](elms []T, fn func(T) V) []V {
	outputs := make([]V, len(elms), cap(elms))
	for i, elm := range elms {
		outputs[i] = fn(elm)
	}
	return outputs
}

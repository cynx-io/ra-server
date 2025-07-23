package helper

func PtrOrDefault[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}

func Containsint32(slice []int32, item int32) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

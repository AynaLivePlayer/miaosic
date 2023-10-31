package providers

func CheckPageParam(page, size int) bool {
	if page < 1 || size < 1 {
		return false
	}
	return true
}

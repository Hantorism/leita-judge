package utils

func All(s []bool) bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}

func AllString(s ...string) bool {
	for _, v := range s {
		if v == "" {
			return false
		}
	}
	return true
}

func Sum[T ~int | ~int64](s []T) T {
	var sum T
	for _, v := range s {
		sum += v
	}
	return sum
}

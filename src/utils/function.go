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

package function

func All(s []bool) bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}

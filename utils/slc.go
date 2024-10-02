package utils

// SlideStrEqual ...
func SlideStrEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// ContainStr ...
func ContainStr(ss []string, s string) bool {
	for i := range ss {
		if s == ss[i] {
			return true
		}
	}
	return false
}

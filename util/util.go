package util

func ContainsString(slice []string, str string) bool {
	for _, sliceStr := range slice {
		if sliceStr == str {
			return true
		}
	}
	return false
}
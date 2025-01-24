package ultraocr

func isNil(value any) bool {
	switch value := value.(type) {
	case nil:
		return true
	case []map[string]any:
		return len(value) == 0
	case map[string]any:
		return len(value) == 0
	default:
		return false
	}
}

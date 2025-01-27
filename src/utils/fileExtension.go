package utils

func FileExtension(language string) string {
	switch language {
	case "C":
		return "c"
	case "CPP":
		return "cpp"
	case "GO":
		return "go"
	case "JAVA":
		return "java"
	case "JAVASCRIPT":
		return "js"
	case "KOTLIN":
		return "kt"
	case "PYTHON":
		return "py"
	case "SWIFT":
		return "swift"
	default:
		return "error"
	}
}

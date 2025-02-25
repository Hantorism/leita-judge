package utils

import "strings"

func ReplaceSubmitId(args []string, submitID string) []string {
	replaced := make([]string, len(args))
	for i, arg := range args {
		replaced[i] = strings.ReplaceAll(arg, "{SUBMITID}", submitID)
	}
	return replaced
}
package utils

import (
	"strconv"
	"strings"
)

func ReplaceSubmitId(args []string, submitID int) []string {
	replaced := make([]string, len(args))
	for i, arg := range args {
		replaced[i] = strings.ReplaceAll(arg, "{SUBMITID}", strconv.Itoa(submitID))
	}
	return replaced
}
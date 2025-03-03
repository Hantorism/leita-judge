package utils

import (
	"strconv"
	"strings"
)

func ReplaceCommand(args []string, judgeType string, submitID int) []string {
	replaced := make([]string, len(args))
	for i, arg := range args {
		arg = strings.ReplaceAll(arg, "{JUDGE_TYPE}", judgeType)
		replaced[i] = strings.ReplaceAll(arg, "{SUBMIT_ID}", strconv.Itoa(submitID))
	}
	return replaced
}

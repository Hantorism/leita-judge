package models

type JudgeResult struct {
	ProblemId    int    `json:"problem_id"`
	Result       string `json:"result"`
	SizeOfCode   int    `json:"size_of_code"`
	Language     string `json:"language"`
	Memory       int    `json:"memory"`
	Time         int    `json:"time"`
	UserId       int    `json:"user_id"`
	UsedLanguage string `json:"used_language"`
	UsedMemory   int    `json:"used_memory"`
	UsedTime     int    `json:"used_time"`
}

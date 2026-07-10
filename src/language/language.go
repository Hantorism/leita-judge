package language

import (
	"strconv"
	"strings"
)

const FileName = "Main"

type Command struct {
	BuildCmd  []string
	RunCmd    []string
	DeleteCmd []string
}

var Commands = map[string]Command{
	"C": {
		BuildCmd:  []string{"gcc", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".c", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName, "-O2", "-Wall", "-lm", "-static", "-std=gnu99"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
	},
	"CPP": {
		BuildCmd:  []string{"g++", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".cpp", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName, "-O2", "-Wall", "-lm", "-static", "-std=gnu++17"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
	},
	"JAVA": {
		BuildCmd:  []string{"javac", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-encoding", "UTF-8", "-d", "bin", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".java"},
		RunCmd:    []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-cp", "bin", FileName},
		DeleteCmd: []string{"rm", "-r", "{JUDGE_TYPE}/{SUBMIT_ID}/bin"},
	},
	"PYTHON": {
		BuildCmd:  []string{},
		RunCmd:    []string{"python3", "-W", "ignore", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".py"},
		DeleteCmd: []string{},
	},
	"JAVASCRIPT": {
		BuildCmd:  []string{},
		RunCmd:    []string{"node", "--stack-size=65536", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".js"},
		DeleteCmd: []string{},
	},
	"GO": {
		BuildCmd:  []string{"go", "build", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName, "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".go"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
	},
	"KOTLIN": {
		BuildCmd:  []string{"kotlinc", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-include-runtime", "-d", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".jar", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".kt"},
		RunCmd:    []string{"java", "-Xms1024m", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-jar", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".jar"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".jar"},
	},
	"SWIFT": {
		BuildCmd:  []string{"swiftc", "-O", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName, "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".swift"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
	},
}

// memoryMarginKB는 언어 런타임의 고정 메모리 비용을 보상하기 위해 문제 메모리
// 제한에 가산하는 값이다. a+b 프로그램 실측 베이스라인(C/C++ ~0.5MB, Go ~4.8MB,
// Python ~5MB, JavaScript ~9MB, JVM ~38MB)에 여유를 더해 정했다.
var memoryMarginKB = map[string]int{
	"C":          16 * 1024,
	"CPP":        16 * 1024,
	"GO":         32 * 1024,
	"PYTHON":     32 * 1024,
	"JAVASCRIPT": 64 * 1024,
	"JAVA":       128 * 1024,
	"KOTLIN":     128 * 1024,
	"SWIFT":      16 * 1024,
}

// MemoryLimitWithMargin은 문제 메모리 제한(KB)에 언어별 마진을 더해 반환한다.
// 이 값이 cgroup memory.max로 설정되어, 초과 시 OOM kill → MEMORY_OUT 판정이 된다.
// 마진 이내의 초과는 런타임 고정비로 간주해 허용하는 것이 정책 의도다.
func MemoryLimitWithMargin(language string, memoryLimitKB int) int {
	if memoryLimitKB <= 0 {
		return memoryLimitKB
	}
	return memoryLimitKB + memoryMarginKB[language]
}

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

func ReplaceCommand(args []string, judgeType string, submitID int) []string {
	replacer := strings.NewReplacer(
		"{JUDGE_TYPE}", judgeType,
		"{SUBMIT_ID}", strconv.Itoa(submitID),
	)
	replaced := make([]string, len(args))
	for i, arg := range args {
		replaced[i] = replacer.Replace(arg)
	}
	return replaced
}

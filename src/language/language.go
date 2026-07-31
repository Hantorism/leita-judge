package language

import (
	"strconv"
	"strings"
)

const (
	FileName = "Main"

	// CsprojTemplatePath는 C# 이미지에 미리 넣어 두는 프로젝트 템플릿 위치다.
	// 파일명이 곧 어셈블리 이름이 되므로 Main.csproj여야 Main.dll이 나온다.
	CsprojTemplatePath = "/template/" + FileName + ".csproj"

	// NodeTypeRoot는 TypeScript 이미지에 전역 설치된 @types 위치다.
	NodeTypeRoot = "/usr/local/lib/node_modules/@types"
)

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
	// JVM 계열의 RunCmd에는 -Xms를 주지 않는다. 초기 힙을 크게 잡으면 JVM이 힙 예산이
	// 넉넉하다고 판단해 GC를 미루다가 cgroup 상한에 먼저 부딪혀, 정상 코드가 MEMORY_OUT으로
	// 오판정된다(실측: 동일 GC 부하 워크로드가 -Xms1024m에서는 224MB, 제거 시 96MB 필요).
	// 반대로 -Xmx는 문제 제한보다 크게 유지해야 힙 한계 대신 cgroup OOM이 먼저 발생해
	// MEMORY_OUT 판정이 일관된다(줄이면 OutOfMemoryError → RUNTIME_ERROR로 샌다).
	"JAVA": {
		BuildCmd:  []string{"javac", "-J-Xms1024m", "-J-Xmx1920m", "-J-Xss512m", "-encoding", "UTF-8", "-d", "bin", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".java"},
		RunCmd:    []string{"java", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-cp", "bin", FileName},
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
		RunCmd:    []string{"java", "-Xmx1920m", "-Xss512m", "-Dfile.encoding=UTF-8", "-XX:+UseSerialGC", "-jar", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".jar"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".jar"},
	},
	"SWIFT": {
		BuildCmd:  []string{"swiftc", "-O", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName, "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".swift"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
	},
	// Rust 에디션은 2021로 고정한다. 2024는 static mut 전역에 참조가 생기면 하드 에러라
	// PS에서 흔한 전역 배열 관용구가 깨진다. 전환은 이 플래그만 바꾸면 되고 버퍼 값에는 영향이 없다.
	"RUST": {
		BuildCmd:  []string{"rustc", "--edition", "2021", "-O", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName, "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".rs"},
		RunCmd:    []string{"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName},
	},
	// C#은 프로젝트 파일이 있어야 빌드된다. 이미지에 넣어 둔 템플릿(CsprojTemplatePath)을
	// 제출 디렉터리로 복사한 뒤 빌드하며, 템플릿 이름이 곧 산출물 이름(Main.dll)이 된다.
	"CS": {
		BuildCmd: []string{"sh", "-c",
			"cp " + CsprojTemplatePath + " {JUDGE_TYPE}/{SUBMIT_ID}/ && " +
				"dotnet build {JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".csproj -c Release -o {JUDGE_TYPE}/{SUBMIT_ID}/out --nologo -v q"},
		RunCmd: []string{"dotnet", "{JUDGE_TYPE}/{SUBMIT_ID}/out/" + FileName + ".dll"},
		DeleteCmd: []string{"rm", "-rf",
			"{JUDGE_TYPE}/{SUBMIT_ID}/out", "{JUDGE_TYPE}/{SUBMIT_ID}/obj", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".csproj"},
	},
	// TypeScript는 tsc로 JS를 만든 뒤 JavaScript와 동일한 런타임에서 실행한다.
	// 사용자 코드가 require/process 등을 쓰므로 전역 @types/node를 타입 루트로 지정해야
	// 정상 코드가 타입 에러(COMPILE_ERROR)로 오판정되지 않는다.
	"TYPESCRIPT": {
		BuildCmd: []string{"tsc", "--target", "ES2020", "--module", "commonjs",
			"--typeRoots", NodeTypeRoot, "--types", "node",
			"{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".ts"},
		RunCmd:    []string{"node", "--stack-size=65536", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".js"},
		DeleteCmd: []string{"rm", "{JUDGE_TYPE}/{SUBMIT_ID}/" + FileName + ".js"},
	},
}

// limitBuffer는 언어 런타임의 고정 비용을 보상해 언어 간 형평성을 맞추는 보정값이다.
// 시간은 인터프리터의 느림이 작업량에 비례하므로 배수+가산으로, 메모리는 런타임
// 고정비가 문제 제한과 무관한 상수이므로 고정 가산으로만 보정한다.
type limitBuffer struct {
	timeMultiplier int // 시간 제한 배수 (1이면 배수 없음)
	timeAddMs      int // 시간 제한 가산 (ms)
	memoryAddKB    int // 메모리 제한 가산 (KB)
}

// limitBuffers의 시간 보정은 BOJ 정책(help.acmicpc.net/language/info)을 따르고,
// 메모리 보정은 실측값으로 정했다. 30MB를 실제로 사용하는 풀이의 cgroup 오버헤드는
// C/C++ 0.3MB, Python 3.9MB, Go 6.1MB, JavaScript 9.4MB, JVM 39.2MB였다.
var limitBuffers = map[string]limitBuffer{
	// 네이티브: 문제 제한 자체가 C/C++ 기준으로 설계되므로 보정하지 않는다.
	// 마진을 주면 메모리 최적화 문제(16MB 제한 등)의 출제 의도가 깨진다.
	"C":   {timeMultiplier: 1},
	"CPP": {timeMultiplier: 1},
	// Go: 오버헤드는 6MB지만 GC가 라이브 힙 2배까지 늘어난 뒤 수거(GOGC=100)해 여유가 필요하다.
	"GO": {timeMultiplier: 1, timeAddMs: 2000, memoryAddKB: 32 * 1024},
	// 인터프리터: 네이티브 대비 수십 배 느리다. JavaScript는 런타임 하한이 12MB라 마진이 필수.
	"PYTHON":     {timeMultiplier: 3, timeAddMs: 2000, memoryAddKB: 32 * 1024},
	"JAVASCRIPT": {timeMultiplier: 3, timeAddMs: 2000, memoryAddKB: 32 * 1024},
	// JVM: 베이스라인 38MB + SerialGC 복사 공간. 실측상 +64MB가 최소 통과선이라 2배로 잡았다.
	"JAVA":   {timeMultiplier: 2, timeAddMs: 1000, memoryAddKB: 128 * 1024},
	"KOTLIN": {timeMultiplier: 2, timeAddMs: 1000, memoryAddKB: 128 * 1024},
	// Swift/Rust: 네이티브라 속도는 C와 동급(실측 1.0×)이고 오버헤드도 1~2MB에 그친다.
	// 다만 문제 제한이 이 언어들 기준으로 설계되지는 않으므로 최소 쿠션만 둔다.
	"SWIFT": {timeMultiplier: 1, memoryAddKB: 16 * 1024},
	"RUST":  {timeMultiplier: 1, memoryAddKB: 16 * 1024},
	// C#: GC 런타임이지만 오버헤드가 5.7MB(GC 압박 시 10.8MB)로 JVM(39.2MB)보다 훨씬 가볍다.
	// 시간은 런타임 기동(17ms)과 JIT 웜업을 감안해 JVM과 같은 보정을 준다.
	"CS": {timeMultiplier: 2, timeAddMs: 1000, memoryAddKB: 64 * 1024},
	// TypeScript: JavaScript와 동일한 node 런타임이라 실측(하한 12MB, 오버헤드 9.2MB)도 같다.
	"TYPESCRIPT": {timeMultiplier: 3, timeAddMs: 2000, memoryAddKB: 32 * 1024},
}

// bufferOf는 등록되지 않은 언어에 대해 보정 없음(배수 1, 가산 0)을 보장한다.
func bufferOf(language string) limitBuffer {
	buffer, ok := limitBuffers[language]
	if !ok {
		return limitBuffer{timeMultiplier: 1}
	}
	return buffer
}

// TimeLimitWithBuffer는 문제 시간 제한(ms)에 언어별 버퍼를 적용해 반환한다.
// 이 값이 실행 데드라인이 되며, 초과하면 TIME_OUT으로 판정된다.
func TimeLimitWithBuffer(language string, timeLimitMs int) int {
	if timeLimitMs <= 0 {
		return timeLimitMs
	}
	buffer := bufferOf(language)
	return timeLimitMs*buffer.timeMultiplier + buffer.timeAddMs
}

// MemoryLimitWithBuffer는 문제 메모리 제한(KB)에 언어별 버퍼를 더해 반환한다.
// 이 값이 cgroup memory.max로 설정되어, 초과 시 OOM kill → MEMORY_OUT 판정이 된다.
// 버퍼 이내의 초과는 런타임 고정비로 간주해 허용하는 것이 정책 의도다.
func MemoryLimitWithBuffer(language string, memoryLimitKB int) int {
	if memoryLimitKB <= 0 {
		return memoryLimitKB
	}
	return memoryLimitKB + bufferOf(language).memoryAddKB
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
	case "RUST":
		return "rs"
	case "CS":
		return "cs"
	case "TYPESCRIPT":
		return "ts"
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

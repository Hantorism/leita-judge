package language

import "testing"

func TestFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		language string
		want     string
	}{
		{"C", "C", "c"},
		{"CPP", "CPP", "cpp"},
		{"GO", "GO", "go"},
		{"JAVA", "JAVA", "java"},
		{"JAVASCRIPT", "JAVASCRIPT", "js"},
		{"KOTLIN", "KOTLIN", "kt"},
		{"PYTHON", "PYTHON", "py"},
		{"SWIFT", "SWIFT", "swift"},
		{"RUST", "RUST", "rs"},
		{"CS", "CS", "cs"},
		{"TYPESCRIPT", "TYPESCRIPT", "ts"},
		{"unknown language", "HASKELL", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FileExtension(tt.language)
			if got != tt.want {
				t.Errorf("FileExtension(%q) = %q, want %q", tt.language, got, tt.want)
			}
		})
	}
}

func TestSupports(t *testing.T) {
	supported := []string{"C", "CPP", "GO", "JAVA", "JAVASCRIPT", "KOTLIN", "PYTHON", "SWIFT", "RUST", "CS", "TYPESCRIPT"}
	for _, lang := range supported {
		t.Run(lang+"는 지원한다", func(t *testing.T) {
			if !Supports(lang) {
				t.Errorf("Supports(%q) = false, want true", lang)
			}
		})
	}

	for _, lang := range []string{"HASKELL", "typescript", "C#", ""} {
		t.Run("미지원: "+lang, func(t *testing.T) {
			if Supports(lang) {
				t.Errorf("Supports(%q) = true, want false", lang)
			}
		})
	}
}

// 모든 지원 언어는 실행 커맨드와 버퍼 항목을 반드시 가져야 한다.
// 빈 실행 커맨드는 executor에서 인덱스 패닉으로 이어지고,
// 버퍼 항목이 없으면 보정 없이 채점되어 언어 간 형평성이 깨진다.
func TestEverySupportedLanguageIsFullyConfigured(t *testing.T) {
	for lang, command := range Commands {
		t.Run(lang, func(t *testing.T) {
			if len(command.RunCmd) == 0 {
				t.Errorf("%s: RunCmd가 비어 있다", lang)
			}
			if FileExtension(lang) == "error" {
				t.Errorf("%s: FileExtension이 정의되지 않았다", lang)
			}
			if _, ok := limitBuffers[lang]; !ok {
				t.Errorf("%s: limitBuffers 항목이 없다", lang)
			}
		})
	}
}

func TestReplaceCommand(t *testing.T) {
	args := []string{"gcc", "{JUDGE_TYPE}/{SUBMIT_ID}/Main.c", "-o", "{JUDGE_TYPE}/{SUBMIT_ID}/Main"}

	got := ReplaceCommand(args, "submit", 42)

	want := []string{"gcc", "submit/42/Main.c", "-o", "submit/42/Main"}

	if len(got) != len(want) {
		t.Fatalf("ReplaceCommand() returned %d args, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ReplaceCommand()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

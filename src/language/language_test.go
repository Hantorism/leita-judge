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
		{"unknown language", "RUST", "error"},
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

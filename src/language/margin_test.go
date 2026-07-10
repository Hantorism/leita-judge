package language

import "testing"

func TestMemoryLimitWithMargin(t *testing.T) {
	tests := []struct {
		name     string
		language string
		limitKB  int
		want     int
	}{
		{name: "C는 +16MB", language: "C", limitKB: 262144, want: 262144 + 16*1024},
		{name: "CPP는 +16MB", language: "CPP", limitKB: 262144, want: 262144 + 16*1024},
		{name: "GO는 +32MB", language: "GO", limitKB: 262144, want: 262144 + 32*1024},
		{name: "PYTHON은 +32MB", language: "PYTHON", limitKB: 262144, want: 262144 + 32*1024},
		{name: "JAVASCRIPT는 +64MB", language: "JAVASCRIPT", limitKB: 262144, want: 262144 + 64*1024},
		{name: "JAVA는 +128MB", language: "JAVA", limitKB: 262144, want: 262144 + 128*1024},
		{name: "KOTLIN은 +128MB", language: "KOTLIN", limitKB: 262144, want: 262144 + 128*1024},
		{name: "SWIFT는 +16MB", language: "SWIFT", limitKB: 262144, want: 262144 + 16*1024},
		{name: "미지원 언어는 마진 없음", language: "RUST", limitKB: 262144, want: 262144},
		{name: "제한이 0이면 그대로 (제한 미설정 의미 유지)", language: "JAVA", limitKB: 0, want: 0},
		{name: "제한이 음수면 그대로", language: "JAVA", limitKB: -1, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MemoryLimitWithMargin(tt.language, tt.limitKB); got != tt.want {
				t.Errorf("MemoryLimitWithMargin(%q, %d) = %d, want %d", tt.language, tt.limitKB, got, tt.want)
			}
		})
	}
}

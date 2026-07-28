package language

import "testing"

func TestMemoryLimitWithBuffer(t *testing.T) {
	const limitKB = 262144 // 256MB

	tests := []struct {
		name     string
		language string
		limitKB  int
		want     int
	}{
		{name: "C는 보정 없음 (문제 제한이 이미 C 기준)", language: "C", limitKB: limitKB, want: limitKB},
		{name: "CPP는 보정 없음", language: "CPP", limitKB: limitKB, want: limitKB},
		{name: "GO는 +32MB", language: "GO", limitKB: limitKB, want: limitKB + 32*1024},
		{name: "PYTHON은 +32MB", language: "PYTHON", limitKB: limitKB, want: limitKB + 32*1024},
		{name: "JAVASCRIPT는 +32MB", language: "JAVASCRIPT", limitKB: limitKB, want: limitKB + 32*1024},
		{name: "JAVA는 +128MB", language: "JAVA", limitKB: limitKB, want: limitKB + 128*1024},
		{name: "KOTLIN은 +128MB", language: "KOTLIN", limitKB: limitKB, want: limitKB + 128*1024},
		{name: "SWIFT는 +16MB", language: "SWIFT", limitKB: limitKB, want: limitKB + 16*1024},
		{name: "미지원 언어는 보정 없음", language: "RUST", limitKB: limitKB, want: limitKB},
		{name: "제한이 0이면 그대로 (제한 미설정 의미 유지)", language: "JAVA", limitKB: 0, want: 0},
		{name: "제한이 음수면 그대로", language: "JAVA", limitKB: -1, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MemoryLimitWithBuffer(tt.language, tt.limitKB); got != tt.want {
				t.Errorf("MemoryLimitWithBuffer(%q, %d) = %d, want %d", tt.language, tt.limitKB, got, tt.want)
			}
		})
	}
}

func TestTimeLimitWithBuffer(t *testing.T) {
	const limitMs = 1000

	tests := []struct {
		name     string
		language string
		limitMs  int
		want     int
	}{
		{name: "C는 보정 없음", language: "C", limitMs: limitMs, want: 1000},
		{name: "CPP는 보정 없음", language: "CPP", limitMs: limitMs, want: 1000},
		{name: "GO는 +2초 (배수 없음)", language: "GO", limitMs: limitMs, want: 3000},
		{name: "PYTHON은 ×3+2초", language: "PYTHON", limitMs: limitMs, want: 5000},
		{name: "JAVASCRIPT는 ×3+2초", language: "JAVASCRIPT", limitMs: limitMs, want: 5000},
		{name: "JAVA는 ×2+1초", language: "JAVA", limitMs: limitMs, want: 3000},
		{name: "KOTLIN은 ×2+1초", language: "KOTLIN", limitMs: limitMs, want: 3000},
		{name: "SWIFT는 보정 없음", language: "SWIFT", limitMs: limitMs, want: 1000},
		{name: "미지원 언어는 보정 없음 (배수 0이 되면 안 됨)", language: "RUST", limitMs: limitMs, want: 1000},
		{name: "배수는 제한에 비례한다 (PYTHON 2초 → 8초)", language: "PYTHON", limitMs: 2000, want: 8000},
		{name: "제한이 0이면 그대로", language: "PYTHON", limitMs: 0, want: 0},
		{name: "제한이 음수면 그대로", language: "PYTHON", limitMs: -1, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeLimitWithBuffer(tt.language, tt.limitMs); got != tt.want {
				t.Errorf("TimeLimitWithBuffer(%q, %d) = %d, want %d", tt.language, tt.limitMs, got, tt.want)
			}
		})
	}
}

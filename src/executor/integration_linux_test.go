//go:build linux

package executor

import (
	"testing"

	"leita/src/cgroup"
	"leita/src/entity"
)

// TestRunMeasuresMemory는 실제 cgroupfs가 쓰기 가능한 환경(--privileged 컨테이너)
// 에서만 동작하는 통합 테스트다. 그 외 환경에서는 skip된다.
// 계획서 Phase 5의 검증 기준: dd가 30MB 버퍼를 할당하므로 측정값이 그 근처여야 한다.
func TestRunMeasuresMemory(t *testing.T) {
	monitor, err := cgroup.Setup()
	if err != nil {
		t.Skipf("cgroup을 쓸 수 없는 환경이라 skip: %v", err)
	}

	exec := NewOsExecutor(monitor)
	result, _, usedTime, usedMemory, err := exec.Run(
		[]string{"dd", "if=/dev/zero", "of=/dev/null", "bs=30M", "count=1"},
		nil,
		5000,
		262144, // 256MB
	)
	if err != nil {
		t.Fatalf("Run 실패: %v", err)
	}
	if result != entity.JudgeCorrect {
		t.Errorf("result = %v, want CORRECT", result)
	}
	if usedTime <= 0 {
		t.Errorf("usedTime = %d, 0보다 커야 함", usedTime)
	}
	// dd 버퍼 30MB(30720KB) + 고정비. 측정이 동작하면 30MB 이상,
	// 상한은 여유 있게 3배로 잡는다.
	if usedMemory < 30720 || usedMemory > 92160 {
		t.Errorf("usedMemory = %dKB, want 30720~92160KB (30MB 버퍼 + 고정비)", usedMemory)
	}
}

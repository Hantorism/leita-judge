//go:build !linux

package cgroup

import "os/exec"

// Apply는 non-Linux에서 no-op이다. macOS에서는 컴파일(IDE/gopls) 호환만 보장하며,
// 서버 실행과 테스트는 전부 Docker(Linux)에서 수행한다.
func (s *fsSession) Apply(cmd *exec.Cmd) error {
	return nil
}

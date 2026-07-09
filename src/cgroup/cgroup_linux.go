//go:build linux

package cgroup

import (
	"os/exec"
	"syscall"
)

// Apply는 clone3(CLONE_INTO_CGROUP)로 프로세스가 태어나는 순간부터 세션 cgroup에
// 속하게 한다. 시작 후 cgroup.procs에 PID를 쓰는 방식은 초기 할당 메모리가
// 누락되는 레이스가 있어 쓰지 않는다.
func (s *fsSession) Apply(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.UseCgroupFD = true
	cmd.SysProcAttr.CgroupFD = int(s.fd.Fd())
	return nil
}

package cgroup

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
)

const (
	// DefaultRoot는 cgroup v2 파일시스템의 마운트 지점이다.
	DefaultRoot = "/sys/fs/cgroup"

	// baseDirName 아래에 파드별 격리 계층({pod})이 생기고, 그 아래에 세션 cgroup이 생긴다.
	// 최종 경로: {root}/leita-judge/{pod}/{session}
	baseDirName = "leita-judge"

	// leafName은 로컬 Docker(private cgroupns)에서 "no internal processes" 규칙을
	// 피하기 위해 서버 프로세스를 옮겨 둘 leaf cgroup 이름이다.
	leafName = "main"
)

// Monitor는 채점 프로세스 1회 실행에 대한 메모리 측정 세션을 발급한다.
type Monitor interface {
	NewSession(name string, memoryLimitKB int) (Session, error)
}

// Session은 실행 1회를 감싸는 일회용 cgroup이다. cmd 시작 전에 Apply를 호출하고,
// 종료 후 PeakMemoryKB/OOMKilled를 읽은 뒤 Close로 정리한다.
type Session interface {
	Apply(cmd *exec.Cmd) error
	PeakMemoryKB() (int64, error)
	OOMKilled() (bool, error)
	Close() error
}

// Setup은 cgroup 환경을 감지하고 측정 준비를 한다. 실패하면 측정하지 않는
// 폴백 Monitor와 원인 에러를 함께 반환하므로 서버는 항상 기동할 수 있다.
func Setup() (Monitor, error) {
	return SetupAt(DefaultRoot)
}

// SetupAt은 테스트에서 가짜 cgroupfs 루트를 주입할 수 있도록 분리된 Setup 본체다.
func SetupAt(root string) (Monitor, error) {
	controllers, err := os.ReadFile(filepath.Join(root, "cgroup.controllers"))
	if err != nil {
		return NewFallbackMonitor(), fmt.Errorf("cgroup v2를 사용할 수 없습니다: %w", err)
	}
	if !hasController(string(controllers), "memory") {
		return NewFallbackMonitor(), errors.New("cgroup memory 컨트롤러를 사용할 수 없습니다")
	}

	if err := escapeRootIfNeeded(root); err != nil {
		return NewFallbackMonitor(), err
	}

	pod, err := os.Hostname()
	if err != nil {
		return NewFallbackMonitor(), fmt.Errorf("파드 식별자(hostname)를 얻을 수 없습니다: %w", err)
	}

	base := filepath.Join(root, baseDirName)
	podDir := filepath.Join(base, pod)
	for _, dir := range []string{base, podDir} {
		if err := os.Mkdir(dir, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
			return NewFallbackMonitor(), fmt.Errorf("cgroup 생성 실패 (%s): %w", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "cgroup.subtree_control"), []byte("+memory"), 0o644); err != nil {
			return NewFallbackMonitor(), fmt.Errorf("memory 컨트롤러 위임 실패 (%s): %w", dir, err)
		}
	}

	cleanupLeftovers(podDir)

	return &fsMonitor{dir: podDir}, nil
}

// readSelfCgroup은 테스트에서 /proc/self/cgroup 내용을 주입할 수 있도록 변수다.
var readSelfCgroup = func() ([]byte, error) {
	return os.ReadFile("/proc/self/cgroup")
}

// inRootCgroup은 서버 프로세스가 (현재 cgroup 네임스페이스 기준) 루트 cgroup에
// 있는지 판별한다. 루트 cgroup.procs의 PID 목록과 비교하는 방식은 host cgroupns
// 에서 PID 네임스페이스가 달라 오탐할 수 있으므로 /proc/self/cgroup을 쓴다.
func inRootCgroup() (bool, error) {
	data, err := readSelfCgroup()
	if err != nil {
		return false, fmt.Errorf("/proc/self/cgroup을 읽을 수 없습니다: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if path, ok := strings.CutPrefix(line, "0::"); ok {
			return strings.TrimSpace(path) == "/", nil
		}
	}
	return false, errors.New("cgroup v2 항목(0::)을 찾을 수 없습니다")
}

// escapeRootIfNeeded는 로컬 Docker(private cgroupns)처럼 서버 프로세스가 루트
// cgroup에 있는 경우에만 "no internal processes" 규칙을 피하기 위해 루트의 모든
// 프로세스를 leaf로 옮기고 루트에 memory 컨트롤러를 위임한다. k8s(privileged,
// host cgroupns)에서는 서버가 kubepods 계층 안에 있고 루트는 systemd가 이미
// 위임해 두었으므로 아무것도 하지 않는다.
func escapeRootIfNeeded(root string) error {
	inRoot, err := inRootCgroup()
	if err != nil {
		return err
	}
	if !inRoot {
		return nil
	}

	leaf := filepath.Join(root, leafName)
	if err := os.Mkdir(leaf, 0o755); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("leaf cgroup 생성 실패: %w", err)
	}

	procs, err := os.ReadFile(filepath.Join(root, "cgroup.procs"))
	if err != nil {
		return fmt.Errorf("루트 cgroup.procs를 읽을 수 없습니다: %w", err)
	}
	for _, pid := range strings.Fields(string(procs)) {
		// 이미 종료된 프로세스는 실패해도 무방하다.
		_ = os.WriteFile(filepath.Join(leaf, "cgroup.procs"), []byte(pid), 0o644)
	}
	// 자기 자신은 반드시 옮겨져야 한다 (누락 시 subtree_control 위임이 실패한다).
	if err := os.WriteFile(filepath.Join(leaf, "cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		return fmt.Errorf("서버 프로세스를 leaf cgroup으로 옮기지 못했습니다: %w", err)
	}

	if err := os.WriteFile(filepath.Join(root, "cgroup.subtree_control"), []byte("+memory"), 0o644); err != nil {
		return fmt.Errorf("루트 memory 컨트롤러 위임 실패: %w", err)
	}
	return nil
}

// cleanupLeftovers는 같은 파드(hostname 동일)의 이전 컨테이너가 남긴 세션
// 디렉터리만 정리한다. 다른 파드의 디렉터리는 같은 노드에서 채점 중인 세션일 수
// 있으므로 절대 건드리지 않는다.
func cleanupLeftovers(podDir string) {
	entries, err := os.ReadDir(podDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(podDir, entry.Name())
		killCgroup(dir)
		if err := removeDirWithRetry(dir); err != nil {
			log.Warn("잔존 cgroup 정리 실패: ", err)
		}
	}
}

// removeDirWithRetry는 cgroup 디렉터리를 재시도하며 삭제한다. cgroup.kill의
// 프로세스 종료가 비동기라 직후의 rmdir은 EBUSY로 실패할 수 있다.
func removeDirWithRetry(dir string) error {
	var err error
	for i := 0; i < 5; i++ {
		err = os.Remove(dir)
		if err == nil || errors.Is(err, os.ErrNotExist) {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}

// killCgroup은 cgroup.kill이 존재하면(실제 cgroupfs) 하위 프로세스를 일괄 종료한다.
func killCgroup(dir string) {
	kill := filepath.Join(dir, "cgroup.kill")
	if _, err := os.Stat(kill); err != nil {
		return
	}
	if err := os.WriteFile(kill, []byte("1"), 0o644); err != nil {
		log.Warn("cgroup.kill 실패: ", err)
	}
}

type fsMonitor struct {
	dir string
}

func (m *fsMonitor) NewSession(name string, memoryLimitKB int) (Session, error) {
	dir := filepath.Join(m.dir, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		return nil, fmt.Errorf("세션 cgroup 생성 실패: %w", err)
	}

	if memoryLimitKB > 0 {
		// 호출자(service)가 문제 제한 + 언어별 마진을 계산해서 넘긴다.
		// 이 값을 초과하면 OOM kill → MEMORY_OUT으로 판정된다.
		limit := int64(memoryLimitKB) * 1024
		if err := os.WriteFile(filepath.Join(dir, "memory.max"), []byte(strconv.FormatInt(limit, 10)), 0o644); err != nil {
			log.Warn("memory.max 설정 실패 (측정은 계속 진행): ", err)
		}
		// 스왑으로 메모리 상한이 무력화되지 않도록 막는다.
		if err := os.WriteFile(filepath.Join(dir, "memory.swap.max"), []byte("0"), 0o644); err != nil {
			log.Warn("memory.swap.max 설정 실패 (측정은 계속 진행): ", err)
		}
	}

	fd, err := os.Open(dir)
	if err != nil {
		_ = os.Remove(dir)
		return nil, fmt.Errorf("세션 cgroup 열기 실패: %w", err)
	}
	return &fsSession{dir: dir, fd: fd}, nil
}

type fsSession struct {
	dir string
	fd  *os.File
}

func (s *fsSession) PeakMemoryKB() (int64, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, "memory.peak"))
	if err != nil {
		return 0, fmt.Errorf("memory.peak 읽기 실패: %w", err)
	}
	return parsePeakKB(data)
}

func (s *fsSession) OOMKilled() (bool, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, "memory.events"))
	if err != nil {
		return false, fmt.Errorf("memory.events 읽기 실패: %w", err)
	}
	return parseOOMKilled(data)
}

func (s *fsSession) Close() error {
	killCgroup(s.dir)
	closeErr := s.fd.Close()
	if err := removeDirWithRetry(s.dir); err != nil {
		return fmt.Errorf("세션 cgroup 삭제 실패: %w", err)
	}
	return closeErr
}

func parsePeakKB(data []byte) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("memory.peak 파싱 실패: %w", err)
	}
	return n / 1024, nil
}

func parseOOMKilled(data []byte) (bool, error) {
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[0] != "oom_kill" {
			continue
		}
		n, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return false, fmt.Errorf("memory.events 파싱 실패: %w", err)
		}
		return n > 0, nil
	}
	return false, nil
}

func hasController(controllers, name string) bool {
	for _, c := range strings.Fields(controllers) {
		if c == name {
			return true
		}
	}
	return false
}

// NewFallbackMonitor는 cgroup을 쓸 수 없는 환경(macOS 컴파일 호환, cgroup 미지원
// 컨테이너)에서 측정 없이(0 반환) 동작하는 Monitor를 반환한다.
func NewFallbackMonitor() Monitor {
	return fallbackMonitor{}
}

// NewFallbackSession은 세션 생성에 실패했을 때 실행 흐름을 유지하기 위한
// no-op 세션을 반환한다.
func NewFallbackSession() Session {
	return fallbackSession{}
}

type fallbackMonitor struct{}

func (fallbackMonitor) NewSession(string, int) (Session, error) {
	return fallbackSession{}, nil
}

type fallbackSession struct{}

func (fallbackSession) Apply(*exec.Cmd) error        { return nil }
func (fallbackSession) PeakMemoryKB() (int64, error) { return 0, nil }
func (fallbackSession) OOMKilled() (bool, error)     { return false, nil }
func (fallbackSession) Close() error                 { return nil }

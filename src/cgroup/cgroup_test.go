package cgroup

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// fakeRoot는 t.TempDir()에 가짜 cgroupfs 루트를 만들고, /proc/self/cgroup 내용을
// 주입해 환경을 흉내 낸다. 기본은 host cgroupns(k8s privileged) 상황이며,
// fakeSelfCgroup으로 로컬 Docker(private cgroupns, "0::/") 상황을 만들 수 있다.
func fakeRoot(t *testing.T, controllers string, procsPIDs []string) string {
	t.Helper()
	fakeSelfCgroup(t, "0::/kubepods.slice/kubepods-pod.slice/crio-abc.scope\n")
	root := t.TempDir()
	writeFakeFile(t, filepath.Join(root, "cgroup.controllers"), controllers)
	writeFakeFile(t, filepath.Join(root, "cgroup.procs"), strings.Join(procsPIDs, "\n"))
	return root
}

func fakeSelfCgroup(t *testing.T, content string) {
	t.Helper()
	original := readSelfCgroup
	readSelfCgroup = func() ([]byte, error) { return []byte(content), nil }
	t.Cleanup(func() { readSelfCgroup = original })
}

func writeFakeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFakeFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func podDirOf(t *testing.T, root string) string {
	t.Helper()
	pod, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(root, baseDirName, pod)
}

func TestParsePeakKB(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    int64
		wantErr bool
	}{
		{name: "정상 값(바이트)을 KB로 변환한다", data: "39324672\n", want: 38403},
		{name: "1024 미만은 0KB", data: "1000", want: 0},
		{name: "0", data: "0\n", want: 0},
		{name: "숫자가 아니면 에러", data: "max\n", wantErr: true},
		{name: "빈 파일이면 에러", data: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePeakKB([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseOOMKilled(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    bool
		wantErr bool
	}{
		{
			name: "oom_kill > 0이면 true",
			data: "low 0\nhigh 0\nmax 12\noom 3\noom_kill 1\noom_group_kill 0\n",
			want: true,
		},
		{
			name: "oom_kill == 0이면 false (oom 카운트만 있으면 kill 아님)",
			data: "low 0\nhigh 0\nmax 0\noom 3\noom_kill 0\n",
			want: false,
		},
		{name: "oom_kill 라인이 없으면 false", data: "low 0\nhigh 0\n", want: false},
		{name: "빈 파일이면 false", data: "", want: false},
		{name: "oom_kill 값이 숫자가 아니면 에러", data: "oom_kill abc\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOOMKilled([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetupAt(t *testing.T) {
	self := strconv.Itoa(os.Getpid())

	t.Run("memory 컨트롤러가 없으면 폴백", func(t *testing.T) {
		root := fakeRoot(t, "cpuset cpu io", nil)
		monitor, err := SetupAt(root)
		if err == nil {
			t.Fatal("에러를 기대했지만 nil")
		}
		if _, ok := monitor.(fallbackMonitor); !ok {
			t.Errorf("폴백 Monitor를 기대했지만 %T", monitor)
		}
	})

	t.Run("cgroupfs가 없으면 폴백 (macOS 등)", func(t *testing.T) {
		monitor, err := SetupAt(filepath.Join(t.TempDir(), "no-such-dir"))
		if err == nil {
			t.Fatal("에러를 기대했지만 nil")
		}
		if _, ok := monitor.(fallbackMonitor); !ok {
			t.Errorf("폴백 Monitor를 기대했지만 %T", monitor)
		}
	})

	t.Run("host cgroupns(k8s): 프로세스 이동 없이 파드 계층만 생성", func(t *testing.T) {
		root := fakeRoot(t, "cpuset cpu io memory", []string{"1", "42"})
		monitor, err := SetupAt(root)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := monitor.(*fsMonitor); !ok {
			t.Fatalf("fsMonitor를 기대했지만 %T", monitor)
		}
		if _, err := os.Stat(podDirOf(t, root)); err != nil {
			t.Errorf("파드 계층 디렉터리가 없음: %v", err)
		}
		if _, err := os.Stat(filepath.Join(root, leafName)); err == nil {
			t.Error("host cgroupns에서는 leaf cgroup을 만들면 안 됨")
		}
		got := readFakeFile(t, filepath.Join(root, baseDirName, "cgroup.subtree_control"))
		if got != "+memory" {
			t.Errorf("base subtree_control = %q, want %q", got, "+memory")
		}
		got = readFakeFile(t, filepath.Join(podDirOf(t, root), "cgroup.subtree_control"))
		if got != "+memory" {
			t.Errorf("pod subtree_control = %q, want %q", got, "+memory")
		}
	})

	t.Run("private cgroupns(로컬 Docker): 루트 프로세스를 leaf로 옮기고 위임", func(t *testing.T) {
		root := fakeRoot(t, "memory pids", []string{self, "99"})
		fakeSelfCgroup(t, "0::/\n")
		if _, err := SetupAt(root); err != nil {
			t.Fatal(err)
		}
		// 가짜 파일시스템에서는 매 쓰기가 덮어쓰므로 마지막 쓰기(자기 PID)가 남는다.
		// 실제 cgroupfs는 쓰기마다 프로세스가 이동한다. 여기서는 쓰기 시도 여부만 검증.
		leafProcs := readFakeFile(t, filepath.Join(root, leafName, "cgroup.procs"))
		if leafProcs != self {
			t.Errorf("leaf cgroup.procs = %q, want %q (자기 PID가 반드시 이동돼야 함)", leafProcs, self)
		}
		got := readFakeFile(t, filepath.Join(root, "cgroup.subtree_control"))
		if got != "+memory" {
			t.Errorf("root subtree_control = %q, want %q", got, "+memory")
		}
	})

	t.Run("시작 시 자기 파드의 잔존 세션만 정리한다", func(t *testing.T) {
		root := fakeRoot(t, "memory", nil)
		stale := filepath.Join(podDirOf(t, root), "stale-session")
		otherPod := filepath.Join(root, baseDirName, "other-pod", "active-session")
		for _, dir := range []string{stale, otherPod} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := SetupAt(root); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(stale); err == nil {
			t.Error("자기 파드의 잔존 세션이 정리되지 않음")
		}
		if _, err := os.Stat(otherPod); err != nil {
			t.Error("다른 파드의 세션을 삭제함 — 절대 건드리면 안 됨")
		}
	})
}

func TestNewSession(t *testing.T) {
	newMonitor := func(t *testing.T) (*fsMonitor, string) {
		root := fakeRoot(t, "memory", nil)
		monitor, err := SetupAt(root)
		if err != nil {
			t.Fatal(err)
		}
		return monitor.(*fsMonitor), root
	}

	t.Run("메모리 제한이 있으면 안전망 상한과 스왑 금지를 기록한다", func(t *testing.T) {
		monitor, root := newMonitor(t)
		session, err := monitor.NewSession("1", 65536) // 64MB
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(podDirOf(t, root), "1")
		wantMax := strconv.FormatInt(int64(65536)*1024+safetyMarginBytes, 10)
		if got := readFakeFile(t, filepath.Join(dir, "memory.max")); got != wantMax {
			t.Errorf("memory.max = %q, want %q", got, wantMax)
		}
		if got := readFakeFile(t, filepath.Join(dir, "memory.swap.max")); got != "0" {
			t.Errorf("memory.swap.max = %q, want %q", got, "0")
		}
		// 가짜 파일시스템에서는 memory.max 등 일반 파일이 남아 rmdir이 실패하므로
		// Close의 삭제 성공 여부는 제한 없는 세션 테스트에서 검증한다.
		_ = session
	})

	t.Run("메모리 제한이 없으면 제한 파일을 만들지 않고, Close가 cgroup을 삭제한다", func(t *testing.T) {
		monitor, root := newMonitor(t)
		session, err := monitor.NewSession("2", 0)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(podDirOf(t, root), "2")
		if _, err := os.Stat(filepath.Join(dir, "memory.max")); err == nil {
			t.Error("제한이 없는데 memory.max가 생성됨")
		}
		if err := session.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(dir); err == nil {
			t.Error("Close 후에도 세션 cgroup이 남아 있음")
		}
	})

	t.Run("PeakMemoryKB와 OOMKilled는 세션 cgroup의 파일을 읽는다", func(t *testing.T) {
		monitor, root := newMonitor(t)
		session, err := monitor.NewSession("3", 0)
		if err != nil {
			t.Fatal(err)
		}
		dir := filepath.Join(podDirOf(t, root), "3")
		writeFakeFile(t, filepath.Join(dir, "memory.peak"), "583168\n") // 569KB
		writeFakeFile(t, filepath.Join(dir, "memory.events"), "oom 1\noom_kill 1\n")

		peak, err := session.PeakMemoryKB()
		if err != nil {
			t.Fatal(err)
		}
		if peak != 569 {
			t.Errorf("peak = %d, want 569", peak)
		}
		oom, err := session.OOMKilled()
		if err != nil {
			t.Fatal(err)
		}
		if !oom {
			t.Error("OOMKilled = false, want true")
		}
	})

	t.Run("같은 이름의 세션은 중복 생성할 수 없다", func(t *testing.T) {
		monitor, _ := newMonitor(t)
		if _, err := monitor.NewSession("dup", 0); err != nil {
			t.Fatal(err)
		}
		if _, err := monitor.NewSession("dup", 0); err == nil {
			t.Error("중복 세션 생성이 성공하면 안 됨")
		}
	})
}

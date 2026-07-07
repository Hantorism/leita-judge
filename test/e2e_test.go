//go:build e2e

package test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"leita/src/entity"
	"leita/src/route"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func startTestServer(t *testing.T) string {
	t.Helper()

	_ = godotenv.Load(".env")

	app := fiber.New()
	if err := route.RegisterRoutes(app); err != nil {
		t.Fatalf("라우트 등록 실패: %v", err)
	}

	portCh := make(chan string, 1)
	app.Hooks().OnListen(func(ld fiber.ListenData) error {
		portCh <- ld.Port
		return nil
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(":0", fiber.ListenConfig{DisableStartupMessage: true})
	}()

	var port string
	select {
	case port = <-portCh:
	case err := <-errCh:
		t.Fatalf("서버 기동 실패: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("서버가 5초 내에 기동되지 않았다")
	}

	t.Cleanup(func() {
		if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
			t.Errorf("서버 종료 실패: %v", err)
		}
	})

	return "http://127.0.0.1:" + port
}

func TestRunProblemE2E(t *testing.T) {
	baseURL := startTestServer(t)

	code := "n = int(input())\nprint(n * 2)\n"
	reqBody := entity.RunProblemRequest{
		Language: "PYTHON",
		Code:     base64.StdEncoding.EncodeToString([]byte(code)),
		Limit:    entity.Limit{Time: 3000, Memory: 65536},
		TestCases: []entity.TestCase{
			{
				Input:  base64.StdEncoding.EncodeToString([]byte("3")),
				Output: base64.StdEncoding.EncodeToString([]byte("6")),
			},
			{
				Input:  base64.StdEncoding.EncodeToString([]byte("5")),
				Output: base64.StdEncoding.EncodeToString([]byte("10")),
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("요청 payload 직렬화 실패: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(baseURL+"/api/problem/run/1", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("요청 실패: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("응답 바디 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d\nbody: %s", resp.StatusCode, http.StatusOK, respBody)
	}

	var results []entity.RunProblemResponse
	if err := json.Unmarshal(respBody, &results); err != nil {
		t.Fatalf("응답 JSON 파싱 실패: %v\nbody: %s", err, respBody)
	}

	if len(results) != len(reqBody.TestCases) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(reqBody.TestCases))
	}

	wantOutputs := []string{
		base64.StdEncoding.EncodeToString([]byte("6")),
		base64.StdEncoding.EncodeToString([]byte("10")),
	}

	for i, result := range results {
		if result.Result != entity.JudgeCorrect.String() {
			t.Errorf("results[%d].Result = %q, want %q (error: %s)", i, result.Result, entity.JudgeCorrect.String(), result.Error)
		}
		if result.Output != wantOutputs[i] {
			t.Errorf("results[%d].Output = %q, want %q", i, result.Output, wantOutputs[i])
		}
	}
}

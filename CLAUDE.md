# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Leita Judge는 온라인 저지(online judge) 시스템의 코드 실행/채점 백엔드다. Go + Fiber v3로 작성되었으며, 제출된 소스 코드를 실제로 빌드/실행하고 테스트케이스와 비교하여 채점 결과를 반환한다. C, C++, Java, Python, JavaScript, Go, Kotlin, Swift, Rust, C#, TypeScript를 지원한다(언어 식별자는 `RUST`/`CS`/`TYPESCRIPT`).

## Commands

```bash
# 의존성 설치
go mod download

# 로컬 실행. .env(선택, 없으면 환경변수로 주입)와 **실행 중인 Redis**가 필요하다 —
# 채점 결과를 Redis Stream으로 발행하므로 연결에 실패하면 기동 자체가 중단된다(main.go Fatal).
# 접속 정보는 REDIS_HOST(기본 localhost)/REDIS_PORT(기본 6379)/REDIS_PASSWORD로 설정한다.
docker run -d --name judge-redis -p 6379:6379 redis:7-alpine
go run .

# 빌드
go build -o server .

# Swagger 문서 재생성 (go.mod에 tool로 선언되어 있음, main.go 상단 주석 기반으로 docs/ 생성)
go tool swag init

# 모든 의존성을 최신 minor/patch 버전으로 업데이트 (./... 없이 go get -u만 실행하면 루트 패키지만 대상이 됨)
go get -u ./...

# 업데이트 후 go.mod/go.sum 정리 (사용하지 않는 의존성 제거, 필요한 것 추가)
go mod tidy

# 테스트 실행
go test ./src/...

# 테스트 실행 (테스트별 상세 결과 출력)
go test -v ./src/...

# 테스트 실행 (커버리지 % 표시)
go test -cover ./src/...

# 유닛 테스트를 Linux(운영과 동일 커널 인터페이스)에서 실행. --privileged를 주면
# 실제 cgroup을 쓰는 통합 테스트(src/executor/integration_linux_test.go)까지 돌고,
# 없으면 통합 테스트는 자동 skip된다.
docker run --rm --privileged -v "$PWD":/workspace -v "$(go env GOMODCACHE)":/go/pkg/mod \
  -w /workspace golang:1.26.2-alpine3.23 go test ./src/...

# 운영(k8s host cgroupns) 경로를 로컬에서 재현하려면 서버를 이렇게 띄운다
# docker run --privileged --cgroupns=host <run 스테이지 이미지> ...

# E2E 테스트 실행 (실제 서버를 로컬에 띄우고 진짜 HTTP 요청으로 검증.
# 로컬에 언어별 컴파일러/런타임 + OCI 자격증명 + 실행 중인 Redis 필요 — 위 로컬 실행과 동일)
go test -tags=e2e ./test/...
```

## 테스트 작성 규칙

"어떤 함수에 어떤 입력을 넣으면 처리를 거쳐서 어떤 출력이 나와야 한다" 형식으로 유닛 테스트 작성을 요청하면, 그 스펙을 바로 Go 테스트 코드로 옮긴다. [src/language/language_test.go](src/language/language_test.go), [src/service/problem/service_test.go](src/service/problem/service_test.go)에 이미 있는 패턴을 따른다:
- 테이블 드리븐 테스트(케이스를 구조체 슬라이스로 나열하고 `t.Run`으로 서브테스트 실행)
- unexported 식별자(`allTrue`, `checkDifference` 등)를 테스트해야 하면 `package problem_test`가 아닌 `package problem`(내부 테스트)으로 작성
- 채점 로직 테스트에서 문제의 테스트케이스 픽스처는 **최소 5개**로 구성한다 (실제 문제가 최소 5개 테스트케이스를 보장하는 정책과 일치시키기 위함)

"서버를 로컬에 띄우고 curl/API 호출로 어떤 요청을 하면 어떤 응답이 나와야 한다" 형식으로 e2e 테스트 작성을 요청하면 [test/e2e_test.go](test/e2e_test.go)의 패턴을 따라 E2E 테스트로 작성한다: `//go:build e2e` 빌드 태그로 일반 테스트와 분리하고, `route.RegisterRoutes`로 실제 서비스(진짜 Executor/파일저장소/OCI/Redis)를 그대로 띄운 뒤 `:0`으로 랜덤 포트를 확보(`OnListen` 훅으로 캡처)해 진짜 `net/http` 클라이언트로 요청한다. `submit` 엔드포인트는 OCI 버킷에 실제 문제 데이터가 있어야 검증 가능하므로, 요청 페이로드만으로 자기완결적인 `run` 엔드포인트 위주로 작성한다. 실행할 때마다 `run/{submitId}/` 디렉터리가 실제로 남고(응답에 submitId가 없어 자동 정리 불가) 이건 정상이니 놀라지 않는다.

## Architecture

레이어드 아키텍처: `route → handler → service → repository → datasource`. 각 레이어는 상위 레이어가 정의한 인터페이스에 의존하며, 각 패키지의 `New*` 생성자가 하위 레이어를 조립해 최상위(`main.go`)까지 전달된다.

```
main.go
  → route.RegisterRoutes            (src/route)
    → handler.NewHandler            (src/handler)
      → service.NewService          (src/service)
        → repository.NewRepository (src/repository)
          → datasource.NewDataSource (src/datasource, OCI Object Storage 클라이언트)
```

- **entity** ([src/entity/problem.go](src/entity/problem.go)) — 요청/응답 DTO와 `JudgeResultEnum`(CORRECT/WRONG/COMPILE_ERROR/RUNTIME_ERROR/MEMORY_OUT/TIME_OUT) 등 도메인 타입 전체가 여기 모여 있다.
- **language** ([src/language/language.go](src/language/language.go)) — 언어별 빌드/실행/삭제 커맨드를 `{JUDGE_TYPE}`/`{SUBMIT_ID}` 플레이스홀더가 있는 템플릿으로 정의(`Commands` 맵). `ReplaceCommand`로 실제 경로를 채워 넣는다. 소스 파일명은 항상 `Main.{ext}`(`FileName` 상수)로 고정된다. 언어별 **시간/메모리 버퍼**도 여기서 정의한다(`limitBuffers` 테이블 + `TimeLimitWithBuffer`/`MemoryLimitWithBuffer`). 런타임 고정비를 보상해 언어 간 형평성을 맞추는 값으로, 버퍼 이내의 초과는 정책상 허용이다.

| 언어 | 시간 | 메모리 |
|---|---|---|
| C, CPP | 보정 없음 | 보정 없음 |
| SWIFT, RUST | 보정 없음 | +16MB |
| GO | +2초 | +32MB |
| PYTHON, JAVASCRIPT, TYPESCRIPT | ×3 + 2초 | +32MB |
| CS | ×2 + 1초 | +64MB |
| JAVA, KOTLIN | ×2 + 1초 | +128MB |

**시간은 배수+가산, 메모리는 고정 가산**인 이유: 인터프리터의 느림은 작업량에 비례하지만, 메모리 오버헤드는 문제 제한과 무관한 상수(런타임 고정비 + GC 여유)다. C/C++에 보정을 주지 않는 것은 문제의 제한 자체가 C/C++ 기준으로 설계되기 때문이며, 마진을 주면 메모리 최적화 문제의 출제 의도가 깨진다. 시간 보정값은 BOJ 정책을, 메모리 보정값은 실측 오버헤드(C/C++ 0.3MB, Rust 1.0MB, Swift 1.8MB, Python 3.9MB, C# 5.7MB, Go 6.1MB, JS/TS 9.2~9.4MB, JVM 39.2MB)를 근거로 한다.

**JVM 계열 실행 커맨드에는 `-Xms`를 주지 않는다.** 초기 힙을 크게 잡으면 JVM이 힙 예산이 넉넉하다고 판단해 GC를 미루다가 cgroup 상한에 먼저 부딪혀, 정상 코드가 `MEMORY_OUT`으로 오판정된다(실측: 동일 GC 부하 워크로드가 `-Xms1024m`에서는 224MB를 요구했고 제거 후 96MB로 줄었다). 반대로 `-Xmx`는 문제 제한보다 크게 유지해야 힙 한계 대신 cgroup OOM이 먼저 발생해 판정이 일관된다 — 줄이면 `OutOfMemoryError`가 나 `RUNTIME_ERROR`로 샌다.
- **cgroup** ([src/cgroup/cgroup.go](src/cgroup/cgroup.go)) — cgroup v2 기반 메모리 측정 계층. 테스트케이스 실행마다 `/sys/fs/cgroup/leita-judge/{pod}/{seq}`에 일회용 cgroup을 만들고, 채점 프로세스를 clone3(`CLONE_INTO_CGROUP`)로 그 안에서 시작시킨 뒤 종료 후 `memory.peak`을 읽는다(KB 단위). `{pod}`는 `os.Hostname()`(k8s 파드명/Docker 컨테이너 ID)으로, 같은 노드의 여러 judge 파드 간 격리 계층이다. `Setup()`이 시작 시 환경을 감지한다: `/proc/self/cgroup`이 `0::/`이면(로컬 Docker, private cgroupns) 루트 프로세스를 `main/` leaf로 옮겨 "no internal processes" 규칙을 회피하고, 아니면(k8s privileged + host cgroupns) 파드 계층만 만든다. 시작 시 자기 파드의 잔존 세션만 정리하며 다른 파드 디렉터리는 절대 건드리지 않는다. cgroup을 쓸 수 없으면(비-privileged 컨테이너, macOS) 경고 로그 후 측정값 0을 반환하는 폴백으로 기동한다. **측정 범위는 코드 실행(stdin~stdout)만이며 Build/Delete는 cgroup 밖에서 실행되어 절대 포함되지 않는다.**
- **executor** ([src/executor/executor.go](src/executor/executor.go)) — `Executor` 인터페이스(Build/Run/Delete)의 실제 구현인 `OsExecutor`가 `os/exec`로 컴파일러/런타임 프로세스를 직접 구동한다. `Run`은 `context.WithTimeout`으로 제한 시간을 강제하고, cgroup 세션으로 메모리를 측정한다. 시간·메모리 제한 모두 service가 **언어별 버퍼를 적용한 값**(`language.TimeLimitWithBuffer`/`MemoryLimitWithBuffer`)을 넘겨주며, 메모리 초과 시 OOM kill로 `MEMORY_OUT`이 된다. 응답의 `usedTime`/`usedMemory`는 버퍼와 무관한 실측값이다 — 버퍼는 판정에만 쓰인다. 판정 우선순위는 **OOM(`JudgeMemoryOut`) → 타임아웃(`JudgeTimeOut`) → 비정상 종료(`JudgeRuntimeError`)** 순서다 — OOM kill은 SIGKILL이라 순서를 바꾸면 메모리 초과가 RUNTIME_ERROR나 TIME_OUT으로 오분류된다.
- **repository/file** ([src/repository/file/repository.go](src/repository/file/repository.go)) — 로컬 파일시스템에 소스 코드/테스트케이스를 저장하는 `Repository` 인터페이스(`LocalRepository` 구현). 경로 규칙은 `{judgeType}/{submitId}/Main.{ext}`, `{judgeType}/{submitId}/in/{i}.in`, `{judgeType}/{submitId}/out/{i}.out`.
- **datasource/redis** ([src/datasource/redis.go](src/datasource/redis.go)) — 채점 결과 발행용 Redis 클라이언트. `REDIS_HOST`(기본 `localhost`)/`REDIS_PORT`(기본 `6379`)/`REDIS_PASSWORD`로 설정하며, **기동 시 연결에 실패하면 서버가 뜨지 않는다**(`main.go`에서 Fatal).
- **repository/problem** + **datasource** ([src/datasource/objectstorage.go](src/datasource/objectstorage.go)) — OCI Object Storage를 감싸는 계층. 문제의 정답 테스트케이스를 `problems/{problemId}/testcases`에서 읽어오고, 제출된 코드를 `submits/{submitId}/...`에 base64로 저장한다. **정책: 모든 문제는 최소 5개의 테스트케이스를 보장한다** — 문제 퀄리티를 위한 정책으로 문제 데이터 등록 단계에서 강제되며, 이 레포 코드에는 별도 검증 로직이 없다.
- **service/problem** ([src/service/problem/service.go](src/service/problem/service.go)) — 채점 핵심 로직. 저장 → 빌드 → (테스트케이스별) 실행 → 출력 바이트 비교(`bytes.Equal`) → 결과 집계 순으로 진행하고, 빌드/삭제 실패 시에도 실행 파일 삭제(`Delete`)가 `defer`로 항상 시도된다.

### Judge 타입: `submit` vs `run`

같은 `Service`가 두 가지 판정 흐름을 처리하며, `judgeType` 문자열("submit"/"run")로 파일 경로와 동작이 갈린다.

- **submit** (`SubmitProblem`) — **비동기다.** 핸들러는 고루틴으로 채점을 시작하고 즉시 `{"status":"RECEIVED"}`를 응답하며, 결과는 채점이 끝난 뒤 `PublishJudgeResult`가 Redis Stream(`REDIS_STREAM_KEY`, 기본 `judge-result-stream`)으로 발행한다. 동시 채점은 `judgeSemaphore`(현재 4개)로 제한하고, 고루틴 패닉은 recover해 `UNKNOWN` 결과를 발행하므로 요청이 유실되지 않는다. `submitId`는 요청에서 받은 실제 제출 ID. 테스트케이스는 OCI Object Storage의 문제 데이터에서 가져오고, 채점 후 제출 코드를 Object Storage에 base64로 영구 저장한다. 평균 사용 시간·사용 메모리는 첫 번째 테스트케이스(워밍업으로 간주 — 시간은 JIT/캐시 예열, 메모리는 페이지 캐시 첫 적재 비용)를 제외하고 계산한다(`judgeSubmit`의 `averageExcludingWarmup`).
- **run** (`RunProblem`) — 코드 테스트/실행 목적. `submitId`는 매 요청마다 12자리 난수로 생성하고, 테스트케이스는 요청 바디에 실려온 것을 그대로 사용한다. Object Storage에 결과를 저장하지 않고, 각 테스트케이스의 실제 출력(base64)까지 응답에 포함한다.

두 경우 모두 로컬 디렉터리 `submit/{id}/`, `run/{id}/`가 작업 공간으로 쓰이며(`.gitignore`에 포함되어 커밋되지 않음), 채점 완료 후 빌드 산출물만 `Delete` 커맨드로 정리되고 소스/입출력 파일은 남는다.

### 배포 구조

`deploy/Dockerfile-{language}`가 언어별로 하나씩 존재한다. 모든 이미지가 동일한 Go 서버 바이너리를 빌드하지만, 최종 런타임 스테이지 베이스 이미지만 언어별 컴파일러/런타임(gcc, jdk, node, python 등)으로 다르게 지정된다 — 즉 언어별로 별도 컨테이너를 띄워 해당 언어의 코드만 처리하는 구조.

런타임 설치만으로 끝나지 않는 언어가 둘 있다. **C#**(`Dockerfile-cs`)은 프로젝트 파일이 있어야 빌드되므로 `/template/Main.csproj`를 이미지에 구워 두고(NuGet 캐시도 함께 워밍) 제출마다 복사해 쓴다. **TypeScript**(`Dockerfile-typescript`)는 `tsc`와 함께 **`@types/node`를 전역 설치해야 한다** — 타입 정의가 없으면 `require`/`process`를 쓰는 정상 코드가 타입 에러로 잡혀 COMPILE_ERROR가 된다. 두 경로 모두 `language.CsprojTemplatePath`/`language.NodeTypeRoot` 상수와 이미지 내 실제 위치가 일치해야 한다.

**언어를 추가할 때 손볼 곳**: `src/language/language.go`의 `Commands`·`FileExtension`·`limitBuffers` 세 곳(빠뜨리면 `TestEverySupportedLanguageIsFullyConfigured`가 실패한다), `deploy/Dockerfile-{language}`, 그리고 레포 밖 GitOps의 Deployment(`privileged: true` 필수)·Service·Tekton 파이프라인의 `dockerfile-paths`/`deployment-files` 배열이다. 마지막 배열과 실제 파일명이 어긋나면 judge CI 전체가 실패한다.

메모리 측정은 컨테이너 안에서 `/sys/fs/cgroup`이 rw여야 동작한다. k8s에서는 judge Deployment에 `securityContext.privileged: true`가 필요하며(GitOps 레포에서 관리), 없으면 서버는 정상 동작하되 `usedMemory`가 항상 0이다(폴백 모드). 노드 요구사항: cgroup v2 + 커널 5.19+(`memory.peak`).

### API

Fiber 앱은 `/api/problem/submit/:problemId`, `/api/problem/run/:problemId` (둘 다 POST)를 노출한다. **`submit`은 접수 응답만 즉시 돌려주고 결과는 Redis Stream으로 발행**하는 반면, `run`은 채점이 끝날 때까지 기다렸다가 케이스별 결과를 응답 본문에 담아 반환한다. 또한 `/swagger.json` 및 `/api/swagger/*`로 Swagger UI를 제공한다(`docs/`는 `go tool swag init`으로 생성되는 파일이므로 수동 편집 금지).

## Commit Message Convention

커밋 메시지는 아래 타입 중 하나로 시작한다(README.md 기준):

| Type | Description |
|------|------|
| feat | 새로운 기능 추가 |
| fix | 버그 수정 |
| refactor | 코드의 로직 변경, 기능 개선 |
| delete | 코드나 파일 삭제 |
| move | 파일 이동, 파일명/디렉터리명 변경 |
| test | 테스트/벤치마크 |
| style | 변수명/함수명 변경, 코드 포맷 변경 |
| design | UI 디자인 변경 |
| deploy | 배포 관련 작업 |
| docs | 문서 관련 작업 |
| chore | 기타 |

형식은 다음과 같다:

```
type: 1줄 요약

상세 설명
```

- 첫 줄은 `type: 요약` 형태로, 변경 사항을 간결하게 한 줄로 요약한다.
- 한 줄 비우고, 본문에 변경 이유와 상세 내용을 작성한다.
- 커밋 메시지는 영어로 작성한다.

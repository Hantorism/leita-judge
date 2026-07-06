# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Leita Judge는 온라인 저지(online judge) 시스템의 코드 실행/채점 백엔드다. Go + Fiber v3로 작성되었으며, 제출된 소스 코드를 실제로 빌드/실행하고 테스트케이스와 비교하여 채점 결과를 반환한다. C, C++, Java, Python, JavaScript, Go, Kotlin, Swift를 지원한다.

## Commands

```bash
# 의존성 설치
go mod download

# 로컬 실행 (.env 파일 필요)
go run .

# 빌드
go build -o server .

# Swagger 문서 재생성 (go.mod에 tool로 선언되어 있음, main.go 상단 주석 기반으로 docs/ 생성)
go tool swag init

# 모든 의존성을 최신 minor/patch 버전으로 업데이트 (./... 없이 go get -u만 실행하면 루트 패키지만 대상이 됨)
go get -u ./...

# 업데이트 후 go.mod/go.sum 정리 (사용하지 않는 의존성 제거, 필요한 것 추가)
go mod tidy
```

테스트 코드(`*_test.go`)는 현재 저장소에 존재하지 않는다.

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
- **language** ([src/language/language.go](src/language/language.go)) — 언어별 빌드/실행/삭제 커맨드를 `{JUDGE_TYPE}`/`{SUBMIT_ID}` 플레이스홀더가 있는 템플릿으로 정의(`Commands` 맵). `ReplaceCommand`로 실제 경로를 채워 넣는다. 소스 파일명은 항상 `Main.{ext}`(`FileName` 상수)로 고정된다.
- **executor** ([src/executor/executor.go](src/executor/executor.go)) — `Executor` 인터페이스(Build/Run/Delete)의 실제 구현인 `OsExecutor`가 `os/exec`로 컴파일러/런타임 프로세스를 직접 구동한다. `Run`은 `context.WithTimeout`으로 제한 시간을 강제하며 타임아웃은 `JudgeTimeOut`, 비정상 종료는 `JudgeRuntimeError`로 매핑된다.
- **repository/file** ([src/repository/file/repository.go](src/repository/file/repository.go)) — 로컬 파일시스템에 소스 코드/테스트케이스를 저장하는 `Repository` 인터페이스(`LocalRepository` 구현). 경로 규칙은 `{judgeType}/{submitId}/Main.{ext}`, `{judgeType}/{submitId}/in/{i}.in`, `{judgeType}/{submitId}/out/{i}.out`.
- **repository/problem** + **datasource** ([src/datasource/objectstorage.go](src/datasource/objectstorage.go)) — OCI Object Storage를 감싸는 계층. 문제의 정답 테스트케이스를 `problems/{problemId}/testcases`에서 읽어오고, 제출된 코드를 `submits/{submitId}/...`에 base64로 저장한다.
- **service/problem** ([src/service/problem/service.go](src/service/problem/service.go)) — 채점 핵심 로직. 저장 → 빌드 → (테스트케이스별) 실행 → 출력 바이트 비교(`bytes.Equal`) → 결과 집계 순으로 진행하고, 빌드/삭제 실패 시에도 실행 파일 삭제(`Delete`)가 `defer`로 항상 시도된다.

### Judge 타입: `submit` vs `run`

같은 `Service`가 두 가지 판정 흐름을 처리하며, `judgeType` 문자열("submit"/"run")로 파일 경로와 동작이 갈린다.

- **submit** (`SubmitProblem`) — `submitId`는 요청에서 받은 실제 제출 ID. 테스트케이스는 OCI Object Storage의 문제 데이터에서 가져오고, 채점 후 제출 코드를 Object Storage에 base64로 영구 저장한다. 평균 사용 시간은 첫 번째 테스트케이스(워밍업으로 간주)를 제외하고 계산한다(`judgeSubmit`).
- **run** (`RunProblem`) — 코드 테스트/실행 목적. `submitId`는 매 요청마다 12자리 난수로 생성하고, 테스트케이스는 요청 바디에 실려온 것을 그대로 사용한다. Object Storage에 결과를 저장하지 않고, 각 테스트케이스의 실제 출력(base64)까지 응답에 포함한다.

두 경우 모두 로컬 디렉터리 `submit/{id}/`, `run/{id}/`가 작업 공간으로 쓰이며(`.gitignore`에 포함되어 커밋되지 않음), 채점 완료 후 빌드 산출물만 `Delete` 커맨드로 정리되고 소스/입출력 파일은 남는다.

### 배포 구조

`deploy/Dockerfile-{language}`가 언어별로 하나씩 존재한다. 모든 이미지가 동일한 Go 서버 바이너리를 빌드하지만, 최종 런타임 스테이지 베이스 이미지만 언어별 컴파일러/런타임(gcc, jdk, node, python 등)으로 다르게 지정된다 — 즉 언어별로 별도 컨테이너를 띄워 해당 언어의 코드만 처리하는 구조. Swift는 현재 Dockerfile 전체가 주석 처리되어 비활성화 상태다.

### API

Fiber 앱은 `/api/problem/submit/:problemId`, `/api/problem/run/:problemId` (둘 다 POST)를 노출하며, `/swagger.json` 및 `/api/swagger/*`로 Swagger UI를 제공한다(`docs/`는 `go tool swag init`으로 생성되는 파일이므로 수동 편집 금지).

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

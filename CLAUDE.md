# arcusctl

[Arcus](https://github.com/naver/arcus) 캐시 클러스터를 배포·운영하는 Go CLI 도구. ZooKeeper 앙상블과 Arcus 클러스터를 YAML
토폴로지로 선언하고, SSH/SCP로 원격 호스트에 배포한다.

## 빌드 / 검증

```sh
go build ./...
go vet ./...
go test ./...
golangci-lint run          # .golangci.yml, CI는 v2.12.1 사용
```

Go 1.25.0. CI는 `.github/workflows/ci.yml`에서 lint → test → build 순으로 돈다.

## 명령어 체계

```
arcusctl <resource> <command> [<name>] [flags]
```

| resource  | 커맨드                                                                                     |
|-----------|--------------------------------------------------------------------------------------------|
| `zk`      | `deploy <version> <topology.yml>`, `start`/`stop`/`delete <name>`, `list`, `status <name>` |
| `cluster` | 위와 동일 (`<name>`은 servicecode)                                                         |
| `acl`     | `admin`, `user`, `group` 하위 커맨드 (ZooKeeper SASL/SCRAM ACL 관리)                       |

전역 플래그: `--verbose/-v`, `--config-file`.

## 패키지 구조

```
main.go              cmd.Execute() 호출만
cmd/                 cobra 커맨드 정의 — 얇게 유지. 인자 파싱과 검증만.
  root.go            rootCmd, 전역 플래그, cobra.OnInitialize(internal.InitConfig)
  zk/ cluster/ acl/  리소스별 서브커맨드
internal/            모든 실제 로직. cmd/에서 로직을 구현하지 않는다.
  config.go          viper 설정 로드 (Flags, Config 전역)
  const.go           ZooKeeper znode 경로 상수
  util.go            ZK 커넥션/znode 헬퍼, 표준입력 읽기
  topology/          YAML 토폴로지 타입 + Validate() + Load*()
  store/             ~/.arcusctl 하위 배포 메타데이터 영속화
  ssh/               ssh/scp 서브프로세스 실행 래퍼
  zk/                ZooKeeper 앙상블 배포·기동·상태
  cluster/           Arcus 클러스터 배포·기동·상태
  scram/             SCRAM 자격증명 생성
```

새 기능은 `internal/<도메인>/`에 구현하고 `cmd/`에서는 인자를 넘겨 호출만 한다.

## 설계 원칙

- **선언형 배포** — 형상은 YAML로 정의하고, 버전은 명령 인자로 받는다.
- **이름 기반 조작** — deploy 이후에는 ensemble-name / servicecode로 대상을 식별한다.
- **실패 시 롤백하지 않는다** — 여러 호스트에 순차 적용하다 중간에 실패하면 롤백하지 않고 즉시 중단한다. 대신 **실패 단계 / 원인 / 수동 복구 방법**을 출력하고
  종료한다. 이것은 확정된 결정이며 PR마다 재논의하지 않는다.
- **파괴적 동작은 확인을 받는다** — delete/stop 계열은 `internal.Confirm` 또는
  `--force`를 거친다.
- **plan 후 확인** — deploy는 적용 전에 대상 호스트·포트·디렉터리 표를 출력하고
  `[y/N]` 확인을 받는다.

## 배포 메타데이터

```
~/.arcusctl/clusters/
  zk/<ensemble-name>/{meta.yml, topology.yml}
  arcus/<servicecode>/{meta.yml, topology.yml}
```

`internal/store`가 관리한다. deploy 시 토폴로지 원본을 그대로 저장해 두고, 이후 커맨드는 이름으로 이걸 읽어 동작한다.

## 코딩 규약

[Effective Go](https://go.dev/doc/effective_go)를 기준으로 한다. 별도의 팀 컨벤션은 두지 않는다.

### 표기

제품명은 **Arcus**로 쓴다. `ARCUS`(전부 대문자)로 쓰지 않는다. 커맨드·패키지·모듈 이름은 소문자 `arcus` / `arcusctl`을 유지한다.

### 에러 처리

- `internal/`의 모든 함수는 **error를 반환한다.** panic하지 않는다.
- 에러를 감쌀 때는 `fmt.Errorf("...: %w", err)`로 원인을 보존한다.
- 에러를 `_`로 버리지 않는다. 무시가 의도라면 이유를 주석으로 남긴다.
- 사용자에게 보이는 에러 메시지는 **무엇이 / 왜 실패했고 / 어떻게 복구하는지**를 담는다. 원격 호스트가 관련되면 어느 호스트인지 반드시 포함한다.
- **cobra 커맨드는 `RunE`로 error를 반환한다.** `cmd.Execute()`가 stderr 출력과 exit code 1을 담당한다. 커맨드 안에서
  `panic()`이나 `os.Exit()`을 부르지 않는다.
  > 현재 `cmd/` 일부와 `internal/config.go`, `internal/util.go`에 `panic(err)`이
  > 남아 있다. 이것은 **정리 대상인 과거 패턴**이며 새 코드가 따라갈 기준이 아니다.
  > 건드리는 김에 점진적으로 `RunE` + error 반환으로 옮긴다.

### 원격 실행 (internal/ssh)

- `ssh.Run`이 넘기는 커맨드 문자열은 **원격 셸이 해석한다.** 경로·이름 등 사용자 입력을 그대로 끼워 넣지 말고 인용/이스케이프한다.
- 모든 원격 호출에는 타임아웃이 걸려 있어야 한다 (`ConnectTimeout`).
- `exec.ExitError`의 exit code를 해석할 때 "명령 실패"와 "연결 실패"를 구분한다.

### ZooKeeper

- `zk.Conn`은 반드시 `defer conn.Close()`로 닫는다.
- znode 생성·삭제는 멱등해야 한다 (`zk.ErrNodeExists`, `zk.ErrNoNode`를 성공으로 취급).
- 재귀 헬퍼 (`EnsureZNode`, `DeleteZNode`)를 수정할 때 부모 경로를 넘기는지 확인한다.

### 테스트

- 순수 로직 (토폴로지 파싱·검증, 주소 파싱, 계산)은 테이블 기반 테스트를 붙인다.
- SSH나 실제 ZooKeeper가 필요한 코드에는 단위 테스트를 요구하지 않는다.
- 기존 예시: `internal/util_test.go`, `internal/scram/scram_test.go`.

## 문서

`docs/` 아래에 커맨드별 가이드가 있다. 사용자에게 보이는 플래그나 YAML 스키마를 바꾸면 해당 문서와 `zk-sample-topology.yml` /
`cluster-sample-topology.yml`도 같이 갱신한다.

## 코드 리뷰

PR 리뷰 규칙은 [`.github/REVIEW.md`](.github/REVIEW.md)에 있다.
`.github/workflows/claude-code-review.yml`이 PR마다 이 파일을 읽어 리뷰한다. 리뷰 기준을 바꾸고 싶으면 워크플로가 아니라 `REVIEW.md`
를 수정한다.

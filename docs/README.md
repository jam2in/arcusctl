# Arcusctl 문서

`arcusctl`은 Arcus 캐시 클러스터를 선언형으로 배포하고 운영하기 위한 Go 기반 CLI 도구입니다. 운영자는 원하는 ZooKeeper Ensemble과 Arcus Memcached 서버 구성을 명령으로 선언하고, `arcusctl`이 ZooKeeper의 Arcus znode 구조와 서버 목록을 갱신하도록 사용할 수 있습니다.

## 먼저 보는 사용 흐름

### 1. 설치 확인

명령 형식:
```sh
arcusctl version
```

실행 예시:
```sh
./arcusctl version
```

### 2. ZooKeeper 연결 설정

명령 형식:
```sh
arcusctl --config-file <config.yaml> <command>
```

실행 예시:
```sh
./arcusctl --config-file ./config.yaml acl group list
```

`arcusctl`은 실행 전에 ZooKeeper 접속 정보를 읽습니다. 설정 파일 탐색 순서, `--config-file`, `ARCUSCTL_` 환경 변수 사용법은 [설정 가이드](./config-file.md)를 먼저 확인하세요.

### 3. 명령 도움말 확인

명령 형식:
```sh
arcusctl <command> --help
```

실행 예시:
```sh
./arcusctl connect --help
```

## 문서 목록

| 문서 | 설명 |
| --- | --- |
| [설정 가이드](./config-file.md) | ZooKeeper 연결 설정, 설정 파일 탐색 순서, 환경 변수 재정의 방법 |
| [Memcached 서버 관리 가이드](./command-memcached.md) | 서비스 코드별 서버 등록, 설정, 조회, 시작/중지 흐름 |
| [ACL 사용 가이드](./command-acl.md) | Arcus ACL 그룹과 사용자 관리 명령 |

## 권장 읽기 순서

1. 루트 [`README.md`](../README.md)에서 설치 방법과 기본 실행 방식을 확인합니다.
2. [설정 가이드](./config-file.md)에서 ZooKeeper 접속 정보를 준비합니다.
3. [Memcached 서버 관리 가이드](./command-memcached.md)에서 서비스 코드와 서버 관리 흐름을 확인합니다.
4. 필요한 경우 명령별 도움말을 함께 확인합니다.

## 문서 작성 원칙

- 실제 명령 형식을 먼저 제시하고 바로 아래에 실행 예시를 둡니다.
- 설정 파일, 환경 변수, 명령 이름은 코드 스타일로 표기합니다.
- 운영 시 주의가 필요한 제약 조건은 별도 주의 사항으로 분리합니다.
- 새 명령 문서는 `docs/command-<name>.md` 형식으로 추가하여 확장하기 쉽게 유지합니다.

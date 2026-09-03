# arcusctl

`arcusctl`은 [Arcus](https://github.com/naver/arcus) 클러스터와 ZooKeeper 앙상블을 배포하고 관리하기 위한 CLI 도구입니다.

## 설치

### GitHub Releases

[GitHub Releases](https://github.com/jam2in/arcusctl/releases)에서 운영체제와 아키텍처에 맞는 압축 파일을 내려받아 설치할 수 있습니다.

```sh
curl -LO https://github.com/jam2in/arcusctl/releases/download/v<VERSION>/arcusctl-<VERSION>.<OS>-<ARCH>.tar.gz
tar zxvf arcusctl-<VERSION>.<OS>-<ARCH>.tar.gz
cd arcusctl-<VERSION>.<OS>-<ARCH>
./arcusctl version
```

### Go

Go toolchain이 설치된 환경에서는 `go install`을 사용할 수 있습니다.

```sh
go install github.com/jam2in/arcusctl@latest
arcusctl version
```

### Docker

Docker 이미지로 설치 없이 실행할 수도 있습니다.

```sh
docker run --rm jam2in/arcusctl:latest version
```

## 기본 사용

사용 가능한 명령과 옵션은 `-h` 또는 `--help`로 확인할 수 있습니다.

```sh
arcusctl -h
```

Bash에서는 다음과 같이 명령어 자동 완성을 적용할 수 있습니다.

```sh
source <(arcusctl completion bash)
```

## 문서

처음 사용하는 경우 시작하기 문서에서 실행 환경과 기본 운영 흐름을 먼저 확인하세요.

| 문서                                                  | 내용                                         |
|-------------------------------------------------------|----------------------------------------------|
| [시작하기](docs/0-getting-started.md)                 | 실행 환경, 상태 저장 방식 및 기본 운영 흐름  |
| [ZooKeeper 운영 가이드](docs/1-zk-guide.md)           | ZooKeeper 앙상블 배포 및 관리                |
| [Arcus 클러스터 운영 가이드](docs/2-cluster-guide.md) | Community 및 Enterprise 클러스터 배포와 관리 |
| [Arcus ACL 운영 가이드](docs/3-arcus-acl-guide.md)    | ACL 그룹과 사용자 관리                       |

## 예시 파일

설정 및 토폴로지 파일은 다음 예시를 복사한 뒤 환경에 맞게 수정할 수 있습니다.

- [설정 파일](examples/config.yaml)
- [ZooKeeper 토폴로지](examples/zk-topology.yaml)
- [Community 클러스터 토폴로지](examples/cluster-community-topology.yaml)
- [Enterprise 클러스터 토폴로지](examples/cluster-enterprise-topology.yaml)

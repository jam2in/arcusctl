# arcusctl

`arcusctl`은 [ARCUS](https://github.com/naver/arcus)의 캐시 클러스터를 선언형으로 배포·관리하기 위한 Go 기반 CLI 도구입니다. ZooKeeper Ensemble 접속 정보를 기반으로 Arcus znode 구조와 Memcached 서버 구성을 관리합니다.

## Getting started

### Install

아래와 같은 방법으로 `arcusctl`을 설치할 수 있습니다.

- [Release Page](https://github.com/jam2in/arcusctl/releases)에서 환경에 맞는 pre-built binary를 다운로드합니다.

  명령 형식:
  ```sh
  curl -LO https://github.com/jam2in/arcusctl/releases/download/v<VERSION>/arcusctl-<VERSION>.<OS>-<ARCH>.tar.gz
  tar zxvf arcusctl-<VERSION>.<OS>-<ARCH>.tar.gz
  cd arcusctl-<VERSION>.<OS>-<ARCH>
  ./arcusctl version
  ```

  실행 예시:
  ```sh
  curl -LO https://github.com/jam2in/arcusctl/releases/download/v0.1.0/arcusctl-0.1.0.linux-amd64.tar.gz
  tar zxvf arcusctl-0.1.0.linux-amd64.tar.gz
  cd arcusctl-0.1.0.linux-amd64
  ./arcusctl version
  ```

- Go toolchain이 설치된 환경에서는 `go install`을 사용할 수 있습니다.

  명령 형식:
  ```sh
  go install github.com/jam2in/arcusctl@latest
  arcusctl version
  ```

  실행 예시:
  ```sh
  go install github.com/jam2in/arcusctl@latest
  arcusctl version
  ```

- Docker image를 사용하여 실행할 수 있습니다.

  명령 형식:
  ```sh
  docker run --rm jam2in/arcusctl:latest <command>
  ```

  실행 예시:
  ```sh
  docker run --rm jam2in/arcusctl:latest version
  ```

### Configure ZooKeeper

`arcusctl`은 ZooKeeper에 저장된 Arcus 클러스터 정보를 사용하므로 명령 실행 전에 ZooKeeper 접속 정보를 설정해야 합니다.

설정 파일 형식:
```yaml
zookeeper: "<host1:port1>,<host2:port2>,<host3:port3>"
```

실행 예시:
```yaml
zookeeper: "10.0.0.1:2181,10.0.0.2:2181,10.0.0.3:2181"
```

설정 파일을 직접 지정하는 명령 형식:
```sh
arcusctl --config-file <config.yaml> <command>
```

실행 예시:
```sh
./arcusctl --config-file ./config.yaml acl group list
```

자세한 설정 파일 탐색 순서와 `ARCUSCTL_` 환경 변수 사용법은 [설정 가이드](docs/config-file.md)를 참고하세요.

### Usage

- help 옵션으로 사용 가능한 명령과 사용 방법을 확인할 수 있습니다.

  명령 형식:
  ```sh
  arcusctl --help
  arcusctl <command> --help
  ```

  실행 예시:
  ```sh
  ./arcusctl --help
  ./arcusctl connect --help
  ```

- 아래와 같이 shell completion을 적용할 수 있습니다.

  명령 형식:
  ```sh
  source <(arcusctl completion <shell>)
  ```

  실행 예시:
  ```sh
  source <(./arcusctl completion bash)
  ```

- 명령별 가이드는 [문서 안내](docs/README.md)에서 확인할 수 있습니다.

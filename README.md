# arcusctl

`arcusctl`은 [ARCUS](https://github.com/naver/arcus)의 운영을 위해 필요한 기능을 제공하는 CLI 도구입니다.

## Getting started

### Install

아래와 같은 방법으로 arcusctl 설치할 수 있습니다.

- [Release Page](https://github.com/jam2in/arcusctl/releases)에서 맞는 pre-built binary 다운로드할 수 있습니다.
  ```sh
  curl -LO https://github.com/jam2in/arcusctl/releases/download/v<VERSION>/arcusctl-<VERSION>.<OS>-<ARCH>.tar.gz
  tar zxvf arcusctl-<VERSION>.<OS>-<ARCH>.tar.gz
  cd arcusctl-<VERSION>.<OS>-<ARCH>
  ./arcusctl version
  ```

- Go toolchain이 설치된 환경에서는 `go install` 활용할 수 있습니다.
  ```sh
  go install github.com/jam2in/arcusctl@latest
  arcusctl version
  ```

- 또는 docker image 활용하여 실행 가능합니다.
  ```sh
  docker run --rm jam2in/arcusctl:latest version
  ```

### Usage

- help 옵션으로 arcusctl에서 사용 가능한 명령과 사용 방법을 확인할 수 있습니다.
  ```sh
  ./arcusctl -h
  ```

- 아래와 같이 completion 적용할 수 있습니다.
  ```sh
  source <(./arcusctl completion bash)
  ```

- 그 외 각 명령에 대한 가이드는 [문서](docs)에서 확인할 수 있습니다.


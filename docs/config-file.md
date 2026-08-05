# 설정 가이드

`arcusctl`은 ZooKeeper Ensemble에 연결하여 Arcus 클러스터 정보를 읽고 갱신합니다. 따라서 설치 후 가장 먼저 ZooKeeper 접속 정보를 설정해야 합니다.

## 1. 설정 파일 작성

설정 파일 형식:
```yaml
zookeeper: "<host1:port1>,<host2:port2>,<host3:port3>"
```

작성 예시:
```yaml
zookeeper: "10.0.0.1:2181,10.0.0.2:2181,10.0.0.3:2181"
# 도메인 주소도 사용할 수 있습니다.
# zookeeper: "zookeeper.example.com:2181"
```

## 2. 설정 파일 지정

명령 형식:
```sh
arcusctl --config-file <config.yaml> <command>
```

실행 예시:
```sh
./arcusctl --config-file ./config.yaml acl group list
```

`--config-file`을 사용하면 지정한 파일을 직접 읽습니다. 운영 자동화 스크립트에서는 설정 파일 위치를 명확히 남기기 위해 이 방식을 권장합니다.

## 3. 기본 설정 파일 탐색 순서

`--config-file`을 지정하지 않으면 `arcusctl`은 다음 순서로 `config.yaml` 파일을 찾습니다.

1. `arcusctl` 실행 파일이 위치한 디렉토리
2. 현재 작업 디렉토리

> [!NOTE]
> `--config-file`로 파일을 직접 지정하는 경우를 제외하면 설정 파일명은 `config.yaml`이어야 합니다.

## 4. 환경 변수로 재정의

환경 변수 형식:
```sh
ARCUSCTL_ZOOKEEPER="<host:port>" arcusctl <command>
```

실행 예시:
```sh
ARCUSCTL_ZOOKEEPER="localhost:2181" ./arcusctl acl group list
```

`ARCUSCTL_` 접두사를 사용한 환경 변수는 설정 파일의 값보다 우선합니다. 임시 테스트나 CI 실행처럼 설정 파일을 수정하지 않고 접속 대상을 바꿔야 할 때 사용할 수 있습니다.

## 5. 설정 확인 흐름

명령 형식:
```sh
arcusctl --config-file <config.yaml> <command> --help
```

실행 예시:
```sh
./arcusctl --config-file ./config.yaml connect --help
```

> [!CAUTION]
> ZooKeeper 주소가 비어 있거나 접근할 수 없으면 ZooKeeper를 사용하는 명령은 실패합니다. 방화벽, 포트, Ensemble 주소, 설정 파일 탐색 위치를 먼저 확인하세요.

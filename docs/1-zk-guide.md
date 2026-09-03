# ZooKeeper 운영 가이드

이 문서는 arcusctl을 사용해 ZooKeeper 앙상블을 배포하고 시작, 상태 확인, 중지 및 삭제하는 방법을 설명합니다.

배포에 필요한 환경과 운영 장비의 상태 저장 방식은 [arcusctl 시작하기](0-getting-started.md)를 먼저 참고하세요.

## 운영 흐름

ZooKeeper 앙상블은 다음 순서로 배포하고 관리합니다.

```text
토폴로지 작성 -> deploy -> start -> status / list -> stop -> delete
```

> [!IMPORTANT]
> `deploy`와 `start`는 분리되어 있습니다. 배포가 성공해도 ZooKeeper 서버는 자동으로 시작되지 않습니다.

## 명령 요약

| 명령                                                | 주요 동작                                                         |
|-----------------------------------------------------|-------------------------------------------------------------------|
| `arcusctl zk deploy <version> <topology.yml>`       | ZooKeeper 바이너리와 서버별 설정 및 데이터 배포                   |
| `arcusctl zk start <ensemble-name> [--node <myid>]` | 앙상블 내 ZooKeeper 서버 시작                                     |
| `arcusctl zk status <ensemble-name>`                | 앙상블 내 ZooKeeper 서버의 실행 상태와 역할 출력                  |
| `arcusctl zk list`                                  | 관리 중인 앙상블 목록 출력                                        |
| `arcusctl zk stop <ensemble-name> [--node <myid>]`  | 앙상블 내 ZooKeeper 서버 중지                                     |
| `arcusctl zk delete <ensemble-name> [--purge]`      | 앙상블 설정과 데이터 삭제 및 선택적으로 버전별 설치 디렉터리 삭제 |

## 토폴로지 작성

저장소의 [ZooKeeper 토폴로지 예시](../examples/zk-topology.yaml)를 복사한 뒤 환경에 맞게 수정할 수 있습니다.

```yaml
name: my-ensemble
path: /home/arcus/apps/zookeeper

servers:
  - myid: 1
    address: 192.0.2.11:2181:2888:3888
    config:
      data_log_dir: /data/zookeeper/txlog-1

  - myid: 2
    address: 192.0.2.12:2181:2888:3888

  - myid: 3
    address: 192.0.2.13:2181:2888:3888

global_config:
  tick_time: 2000
  init_limit: 10
  sync_limit: 5
  data_dir: /var/lib/zk/data
  data_log_dir: /var/lib/zk/datalog
  properties:
    maxClientCnxns: "60"
    autopurge.snapRetainCount: "10"
    autopurge.purgeInterval: "24"
```

### 기본 필드

| 필드                | 설명                                                                  | 필수   |
|---------------------|-----------------------------------------------------------------------|--------|
| `name`              | 앙상블을 식별하는 고유 이름                                           | 예     |
| `path`              | ZooKeeper를 설치할 원격 장비의 기준 경로                              | 예     |
| `servers`           | 하나 이상의 ZooKeeper 서버 목록                                       | 예     |
| `servers[].myid`    | 앙상블 안에서 중복되지 않는 ZooKeeper 서버 ID                         | 예     |
| `servers[].address` | `<host>:<client-port>:<quorum-port>:<election-port>` 형식의 고유 주소 | 예     |
| `servers[].config`  | 특정 ZooKeeper 서버에 추가하거나 덮어쓸 설정                          | 아니요 |
| `global_config`     | 모든 ZooKeeper 서버에 공통으로 적용할 설정                            | 아니요 |

### ZooKeeper 설정

| 필드           | 설명                                                       | 기본값                        |
|----------------|------------------------------------------------------------|-------------------------------|
| `tick_time`    | ZooKeeper의 기본 `tick` 시간                               | `2000`                        |
| `init_limit`   | follower가 leader에 연결하고 동기화할 수 있는 `tick` 수    | `10`                          |
| `sync_limit`   | follower와 leader 사이의 요청 및 응답에 허용되는 `tick` 수 | `5`                           |
| `data_dir`     | 스냅샷과 `myid`를 저장할 기준 디렉터리                     | `<path>/data/<ensemble-name>` |
| `data_log_dir` | 트랜잭션 로그를 저장할 기준 디렉터리                       | 해당 서버의 `data_dir`        |
| `properties`   | 추가할 `zoo.cfg` 속성                                      | 없음                          |

ZooKeeper 서버별 `servers[].config`는 `global_config`에 병합됩니다.

- ZooKeeper 서버별 설정에 값이 있으면 같은 이름의 전역 설정을 덮어씁니다.
- `properties`는 키 단위로 추가하거나 덮어씁니다.
- 전역 설정과 서버별 설정에서 모두 `data_dir`을 생략하면 `<path>/data/<ensemble-name>`을 사용합니다.
- 전역 설정과 서버별 설정에서 모두 `data_log_dir`을 생략하면 해당 ZooKeeper 서버의 최종 `data_dir`을 사용합니다.

각 ZooKeeper 서버의 실제 데이터 및 로그 경로에는 `zk<myid>` 하위 디렉터리가 추가됩니다. 

```text
<data_dir>/zk<myid>
<data_log_dir>/zk<myid>
```

기본 설정을 사용하는 경우 실제 경로는 다음과 같습니다.

```text
<path>/data/<ensemble-name>/zk<myid>
```

> 이 문서에서 `<ensemble-name>`은 토폴로지의 `name`에 지정한 값을 의미합니다.

## `zk deploy`

새 ZooKeeper 앙상블을 배포합니다.

### 명령 형식

```sh
arcusctl zk deploy <version> <topology.yml>
```

### 실행 예시

```sh
arcusctl zk deploy 3.5.9 zk-topology.yml
```

배포를 시작하기 전에 배포 대상과 설치 경로를 다음과 같이 출력하고 진행 여부를 확인합니다.

```text
ZooKeeper ensemble "my-ensemble" will be deployed (version: 3.5.9)

MYID  HOST          PORTS           DIRECTORIES

1     192.0.2.11    2181/2888/3888  /home/arcus/apps/zookeeper/3.5.9
2     192.0.2.12    2181/2888/3888  /home/arcus/apps/zookeeper/3.5.9
3     192.0.2.13    2181/2888/3888  /home/arcus/apps/zookeeper/3.5.9

Attention:
  1. If the topology is not what you expected, check your yaml file.
  2. Please confirm there is no port/directory conflicts in same host.
Proceed with deployment? (y/N):
```

`deploy`는 토폴로지의 유효성과 앙상블 이름의 중복 여부를 검증합니다. 사용자가 배포를 승인하면 다음 작업을 수행합니다.

- 원격 장비에 버전별 ZooKeeper 바이너리를 설치하거나 기존 설치를 재사용합니다.
- ZooKeeper 서버별 설정과 데이터 및 로그 디렉터리를 생성합니다.
- 운영 장비에 배포 버전, 메타데이터 및 토폴로지를 저장합니다.

### 아카이브 준비

Apache ZooKeeper 아카이브는 운영 장비의 다음 위치에 내려받거나 기존 파일을 재사용합니다.

```text
<home>/images/zookeeper/apache-zookeeper-<version>-bin.tar.gz
```

배포 정보는 다음 위치에 저장됩니다.

```text
<home>/clusters/zookeeper/<ensemble-name>/
├── meta.yml
└── topology.yml
```

### 원격 장비의 설치 경로

ZooKeeper 바이너리는 원격 장비의 `<path>/<version>`에 설치됩니다. 앙상블별 설정과 기본 데이터는 `<path>` 아래에 별도로 생성됩니다.

```text
<path>/
├── <version>/
│   ├── bin/
│   │   └── zkServer.sh
│   ├── conf/
│   ├── docs/
│   ├── lib/
│   ├── logs/
│   └── zookeeper-<version>.tar.gz
├── conf/
│   └── <ensemble-name>/
│       └── zk<myid>/
│           ├── zoo.cfg
│           └── zoo.cfg.dynamic
└── data/
    └── <ensemble-name>/
        └── zk<myid>/
            └── myid
```

위의 `data` 구조는 `data_dir`을 생략한 경우입니다. `data_dir` 또는 `data_log_dir`을 지정하면 각각 다음 경로를 사용합니다.

```text
<data_dir>/zk<myid>
<data_log_dir>/zk<myid>
```

`data_log_dir`을 생략하면 `data_dir`과 동일한 경로를 사용하므로 스냅샷과 트랜잭션 로그가 같은 `<data_dir>/zk<myid>` 디렉터리에 저장됩니다.

같은 원격 장비의 동일한 `<path>/<version>`은 여러 앙상블이 공유할 수 있습니다.
`<path>/<version>/bin/zkServer.sh`가 이미 있으면 버전별 바이너리를 재사용하고 앙상블별 설정과 데이터를 생성합니다.

다만 운영 환경에서는 원격 장비당 하나의 ZooKeeper 서버를 구성하는 것이 일반적이며, **같은 원격 장비에서 여러 ZooKeeper 서버를 구동하는 방식은 권장하지 않습니다.**

### 배포 실패 시 확인

> [!WARNING]
> 배포 중 오류가 발생해도 완료된 작업은 자동으로 롤백되지 않습니다.

오류 메시지에 표시된 원격 장비에서 다음 경로가 생성되었는지 확인하세요.

```text
<path>/<version>
<path>/conf/<ensemble-name>/zk<myid>
<data_dir>/zk<myid>
<data_log_dir>/zk<myid>
```

운영 장비에 배포 정보가 일부 저장되었을 수 있으므로 다음 경로도 함께 확인하세요.

```text
<home>/clusters/zookeeper/<ensemble-name>
```

`<path>/<version>`은 다른 앙상블이 공유하여 사용할 수 있으므로 삭제하기 전에 사용 여부를 확인하세요. 데이터와 로그 경로는 보존 필요 여부를 확인한 후 정리하세요.

## `zk start`

ZooKeeper 서버를 시작합니다.

### 명령 형식

앙상블 전체 또는 `--node`로 지정한 ZooKeeper 서버를 시작합니다.

```sh
arcusctl zk start <ensemble-name> [--node <myid>]
```

### 실행 예시

전체 앙상블을 시작합니다.

```sh
arcusctl zk start my-ensemble
```

`myid`가 `2`인 ZooKeeper 서버만 시작합니다.

```sh
arcusctl zk start my-ensemble --node 2
```

전체 앙상블을 시작하던 중 오류가 발생하면 일부 ZooKeeper 서버만 실행된 상태로 남을 수 있습니다. 
`status`로 전체 상태를 확인한 후 실행되지 않은 서버를 다시 시작하세요.

## `zk status`

ZooKeeper 앙상블의 상태를 확인합니다.

### 명령 형식

```sh
arcusctl zk status <ensemble-name>
```

### 실행 예시

```sh
arcusctl zk status my-ensemble
```

앙상블의 각 ZooKeeper 서버에 대해 원격 장비 주소와 `myid`를 표시한 후, 
해당 서버에서 `zkServer.sh status <zoo.cfg>`를 실행한 결과를 출력합니다.

다음은 정상적으로 실행 중인 앙상블의 주요 출력만 나타낸 예시입니다.

```text
Ensemble: my-ensemble (version: 3.5.9)

=== 192.0.2.11 (myid=1) ===
Using config: /home/arcus/apps/zookeeper/conf/my-ensemble/zk1/zoo.cfg
Client port found: 2181. Client address: localhost. Client SSL: false.
Mode: follower

=== 192.0.2.12 (myid=2) ===
Using config: /home/arcus/apps/zookeeper/conf/my-ensemble/zk2/zoo.cfg
Client port found: 2181. Client address: localhost. Client SSL: false.
Mode: leader

=== 192.0.2.13 (myid=3) ===
Using config: /home/arcus/apps/zookeeper/conf/my-ensemble/zk3/zoo.cfg
Client port found: 2181. Client address: localhost. Client SSL: false.
Mode: follower
```

## `zk list`

현재 운영 장비에서 관리하는 ZooKeeper 앙상블 목록을 출력합니다.

### 명령 및 실행 예시

```sh
arcusctl zk list
```

현재 운영 장비에 저장된 메타데이터를 기준으로 ZooKeeper 앙상블 목록을 다음과 같이 출력합니다.

```text
NAME            VERSION  DEPLOYED AT
dev-ensemble    3.4.6    2026-08-14 12:30:14
prod-ensemble   3.4.5    2026-06-21 01:00:03
test-ensemble   3.5.9    2026-09-01 14:44:03
```

출력 항목은 다음과 같습니다. 

- 앙상블 이름
- ZooKeeper 버전
- 배포 시간

`list`는 실행 중인 ZooKeeper 서버를 자동으로 탐색하지 않습니다.

## `zk stop`

ZooKeeper 서버를 중지합니다.

### 명령 형식

앙상블 전체 또는 `--node`로 지정한 ZooKeeper 서버를 중지합니다.

```sh
arcusctl zk stop <ensemble-name> [--node <myid>]
```

### 실행 예시

전체 앙상블을 중지합니다.

```sh
arcusctl zk stop my-ensemble
```

`myid`가 `2`인 ZooKeeper 서버만 중지합니다.

```sh
arcusctl zk stop my-ensemble --node 2
```

전체 앙상블을 중지하던 중 오류가 발생하면 일부 ZooKeeper 서버가 계속 실행 중일 수 있습니다. `status`로 전체 상태를 확인한 후 다시 중지하세요.

## `zk delete`

ZooKeeper 앙상블을 삭제합니다.

### 명령 형식

```sh
arcusctl zk delete <ensemble-name> [--purge]
```

- 기본 `delete`는 앙상블의 설정과 데이터를 삭제하지만 버전별 설치 디렉터리는 보존합니다.
  - 운영 장비에서 `meta.yml`과 `topology.yml`을 삭제합니다.
  - 원격 장비에서 ZooKeeper 앙상블 설정과 데이터 및 로그 파일을 삭제합니다.
- `--purge` 옵션을 지정하면 운영 장비에 저장된 배포 정보를 기준으로 다른 앙상블이 같은 원격 장비의 `<path>/<version>`을 사용하는지 확인합니다.
    - 해당 설치를 공유하는 다른 앙상블이 없는 경우 설치 디렉터리까지 삭제합니다.

### 실행 예시

```sh
arcusctl zk delete my-ensemble
arcusctl zk delete my-ensemble --purge
```

### 삭제 범위

삭제하기 전에 모든 ZooKeeper 서버가 중지되어 있는지 확인하고 사용자에게 진행 여부를 확인합니다.

기본 `delete`로 원격 장비에서 제거되는 경로는 다음과 같습니다.

```text
<path>/conf/<ensemble-name>/zk<myid>
<data_dir>/zk<myid>
<data_log_dir>/zk<myid>
```

`--purge`로 추가되는 삭제 경로는 다음과 같습니다.

```text
<path>/<version>
```

운영 장비에서는 다음 디렉터리를 제거합니다.

```text
<home>/clusters/zookeeper/<ensemble-name>
```

기본 `delete`와 `delete --purge` 모두 운영 장비의 `images`에 저장된 아카이브를 삭제하지 않습니다.

```text
<home>/images/zookeeper/apache-zookeeper-<version>-bin.tar.gz
```

> [!CAUTION]
> `zk delete`는 앙상블의 설정과 데이터를 영구적으로 삭제하며, 삭제 작업은 자동으로 롤백되지 않습니다.
>
> 작업 중 오류가 발생하면 일부 원격 장비의 설정과 데이터만 삭제된 상태로 남을 수 있습니다.
>
> 운영 환경에서는 데이터 보존 여부와 삭제 경로가 다른 앙상블이나 프로그램과 공유되지 않는지 확인한 후 실행하세요.

# Arcus 클러스터 운영 가이드

이 문서는 arcusctl을 사용해 community 또는 enterprise Arcus 클러스터를 배포하고 시작, 상태 확인, 중지 및 삭제하는 방법을 설명합니다.

배포에 필요한 환경과 운영 장비의 상태 저장 방식은 [arcusctl 시작하기](0-getting-started.md)를 먼저 참고하세요.

## 운영 흐름

Arcus 클러스터는 다음 순서로 배포하고 관리합니다.

```text
ZooKeeper 준비 -> 토폴로지 작성 -> deploy -> start -> status / list -> stop -> delete
```

> [!IMPORTANT]
> 클러스터 토폴로지에는 실행 중인 ZooKeeper 앙상블 주소가 필요합니다.
>
> `deploy`와 `start`는 분리되어 있습니다. 배포가 성공해도 캐시 서버는 자동으로 시작되지 않습니다.

## 명령 요약

| 명령                                               | 주요 동작                                                |
|----------------------------------------------------|----------------------------------------------------------|
| `arcusctl cluster deploy <version> <topology.yml>` | Arcus 바이너리 설치와 ZooKeeper 등록 정보 생성           |
| `arcusctl cluster start <servicecode> [options]`   | 클러스터 내 캐시 서버 시작                               |
| `arcusctl cluster status <servicecode>`            | 클러스터 내 캐시 서버 실행 및 ZooKeeper 등록 상태 출력   |
| `arcusctl cluster list`                            | 관리 중인 클러스터 목록 출력                             |
| `arcusctl cluster stop <servicecode> [options]`    | 클러스터 내 캐시 서버 중지                               |
| `arcusctl cluster delete <servicecode> [--purge]`  | 클러스터 등록 정보 삭제 및 선택적으로 설치 디렉터리 삭제 |

## 토폴로지 작성

Edition에 맞는 토폴로지 예시를 복사한 뒤 환경에 맞게 수정할 수 있습니다.

- [Community 토폴로지 예시](../examples/cluster-community-topology.yaml)
- [Enterprise 토폴로지 예시](../examples/cluster-enterprise-topology.yaml)

### Community 예시

```yaml
servicecode: my-cluster
path: /home/arcus/apps/arcus-memcached
zookeeper: 192.0.2.11:2181,192.0.2.12:2181,192.0.2.13:2181

servers:
  - address: 192.0.2.21:11211

  - address: 192.0.2.22:11211

  - address: 192.0.2.23:11211
    config:
      options: "-m 128 -v"

globalConfig:
  options: "-t 4 -c 1024 -b 1024 -B auto -m 64"
```

### Enterprise 예시

```yaml
servicecode: my-repl-cluster
path: /home/arcus/apps/arcus-memcached
zookeeper: 192.0.2.11:2181,192.0.2.12:2181,192.0.2.13:2181

servers:
  - address: 192.0.2.21:11211
    group:
      name: g1
      role: master
      port: 33533

  - address: 192.0.2.22:11211
    group:
      name: g1
      role: slave
      port: 33533

globalConfig:
  options: "-t 8 -m 128"
```

### 기본 필드

| 필드                       | 설명                                          | 필수   |
|----------------------------|-----------------------------------------------|--------|
| `servicecode`              | 클러스터를 식별하는 고유 이름                 | 예     |
| `path`                     | Arcus를 설치할 원격 장비의 기준 경로          | 예     |
| `zookeeper`                | 쉼표로 구분한 ZooKeeper 앙상블 주소           | 예     |
| `servers`                  | 하나 이상의 캐시 서버 목록                    | 예     |
| `servers[].address`        | `<host>:<port>` 형식의 고유한 캐시 서버 주소  | 예     |
| `servers[].config.options` | 특정 캐시 서버에 추가할 실행 옵션             | 아니요 |
| `globalConfig.options`     | 모든 캐시 서버에 공통으로 적용할 실행 옵션    | 아니요 |
| `servers[].group`          | Enterprise 캐시 서버의 replication group 설정 | 조건부 |

하나의 토폴로지에 Community edition 캐시 서버와 Enterprise edition 캐시 서버를 함께 정의할 수 없습니다.

> [!WARNING]
> `-p`, `-P`, `-z`, `-d` 옵션은 arcusctl이 직접 구성합니다. 토폴로지의 `options`에는 해당 옵션을 지정하지 마세요.

### Group 필드

Enterprise edition에서는 모든 캐시 서버에 `group`을 지정해야 합니다.

| 필드   | 설명                      |
|--------|---------------------------|
| `name` | Replication group 이름    |
| `role` | `master` 또는 `slave`     |
| `port` | Replication에 사용할 포트 |

Enterprise 토폴로지에는 다음 규칙이 적용됩니다.

- 각 group에는 `master`가 정확히 하나 있어야 합니다.
- 각 group에는 `slave`를 하나만 지정하거나 생략할 수 있습니다.
- Replication 포트는 `1`부터 `65535` 사이여야 합니다.

## `cluster deploy`

새 Arcus 클러스터를 배포합니다.

### 명령 형식

```sh
arcusctl cluster deploy <version> <topology.yml>
```

### 실행 예시

배포를 시작하기 전에 배포 대상과 설치 경로를 출력하고 진행 여부를 확인합니다.

#### Community 

```sh
arcusctl cluster deploy 1.16.1 community-topology.yml
```

```text
Arcus cluster "my-cluster" will be deployed (edition: community, version: 1.16.1)

ADDRESS             DIRECTORY

192.0.2.21:11211    /home/arcus/apps/arcus-memcached/1.16.1
192.0.2.22:11211    /home/arcus/apps/arcus-memcached/1.16.1
192.0.2.23:11211    /home/arcus/apps/arcus-memcached/1.16.1

Attention:
  1. If the topology is not what you expected, check your yaml file.
  2. Please confirm there is no port/directory conflicts in same host.
Proceed with deployment? (y/N):
```

#### Enterprise

```sh
arcusctl cluster deploy 0.12.1-E enterprise-topology.yml
```

Enterprise 클러스터는 배포 대상의 replication group, role도 함께 출력합니다.

```text
Arcus cluster "my-repl-cluster" will be deployed (edition: enterprise, version: 0.12.1-E)

GROUP  ROLE    ADDRESS             DIRECTORY

g1     master  192.0.2.21:11211    /home/arcus/apps/arcus-memcached/0.12.1-E
g1     slave   192.0.2.22:11211    /home/arcus/apps/arcus-memcached/0.12.1-E
g2     master  192.0.2.23:11211    /home/arcus/apps/arcus-memcached/0.12.1-E
g2     slave   192.0.2.24:11211    /home/arcus/apps/arcus-memcached/0.12.1-E

Attention:
  1. If the topology is not what you expected, check your yaml file.
  2. Please confirm there is no port/directory conflicts in same host.
Proceed with deployment? (y/N):
```

`deploy`는 토폴로지 유효성을 검증하고, 운영 장비와 ZooKeeper에 동일한 서비스코드가 이미 등록되어 있는지 확인합니다.
사용자가 배포를 승인하면 다음 작업을 수행합니다. 

- 원격 장비에 버전별 Arcus 바이너리를 설치하거나 기존 설치를 재사용합니다.
- ZooKeeper에 클러스터와 캐시 서버의 등록 정보를 생성합니다.
- 운영 장비에 배포 버전, 메타데이터 및 토폴로지를 저장합니다.

### 아카이브 준비

Community 아카이브는 다음 위치에 내려받거나 기존 파일을 재사용합니다.

```text
<home>/images/arcus-community/arcus-memcached-<version>.tar.gz
```

Enterprise 아카이브는 자동으로 내려받지 않습니다. 배포 전에 다음 위치에 직접 준비해야 합니다.

```text
<home>/images/arcus-enterprise/arcus-memcached-<version>.tar.gz
```

배포 정보는 다음 위치에 저장됩니다.

```text
<home>/clusters/arcus/<servicecode>/
├── meta.yml
└── topology.yml
```

### 원격 장비의 설치 경로

Arcus는 원격 장비의 다음 경로에 버전별로 설치됩니다.

```text
<path>/<version>/
├── bin/memcached
├── lib/
├── src/
└── arcus-memcached-<version>.tar.gz
```

같은 토폴로지에 동일한 원격 장비의 캐시 서버가 여러 개 있으면 빌드는 원격 장비마다 한 번만 수행합니다.

대상 경로에 `bin/memcached`가 이미 존재하면 기존 설치를 재사용합니다.

### ZooKeeper 등록

Edition에 따라 다음 ZNode를 생성합니다.

Community edition은 `/arcus`를 root로 사용합니다.

```text
/arcus/cache_list/<servicecode>
/arcus/client_list/<servicecode>
/arcus/cache_server_mapping/<host>:<port>/<servicecode>
```

Enterprise edition은 `/arcus_repl`을 root로 사용합니다.

```text
/arcus_repl/cache_list/<servicecode>
/arcus_repl/client_list/<servicecode>
/arcus_repl/cache_server_mapping/<host>:<port>/<servicecode>^<group-name>^<host>:<replication-port>
/arcus_repl/group_list/<servicecode>/<group-name>
```

### 배포 실패 시 확인

> [!WARNING]
> 배포 중 오류가 발생해도 완료된 작업은 자동으로 롤백되지 않습니다.

오류 메시지에 표시된 원격 장비에서 다음 경로가 생성되었는지 확인하세요.

```text
<path>/<version>
```

ZooKeeper에 클러스터 등록 정보가 일부 생성되었는지 확인하고, 운영 장비의 다음 경로도 함께 확인하세요.

```text
<home>/clusters/arcus/<servicecode>
```

`<path>/<version>`은 같은 원격 장비의 다른 클러스터가 공유할 수 있으므로 삭제하기 전에 사용 여부를 확인하세요.

## `cluster start`

클러스터 전체 또는 일부 캐시 서버를 시작합니다.

### 명령 형식

```sh
arcusctl cluster start <servicecode> [options]
```

- Community 클러스터는 `--node <host>:<port>`로 특정 캐시 서버를 지정할 수 있습니다.
- Enterprise 클러스터는 `--group <group-name>`으로 특정 replication group을 지정할 수 있습니다.
- 옵션을 생략하면 클러스터의 모든 캐시 서버를 시작합니다.

### 실행 예시

#### Community

```sh
arcusctl cluster start my-cluster
arcusctl cluster start my-cluster --node 192.0.2.22:11211
```

Community 클러스터에서는 `--group`을 사용할 수 없습니다.

#### Enterprise

```sh
arcusctl cluster start my-repl-cluster
arcusctl cluster start my-repl-cluster --group g1
```

Enterprise 클러스터에서는 `--node`를 사용할 수 없습니다.

Enterprise 클러스터는 `master`를 먼저 시작한 뒤 `slave`를 시작합니다.

실행 중 오류가 발생하면 일부 캐시 서버만 시작된 상태로 남을 수 있습니다. `status`로 전체 상태를 확인한 후 다시 시작하세요.

## `cluster status`

Arcus 클러스터의 상태를 확인합니다.

### 명령 형식

```sh
arcusctl cluster status <servicecode>
```

### 실행 예시

#### Community

```sh
arcusctl cluster status my-cluster
```

캐시 서버별 프로세스 실행 상태와 ZooKeeper 등록 여부를 다음과 같이 출력합니다.

```text
Arcus cluster "my-cluster" (edition: community, version: 1.16.1)

ADDRESS           PROCESS_STATUS  ZK_REGISTERED
192.0.2.21:11211  running         yes
192.0.2.22:11211  running         yes
192.0.2.23:11211  running         yes
```

#### Enterprise

```sh
arcusctl cluster status my-repl-cluster
```

Enterprise 클러스터는 replication group과 각 캐시 서버의 역할도 함께 출력합니다.

```text
Arcus cluster "my-repl-cluster" (edition: enterprise, version: 0.12.1-E)

GROUP  ROLE    ADDRESS           PROCESS_STATUS  ZK_REGISTERED
g1     master  192.0.2.21:11211  running         yes
g1     slave   192.0.2.22:11211  running         yes
g2     master  192.0.2.23:11211  running         yes
g2     slave   192.0.2.24:11211  running         yes
```

출력 상태는 다음과 같습니다.

| 상태             | 의미                                   | 값                   |
|------------------|----------------------------------------|----------------------|
| `PROCESS_STATUS` | 원격 장비에서 해당 캐시 서버 실행 여부 | `running`, `stopped` |
| `ZK_REGISTERED`  | ZooKeeper에 캐시 서버 매핑 존재 여부   | `yes`, `no`          |

## `cluster list`

현재 운영 장비에서 관리하는 Arcus 클러스터 목록을 출력합니다.

### 명령 및 실행 예시

```sh
arcusctl cluster list
```

현재 운영 장비에 저장된 배포 정보를 기준으로 Arcus 클러스터 목록을 다음과 같이 출력합니다.

```text
SERVICECODE      VERSION   EDITION     NODES  GROUPS  DEPLOYED_AT
my-cluster       1.16.1    community   3      -       2026-08-01 12:30:02
my-repl-cluster  0.12.1-E  enterprise  4      2       2026-09-01 01:00:13
```

다음 정보를 표시합니다.

- 서비스코드
- Arcus 버전
- Edition
- 캐시 서버 수
- replication group 수
- 배포 시간

`list`는 실행 중인 캐시 서버를 자동으로 탐색하지 않습니다.

## `cluster stop`

클러스터 전체 또는 일부 캐시 서버를 중지합니다.

### 명령 형식

```sh
arcusctl cluster stop <servicecode> [options]
```

- Community 클러스터는 `--node <host>:<port>`로 특정 캐시 서버를 지정할 수 있습니다.
- Enterprise 클러스터는 `--group <group-name>`으로 특정 replication group을 지정할 수 있습니다.
- 옵션을 생략하면 클러스터의 모든 캐시 서버를 중지합니다.

### 실행 예시

#### Community

```sh
arcusctl cluster stop my-cluster
arcusctl cluster stop my-cluster --node 192.0.2.22:11211
```

Community 클러스터에서는 `--group`을 사용할 수 없습니다.

#### Enterprise

```sh
arcusctl cluster stop my-repl-cluster
arcusctl cluster stop my-repl-cluster --group g1
```

Enterprise 클러스터에서는 `--node`를 사용할 수 없습니다.

Enterprise 클러스터는 `slave`를 먼저 중지한 뒤 `master`를 중지합니다.

실행 중 오류가 발생하면 일부 캐시 서버가 계속 실행 중일 수 있습니다. `status`로 전체 상태를 확인한 후 다시 중지하세요.

## `cluster upgrade`

> [!NOTE]
> 구현 전 인터페이스 초안입니다. 세부 동작과 출력 형식은 구현 과정에서 바뀔 수 있습니다.

지정한 캐시 노드 또는 그룹을 목표 버전으로 순차 업그레이드합니다.

### 명령 형식

```sh
arcusctl cluster upgrade <servicecode> --version <version> <target>
```

```text
<target>: --node <host>:<port>[,...] | --group <group-name>[,...] | --all
```

- `--version`은 필수이며, 적용할 버전을 지정합니다.
- `<target>`은 필수이며, `--node`, `--group`, `--all` 중 정확히 하나를 지정해야 합니다. 실수로 클러스터 전체를 일괄 업그레이드하는 것을 막기 위해 대상을 생략할 수 없습니다.
- `--node`, `--group`, `--all`은 함께 사용할 수 없습니다. 여러 대상을 선택할 때는 comma로 구분합니다.
- `--all`은 클러스터의 모든 캐시 서버를 대상으로 하되, 이미 목표 버전으로 정상 동작하는 대상은 건너뛰고 남은 대상만 업그레이드합니다. 일부 대상을 먼저 업그레이드한 뒤 나머지를 마무리할 때 사용합니다.
- 실서비스 환경에서는 다음 방식으로 업그레이드하는 것을 권장합니다.
  - 클러스터 노드/그룹 개수 기준으로 1% -> 20% -> 100% 로 점차 확대하여 업그레이드합니다. 각 업그레이드 시점 사이에는 일정 기간을 두고 정상 동작하는지 확인합니다.
  - Community 클러스터의 노드 업그레이드 시 하나씩 `업그레이드 -> 상태 확인 -> 일정 시간 대기` 순서로 진행합니다. Enterprise 클러스터와 달리 업그레이드 시 데이터가 사라지기 때문에 동시에 여러 노드가 재구동되는 경우 캐시 미스가 발생할 확률이 높아지기 때문입니다.
- 테스트 환경에서는 편의를 위해 여러 노드/그룹을 한꺼번에 업그레이드할 수 있습니다.

### 실행 예시

#### Community

```sh
arcusctl cluster upgrade my-cluster --version 1.16.2 --node 192.0.2.21:11211
arcusctl cluster upgrade my-cluster --version 1.16.2 --node 192.0.2.21:11211,192.0.2.22:11211
arcusctl cluster upgrade my-cluster --version 1.16.2 --all
```

- `--node`로 업그레이드할 캐시 서버를 지정합니다. 여러 캐시 서버를 지정하면 나열한 순서대로 업그레이드합니다.
- `--group`은 사용할 수 없습니다.

각 캐시 서버는 다음 순서로 업그레이드합니다.

```
캐시 서버 stop -> 목표 버전으로 start -> 실행 버전 및 클러스터 편입 확인(cache_list)
```

> [!NOTE]
> 업그레이드 작업은 한 번에 모든 노드를 대상으로 하기보다는,
> 일부 노드들만 먼저 업그레이드하여 클러스터가 정상적으로 동작하는지 확인하고
> 나머지 노드들을 최종적으로 업그레이드하는 방식을 권장합니다.

#### Enterprise

```sh
arcusctl cluster upgrade my-repl-cluster --version 0.12.2-E --group g1
arcusctl cluster upgrade my-repl-cluster --version 0.12.2-E --group g1,g2
arcusctl cluster upgrade my-repl-cluster --version 0.12.2-E --all
```

- `--group`으로 업그레이드할 replication group을 지정합니다. 여러 group을 지정하면 나열한 순서대로 업그레이드합니다.
- `--node`는 사용할 수 없습니다.
- 내부적으로 업그레이드 작업 수행 시 그룹 내의 `master` 역할의 노드는 절대 업그레이드하지 않습니다. 반드시 switchover 명령을 통해 slave 역할로 변경한 후 업그레이드 과정을 거칩니다. 단, master만 있는 그룹의 경우 역할 변경 없이 업그레이드가 진행되며 이로 인해 데이터는 유실됩니다.
- 한 그룹 내 노드들의 버전이 일치하지 않는 경우여도 각각 목표 버전으로 업그레이드하도록 합니다.

각 복제 그룹은 다음 순서로 업그레이드합니다.

1. 업그레이드 대상 그룹의 역할과 복제 상태(SYNC_RUN)를 확인합니다.
2. 현재 `slave`를 중지합니다.
3. 목표 버전으로 `slave`를 시작하고 데이터 동기화가 완료(SYNC_RUN)되었는지 확인합니다.
4. 새 버전 `slave`를 `master`로 switchover하고 역할과 복제 상태를 확인합니다.
5. `slave`가 된 기존 `master`에도 2~4번 작업을 반복합니다.
6. Group 내 모든 캐시 서버가 목표 버전으로 정상 동작하는지 확인한 뒤 다음 group으로 진행합니다.

> [!NOTE]
> 업그레이드 작업은 한 번에 모든 그룹을 대상으로 하기보다는,
> 일부 그룹들만 먼저 업그레이드하여 클러스터가 정상적으로 동작하는지 확인하고
> 나머지 그룹들을 최종적으로 업그레이드하는 방식을 권장합니다.

### Plan 출력

업그레이드를 시작하기 전에 대상별 현재 버전, 목표 버전, 수행할 동작을 출력하고 진행 여부를 확인합니다.
이미 목표 버전으로 동작하는 노드는 `skip` 표시하므로, 실제로 업그레이드될 노드 목록을 확인할 수 있습니다.

#### Community

```sh
arcusctl cluster upgrade my-cluster --version 1.16.2 --all
```

```text
Arcus cluster "my-cluster" will be upgraded.

ADDRESS           CURRENT  TARGET  ACTION
192.0.2.21:11211  1.16.2   1.16.2  skip
192.0.2.22:11211  1.16.1   1.16.2  upgrade
192.0.2.23:11211  1.16.1   1.16.2  upgrade

Proceed with upgrade? (y/N):
```

#### Enterprise

```sh
arcusctl cluster upgrade my-repl-cluster --version 0.12.2-E --all
```

Enterprise 클러스터는 replication group 단위로 업그레이드 가능합니다.
이미 목표 버전으로 동작하는 노드는 마찬가지로 `skip` 표시하므로 실제로 업그레이드될 그룹 목록을 확인할 수 있습니다.

```text
Arcus cluster "my-repl-cluster" will be upgraded.

GROUP  ROLE    ADDRESS           CURRENT   TARGET    ACTION
g1     master  192.0.2.22:11211  0.12.2-E  0.12.2-E  skip
g1     slave   192.0.2.21:11211  0.12.2-E  0.12.2-E  skip
g2     master  192.0.2.23:11211  0.12.1-E  0.12.2-E  upgrade
g2     slave   192.0.2.24:11211  0.12.1-E  0.12.2-E  upgrade

Proceed with upgrade? (y/N):
```

### 바이너리 준비 시점

upgrade 명령 수행 시 목표 버전 바이너리는 캐시 서버 중지 작업 전에 모든 대상 노드/그룹의 원격 장비에 준비해둡니다.

```text
plan 출력 및 확인 -> 대상 원격 장비 전체에 목표 버전 설치 -> 대상별 순차 업그레이드 시작
```

- 빌드는 `deploy`와 같은 방식으로 원격 장비마다 한 번만 수행하고, `<path>/<version>/bin/memcached`가 이미 있으면 재사용합니다.
- 설치가 하나라도 실패하면 캐시 노드 상태를 변경하지 않은 채 종료됩니다.
- 기존 버전 설치는 삭제하지 않습니다. 만약 해당 버전을 더 이상 사용하지 않는 경우 직접 제거해주어야 합니다.
- 이전 버전으로 되돌려야 하는 경우에는 upgrade 명령에 이전 버전을 지정해 사용합니다.

### 공통 동작

- 버전별 설치 경로(`<path>/<version>`)를 사용하고, 캐시 서버별 실행 버전을 관리합니다.
- 이미 목표 버전으로 정상 동작하는 캐시 노드/그룹은 건너뜁니다.
- 각 단계의 결과와 캐시 서버별 버전을 운영 장비의 메타데이터 파일(meta.yml)에 기록합니다.

### 업그레이드 실패 시 확인

> [!WARNING]
> 업그레이드 중 오류가 발생해도 완료된 작업은 자동으로 롤백되지 않습니다.

시작, 클러스터 편입, 동기화, switchover 확인 중 하나라도 실패하면 추가 작업을 중단하고 완료·실패·미처리 대상을 출력합니다.
원인을 해결한 뒤 같은 명령을 다시 실행하면 해당 상태를 기준으로 업그레이드 작업을 다시 수행합니다.

### 기존 명령과의 연계

`upgrade`를 지원하면서 클러스터 안에 여러 버전이 함께 존재할 수 있습니다. 이에 맞춰 기존 명령의 버전 처리도 다음과 같이 바뀝니다.

| 명령 | 변경 내용 |
|-|-|
| `start`     | 캐시 서버별로 기록된 버전으로 시작합니다. |
| `stop`      | PID 파일을 캐시 서버별로 기록된 버전 기반으로 찾은 후 중지합니다. |
| `delete`    | 실행 여부 확인과 `--purge` 대상 경로를 캐시 서버별 버전 기준으로 판단합니다. |
| `list`      | 버전 정보 표시를 제거합니다. |
| `status`    | 캐시 서버별 버전 정보를 표시합니다. |

#### `status` 출력

`status`는 캐시 서버별 버전을 `VERSION` 열에 표시합니다.

```text
Arcus cluster "my-repl-cluster"

GROUP  ROLE    ADDRESS           VERSION   PROCESS_STATUS  ZK_REGISTERED
g1     master  192.0.2.22:11211  0.12.2-E  running         yes
g1     slave   192.0.2.21:11211  0.12.2-E  running         yes
g2     master  192.0.2.23:11211  0.12.1-E  running         yes
g2     slave   192.0.2.24:11211  0.12.1-E  running         yes
```

- 실행 중인 캐시 서버는 프로세스에 `version` 명령을 보내 실제 실행 버전을 표시합니다.
- 중지된 캐시 서버는 메타데이터에 기록된 버전, 즉 다음 `start` 시 사용할 버전을 표시합니다.
- 실제 실행 버전과 기록된 버전이 다르면 경고를 함께 출력합니다.

## `cluster delete`

Arcus 클러스터를 삭제합니다.

### 명령 형식

```sh
arcusctl cluster delete <servicecode> [--purge]
```

- 기본 `delete`는 Arcus 클러스터를 제거하지만, 원격 장비의 버전별 설치 디렉터리는 보존합니다.
  - Arcus 클러스터의 ZNode를 삭제합니다.
  - 운영 장비에서 `meta.yml`과 `topology.yml`을 삭제합니다.
- `--purge` 옵션을 지정하면 운영 장비에 저장된 배포 정보를 기준으로 다른 클러스터가 같은 원격 장비의 `<path>/<version>`을 사용하는지 확인합니다.
  - 해당 설치를 공유하는 다른 클러스터가 없는 경우 설치 디렉터리까지 삭제합니다.


### 실행 예시

```sh
arcusctl cluster delete my-cluster
arcusctl cluster delete my-cluster --purge
```

### 삭제 범위

삭제하기 전에 모든 캐시 서버가 중지되어 있는지 확인하고 사용자에게 진행 여부를 확인합니다.

기본 `delete`는 ZooKeeper에서 다음 클러스터 등록 정보를 제거합니다. `<root>`는 Community 클러스터의 `/arcus` 또는 Enterprise 클러스터의 `/arcus_repl`을 의미합니다.

```text
<root>/cache_list/<servicecode>
<root>/client_list/<servicecode>
<root>/cache_server_mapping/<host>:<port>
```

Enterprise 클러스터는 다음 정보도 제거합니다.

```text
/arcus_repl/group_list/<servicecode>/<group-name>
```

기본 `delete`는 원격 장비의 버전별 설치 디렉터리를 보존합니다.

`--purge`로 추가되는 삭제 경로는 다음과 같습니다.

```text
<path>/<version>
```

운영 장비에서는 다음 디렉터리를 제거합니다.

```text
<home>/clusters/arcus/<servicecode>
```

기본 `delete`와 `delete --purge` 모두 운영 장비의 `images`에 저장된 아카이브를 삭제하지 않습니다.

```text
<home>/images/arcus-community/arcus-memcached-<version>.tar.gz
<home>/images/arcus-enterprise/arcus-memcached-<version>.tar.gz
```

> [!CAUTION]
> 삭제 작업은 자동으로 롤백되지 않습니다. 작업 중 오류가 발생하면 ZooKeeper 등록 정보나 운영 장비의 관리 정보가 일부만 삭제된 상태로 남을 수 있습니다.
>
> `--purge`를 실행하기 전에 설치 경로가 다른 클러스터나 프로그램에서 사용되지 않는지 확인하세요.

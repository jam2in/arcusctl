# arcusctl 시작하기

이 문서는 ZooKeeper 앙상블과 Arcus 클러스터를 처음 배포하기 전에 준비해야 할 실행 환경과 arcusctl의 상태 관리 방식을 설명합니다.

## 동작 방식

arcusctl은 원격 장비에 별도 에이전트를 설치하지 않습니다. 운영 장비에서 SSH와 SCP를 사용해 토폴로지에 선언된 원격 장비를 직접 관리합니다.

```text
운영 장비
  ├─ SSH/SCP ───────> ZooKeeper 서버가 실행되는 원격 장비
  ├─ SSH/SCP ───────> 캐시 서버가 실행되는 원격 장비
  └─ ZooKeeper 연결 ─> Arcus 클러스터 ZNode 등록·조회·삭제
```

## 용어

| 용어           | 의미                                                       |
|----------------|------------------------------------------------------------|
| 운영 장비      | arcusctl을 실행하고 메타데이터 및 아카이브를 저장하는 장비 |
| 원격 장비      | arcusctl이 SSH와 SCP로 관리하는 장비                       |
| ZooKeeper 서버 | `myid`로 구분되는 ZooKeeper 프로세스                       |
| 캐시 서버      | `<host>:<port>`로 구분되는 Arcus memcached 프로세스        |

## 실행 환경 준비

### 운영 장비

arcusctl을 실행하는 운영 장비에는 다음 항목이 필요합니다.

| 항목   | 용도                                           |
|--------|------------------------------------------------|
| `ssh`  | 원격 장비에서 명령 실행                        |
| `scp`  | 아카이브와 설정 파일 전송                      |
| `wget` | ZooKeeper 및 Arcus Community 아카이브 다운로드 |

arcusctl은 운영체제에 설치된 OpenSSH의 `ssh`, `scp` 명령을 사용하므로 현재 사용자의 `~/.ssh/config` 설정 항목이 원격 장비 연결에 적용됩니다.

또한, SSH 연결을 비대화형으로 실행하므로 배포하기 전에 다음 명령이 별도의 사용자 입력 없이 성공하는지 확인하세요.

```sh
ssh -o BatchMode=yes -o ConnectTimeout=10 <host> hostname
```

> [!NOTE]
> 토폴로지의 `host`에는 IP 주소 또는 hostname을 사용할 수 있습니다.
> hostname을 사용하는 경우 운영 장비와 해당 주소를 사용하는 모든 원격 장비에서 이름을 해석할 수 있고 서로 통신할 수 있어야 합니다.

### 원격 장비

모든 원격 장비에는 OpenSSH 설정에 지정된 포트로 접속할 수 있어야 합니다. 별도의 Port 설정이 없다면 기본 22번 포트를 사용합니다.

공통 요구사항은 다음과 같습니다.

- 토폴로지에 지정한 설치 경로를 만들고 수정할 수 있는 권한
- 아카이브 압축 해제를 위한 `tar`
- 토폴로지에 선언한 포트를 사용할 수 있는 환경

대상별 추가 요구사항은 다음과 같습니다.

- ZooKeeper 서버가 실행되는 원격 장비
    - ZooKeeper 버전과 호환되는 Java 런타임
    - 토폴로지의 `path`, `data_dir`을 만들고 수정할 수 있는 권한
    - `data_log_dir`을 별도로 지정했다면 해당 경로를 만들고 수정할 수 있는 권한

- 캐시 서버가 실행되는 원격 장비
    - 소스 아카이브의 의존성을 설치할 수 있는 환경
    - `configure`, `make`, `make install`을 실행할 수 있는 셸, 컴파일러 및 빌드 도구

## 운영 장비의 상태 및 아카이브

`home`은 arcusctl이 운영 장비에 아카이브와 배포 상태를 저장하는 기준 디렉터리입니다. 기본값은 `~/.arcusctl`입니다.

설정 파일의 `home` 또는 `ARCUSCTL_HOME` 환경 변수로 다른 경로를 지정할 수 있습니다.

```text
~/.arcusctl/
├── images/
│   ├── arcus-community/
│   │   └── arcus-memcached-<version>.tar.gz
│   ├── arcus-enterprise/
│   │   └── arcus-memcached-<version>.tar.gz
│   └── zookeeper/
│       └── apache-zookeeper-<version>-bin.tar.gz
└── clusters/
    ├── arcus/
    │   └── <servicecode>/
    │       ├── meta.yml
    │       └── topology.yml
    └── zookeeper/
        └── <ensemble-name>/
            ├── meta.yml
            └── topology.yml
```

각 파일과 디렉터리의 용도는 다음과 같습니다.

- `images/`: 내려받거나 사용자가 준비한 아카이브
- `meta.yml`: 리소스 이름, 버전 및 배포 시간 등의 메타데이터
- `topology.yml`: `deploy`에 사용한 토폴로지 원본

> [!IMPORTANT]
> `start`, `stop`, `status`, `list`, `delete`는 `home`에 저장된 배포 정보를 기준으로 동작합니다.
> 원본 토폴로지 파일을 수정해도 이미 등록된 앙상블이나 클러스터에는 반영되지 않습니다.

## 원격 장비의 설치 구조

토폴로지의 `path`는 설치 기준 경로이며, Arcus Memcached와 ZooKeeper 바이너리는 버전별 디렉터리에 설치됩니다.

**Arcus Memcached**

```text
<path>/
└── <version>/
    ├── bin/
    ├── lib/
    ├── src/
    └── memcached-<port>.pid
```

**ZooKeeper**

```text
<path>/
├── <version>/
│   ├── bin/
│   ├── conf/
│   ├── lib/
│   └── logs/
├── conf/
│   └── <ensemble-name>/
│       └── zk<myid>/
│           ├── zoo.cfg
│           └── zoo.cfg.dynamic
└── data/
    └── <ensemble-name>/
        └── zk<myid>/
```

ZooKeeper의 `data_dir`을 생략하면 `<path>/data/<ensemble-name>`을 사용합니다.
`data_dir` 또는 `data_log_dir`을 직접 지정하면 각 경로 아래에 `zk<myid>` 디렉터리를 생성합니다.

## 기본 운영 흐름

ZooKeeper와 Arcus 클러스터 모두 설치와 프로세스 실행이 분리되어 있습니다.

```text
토폴로지 작성
  → deploy
  → start
  → status / list
  → stop
  → delete
```

1. 토폴로지 YAML 파일을 작성합니다.
2. `deploy`로 아카이브, 설정 및 관리 정보를 배포합니다.
3. `start`로 프로세스를 실행합니다.
4. `status`로 ZooKeeper 서버 또는 캐시 서버의 상태를 확인합니다.
5. 필요한 경우 `list`로 현재 운영 장비에서 관리하는 리소스 목록을 확인합니다.
6. 삭제하기 전에 `stop`을 실행하고 `status`로 모든 프로세스가 중지되었는지 확인합니다.
7. 리소스 종류와 삭제 범위를 확인한 뒤 `delete`를 실행합니다.

## 삭제 시 주의사항

> [!NOTE]
> `delete`와 `delete --purge` 모두 운영 장비의 `images`에 저장된 아카이브를 삭제하지 않습니다.

기본 `delete`는 운영 장비의 관리 정보와 리소스별 설정을 삭제하지만, 원격 장비의 버전별 설치 디렉터리는 보존합니다.
`--purge` 옵션을 지정하면 다른 리소스와 공유하지 않는 설치 디렉터리까지 삭제합니다.

자세한 토폴로지 작성 방법과 명령별 동작은 다음 문서를 참고하세요.

- [ZooKeeper 운영 가이드](zk-guide.md)
- [Arcus 클러스터 운영 가이드](cluster-guide.md)

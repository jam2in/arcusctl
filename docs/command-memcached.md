# Memcached 서버 관리 가이드

`memcached` 명령은 서비스 코드 단위로 Arcus Memcached 서버 목록과 전역 설정을 관리하는 명령 그룹입니다. 명령 실행 전 ZooKeeper 연결 설정이 필요하며, 설정 방법은 [설정 가이드](./config-file.md)를 참고하세요.

> [!NOTE]
> 현재 코드에서는 `memcached` 명령 그룹이 루트 명령에 연결되어 있지 않습니다. 이 문서는 구현된 하위 명령의 사용 흐름을 기준으로 정리한 운영 문서이며, 명령을 노출하려면 루트 명령 등록이 필요합니다.

## 기본 사용 패턴

명령 형식:
```sh
arcusctl --config-file <config.yaml> memcached <subcommand> [arguments...]
```

실행 예시:
```sh
./arcusctl --config-file ./config.yaml memcached list
```

서비스 코드는 하나의 Arcus 캐시 클러스터를 구분하는 논리 이름입니다. 서버 주소는 `<ip>:<port>` 형식으로 지정합니다.

## 1. 서비스 코드 전역 설정

명령 형식:
```sh
arcusctl memcached config <serviceCode> [options...]
```

실행 예시:
```sh
./arcusctl memcached config sample '-E /path/to/engine.so -m 1024 -c 10000'
# Global config for service 'sample' has been updated
```

전역 설정은 서비스 코드에 속한 Memcached 서버를 시작할 때 사용할 옵션 문자열입니다.

## 2. 서버 등록

명령 형식:
```sh
arcusctl memcached add <serviceCode> <ip:port>
```

실행 예시:
```sh
./arcusctl memcached add sample 10.0.0.11:11211
# Successfully added server 10.0.0.11:11211 to service code sample
```

## 3. 서버 목록 조회

전체 서비스 코드 요약 조회 명령 형식:
```sh
arcusctl memcached list
```

실행 예시:
```sh
./arcusctl memcached list
# SERVICE CODE              TOTAL    ONLINE   OFFLINE
# ------------------------------------------------------------
# sample                    2        1        1
```

특정 서비스 코드 상세 조회 명령 형식:
```sh
arcusctl memcached list <serviceCode>
```

실행 예시:
```sh
./arcusctl memcached list sample
```

## 4. 서버 시작

전체 서버 시작 명령 형식:
```sh
arcusctl memcached start <serviceCode>
```

실행 예시:
```sh
ARCUS_PATH=/opt/arcus ./arcusctl memcached start sample
#   - Start command sent to 10.0.0.11:11211 successfully.
```

일부 서버 시작 명령 형식:
```sh
arcusctl memcached start <serviceCode> <ip:port>
```

실행 예시:
```sh
ARCUS_PATH=/opt/arcus ./arcusctl memcached start sample 10.0.0.11:11211
```

> [!CAUTION]
> 서버 시작은 `ARCUS_PATH` 환경 변수와 ZooKeeper 주소를 사용합니다. 실행 환경에서 Arcus 설치 경로와 ZooKeeper 주소가 올바르게 설정되어 있는지 확인해야 합니다.

## 5. 서버 중지

전체 서버 중지 명령 형식:
```sh
arcusctl memcached stop <serviceCode>
```

실행 예시:
```sh
ARCUS_PATH=/opt/arcus ./arcusctl memcached stop sample
#   - Stop command sent to 10.0.0.11:11211 successfully.
```

일부 서버 중지 명령 형식:
```sh
arcusctl memcached stop <serviceCode> <ip:port>
```

실행 예시:
```sh
ARCUS_PATH=/opt/arcus ./arcusctl memcached stop sample 10.0.0.11:11211
```

## 6. 서버 또는 서비스 코드 제거

서버 제거 명령 형식:
```sh
arcusctl memcached remove <serviceCode> <ip:port>
```

실행 예시:
```sh
./arcusctl memcached remove sample 10.0.0.11:11211
# Successfully remove to service code sample
```

서비스 코드 제거 명령 형식:
```sh
arcusctl memcached remove <serviceCode>
```

실행 예시:
```sh
./arcusctl memcached remove sample
# Successfully remove to service code sample
```

> [!CAUTION]
> 서비스 코드 제거는 해당 서비스 코드의 서버 목록과 설정을 ZooKeeper에서 삭제하는 작업입니다. 운영 환경에서는 제거 대상 서비스 코드와 서버 주소를 다시 확인하세요.

# Arcus ACL 운영 가이드

이 문서는 arcusctl을 사용해 Arcus ACL 그룹과 사용자를 생성, 조회, 변경 및 삭제하는 방법을 설명합니다.

> [!IMPORTANT]
> 이 문서는 SASL 인증과 ZooKeeper 기반 ACL 연동이 구성된 Arcus 환경을 대상으로 합니다.
>
> Arcus 캐시 서버는 SASL 지원을 포함하여 빌드되어야 하며, 실행 시 SASL 인증이 활성화되어 있어야 합니다.
> 또한 ZooKeeper에 서비스코드와 ACL 그룹의 매핑 정보가 구성되어 있어야 합니다.

Arcus ACL의 개념과 동작 방식은
[캐시 서버 문서](https://github.com/naver/arcus-memcached/blob/master/docs/administration/sasl.md)를 참고하세요.

## 사전 준비

### ACL 사용 환경

ACL 명령은 ZooKeeper의 ACL 그룹과 사용자 정보를 직접 관리합니다.
이 명령은 ZooKeeper의 정보만 변경하며 캐시 서버의 SASL 인증을 활성화하거나 서비스코드와 ACL 그룹을 연결하지 않습니다.

### ZooKeeper 연결 설정

ACL 명령은 기본적으로 `localhost:2181`의 ZooKeeper에 연결합니다. 
다른 ZooKeeper 앙상블에 연결하려면 설정 파일이나 `ARCUSCTL_ZOOKEEPER` 환경 변수로 연결 정보를 지정할 수 있습니다.

### 설정 파일

설정 파일의 `zookeeper`에 ZooKeeper 앙상블 주소를 지정합니다. 저장소의 [설정 파일 예시](../examples/config.yaml)를 복사한 뒤 환경에 맞게 수정할 수 있습니다.

```yaml
zookeeper: 192.0.2.11:2181,192.0.2.12:2181,192.0.2.13:2181
```

### 설정 파일 탐색 순서

`--config-file` 옵션을 사용하면 설정 파일을 직접 지정할 수 있습니다.

```sh
arcusctl --config-file /path/to/config.yaml acl group list
```

`--config-file`을 생략하면 다음 순서로 `config.yaml` 파일을 탐색합니다.

1. `arcusctl` 실행 파일이 있는 디렉터리
2. 현재 작업 디렉터리

실행 파일 디렉터리와 현재 작업 디렉터리에 설정 파일이 모두 있으면 실행 파일 디렉터리의 설정 파일이 우선 적용됩니다.

### 환경 변수

`ARCUSCTL_ZOOKEEPER` 환경 변수가 설정되어 있으면 설정 파일의 `zookeeper` 값보다 환경 변수의 연결 정보를 우선 사용합니다.

```sh
ARCUSCTL_ZOOKEEPER="192.0.2.11:2181,192.0.2.12:2181,192.0.2.13:2181" \
arcusctl acl group list
```

## 운영 흐름

ACL 그룹과 사용자는 다음 순서로 생성하고 관리합니다.

```text
ZooKeeper 연결 설정 -> 그룹 생성 -> 사용자 생성 -> 조회 및 변경 -> 사용자 삭제 -> 그룹 삭제
```

그룹을 생성할 때 그룹 관리자를 지정합니다. 이후 사용자의 생성, 비밀번호 변경, 권한 변경 및 삭제에는 해당 관리자의 인증 정보가 필요합니다.

> [!IMPORTANT]
> 그룹에 사용자가 남아 있으면 그룹을 삭제할 수 없습니다. 그룹을 삭제하기 전에 모든 사용자를 먼저 삭제하세요.

## 명령 요약

| 명령                                                                    | 주요 동작                         |
|-------------------------------------------------------------------------|-----------------------------------|
| `arcusctl acl group add <group-name>`                                   | ACL 그룹 생성 및 그룹 관리자 설정 |
| `arcusctl acl group list`                                               | ACL 그룹 목록 출력                |
| `arcusctl acl admin passwd <group-name>`                                | 그룹 관리자 비밀번호 변경         |
| `arcusctl acl group remove <group-name>`                                | 비어 있는 ACL 그룹 삭제           |
| `arcusctl acl user add <group-name> <user-name> <permissions> [logAll]` | ACL 사용자와 권한 생성            |
| `arcusctl acl user list <group-name>`                                   | 그룹에 속한 사용자와 권한 출력    |
| `arcusctl acl user passwd <group-name> <user-name>`                     | 사용자 비밀번호 변경              |
| `arcusctl acl user permissions <group-name> <user-name> <permissions>`  | 사용자 권한 변경                  |
| `arcusctl acl user remove <group-name> <user-name>`                     | ACL 사용자 삭제                   |

## 그룹 관리

### `acl group add`

새 ACL 그룹을 생성하고 그룹 관리자를 설정합니다.

#### 명령 형식

```sh
arcusctl acl group add <group-name>
```

#### 실행 예시

```sh
arcusctl acl group add sample
```

그룹을 관리할 관리자 이름과 비밀번호를 입력합니다. 비밀번호는 화면에 표시되지 않습니다.

```text
admin name: manager
password:
repeat password:
OK
```

그룹 관리자는 해당 그룹의 사용자를 생성하거나 변경하고 삭제할 때 사용됩니다.

### `acl group list`

ZooKeeper에 저장된 ACL 그룹 목록을 출력합니다.

#### 명령 및 실행 예시

```sh
arcusctl acl group list
```

```text
  * sample
  * dev
  * prod
Total: 3
```

### `acl admin passwd`

그룹 관리자 비밀번호를 변경합니다.

#### 명령 형식

```sh
arcusctl acl admin passwd <group-name>
```

#### 실행 예시

```sh
arcusctl acl admin passwd sample
```

관리자 이름과 기존 비밀번호를 입력한 뒤 새로운 비밀번호를 두 번 입력합니다.

```text
admin name: manager
admin password:
new password:
repeat new password:
OK
```

관리자 비밀번호를 변경하면 해당 그룹과 그룹에 속한 모든 사용자 ZNode의 ZooKeeper ACL이 새로운 비밀번호로 변경됩니다.

### `acl group remove`

사용자가 없는 ACL 그룹을 삭제합니다.

#### 명령 형식

```sh
arcusctl acl group remove <group-name>
```

#### 실행 예시

```sh
arcusctl acl group remove sample
```

```text
OK
```

그룹에 사용자가 남아 있으면 삭제에 실패합니다. 이 명령은 관리자 인증 정보나 별도의 확인을 요청하지 않고 그룹을 즉시 삭제합니다.

## 사용자 관리

사용자의 생성, 비밀번호 변경, 권한 변경 및 삭제는 그룹 관리자만 수행할 수 있습니다.

### `acl user add`

그룹에 새로운 사용자와 권한을 추가합니다.

#### 명령 형식

```sh
arcusctl acl user add <group-name> <user-name> <permissions> [logAll]
```

`permissions`에는 쉼표로 구분한 권한 목록을 지정합니다. 사용할 수 있는 권한은
[캐시 서버 문서의 Permissions](https://github.com/naver/arcus-memcached/blob/master/docs/administration/sasl.md#권한-부여)를 참고하세요.

마지막 인자에 `logAll`을 지정하면 해당 사용자가 수행한 명령을 감사 로그에 기록하도록 설정합니다.

#### 실행 예시

```sh
arcusctl acl user add sample app kv,list,set,map,btree,attr,scan,flush
```

그룹 관리자 인증 정보와 새 사용자의 비밀번호를 입력합니다.

```text
admin name: manager
admin password:
user password:
repeat user password:
OK
```

모든 명령을 감사 로그에 기록할 사용자는 다음과 같이 생성합니다.

```sh
arcusctl acl user add sample operator attr,scan,flush,admin logAll
```

### `acl user list`

특정 그룹에 속한 사용자와 권한을 출력합니다.

#### 명령 형식

```sh
arcusctl acl user list <group-name>
```

#### 실행 예시

```sh
arcusctl acl user list sample
```

```text
  * app [kv list set map btree attr scan flush]
  * operator [attr scan flush admin] logAll
Total: 2
```

### `acl user passwd`

기존 사용자의 비밀번호를 변경합니다.

#### 명령 형식

```sh
arcusctl acl user passwd <group-name> <user-name>
```

#### 실행 예시

```sh
arcusctl acl user passwd sample app
```

그룹 관리자 인증 정보와 사용자의 새로운 비밀번호를 입력합니다.

```text
admin name: manager
admin password:
user password:
repeat user password:
OK
```

### `acl user permissions`

기존 사용자의 권한을 새로 설정합니다.

#### 명령 형식

```sh
arcusctl acl user permissions <group-name> <user-name> <permissions>
```

#### 실행 예시

```sh
arcusctl acl user permissions sample app kv,attr
```

```text
admin name: manager
admin password:
OK
```

사용자 생성 시 `logAll`을 지정했다면 권한을 변경해도 해당 설정은 유지됩니다. `acl user permissions`로 `logAll` 설정을 추가하거나 제거할 수는 없습니다.

### `acl user remove`

기존 사용자를 그룹에서 삭제합니다.

#### 명령 형식

```sh
arcusctl acl user remove <group-name> <user-name>
```

#### 실행 예시

```sh
arcusctl acl user remove sample app
```

```text
admin name: manager
admin password:
OK
```

사용자가 ACL 그룹과 연결된 서비스의 캐시 서버에 접속해 있으면 삭제할 수 없습니다. 해당 사용자의 모든 연결을 종료한 후 다시 실행하세요.

## 비밀번호 규칙

arcusctl에 입력하는 비밀번호는 다음 조건을 충족해야 합니다.

- 12자 이상
- 다음 종류 중 세 가지 이상 포함
    - 알파벳 대문자
    - 알파벳 소문자
    - 숫자
    - 특수문자

새 비밀번호와 확인용 비밀번호가 일치하지 않거나 비밀번호 규칙을 충족하지 않으면 명령이 실패합니다.

## 운영 시 주의사항

- ACL 명령은 지정한 ZooKeeper의 정보를 직접 변경하며 자동으로 롤백하지 않습니다. 명령을 실행하기 전에 연결 대상이 올바른지 확인하세요.
- `acl group remove`는 관리자 인증이나 별도의 확인 절차 없이 그룹을 즉시 삭제합니다.
- `acl group remove`는 서비스코드와 ACL 그룹의 매핑 정보를 삭제하지 않습니다. 그룹을 삭제한 후 관련 매핑 정보를 별도로 확인하세요.
- `acl user remove`를 실행하기 전에 해당 사용자의 캐시 서버 연결이 모두 종료되었는지 확인하세요.
- 사용자 권한은 쉼표로 구분하고 공백 없이 입력하세요.

# Arcus ACL 사용 가이드

Arcus ACL 개요에 대한 설명은 [캐시 서버 문서](https://github.com/naver/arcus-memcached/blob/master/docs/administration/sasl.md) 참고 바랍니다.

## 0) 사전 준비

Arcus ACL 정보를 저장, 관리하는 ZooKeeper에 연결하기 위하여, [설정 파일 또는 환경 변수](./config-file.md)를 통해 zookeeper 주소를 지정해야 합니다.

## 1) 그룹 관리

### 1-1) 그룹 생성 - `acl group add`

새로운 그룹을 생성하는 명령입니다.
```
./arcusctl acl group add <groupname>
```

아래는 `sample` 그룹을 생성하는 예시입니다.
```sh
./arcusctl acl group add sample
# admin name: alice
# password: 
# repeat password: 
# OK
```
- 그룹 생성 시에 그룹 관리자의 username과 password를 설정해야 합니다.
- 생성된 그룹에는 그룹 관리자만 사용자를 추가/제거할 수 있습니다.

### 1-2) 그룹 목록 확인 - `acl group list`

ZooKeeper에 저장된 전체 그룹 목록을 확인하는 명령입니다.
```
./arcusctl acl group list
```

아래는 수행 예시로, `sample`, `dev`, `prod` 그룹이 조회되었습니다.
```sh
./arcusctl acl group list
#   * sample
#   * dev
#   * prod
# Total: 3
```

### 1-3) 관리자 비밀번호 변경 - `acl admin passwd`

그룹의 관리자 비밀번호를 변경하는 명령입니다.
```
./arcusctl acl admin passwd <groupname>
```

아래는 `sample` 그룹의 관리자 비밀번호를 변경하는 예시입니다.
```sh
./arcusctl acl admin passwd sample
# admin name: alice
# password:
# new admin password: 
# repeat new admin password:
# OK
```
- 기존 password 입력 후에 새로운 password를 2회 입력 받습니다.


### 1-4) 그룹 제거 - `acl group remove`

그룹을 제거하는 명령입니다.

```
./arcusctl acl group remove <groupname>
```

아래는 `sample` 그룹을 제거하는 예시입니다.
```sh
./arcusctl acl group remove sample
# OK
```
- 해당 그룹의 하위에 사용자가 아무도 없는 상태여야 합니다.
- 빈 그룹을 제거하기 위해 관리자 username / password를 요구하지 않습니다.

## 2) 사용자 관리

사용자 관리는 그룹의 관리자만 수행할 수 있습니다.

### 2-1) 사용자 추가 - `acl user add`

그룹에 사용자를 추가하는 명령입니다. 사용자 권한인 `<permissions>`은 반드시 지정해야 하며, `logAll` 지정은 선택입니다.

```
./arcusctl acl user add <groupname> <username> <permissions> [logAll]
```

아래는 `sample` 그룹에 `app`과 `operator` 사용자 계정을 추가하는 예시입니다.

```sh
./arcusctl acl user add sample app kv,list,set,map,btree,attr,scan,flush
# admin name: alice
# admin password: 
# user password: 
# repeat user password: 
# OK

./arcusctl acl user add sample operator attr,scan,flush,admin logAll
# admin name: alice
# admin password: 
# user password: 
# repeat user password: 
# OK
```
- 사용자 계정에 부여할 권한을 `,`으로 연결하여 전달합니다. 권한 목록은 [캐시 서버 문서](https://github.com/naver/arcus-memcached/blob/master/docs/administration/sasl.md#permissions)에서 확인할 수 있습니다.
- `logAll` 속성을 추가로 지정하면 해당 계정으로 서버에서 수행하는 모든 명령 이력이 감사 로그로 남게 됩니다.

### 2-2) 사용자 목록 확인 - `acl user list`

특정 그룹에 속한 사용자 목록을 조회하는 명령입니다.
```
./arcusctl acl user list <groupname>
```

아래는 `sample` 그룹에 있는 사용자 목록을 조회하는 예시입니다.
```sh
./arcusctl acl user list sample
#   * app [kv list set map btree attr flush]
#   * operator [attr scan flush admin] logAll
# Total: 2
```

### 2-3) 사용자 비밀번호 변경 - `acl user passwd`

기존 사용자의 비밀번호를 변경하는 명령입니다.

```
./arcusctl acl user passwd <groupname> <username>
```

아래는 `sample` 그룹에서 `app` 계정의 비밀번호를 변경하는 예시입니다.

```sh
./arcusctl acl user passwd sample app
# admin name: alice
# admin password: 
# user password: 
# repeat user password: 
# OK
```

**[주의] ARCUS 서버에 해당 사용자 계정의 연결이 남아 있으면 변경할 수 없고, 아래 오류를 출력합니다.**

```
panic: found namsic in /arcus/client_list/test/devbox_10.178.0.39_9_c_1.15.0-11_20251119064221_1000f221c46000e_sasluser=namsic
```

### 2-4) 사용자 권한 변경 - `acl user permissions`

기존 사용자의 사용 권한을 새로 설정하는 명령입니다.

```
./arcusctl acl user passwd <groupname> <username> <permissions>
```

아래는 `sample` 그룹의 `app` 사용자 계정의 권한을 `kv,attr`로 변경하는 예시입니다.

```sh
./arcusctl acl user permissions sample app kv,attr
# admin name: alice
# admin password: 
# OK
```
- `logAll` 설정은 변경할 수 없으며, 사용자 생성 시에만 설정할 수 있다.

**[주의] ARCUS 서버에 해당 사용자 계정의 연결이 남아 있으면 변경할 수 없고, 아래 오류를 출력합니다.**

```
panic: found namsic in /arcus/client_list/test/devbox_10.178.0.39_9_c_1.15.0-11_20251119064221_1000f221c46000e_sasluser=namsic
```

### 2-5) 사용자 제거 - `acl user remove`

기존 사용자를 제거하는 명령입니다.

```
./arcusctl acl user remove <groupname> <username>
```

아래는 `sample` 그룹의 `app` 사용자 계정을 제거하는 예시입니다.

```sh
./arcusctl acl user remove sample app
# admin name: alice
# admin password: 
# OK
```

**[주의] ARCUS 서버에 해당 사용자 계정의 연결이 남아 있으면 제거할 수 없고, 아래 오류를 출력합니다.**

```
panic: found namsic in /arcus/client_list/test/devbox_10.178.0.39_9_c_1.15.0-11_20251119064221_1000f221c46000e_sasluser=namsic
```

## 3) 기타

### 3-1) 비밀번호 규칙
- 12자 이상
- 아래 중 3가지 이상 포함
  - 알파벳 대문자
  - 알파벳 소문자
  - 숫자
  - 특수기호

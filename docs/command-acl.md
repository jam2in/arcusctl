# Arcus ACL 관리 가이드

Arcus의 ACL 기본 개요에 대한 설명은 [캐시 서버 문서](https://github.com/naver/arcus-memcached/blob/master/docs/administration/sasl.md) 참고 바랍니다.

### 0) 사전 준비

Arcus ACL 정보를 저장, 관리하는 ZooKeeper에 연결하기 위하여, [설정 파일 또는 환경 변수](./config-file.md)를 통해 zookeeper 주소를 지정합니다.

### 1) 그룹 생성 - `acl group add`

```sh
./arcusctl acl group add sample
# admin name: alice
# password: 
# repeat password: 
# OK
```
- 새로운 그룹을 생성합니다.
- 그룹 관리자 username과 password를 설정해야 합니다.
- 생성된 그룹에는 관리자만 사용자를 추가/제거할 수 있습니다.

### 2) 그룹 목록 확인 - `acl group list`

```
./arcusctl acl group list
#   * sample
#   * dev
#   * prod
# Total: 3
```
- ZooKeeper에 저장된 전체 그룹 목록을 확인합니다.

### 3) 사용자 추가 - `acl user add`

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
- 그룹에 사용자를 추가합니다.
- 사용자에게 부여할 권한을 `,`으로 연결하여 전달합니다. 권한 목록은 [캐시 서버 문서](https://github.com/naver/arcus-memcached/blob/master/docs/administration/sasl.md#permissions)에서 확인할 수 있습니다.
- logAll 인자 추가로 지정하면 해당 계정으로 서버에서 수행하는 모든 명령 이력이 감사 로그로 남게 됩니다.

### 4) 사용자 목록 확인 - `acl user list`

```sh
./arcusctl acl user list sample
#   * app [kv list set map btree attr flush]
#   * operator [attr scan flush admin] logAll
# Total: 2
```
- 특정 그룹에 속한 사용자 목록을 조회합니다.

### 5) 사용자 비밀번호 변경 - `acl user passwd`

```sh
./arcusctl acl user passwd sample app
# admin name: alice
# admin password: 
# user password: 
# repeat user password: 
# OK
```
- 기존 사용자의 비밀번호를 변경합니다.

### 6) 사용자 권한 변경 - `acl user permissions`

```sh
./arcusctl acl user permissions sample app kv,attr
# admin name: alice
# admin password: 
# OK
```
- 기존 사용자의 권한을 변경합니다.
- `logAll` 설정은 변경할 수 없으며, 사용자 생성 시에만 설정할 수 있다.

### 7) 사용자 제거 - `acl user remove`

```sh
./arcusctl acl user remove sample app
# admin name: alice
# admin password: 
# OK
```
- 사용자를 그룹에서 제거합니다.

### 8) 관리자 비밀번호 변경 - `acl admin passwd`

```sh
./arcusctl acl admin passwd sample
# admin name: alice
# password:
# new admin password: 
# repeat new admin password:
# OK
```
- 그룹의 관리자 비밀번호를 변경합니다.

### 9) 그룹 제거 - `acl group remove`

```sh
./arcusctl acl group remove sample
# OK
```
- 그룹을 제거합니다.
- 해당 그룹 하위에 사용자가 아무도 없는 상태여야 합니다.
- 빈 그룹을 제거하기 위해 관리자 username / password를 요구하지 않습니다.

## 기타

### 비밀번호 규칙
- 12자 이상
- 아래 중 3가지 이상 포함
  - 알파벳 대문자
  - 알파벳 소문자
  - 숫자
  - 특수기호


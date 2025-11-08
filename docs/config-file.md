# 설정 가이드

## 설정 파일

`arcusctl` 도구는 아래 내용의 `config.yaml` 설정 파일을 사용합니다.

```yaml
zookeeper: "10.0.0.1:2181,10.0.0.2:2181,10.0.0.3:2181"
# Alternatively, you can use a domain address.
# zookeeper: "zookeeper.example.com:2181"
```

## 설정 파일 위치

`arcusctl` 도구는 다음 순서대로 설정 파일을 찾습니다:

1. 사용자 지정 경로 (`--config-file` 옵션)
   ```sh
   ./arcusctl --config-file /path/to/config.yaml
   ```

2. 실행 파일 디렉토리
   - `arcusctl` 바이너리가 위치한 디렉토리

3. 현재 작업 디렉토리
   - 명령어를 실행하는 디렉토리

> [!NOTE]
> `--config-file` 사용하여 파일 직접 지정하는 경우 제외하고 파일명은 `config.yaml`이어야 합니다.

## 환경 변수

`ARCUSCTL_` 접두사를 사용하여 환경 변수로 설정 값을 재정의할 수 있습니다.

**예시:**
```sh
ARCUSCTL_ZOOKEEPER="localhost:2181" ./arcusctl
```

환경 변수는 설정 파일의 값보다 우선적으로 적용됩니다.


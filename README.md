# Caddy v2 Response Logger

Caddy v2용 고급 HTTP 응답 로거 모듈입니다. 요청과 응답에 대한 상세한 로깅을 제공하며, 설정 가능한 다양한 옵션을 지원합니다.

## 기능

-   🔍 **상세한 요청/응답 로깅**: HTTP 메소드, 경로, 상태 코드, 응답 크기, 처리 시간 등
-   📊 **구조화된 로깅**: Zap 라이브러리를 사용한 JSON 형식의 구조화된 로그
-   ⚙️ **설정 가능한 옵션**: 로그 레벨, 본문 포함 여부, 특정 경로/상태 코드 제외 등
-   🚀 **고성능**: 최소한의 오버헤드로 효율적인 로깅
-   🎯 **선택적 로깅**: 특정 경로나 상태 코드 제외 가능
-   📋 **헤더 로깅**: 특정 헤더만 선택적으로 로깅 가능

## 설치

1. 이 저장소를 클론합니다:

```bash
git clone https://github.com/koorukuroo/caddy-response-logger.git
cd caddy-response-logger
```

2. 의존성을 설치합니다:

```bash
go mod tidy
```

3. Caddy와 함께 빌드합니다:

```bash
xcaddy build --with github.com/koorukuroo/caddy-response-logger
```

## 사용 방법

### JSON 설정

```json
{
    "apps": {
        "http": {
            "servers": {
                "default": {
                    "listen": [":8080"],
                    "routes": [
                        {
                            "handle": [
                                {
                                    "handler": "response_logger",
                                    "logger_name": "api_logger",
                                    "log_level": "info",
                                    "include_response_body": true,
                                    "max_body_size": 1048576,
                                    "skip_status_codes": [404, 304],
                                    "skip_paths": ["/health", "/metrics"],
                                    "include_headers": [
                                        "Authorization",
                                        "Content-Type"
                                    ]
                                },
                                {
                                    "handler": "static",
                                    "root": "/var/www/html"
                                }
                            ]
                        }
                    ]
                }
            }
        }
    }
}
```

### Caddyfile 설정

```
:8080 {
    response_logger {
        logger_name api_logger
        log_level info
        include_response_body
        max_body_size 1MB
        skip_status_codes 404 304
        skip_paths /health /metrics
        include_headers Authorization Content-Type
    }

    root * /var/www/html
    file_server
}
```

## 설정 옵션

| 옵션                    | 타입     | 기본값              | 설명                                  |
| ----------------------- | -------- | ------------------- | ------------------------------------- |
| `logger_name`           | string   | `"response_logger"` | 로거 이름                             |
| `log_level`             | string   | `"info"`            | 로그 레벨 (debug, info, warn, error)  |
| `include_request_body`  | bool     | `false`             | 요청 본문 포함 여부                   |
| `include_response_body` | bool     | `false`             | 응답 본문 포함 여부                   |
| `max_body_size`         | int      | `1048576`           | 로그에 포함할 최대 본문 크기 (바이트) |
| `skip_status_codes`     | []int    | `[]`                | 로깅을 제외할 HTTP 상태 코드          |
| `skip_paths`            | []string | `[]`                | 로깅을 제외할 경로                    |
| `include_headers`       | []string | `[]`                | 로그에 포함할 헤더 목록               |

## 로그 형식

로그는 JSON 형식으로 출력되며, 다음과 같은 필드를 포함합니다:

```json
{
    "level": "info",
    "ts": "2024-01-15T10:30:00.000Z",
    "logger": "response_logger",
    "msg": "GET /api/users → 200",
    "method": "GET",
    "path": "/api/users",
    "query": "limit=10&offset=0",
    "status": 200,
    "size": 1024,
    "duration": "15.5ms",
    "remote_addr": "192.168.1.100:52134",
    "user_agent": "Mozilla/5.0...",
    "referer": "https://example.com/",
    "headers": {
        "Authorization": "Bearer token...",
        "Content-Type": "application/json"
    },
    "response_body": "..."
}
```

## 사용 예제

### 기본 로깅

```
response_logger
```

### API 서버용 상세 로깅

```
response_logger {
    logger_name api_server
    log_level debug
    include_request_body
    include_response_body
    max_body_size 2MB
    include_headers Authorization Content-Type X-Request-ID
}
```

### 프로덕션 환경용 설정

```
response_logger {
    logger_name production
    log_level info
    skip_status_codes 200 301 302 304
    skip_paths /health /metrics /favicon.ico
    include_headers X-Real-IP X-Forwarded-For
}
```

## 성능 고려사항

-   본문 로깅을 활성화하면 메모리 사용량이 증가합니다
-   `max_body_size`를 적절히 설정하여 메모리 사용량을 제한하세요
-   프로덕션에서는 `skip_status_codes`와 `skip_paths`를 활용하여 불필요한 로그를 줄이세요
-   로그 레벨을 적절히 설정하여 로그 볼륨을 조절하세요

## 라이센스

MIT License

## 기여

이슈 제보와 풀 리퀘스트를 환영합니다!

package response_logger

import (
    "bytes"
    "context"
    "io"
    "log"
    "net/http"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    caddy.RegisterModule(ResponseLogger{})
}

type ResponseLogger struct{}

func (ResponseLogger) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.response_logger",
        New: func() caddy.Module { return new(ResponseLogger) },
    }
}

func (h ResponseLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    // 프록시 응답을 캡처할 기록기 생성
    rec := caddyhttp.NewResponseRecorder(w, nil, nil)
    buf := new(bytes.Buffer)
    rec.BufferWriter = io.MultiWriter(w, buf) // 복사 + 기록

    err := next.ServeHTTP(rec, r)
    if err != nil {
        return err
    }

    log.Printf("[ResponseLogger] %s %s -> %d, Body: %s\n", r.Method, r.URL.Path, rec.Status(), buf.String())

    return nil
}

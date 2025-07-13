package response_logger

import (
    "bytes"
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

type bodyCaptureWriter struct {
    http.ResponseWriter
    statusCode int
    body       bytes.Buffer
}

func (w *bodyCaptureWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
    w.body.Write(b) // 바디 복사
    return w.ResponseWriter.Write(b)
}

func (h ResponseLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    bw := &bodyCaptureWriter{ResponseWriter: w, statusCode: 200}

    err := next.ServeHTTP(bw, r)
    if err != nil {
        return err
    }

    log.Printf("[ResponseLogger] %s %s -> %d\nBody:\n%s\n", r.Method, r.URL.Path, bw.statusCode, bw.body.String())

    return nil
}

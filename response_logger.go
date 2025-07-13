package response_logger

import (
    "bytes"
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
    rec := caddyhttp.NewResponseRecorder(w, nil, nil)
    err := next.ServeHTTP(rec, r)
    if err != nil {
        return err
    }

    body := rec.Body()
    log.Printf("[ResponseLogger] %s %s -> %d\nBody:\n%s\n", r.Method, r.URL.Path, rec.Status(), string(body))
    return nil
}

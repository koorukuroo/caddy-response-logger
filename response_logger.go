package response_logger

import (
	"bytes"
	"log"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(ResponseLogger{})
	caddy.RegisterHandlerDirective("response_logger", parseCaddyfile)
}

// ResponseLogger is a simple middleware that logs the response body.
type ResponseLogger struct{}

func (ResponseLogger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.response_logger",
		New: func() caddy.Module { return new(ResponseLogger) },
	}
}

func (h ResponseLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Create a response recorder to capture status and body
	rec := caddyhttp.NewResponseRecorder(w, nil, nil)
	buf := new(bytes.Buffer)
	rec.BufferWriter = buf

	// Serve the next handler
	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}

	// Log the status and body
	log.Printf("[response_logger] %s %s -> %d\nBody:\n%s\n", r.Method, r.URL.Path, rec.Status(), buf.String())

	return nil
}

// parseCaddyfile makes response_logger usable in Caddyfile.
func parseCaddyfile(d *caddyfile.Dispenser) (caddyhttp.MiddlewareHandler, error) {
	return ResponseLogger{}, nil
}

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

// ResponseLogger implements a response handler module used within `handle_response`
type ResponseLogger struct{}

// CaddyModule returns the module information.
func (ResponseLogger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.copy_response.response_logger",
		New: func() caddy.Module { return new(ResponseLogger) },
	}
}

// Provision is a no-op
func (ResponseLogger) Provision(_ caddy.Context) error {
	return nil
}

// ResponseHandler processes the HTTP response
func (ResponseLogger) ResponseHandler(next caddyhttp.ResponseHandler) caddyhttp.ResponseHandler {
	return caddyhttp.ResponseHandlerFunc(func(w http.ResponseWriter, r *http.Request, resp *http.Response) error {
		var buf bytes.Buffer
		tee := io.TeeReader(resp.Body, &buf)

		resp.Body = io.NopCloser(tee)

		err := next.ServeHTTP(w, r, resp)

		log.Printf("[ResponseLogger] %s %s → %d\nBody:\n%s\n",
			r.Method, r.URL.Path, resp.StatusCode, buf.String())

		return err
	})
}

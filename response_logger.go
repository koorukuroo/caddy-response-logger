package response_logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(ResponseLogger{})
	httpcaddyfile.RegisterHandlerDirective("response_logger", parseCaddyfile)
}

// parseSize parses a size string (e.g., "1MB", "512KB", "2GB") and returns the size in bytes
func parseSize(sizeStr string) (int, error) {
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Convert to uppercase for case-insensitive matching
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Handle just numeric values (assume bytes)
	if val, err := strconv.Atoi(sizeStr); err == nil {
		return val, nil
	}

	// Extract number and unit
	var num string
	var unit string
	
	for i, char := range sizeStr {
		if char >= '0' && char <= '9' || char == '.' {
			num += string(char)
		} else {
			unit = sizeStr[i:]
			break
		}
	}

	if num == "" {
		return 0, fmt.Errorf("no numeric value found in size string: %s", sizeStr)
	}

	// Parse the numeric part
	val, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %s", num)
	}

	// Convert based on unit
	switch unit {
	case "B", "":
		return int(val), nil
	case "KB":
		return int(val * 1024), nil
	case "MB":
		return int(val * 1024 * 1024), nil
	case "GB":
		return int(val * 1024 * 1024 * 1024), nil
	case "TB":
		return int(val * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// parseCaddyfile parses the Caddyfile configuration for response_logger
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var rl ResponseLogger
	
	// Parse the Caddyfile configuration
	err := rl.UnmarshalCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}
	
	return &rl, nil
}

// ResponseLogger implements an HTTP middleware that logs response details
type ResponseLogger struct {
	// Logger name for structured logging
	LoggerName string `json:"logger_name,omitempty"`
	
	// Log level: debug, info, warn, error
	LogLevel string `json:"log_level,omitempty"`
	
	// Include request body in logs
	IncludeRequestBody bool `json:"include_request_body,omitempty"`
	
	// Include response body in logs
	IncludeResponseBody bool `json:"include_response_body,omitempty"`
	
	// Maximum body size to log (in bytes)
	MaxBodySize int `json:"max_body_size,omitempty"`
	
	// Skip logging for specific status codes
	SkipStatusCodes []int `json:"skip_status_codes,omitempty"`
	
	// Skip logging for specific paths
	SkipPaths []string `json:"skip_paths,omitempty"`
	
	// Headers to include in logs
	IncludeHeaders []string `json:"include_headers,omitempty"`
	
	logger *zap.Logger
}

// CaddyModule returns the module information.
func (ResponseLogger) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.response_logger",
		New: func() caddy.Module { return new(ResponseLogger) },
	}
}

// Provision sets up the module
func (rl *ResponseLogger) Provision(ctx caddy.Context) error {
	// Set defaults
	if rl.LoggerName == "" {
		rl.LoggerName = "response_logger"
	}
	if rl.LogLevel == "" {
		rl.LogLevel = "info"
	}
	if rl.MaxBodySize == 0 {
		rl.MaxBodySize = 1024 * 1024 // 1MB default
	}
	
	// Get logger
	rl.logger = ctx.Logger(rl)
	
	return nil
}

// responseWriter wraps http.ResponseWriter to capture response details
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.body != nil {
		rw.body.Write(b)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// ServeHTTP implements the middleware interface
func (rl *ResponseLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	// Check if we should skip logging for this path
	for _, skipPath := range rl.SkipPaths {
		if strings.Contains(r.URL.Path, skipPath) {
			return next.ServeHTTP(w, r)
		}
	}
	
	start := time.Now()
	
	// Read request body if needed
	var requestBody []byte
	if rl.IncludeRequestBody && r.Body != nil {
		requestBody, _ = io.ReadAll(io.LimitReader(r.Body, int64(rl.MaxBodySize)))
		r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	
	// Wrap response writer
	var responseBody *bytes.Buffer
	if rl.IncludeResponseBody {
		responseBody = &bytes.Buffer{}
	}
	
	rw := &responseWriter{
		ResponseWriter: w,
		statusCode:     200, // default status code
		body:           responseBody,
	}
	
	// Call next handler
	err := next.ServeHTTP(rw, r)
	
	// Check if we should skip logging for this status code
	for _, skipCode := range rl.SkipStatusCodes {
		if rw.statusCode == skipCode {
			return err
		}
	}
	
	// Prepare log fields
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
		zap.Int("status", rw.statusCode),
		zap.Int("size", rw.size),
		zap.Duration("duration", time.Since(start)),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("referer", r.Referer()),
	}
	
	// Add specific headers if requested
	if len(rl.IncludeHeaders) > 0 {
		headers := make(map[string]string)
		for _, headerName := range rl.IncludeHeaders {
			if value := r.Header.Get(headerName); value != "" {
				headers[headerName] = value
			}
		}
		if len(headers) > 0 {
			fields = append(fields, zap.Any("headers", headers))
		}
	}
	
	// Add request body if included
	if rl.IncludeRequestBody && len(requestBody) > 0 {
		fields = append(fields, zap.ByteString("request_body", requestBody))
	}
	
	// Add response body if included
	if rl.IncludeResponseBody && responseBody != nil && responseBody.Len() > 0 {
		body := responseBody.Bytes()
		if len(body) > rl.MaxBodySize {
			body = body[:rl.MaxBodySize]
		}
		fields = append(fields, zap.ByteString("response_body", body))
	}
	
	// Log based on level and status code
	message := fmt.Sprintf("%s %s → %d", r.Method, r.URL.Path, rw.statusCode)
	
	switch {
	case rw.statusCode >= 500:
		rl.logger.Error(message, fields...)
	case rw.statusCode >= 400:
		rl.logger.Warn(message, fields...)
	default:
		switch rl.LogLevel {
		case "debug":
			rl.logger.Debug(message, fields...)
		case "info":
			rl.logger.Info(message, fields...)
		case "warn":
			rl.logger.Warn(message, fields...)
		case "error":
			rl.logger.Error(message, fields...)
		default:
			rl.logger.Info(message, fields...)
		}
	}
	
	return err
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (rl *ResponseLogger) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "logger_name":
				if !d.Args(&rl.LoggerName) {
					return d.ArgErr()
				}
			case "log_level":
				if !d.Args(&rl.LogLevel) {
					return d.ArgErr()
				}
			case "include_request_body":
				rl.IncludeRequestBody = true
			case "include_response_body":
				rl.IncludeResponseBody = true
			case "max_body_size":
				var sizeStr string
				if !d.Args(&sizeStr) {
					return d.ArgErr()
				}
				var err error
				rl.MaxBodySize, err = parseSize(sizeStr)
				if err != nil {
					return d.Errf("invalid size: %v", err)
				}
			case "skip_status_codes":
				args := d.RemainingArgs()
				for _, arg := range args {
					var code int
					if _, err := fmt.Sscanf(arg, "%d", &code); err != nil {
						return d.Errf("invalid status code: %s", arg)
					}
					rl.SkipStatusCodes = append(rl.SkipStatusCodes, code)
				}
			case "skip_paths":
				rl.SkipPaths = append(rl.SkipPaths, d.RemainingArgs()...)
			case "include_headers":
				rl.IncludeHeaders = append(rl.IncludeHeaders, d.RemainingArgs()...)
			default:
				return d.Errf("unknown directive: %s", d.Val())
			}
		}
	}
	return nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*ResponseLogger)(nil)
	_ caddyhttp.MiddlewareHandler = (*ResponseLogger)(nil)
	_ caddyfile.Unmarshaler       = (*ResponseLogger)(nil)
)

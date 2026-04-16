package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger returns a Gin middleware that logs every inbound HTTP request
// using the standard library's structured logger (slog).
//
// Logged before the handler runs: method, path, client IP, and request body (truncated).
// Logged after the handler completes: status code, response size, and latency.
func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Read and re-inject the request body so downstream handlers can still use it.
		var reqBody string
		if ctx.Request.Body != nil {
			rawBody, err := io.ReadAll(ctx.Request.Body)
			if err == nil {
				reqBody = truncate(string(rawBody), 500)
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))
			}
		}

		slog.Info("inbound request",
			slog.String("method", ctx.Request.Method),
			slog.String("path", ctx.Request.URL.Path),
			slog.String("client_ip", ctx.ClientIP()),
			slog.String("body", reqBody),
		)

		// Wrap the response writer to capture the status code and body size.
		blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: ctx.Writer}
		ctx.Writer = blw

		ctx.Next()

		slog.Info("inbound response",
			slog.String("method", ctx.Request.Method),
			slog.String("path", ctx.Request.URL.Path),
			slog.Int("status", ctx.Writer.Status()),
			slog.Int("response_size", blw.body.Len()),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		)
	}
}

// bodyLogWriter wraps gin.ResponseWriter to capture the response body size.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// truncate shortens a string to maxLen characters for safe log output.
func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}

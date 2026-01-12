package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"
)

// Middleware provides HTTP middleware for the Da Vinci API.
type Middleware struct {
	logger Logger
}

// Logger interface for logging (allows using zap or standard log)
type Logger interface {
	Infof(format string, args ...interface{})
}

// StandardLogger wraps standard log package to implement Logger interface
type StandardLogger struct{}

func (l *StandardLogger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// NewMiddleware creates middleware with logger.
// If logger is nil, uses standard log package.
func NewMiddleware(logger Logger) *Middleware {
	if logger == nil {
		logger = &StandardLogger{}
	}
	return &Middleware{logger: logger}
}

// RequestLogger logs request details.
func (m *Middleware) RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// Generate a simple request ID (8 characters)
			reqIDBytes := make([]byte, 4)
			if _, err := rand.Read(reqIDBytes); err != nil {
				// Fallback to timestamp-based ID if crypto/rand fails
				reqIDBytes = []byte{byte(time.Now().UnixNano() & 0xFF), byte((time.Now().UnixNano() >> 8) & 0xFF), byte((time.Now().UnixNano() >> 16) & 0xFF), byte((time.Now().UnixNano() >> 24) & 0xFF)}
			}
			reqID := hex.EncodeToString(reqIDBytes)

			// Store request ID in context
			ctx := context.WithValue(r.Context(), "requestID", reqID)
			r = r.WithContext(ctx)

			// Get client IP
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = forwarded
			} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}

			// Log request
			m.logger.Infof(
				"%s %s uid=%s from=%s",
				r.Method,
				r.URL.Path,
				reqID,
				clientIP,
			)

			// Create a response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Log response
				duration := time.Since(start)
				m.logger.Infof(
					"%s %s uid=%s status=%d latency=%v",
				r.Method,
				r.URL.Path,
					reqID,
				rw.statusCode,
					duration,
				)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware enables CORS for SPA frontend.
// Note: CORS is typically handled by github.com/rs/cors in main.go,
// but this function is kept for compatibility and can be used if needed.
func (m *Middleware) CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter middleware (basic implementation).
func (m *Middleware) RateLimiter(maxRequests int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic rate limiting implementation
			// TODO: Implement actual rate limiting logic
			// For now, just pass through to next handler
			next.ServeHTTP(w, r)
		})
	}
}

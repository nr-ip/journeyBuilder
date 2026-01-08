package handlers

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// Middleware provides Echo middleware for the Da Vinci API.
type Middleware struct {
	logger *zap.SugaredLogger
}

// NewMiddleware creates middleware with logger.
func NewMiddleware(logger *zap.SugaredLogger) *Middleware {
	return &Middleware{logger: logger}
}

// RequestLogger logs request details.
func (m *Middleware) RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			reqID := uuid.New().String()[:8]
			c.Set("requestID", reqID)

			// Log request
			m.logger.Infof(
				"%s %s uid=%s from=%s",
				c.Request().Method,
				c.Request().URL.Path,
				reqID,
				c.RealIP(),
			)

			defer func() {
				duration := time.Since(start)
				status := c.Response().Status
				m.logger.Infof(
					"%s %s uid=%s status=%d latency=%v",
					c.Request().Method,
					c.Request().URL.Path,
					reqID,
					status,
					duration,
				)
			}()

			return next(c)
		}
	}
}

// CORSMiddleware enables CORS for SPA frontend.
func (m *Middleware) CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000", "*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.OPTIONS},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		ExposeHeaders: []string{"X-Request-ID"},
		MaxAge:        86400,
	})
}

// RateLimiter middleware (basic implementation).
func (m *Middleware) RateLimiter(maxRequests int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Basic rate limiting implementation
			// TODO: Implement actual rate limiting logic
			return next(c)
		}
	}
}

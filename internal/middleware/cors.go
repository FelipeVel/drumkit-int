package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware with explicit CORS policy.
// allowedOrigins is a comma-separated list (e.g. "https://app.example.com,https://admin.example.com").
// Pass "*" to allow all origins (development only).
func CORS(allowedOrigins string) gin.HandlerFunc {
	origins := splitTrimmed(allowedOrigins)
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	cfg := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: !containsWildcard(origins),
		MaxAge:           12 * time.Hour,
	}

	return cors.New(cfg)
}

func splitTrimmed(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" {
			return true
		}
	}
	return false
}

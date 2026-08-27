package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Secure Middleware
func Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.FullPath(), "/docs") {
			return
		}

		header := c.Writer.Header()
		header.Set("X-XSS-Protection", "1; mode=block")
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "SAMEORIGIN")
	}
}

package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var corsAllowMethods = strings.Join([]string{
	http.MethodHead,
	http.MethodGet,
	http.MethodPut,
	http.MethodPost,
	http.MethodDelete,
}, ",")

// Cors Middleware
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Writer.Header()
		header.Add("Vary", "Origin")

		preflight := c.Request.Method == http.MethodOptions
		origin := c.Request.Header.Get("Origin")

		if origin == "" {
			if !preflight {
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		header.Set("Access-Control-Allow-Origin", "*")
		if !preflight {
			return
		}

		header.Add("Vary", "Access-Control-Request-Method")
		header.Add("Vary", "Access-Control-Request-Headers")
		header.Set("Access-Control-Allow-Methods", corsAllowMethods)
		if h := c.Request.Header.Get("Access-Control-Request-Headers"); h != "" {
			header.Set("Access-Control-Allow-Headers", h)
		}
		c.AbortWithStatus(http.StatusNoContent)
	}
}

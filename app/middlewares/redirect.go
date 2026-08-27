package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func requestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	if scheme := c.Request.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if scheme := c.Request.Header.Get("X-Forwarded-Protocol"); scheme != "" {
		return scheme
	}
	if c.Request.Header.Get("X-Forwarded-Ssl") == "on" {
		return "https"
	}
	if scheme := c.Request.Header.Get("X-Url-Scheme"); scheme != "" {
		return scheme
	}
	return "http"
}

// HTTPSRedirect Middleware
func HTTPSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if requestScheme(c) != "https" {
			c.Redirect(http.StatusMovedPermanently, "https://"+c.Request.Host+c.Request.RequestURI)
			c.Abort()
		}
	}
}

// NonWWWRedirect Middleware
func NonWWWRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		if strings.HasPrefix(host, "www.") {
			c.Redirect(http.StatusMovedPermanently, requestScheme(c)+"://"+host[4:]+c.Request.RequestURI)
			c.Abort()
		}
	}
}

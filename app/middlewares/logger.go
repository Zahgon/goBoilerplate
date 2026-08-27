package middlewares

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

// Logger Middleware
func Logger() gin.HandlerFunc {
	out, err := os.Create("public/logs.txt")
	if err != nil {
		out = os.Stdout
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("Method=%s, Url=%q, Status=%d, Latency:%s \n",
				param.Method, param.Path, param.StatusCode, param.Latency)
		},
		Output: out,
	})
}

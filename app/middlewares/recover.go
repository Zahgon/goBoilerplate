package middlewares

import (
	"goBoilterplate/app/helpers"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Recover Middleware
func Recover() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(os.Stderr, func(c *gin.Context, _ interface{}) {
		helpers.ErrorJSON(c, http.StatusInternalServerError)
		c.Abort()
	})
}

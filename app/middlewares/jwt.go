package middlewares

import (
	"fmt"
	"goBoilterplate/app/helpers"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtAuthScheme = "Bearer "

// Jwt Middleware
func Jwt() gin.HandlerFunc {
	secret := os.Getenv("APP_KEY")

	return func(c *gin.Context) {
		if strings.Contains(c.FullPath(), "/login") {
			return
		}

		token, err := parseJwt(c, secret)
		if err != nil {
			jwtUnauthorized(c)
			return
		}

		c.Set("token", token)
	}
}

func parseJwt(c *gin.Context, secret string) (*jwt.Token, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, jwtAuthScheme) {
		return nil, fmt.Errorf("missing or malformed jwt")
	}

	tokenString := strings.TrimPrefix(authHeader, jwtAuthScheme)
	if tokenString == "" {
		return nil, fmt.Errorf("missing or malformed jwt")
	}

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func jwtUnauthorized(c *gin.Context) {
	helpers.JSON(c, 401, map[string]interface{}{
		"message": "Unauthorized",
	})
	c.Abort()
}

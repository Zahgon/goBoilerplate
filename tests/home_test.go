package tests

import (
	"encoding/json"
	"goBoilterplate/app/controllers"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const formContentType = "application/x-www-form-urlencoded"

func TestHome(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	controllers.Index(c)

	assert.Equal(t, 200, rec.Code)
}

type Login struct {
	Token string
}

var JWT *Login

func TestLogin(t *testing.T) {
	// User found

	f := make(url.Values)
	f.Set("email", "andres@teachlr.org")
	f.Set("password", "123456")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(f.Encode()))
	req.Header.Set("Content-Type", formContentType)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	controllers.Login(c)

	if err := json.NewDecoder(rec.Body).Decode(&JWT); err != nil {
		panic(err)
	}
	req.Body.Close()

	assert.Equal(t, 200, rec.Code)

	// User not found

	f = make(url.Values)
	f.Set("email", "andres@teachlr.org")
	f.Set("password", "1234567")

	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(f.Encode()))
	req.Header.Set("Content-Type", formContentType)
	rec = httptest.NewRecorder()

	c, _ = gin.CreateTestContext(rec)
	c.Request = req

	controllers.Login(c)

	assert.Equal(t, 404, rec.Code)
}

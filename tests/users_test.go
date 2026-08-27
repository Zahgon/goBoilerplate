package tests

import (
	"encoding/json"
	"goBoilterplate/app/controllers"
	"goBoilterplate/app/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var User *models.User

func TestUserList(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	controllers.UserList(c)

	assert.Equal(t, 200, rec.Code)
}

func TestUserStore(t *testing.T) {
	// New User

	f := make(url.Values)
	f.Set("username", "Andres Fuentes")
	f.Set("email", "andresf@manzanares.com.ve")
	f.Set("password", "123456")
	f.Set("role", "Admin")

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(f.Encode()))
	req.Header.Set("Content-Type", formContentType)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req

	controllers.UserStore(c)

	if err := json.NewDecoder(rec.Body).Decode(&User); err != nil {
		panic(err)
	}
	req.Body.Close()

	assert.Equal(t, 201, rec.Code)

	// Validation Error

	f = make(url.Values)
	f.Set("name", "Andres Fuentes")
	f.Set("password", "123456")
	f.Set("role", "Admin")

	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(f.Encode()))
	req.Header.Set("Content-Type", formContentType)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec = httptest.NewRecorder()

	c, _ = gin.CreateTestContext(rec)
	c.Request = req

	controllers.UserStore(c)

	assert.Equal(t, 422, rec.Code)
}

func TestUserShow(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(User.ID)}}

	controllers.UserShow(c)

	assert.Equal(t, 200, rec.Code)
}

func TestUserUpdate(t *testing.T) {
	// Update User

	f := make(url.Values)
	f.Set("username", "Andres Fuentes A")
	f.Set("email", "andresf@manzanares.com.ve")
	f.Set("password", "123456")
	f.Set("role", "Admin")

	req := httptest.NewRequest(http.MethodPut, "/api/users", strings.NewReader(f.Encode()))
	req.Header.Set("Content-Type", formContentType)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(User.ID)}}

	controllers.UserUpdate(c)

	assert.Equal(t, 200, rec.Code)

	// Validation Error

	f = make(url.Values)
	f.Set("name", "Andres Fuentes")
	f.Set("password", "123456")
	f.Set("role", "Admin")

	req = httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(f.Encode()))
	req.Header.Set("Content-Type", formContentType)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec = httptest.NewRecorder()

	c, _ = gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(User.ID)}}

	controllers.UserUpdate(c)

	assert.Equal(t, 422, rec.Code)
}

func TestUserDelete(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/", nil)
	req.Header.Set("Authorization", "Bearer "+JWT.Token)
	rec := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(User.ID)}}

	controllers.UserDelete(c)

	assert.Equal(t, 200, rec.Code)
}

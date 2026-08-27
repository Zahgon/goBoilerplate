package controllers

import (
	"fmt"
	"goBoilterplate/app/helpers"
	"goBoilterplate/app/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Index godoc
// @Summary Home Page
// @Description Display Home Page
// @Tags Home
// @Produce  json
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router / [get]
func Index(c *gin.Context) {
	helpers.JSON(c, 200, "Welcome to Echo")
}

// Login godoc
// @Summary Login
// @Description Login User in API
// @Tags Auth
// @Produce  json
// @Param email query string true "Email"
// @Param password query string true "Password"
// @Success 200 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/login [post]
func Login(c *gin.Context) {
	login := models.Login{}
	login.Email = c.Request.FormValue("email")
	login.Password = c.Request.FormValue("password")

	err := helpers.Validate(&login)
	if err != nil {
		helpers.JSON(c, 422, err)
		return
	}

	user := models.AuthLogin(login.Email, login.Password)
	if user != nil {
		token, err := helpers.AuthMakeToken(user)
		if err != nil {
			helpers.JSON(c, 500, "Server Error")
			return
		}
		helpers.JSON(c, 200, map[string]string{"token": token})
		return
	}
	helpers.JSON(c, 404, "Not Found")
}

// Logout godoc
// @Summary Logout
// @Description User Logout
// @Tags Auth
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {string} string
// @Failure 401 {string} string
// @Router /api/logout [get]
func Logout(c *gin.Context) {
	user := helpers.AuthGetUser(c)
	if user != nil {
		helpers.JSON(c, 200, user)
		return
	}
	helpers.JSON(c, 401, "Unauthorized")
}

// Test godoc
func Test(c *gin.Context) {
	req := c.Request
	format := `<code> Protocol: %s<br> Host: %s<br> Method: %s<br> Path: %s<br> </code>`
	c.Data(http.StatusOK, "text/html; charset=UTF-8",
		[]byte(fmt.Sprintf(format, req.Proto, req.Host, req.Method, req.URL.Path)))
}

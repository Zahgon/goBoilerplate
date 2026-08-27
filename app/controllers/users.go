package controllers

import (
	"goBoilterplate/app/helpers"
	"goBoilterplate/app/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserList godoc
// @Summary UserList
// @Description Listado de Usuarios
// @Tags User
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} models.User
// @Router /api/users [get]
func UserList(c *gin.Context) {
	users := models.UserProfileList()
	if users != nil {
		helpers.JSON(c, 200, users)
		return
	}
	helpers.JSON(c, 204, "No Content")
}

// UserStore godoc
// @Summary UserStore
// @Description Guardar datos de Usuario
// @Tags User
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "Username"
// @Param email query string true "Email"
// @Param role query string true "Role"
// @Param password query string true "Password"
// @Success 201 {object} models.User
// @Failure 422 {string} string
// @Failure 400 {string} string
// @Router /api/users [post]
func UserStore(c *gin.Context) {
	user := models.User{}
	user.Username = c.Request.FormValue("username")
	user.Email = c.Request.FormValue("email")
	user.Password = c.Request.FormValue("password")
	user.Role = c.Request.FormValue("role")

	err := helpers.Validate(&user)
	if err != nil {
		helpers.JSON(c, 422, err)
		return
	}

	res := models.UserStore(&user)
	if res {
		helpers.JSON(c, 201, user)
		return
	}
	helpers.JSON(c, 400, "Bad Request")
}

// UserShow godoc
// @Summary UserShow
// @Description Consultar Usuario
// @Tags User
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Id"
// @Success 200 {object} models.User
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /api/users/{id} [get]
func UserShow(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		user := models.UserProfileShow(id)
		if user != nil {
			helpers.JSON(c, 200, user)
			return
		}
		helpers.JSON(c, 404, "Not Found")
		return
	}
	helpers.JSON(c, 400, "Bad Request")
}

// UserUpdate godoc
// @Summary UserUpdate
// @Description Actualizar datos de Usuario
// @Tags User
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Id"
// @Param username query string true "Username"
// @Param email query string true "Email"
// @Param role query string true "Role"
// @Param password query string true "Password"
// @Success 200 {object} models.User
// @Failure 422 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /api/users/{id} [put]
func UserUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		user := models.UserShow(id)
		if user != nil {
			user.Username = c.Request.FormValue("username")
			user.Email = c.Request.FormValue("email")
			user.Password = c.Request.FormValue("password")
			user.Role = c.Request.FormValue("role")

			err := helpers.Validate(user)
			if err != nil {
				helpers.JSON(c, 422, err)
				return
			}

			res := models.UserUpdate(user)
			if res {
				helpers.JSON(c, 200, user)
				return
			}
		} else {
			helpers.JSON(c, 404, "Not Found")
			return
		}
	}
	helpers.JSON(c, 400, "Bad Request")
}

// UserDelete godoc
// @Summary UserDelete
// @Description Borrado de Usuario
// @Tags User
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "Id"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Router /api/users/{id} [delete]
func UserDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err == nil {
		res := models.UserDelete(id)
		if res {
			helpers.JSON(c, 200, "Deleted")
			return
		}
		helpers.JSON(c, 404, "Not Found")
		return
	}
	helpers.JSON(c, 400, "Bad Request")
}

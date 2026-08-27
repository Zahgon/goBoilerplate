package main

import (
	"goBoilterplate/app/console"
	"goBoilterplate/app/router"
	"goBoilterplate/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/tylerb/graceful.v1"
)

// @title Golang Echo API
// @version 1.0
// @description API documentation by Swagger

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	app := gin.New()

	db := config.Database()
	defer db.Close()

	config.Redis()
	console.Schedule()
	router.Init(app)

	server := &http.Server{Addr: ":3000", Handler: app}
	graceful.ListenAndServe(server, 5*time.Second)
}

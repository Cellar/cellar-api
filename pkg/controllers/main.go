package controllers

import "github.com/gin-gonic/gin"

// @title Cellar
// @description Simple secret sharing with the infrastructure you already trust
// @contact.name Aria Vesta
// @contact.email dev@ariavesta.com
// @contact.url http://cellar-app.io
// @license.name MIT
// @license.url https://gitlab.com/cellar-app/cellar-api/-/blob/main/LICENSE.txt
// @BasePath /
func Register(router *gin.Engine) {
	router.GET("/health-check", HealthCheck)
}

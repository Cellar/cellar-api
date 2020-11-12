package middleware

import (
	"cellar/pkg/settings"
	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, cfg settings.IConfiguration) {
	configureAppLogging(cfg)
	configureWebLogging(router)
	injectDependencies(router, cfg)
	configureSwagger(cfg)
}

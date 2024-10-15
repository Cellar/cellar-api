package main

import (
	"cellar/pkg/controllers"
	secretsV1 "cellar/pkg/controllers/v1"
	secretsV2 "cellar/pkg/controllers/v2"
	"cellar/pkg/middleware"
	"cellar/pkg/settings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var version string = "0.0.0"

// @title Cellar
// @description Simple secret sharing with the infrastructure you already trust
// @contact.name Aria Vesta
// @contact.email dev@ariavesta.com
// @license.name MIT
// @license.url https://gitlab.com/cellar-app/cellar-api/-/blob/main/LICENSE.txt
func main() {
	router := gin.New()
	settings.SetAppVersion(version)
	cfg := settings.NewConfiguration()
	middleware.Setup(router, cfg)
	addRoutes(router)
	middleware.HandleError("error while starting the server", router.Run(cfg.App().BindAddress()))
}

func addRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER"))

	router.GET("/health-check", controllers.HealthCheck)

	v1 := router.Group("/v1")
	{
		secrets := v1.Group("/secrets")
		{
			secrets.POST("", secretsV1.CreateSecret)
			secrets.POST(":id/access", secretsV1.AccessSecretContent)
			secrets.GET(":id", secretsV1.GetSecretMetadata)
			secrets.DELETE(":id", secretsV1.DeleteSecret)
		}
	}

	v2 := router.Group("/v2")
	{
		secrets := v2.Group("/secrets")
		{
			secrets.POST("", secretsV2.CreateSecret)
			secrets.POST(":id/access", secretsV2.AccessSecretContent)
			secrets.GET(":id", secretsV2.GetSecretMetadata)
			secrets.DELETE(":id", secretsV2.DeleteSecret)
		}
	}
}

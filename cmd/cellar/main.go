package main

import (
	"cellar/pkg/controllers"
	"cellar/pkg/middleware"
	"cellar/pkg/settings"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @contact.name Parker Johansen
// @contact.email johansen.parker@gmail.com
func main() {
	router := gin.New()
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
			secrets.POST("", controllers.CreateSecret)
			secrets.POST(":id/access", controllers.AccessSecretContent)
			secrets.GET(":id", controllers.GetSecretMetadata)
			secrets.DELETE(":id", controllers.DeleteSecret)
		}
	}
}

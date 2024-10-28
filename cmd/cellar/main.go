package main

import (
	"cellar/pkg/controllers"
	v1 "cellar/pkg/controllers/v1"
	v2 "cellar/pkg/controllers/v2"
	"cellar/pkg/middleware"
	"cellar/pkg/settings"
	"golang.org/x/net/webdav"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var version string = "0.0.0"

func main() {
	router := gin.New()
	settings.SetAppVersion(version)
	cfg := settings.NewConfiguration()
	middleware.Setup(router, cfg)
	addRoutes(router)
	middleware.HandleError("error while starting the server", router.Run(cfg.App().BindAddress()))
}

func addRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER", ginSwagger.InstanceName("common")))
	router.GET("/health-check", controllers.HealthCheck)

	router.GET("/swagger/v1/*any", DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER", ginSwagger.InstanceName("v1")))
	v1.Register(router)

	router.GET("/swagger/v2/*any", DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER", ginSwagger.InstanceName("v2")))
	v2.Register(router)

}

// DisablingWrapHandler turn handler off
// if specified environment variable passed.
func DisablingWrapHandler(handler *webdav.Handler, envName string, options ...func(*ginSwagger.Config)) gin.HandlerFunc {
	if os.Getenv(envName) != "" {
		return func(c *gin.Context) {
			// Simulate behavior when route unspecified and
			// return 404 HTTP code
			c.String(http.StatusNotFound, "")
		}
	}

	return ginSwagger.WrapHandler(handler, options...)
}

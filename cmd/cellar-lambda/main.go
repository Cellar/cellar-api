package main

import (
	"cellar/pkg/controllers"
	v1 "cellar/pkg/controllers/v1"
	v2 "cellar/pkg/controllers/v2"
	"cellar/pkg/middleware"
	"cellar/pkg/settings"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/net/webdav"
	"net/http"
	"os"
)

var ginLambda *ginadapter.GinLambda

var version string = "0.0.0"

func init() {
	router := gin.New()
	settings.SetAppVersion(version)
	cfg := settings.NewConfiguration()
	setMultipartMemoryLimit(router, cfg)
	middleware.Setup(router, cfg)
	addRoutes(router)

	ginLambda = ginadapter.New(router)
}

func setMultipartMemoryLimit(router *gin.Engine, cfg settings.IConfiguration) {
	const ginDefaultMultipartMemory = 32 * 1024 * 1024
	const multipartBuffer = 2 * 1024 * 1024

	configuredLimit := int64(cfg.App().MaxFileSizeMB()*1024*1024) + multipartBuffer
	if configuredLimit > ginDefaultMultipartMemory {
		router.MaxMultipartMemory = configuredLimit
	} else {
		router.MaxMultipartMemory = ginDefaultMultipartMemory
	}
}

func addRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER"))
	router.GET("/health-check", controllers.HealthCheck)

	v1.Register(router)

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

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}

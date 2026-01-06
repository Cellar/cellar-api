package v1

import (
	"cellar/pkg/middleware"
	"cellar/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

// @title Cellar
// @description Simple secret sharing with the infrastructure you already trust
// @contact.name Aria Vesta
// @contact.email dev@ariavesta.com
// @contact.url http://cellar-app.io
// @license.name MIT
// @license.url https://gitlab.com/cellar-app/cellar-api/-/blob/main/LICENSE.txt
// @BasePath /v1
func Register(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		secrets := v1.Group("/secrets")
		{
			secrets.POST("", middleware.RateLimit(ratelimit.Tier1), CreateSecret)
			secrets.POST(":id/access", middleware.RateLimit(ratelimit.Tier1), AccessSecretContent)
			secrets.GET(":id", middleware.RateLimit(ratelimit.Tier2), GetSecretMetadata)
			secrets.DELETE(":id", middleware.RateLimit(ratelimit.Tier2), DeleteSecret)
		}
	}
}

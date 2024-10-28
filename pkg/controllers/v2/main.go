package v2

import "github.com/gin-gonic/gin"

// @title Cellar
// @description Simple secret sharing with the infrastructure you already trust
// @contact.name Aria Vesta
// @contact.email dev@ariavesta.com
// @contact.url http://cellar-app.io
// @license.name MIT
// @license.url https://gitlab.com/cellar-app/cellar-api/-/blob/main/LICENSE.txt
// @BasePath /v2
func Register(router *gin.Engine) {
	v2 := router.Group("/v2")
	{
		secrets := v2.Group("/secrets")
		{
			secrets.POST("", CreateSecret)
			secrets.POST(":id/access", AccessSecretContent)
			secrets.GET(":id", GetSecretMetadata)
			secrets.DELETE(":id", DeleteSecret)
		}
	}
}

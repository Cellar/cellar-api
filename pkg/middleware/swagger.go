package middleware

import (
	"cellar/cmd/cellar/docs"
	"cellar/pkg/settings"
	"strings"
)

func configureSwagger(configuration settings.IConfiguration) {
	docs.SwaggerInfo.Host = strings.TrimLeft(configuration.App().ClientAddress(), "http://")
	docs.SwaggerInfo.BasePath = "/"
}

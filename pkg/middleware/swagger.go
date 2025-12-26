package middleware

import (
	"cellar/docs"
	"cellar/pkg/settings"
	"strings"
)

func configureSwagger(configuration settings.IConfiguration) {
	host := configuration.App().ClientAddress()
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = "/"
}

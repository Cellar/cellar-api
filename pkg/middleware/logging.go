package middleware

import (
	"cellar/pkg/settings"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"time"
)

func configureAppLogging(cfg settings.IConfiguration) {
	logLevel, err := cfg.Logging().Level()
	HandleError("Unable to read log level from configuration", err)
	log.SetLevel(logLevel)

	logFormat, err := cfg.Logging().Format()
	HandleError("Unable to read log format for configuration", err)
	log.SetFormatter(logFormat)

	logLocations, err := cfg.Logging().Locations()
	HandleError("Unable to determine log location from configuration", err)

	if len(logLocations) <= 0 {
		log.Warn("No logging locations enabled. Continuing with logging disabled")
	} else {
		for _, logLocation := range logLocations {
			log.SetOutput(logLocation)
		}
	}
}

func configureWebLogging(router *gin.Engine) {

	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			path := c.Request.URL.Path
			start := time.Now()

			c.Next()
			latency := time.Since(start).Milliseconds()
			statusCode := c.Writer.Status()
			clientIP := c.ClientIP()
			clientUserAgent := c.Request.UserAgent()
			referer := c.Request.Referer()
			dataLength := c.Writer.Size()
			if dataLength < 0 {
				dataLength = 0
			}

			ginLogger := log.WithFields(log.Fields{
				"statusCode": statusCode,
				"method":     c.Request.Method,
				"path":       path,
				"latency":    fmt.Sprintf("%d ms", latency),
				"clientIP":   clientIP,
				"referer":    referer,
				"userAgent":  clientUserAgent,
				"dataLength": dataLength,
			})

			if len(c.Errors) > 0 {
				ginLogger.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
			} else {
				msg := fmt.Sprintf(
					"%d %s %s",
					statusCode,
					c.Request.Method,
					path)
				if statusCode > 499 {
					ginLogger.Error(msg)
				} else if statusCode > 399 {
					ginLogger.Warn(msg)
				} else {
					ginLogger.Info(msg)
				}
			}
		}
	}())
}

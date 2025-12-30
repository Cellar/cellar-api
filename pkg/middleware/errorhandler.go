package middleware

import (
	pkgerrors "cellar/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrorHandler is a middleware that handles errors attached to the gin context.
// Controllers should use c.Error(err) to attach errors, and this middleware will
// convert them to appropriate HTTP responses based on error type.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var statusCode int
			switch {
			case pkgerrors.IsContextError(err):
				statusCode = http.StatusRequestTimeout
			case pkgerrors.IsFileTooLargeError(err):
				statusCode = http.StatusRequestEntityTooLarge
			case pkgerrors.IsValidationError(err):
				statusCode = http.StatusBadRequest
			default:
				statusCode = http.StatusInternalServerError
			}

			log.WithFields(log.Fields{
				"status": statusCode,
				"error":  err.Error(),
				"path":   c.Request.URL.Path,
			}).Error("Request error")

			if !c.Writer.Written() {
				c.JSON(statusCode, gin.H{
					"error": err.Error(),
				})
			}
		}
	}
}

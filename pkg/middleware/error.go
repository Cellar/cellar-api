package middleware

import log "github.com/sirupsen/logrus"

// HandleError logs a fatal error with the given message if err is not nil.
// The application will exit after logging.
func HandleError(message string, err error) {
	if err != nil {
		log.WithField("error", err).
			Fatal(message)
	}
}

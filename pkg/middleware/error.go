package middleware

import log "github.com/sirupsen/logrus"

func HandleError(message string, err error) {
	if err != nil {
		log.WithField("error", err).
			Fatal(message)
	}
}

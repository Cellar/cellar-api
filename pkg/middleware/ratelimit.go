package middleware

import (
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/ratelimit"
	"cellar/pkg/settings"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func RateLimit(tier ratelimit.Tier) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet(settings.Key).(settings.IConfiguration)

		if !cfg.RateLimit().Enabled() {
			c.Next()
			return
		}

		rateLimiter, exists := c.Get(ratelimit.Key)
		if !exists {
			log.Warn("rate limiter not found in context, skipping rate limiting")
			c.Next()
			return
		}

		limiter := rateLimiter.(ratelimit.RateLimiter)
		identifier := c.ClientIP()

		result, err := limiter.Allow(c.Request.Context(), identifier, tier)
		if err != nil {
			log.WithError(err).WithField("identifier", identifier).Error("failed to check rate limit")
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			c.Header("Retry-After", strconv.Itoa(result.RetryAfter))
			err := pkgerrors.NewRateLimitError(
				fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", result.RetryAfter),
				result.RetryAfter,
			)
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Next()
	}
}

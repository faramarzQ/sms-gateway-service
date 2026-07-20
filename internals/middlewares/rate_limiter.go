package middlewares

import (
	"fmt"
	"github.com/faramarzQ/sms-gateway-service/internals/http/responses"
	"github.com/faramarzQ/sms-gateway-service/internals/logger"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

const (
	RateLimit = 3
	Window    = time.Minute
)

func RateLimiter(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDHeader := c.GetHeader("X-User-ID")
		if userIDHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				responses.Response{
					Status:  http.StatusUnauthorized,
					Message: "missing X-User-ID header",
				},
			)
			return
		}

		userID, err := strconv.ParseUint(userIDHeader, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				responses.Response{
					Status:  http.StatusUnauthorized,
					Message: "invalid user_id header",
				},
			)
			return
		}

		now := time.Now().UTC()
		key := fmt.Sprintf(
			"rate_limit:sms:%d:%s",
			userID,
			now.Format("200601021504"),
		)

		ctx := c.Request.Context()

		count, err := redis.Incr(ctx, key).Result()
		if err != nil {
			logger.Logger.Error("failed storing rate limit", zap.Error(err))

			c.AbortWithStatusJSON(
				500,
				responses.Response{
					Status: http.StatusInternalServerError,
				},
			)
			return
		}

		// first request in this window
		if count == 1 {
			redis.Expire(ctx, key, Window)
		}

		if count > RateLimit {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				responses.Response{
					Status: http.StatusTooManyRequests,
				},
			)
			return
		}

		c.Next()
	}
}

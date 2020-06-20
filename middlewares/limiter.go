package middlewares

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	limitergin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/websentry/websentry/controllers"
)

func limitReachedHandler(c *gin.Context) {
	controllers.JSONResponse(c, controllers.CodeExceededLimits, "limiter", nil)
}

func keyGetterIP(c *gin.Context) string {
	return c.ClientIP()
}

func keyGetterUserID(c *gin.Context) string {
	return strconv.FormatInt(c.MustGet("userId").(int64), 16)
}

// TODO: use redis

func GetSensitiveLimiter() gin.HandlerFunc {
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  500,
	}
	options := []limitergin.Option{
		limitergin.WithKeyGetter(keyGetterIP),
		limitergin.WithLimitReachedHandler(limitReachedHandler),
	}
	return limitergin.NewMiddleware(limiter.New(store, rate), options...)
}

func GetGeneralLimiter() gin.HandlerFunc {
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  1000,
	}
	options := []limitergin.Option{
		limitergin.WithKeyGetter(keyGetterUserID),
		limitergin.WithLimitReachedHandler(limitReachedHandler),
	}
	return limitergin.NewMiddleware(limiter.New(store, rate), options...)
}

func GetScreenshotLimiter() gin.HandlerFunc {
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  20,
	}
	options := []limitergin.Option{
		limitergin.WithKeyGetter(keyGetterUserID),
		limitergin.WithLimitReachedHandler(limitReachedHandler),
	}
	return limitergin.NewMiddleware(limiter.New(store, rate), options...)
}

func GetSlaveLimiter() gin.HandlerFunc {
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Hour,
		Limit:  5000,
	}
	options := []limitergin.Option{
		limitergin.WithKeyGetter(keyGetterIP),
		limitergin.WithLimitReachedHandler(limitReachedHandler),
	}
	return limitergin.NewMiddleware(limiter.New(store, rate), options...)
}

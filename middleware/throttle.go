package middleware

import (
	"context"
	"deeptrain/connection"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

type Limiter struct {
	Duration int
	Count    int64
}

func (l *Limiter) RateLimit(ctx context.Context, ip string, path string) bool {
	key := fmt.Sprintf("rate%s:%s", path, ip)
	count, err := connection.Cache.Incr(ctx, key).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		connection.Cache.Expire(ctx, key, time.Duration(l.Duration)*time.Second)
	}
	return count > l.Count
}

var limits = map[string]Limiter{
	"login":    {Duration: 60, Count: 5},
	"register": {Duration: 60, Count: 5},
	"verify":   {Duration: 60, Count: 3},
	"resend":   {Duration: 60, Count: 1},
}

func ThrottleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.Request.URL.Path
		for prefix, limit := range limits {
			if strings.HasPrefix(path, "/"+prefix) {
				if limit.RateLimit(c, ip, path) {
					c.JSON(200, gin.H{"status": false, "reason": "You have sent too many requests. Please try again later."})
					c.Abort()
					return
				}
			}
		}
		c.Next()
	}
}

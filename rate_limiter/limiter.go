package rate_limiter

import (
	"shareit/db"
	"strconv"
	"time"
)

// RateLimit limits the number of requests a user can make per minute
func RateLimit(userID int) bool {
    key := "ratelimit_" + strconv.Itoa(userID)
    count, err := db.RedisClient.Incr(db.Ctx, key).Result()
    if err != nil {
        return false
    }

    if count == 1 {
        db.RedisClient.Expire(db.Ctx, key, 1*time.Minute)
    }

    return count <= 100
}

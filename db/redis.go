package db

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// ConnectRedis initializes the Redis client connection
func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "complete-feline-33196.upstash.io:6379", // Redis address
		Password: "AYGsAAIjcDE2MmFmN2MwM2M2NzA0YWMzYjhhODM2N2MzMzgxMjhiN3AxMA", // Redis password
		DB:       0, // use default DB
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip verification of server certificate (useful for self-signed certs, not recommended for production)
		},
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Could not connect to Redis:", err)
	}

	log.Println("Connected to Redis successfully!")
}

// CacheFileMetadata caches file metadata
func CacheFileMetadata(fileID string, metadata string) error {
    return RedisClient.Set(Ctx, fileID, metadata, 10*time.Minute).Err()
}

func GetCachedFileMetadata(fileID string) (string, error) {
    return RedisClient.Get(Ctx, fileID).Result()
}

package config

import (
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient         *redis.Client = nil
	redisAddr           string
	redisPassword       string
	redisIndex          int
	CacheExpiredSeconds int
)

func init() {
	redisAddr = os.Getenv("REDIS_ADDRESS")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	dbIndexStr := os.Getenv("REDIS_DB_INDEX")
	var err error
	redisIndex, err = strconv.Atoi(dbIndexStr)
	if err != nil {
		log.Println("REDIS_DB_INDEX is not set, default 0")
		redisIndex = 0
	}
	expiredSecondsStr := os.Getenv("REDIS_EXPIRED_SECONDS")
	CacheExpiredSeconds, err = strconv.Atoi(expiredSecondsStr)
	if err != nil {
		log.Println("REDIS_EXPIRED_SECONDS is not set, default 3600")
		CacheExpiredSeconds = 3600
	}

	log.Println("REDIS_ADDR:", redisAddr)
	log.Println("REDIS_DB_INDEX:", redisIndex)
	log.Println("REDIS_EXPIRED_SECONDS:", CacheExpiredSeconds)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisIndex,
	})
}

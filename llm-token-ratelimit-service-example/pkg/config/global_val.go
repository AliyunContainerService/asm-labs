package config

import (
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient      *redis.Client = nil
	redisAddr        string
	redisPassword    string
	redisIndex       int
	RateLimitConfigs []RateLimitConfig = []RateLimitConfig{}
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

	log.Println("REDIS_ADDRESS:", redisAddr)
	log.Println("REDIS_DB_INDEX:", redisIndex)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisIndex,
	})

	RateLimitConfigs, err = initRateLimitConfig()
	if err != nil {
		log.Fatalf("initRateLimitConfig failed: %v", err)
	}

	log.Println("RateLimitConfigs:", RateLimitConfigs)
	log.Println("init config done")
}

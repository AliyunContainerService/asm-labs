package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/asm-labs/llm-cache-service-example/pkg/config"
	"github.com/redis/go-redis/v9"
)

var logger = log.Default()

func Lookup(w http.ResponseWriter, r *http.Request) {
	logger.Println("in lookup")

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	logger.Printf("body: %s\n", body)
	if err != nil {
		logger.Printf("error reading body: %s\n", err)
		return
	}

	cacheInfo := &CacheInfo{}
	err = json.Unmarshal(body, cacheInfo)
	if err != nil {
		logger.Printf("unmarshal failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key, err := CacheRequestToCacheKey(&cacheInfo.Request)
	if err != nil {
		logger.Printf("cache request to cache key failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value, err := config.RedisClient.Get(context.TODO(), key).Result()
	if err == redis.Nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("get failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Redis中存储的已经是一个CacheInfo结构的JSON String，无需修改，直接返回。
	_, err = w.Write([]byte(value))
	if err != nil {
		log.Printf("write failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return

}

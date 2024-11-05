package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/asm-labs/llm-cache-service-example/pkg/config"
)

func Update(w http.ResponseWriter, r *http.Request) {
	logger.Println("in update")
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.Printf("read failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
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
		logger.Printf("error in cache key: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 存入Redis的value为CacheInfo结构体，不带Request字段。
	cacheInfo.Request = CacheRequest{}
	value, err := MarshalCacheItem(cacheInfo)
	if err != nil {
		logger.Printf("error in cache value: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = config.RedisClient.Set(context.TODO(), key, value, time.Duration(config.CacheExpiredSeconds)*time.Second).Err()
	if err != nil {
		logger.Printf("error in set redis: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Println(string(body))
	w.WriteHeader(http.StatusOK)
	return
}

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/asm-labs/llm-token-ratelimit-service-example/pkg/config"
)

var logger = log.Default()

type RateLimitResponse struct {
	Allow       bool   `json:"allow"`
	Description string `json:"description"`
}

func RateLimit(w http.ResponseWriter, r *http.Request) {
	rateLimitKeys := getRateLimitKeys(r)
	ratelimitResponse := &RateLimitResponse{
		Allow:       true,
		Description: "ok",
	}
	errList := []error{}
	for _, key := range rateLimitKeys {
		shouldLimit, err := queryAndFillBucket(key)
		if err != nil {
			logger.Printf("queryAndFillBucket failed, key: %s, err: %v\n", key, err)
			errList = append(errList, err)
		} else if shouldLimit {
			ratelimitResponse.Allow = false
			ratelimitResponse.Description = fmt.Sprintf("%s is being rate-limited", key)
			break
		}
	}
	if len(errList) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(ratelimitResponse)
	if err != nil {
		logger.Printf("json.Marshal failed, err: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(body); err != nil {
		logger.Printf("write failed, %v", err)
	}
}

func getRateLimitKeys(r *http.Request) []string {
	if r.URL == nil {
		logger.Println("cannot get request url")
		return []string{}
	}
	rateLimitKeysStr := r.URL.Query().Get("ratelimit_keys")
	if rateLimitKeysStr == "" {
		logger.Println("ratelimit_keys is empty")
		return []string{}
	}
	rateLimitKeys := []string{}
	if err := json.Unmarshal([]byte(rateLimitKeysStr), &rateLimitKeys); err != nil {
		logger.Printf("invalid ratelimit_keys, err: %v\n", err)
		return []string{}
	}
	return rateLimitKeys
}

var luaScript = `
local key = KEYS[1]
local lastFillTimeKey = key .. "_last_fill_time"
local fillInterval = tonumber(ARGV[1])
local tokensPerFill = tonumber(ARGV[2])
local maxTokens = tonumber(ARGV[3])
local expireTime = tonumber(ARGV[4])
local currentTime = tonumber(redis.call("TIME")[1])
if redis.call("EXISTS", key) == 0 or redis.call("EXISTS", lastFillTimeKey) == 0 then
    redis.call("SET", key, tokensPerFill)
    redis.call("EXPIRE", key, expireTime)
    redis.call("SET", lastFillTimeKey, currentTime)
    redis.call("EXPIRE", lastFillTimeKey, expireTime)
else
	local currentValue = redis.call("GET", key)
	local lastFillTime = redis.call("GET", lastFillTimeKey)
    local intervals = math.floor((currentTime - tonumber(lastFillTime)) / fillInterval)
    if intervals > 0 then
        local expect = tonumber(currentValue) + (tokensPerFill * intervals)
        if expect > maxTokens then
            currentValue = maxTokens
        else
            currentValue = expect
        end
        redis.call("SET", key, currentValue)
        redis.call("EXPIRE", key, expireTime)
        redis.call("SET", lastFillTimeKey, currentTime)
        redis.call("EXPIRE", lastFillTimeKey, expireTime)
    end
end

if tonumber(redis.call("GET", key)) > 0 then
	return 0
else
	return 1
end
`

// return: shouldLimit, error
func queryAndFillBucket(key string) (bool, error) {
	rateLimitConfig := config.GetRateLimitConfig(key)
	if rateLimitConfig == nil {
		// no rate limit
		return false, nil
	}
	result, err := config.RedisClient.Eval(context.TODO(), luaScript, []string{key},
		rateLimitConfig.FillIntervalSecond,
		rateLimitConfig.TokensPerFill,
		rateLimitConfig.MaxTokens,
		rateLimitConfig.RedisExpiredSeconds).Result()
	if err != nil {
		return true, err
	}
	switch result := result.(type) {
	case int64:
		{
			if result == 0 {
				return false, nil
			} else {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("invalid result type")
	}
}

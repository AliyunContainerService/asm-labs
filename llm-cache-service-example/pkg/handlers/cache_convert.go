package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

func init() {
	logger.Print("CacheKeyTemplate: {{headers.host}}-{{body.model}}-{{body.messages|@reverse|0.content}}, if you want to change it, just modify this file")
}

// CacheRequest ==> CacheKey
func CacheRequestToCacheKey(req *CacheRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("cache request is nil")
	}

	key := ""
	if req.Headers["host"] != "" {
		key += req.Headers["host"] + "-"
	} else if req.Headers[":authority"] != "" {
		key += req.Headers[":authority"] + "-"
	}

	model := gjson.Get(req.Body, "model").String()
	if model == "" {
		model = "UNKNOW"
	}
	key += model + "-"

	lastMessages := gjson.Get(req.Body, "messages|@reverse|0.content").String()
	if lastMessages == "" {
		return "", fmt.Errorf("last messages is empty")
	}
	key += lastMessages

	return key, nil
}

func MarshalCacheItem(cacheInfo *CacheInfo) (string, error) {
	if cacheInfo == nil {
		return "", fmt.Errorf("cache cacheInfo is nil")
	}
	value, err := json.Marshal(cacheInfo)
	if err != nil {
		logger.Printf("marshal failed: %v\n", err)
		return "", err
	}
	return string(value), nil
}

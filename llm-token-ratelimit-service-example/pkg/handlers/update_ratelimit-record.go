package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/asm-labs/llm-token-ratelimit-service-example/pkg/config"
)

type UpdateRateLimitRecordBody struct {
	RateLimitKeys    []string `json:"ratelimit_keys"`
	PromptTokens     int      `json:"prompt_tokens"`
	CompletionTokens int      `json:"completion_tokens"`
	TotalTokens      int      `json:"total_tokens"`
}

func UpdateRateLimitRecord(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.Printf("read request body failed, %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	bodyObject := &UpdateRateLimitRecordBody{}
	if err := json.Unmarshal(body, bodyObject); err != nil {
		logger.Printf("invalid request body, err: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, key := range bodyObject.RateLimitKeys {
		if err := consumeTokens(key, bodyObject.CompletionTokens); err != nil {
			logger.Printf("consumeTokens failed, key: %s, err: %v\n", key, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

// assume the key is already exists
func consumeTokens(key string, tokens int) error {
	_, err := config.RedisClient.DecrBy(context.TODO(), key, int64(tokens)).Result()
	return err
}

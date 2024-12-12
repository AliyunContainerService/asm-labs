package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func initRateLimitConfig() ([]RateLimitConfig, error) {
	val, ok := os.LookupEnv("RATE_LIMIT_CONFIG")
	if !ok {
		return nil, fmt.Errorf("RATE_LIMIT_CONFIG is not set")
	}
	configs := []RateLimitConfig{}
	err := json.Unmarshal([]byte(val), &configs)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed, invalid RATE_LIMIT_CONFIG: %s", val)
	}

	for i := range configs {
		regexp, err := regexp.Compile(configs[i].RateLimitKeyRegex)
		if err != nil {
			return nil, fmt.Errorf("invalid rate_limit_key_regex: %s", configs[i].RateLimitKeyRegex)
		}
		configs[i].regexp = regexp
	}
	return configs, nil

}

type RateLimitConfig struct {
	RateLimitKeyRegex   string `json:"rate_limit_key_regex"`
	RedisExpiredSeconds int    `json:"redis_expired_seconds"`
	FillIntervalSecond  int    `json:"fill_interval_second"`
	TokensPerFill       int    `json:"tokens_per_fill"`
	MaxTokens           int    `json:"max_tokens"`

	regexp *regexp.Regexp
}

func GetRateLimitConfig(key string) *RateLimitConfig {
	for i := range RateLimitConfigs {
		if RateLimitConfigs[i].match(key) {
			return &RateLimitConfigs[i]
		}
	}
	return nil
}

func (c *RateLimitConfig) match(key string) bool {
	if c == nil || c.regexp == nil {
		return false
	}
	return c.regexp.MatchString(key)
}

package llmproxy

import (
	"encoding/json"
	"strings"
)

func NewLLMRequestBody(jsonStr []byte) (*LLMRequestBody, error) {
	request := &LLMRequestBody{}
	err := json.Unmarshal(jsonStr, request)
	if err != nil {
		// proxywasm.LogErrorf("error in unmarshal LLMRequest: %v", err)
		return request, err
	}
	return request, nil
}

type LLMRequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   *bool     `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (r *LLMRequestBody) GetMessageString() string {
	if r == nil {
		return ""
	}
	messageStr := ""
	for _, message := range r.Messages {
		if strings.ToLower(message.Role) != "user" {
			continue
		}
		messageStr += "\n" + message.Content
	}
	return messageStr
}

type LLMResponseBody struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
	Logprobs     []int   `json:"logprobs"`
}

type CustomIntelligentGuardResponse struct {
	Result *string `json:"result"`
	Reason *string `json:"reason"`
}

package llmproxy

import (
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func getUserMessageString(r *openai.ChatCompletionRequest) string {
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

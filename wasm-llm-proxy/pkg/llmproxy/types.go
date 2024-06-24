package llmproxy

type CustomIntelligentGuardResponse struct {
	Result *string `json:"result"`
	Reason *string `json:"reason"`
}

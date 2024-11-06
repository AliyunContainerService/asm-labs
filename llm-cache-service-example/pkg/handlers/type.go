package handlers

type CacheInfo struct {
	Request  CacheRequest  `json:"request"`
	Response CacheResponse `json:"response"`
}

type CacheRequest struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
type CacheResponse struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

package llmproxy

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
)

func NewLLMProxyConfig(jsonStr []byte) (*LLMProxyConfig, error) {
	config := &LLMProxyConfig{}
	err := json.Unmarshal(jsonStr, config)
	if err != nil {
		proxywasm.LogErrorf("error in unmarshal LLMProxyConfig: %v", err)
		return config, err
	}
	err = config.Init()
	if err != nil {
		proxywasm.LogErrorf("error in init LLMProxyConfig: %v", err)
		return config, err
	}
	return config, nil
}

type LLMProxyConfig struct {
	Hosts            []string          `json:"hosts"`
	API_KEY          string            `json:"api_key"`
	AllowPatterns    []string          `json:"allow_patterns"`
	DenyPatterns     []string          `json:"deny_patterns"`
	IntelligentGuard *IntelligentGuard `json:"intelligent_guard"`

	// private
	allowRegexList []*regexp.Regexp
	denyRegexList  []*regexp.Regexp
}

// cluster: outbound|443||dashscope.aliyuncs.com
// authority header: dashscope.aliyuncs.com
type IntelligentGuard struct {
	Host    *string `json:"host"` // host header
	Port    *uint32 `json:"port"`
	Path    *string `json:"path"`    // default "/compatible-mode/v1/chat/completions"
	Model   *string `json:"model"`   // default qwen-turbo
	API_KEY *string `json:"api_key"` // can not be empty
}

func (c *LLMProxyConfig) Init() error {
	if c.IntelligentGuard != nil {
		if c.IntelligentGuard.Host == nil || *c.IntelligentGuard.Host == "" {
			err := fmt.Errorf("intelligent guard cluster cannot be empty")
			proxywasm.LogErrorf("%v", err)
			return err
		}
		// set default
		if c.IntelligentGuard.Port == nil || *c.IntelligentGuard.Port == 0 {
			c.IntelligentGuard.Port = UInt32Prt(443)
		}

		if c.IntelligentGuard.Path == nil {
			c.IntelligentGuard.Path = StringPtr("/compatible-mode/v1/chat/completions")
		}
		if c.IntelligentGuard.Model == nil {
			c.IntelligentGuard.Model = StringPtr("qwen-turbo")
		}
		if c.IntelligentGuard.API_KEY == nil {
			err := fmt.Errorf("intelligent guard api_key cannot be empty")
			proxywasm.LogErrorf("%v", err)
			return err
		}
	}

	for _, pattern := range c.AllowPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			proxywasm.LogErrorf("error in compile regex %v: %v", pattern, err)
			return err
		}
		c.allowRegexList = append(c.allowRegexList, regex)
	}
	for _, pattern := range c.DenyPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			proxywasm.LogErrorf("error in compile regex %v: %v", pattern, err)
			return err
		}
		c.denyRegexList = append(c.denyRegexList, regex)
	}
	return nil
}

func (c *LLMProxyConfig) RunMessageGuard(llmRequest *LLMRequestBody) error {
	if c == nil {
		proxywasm.LogInfof("nil llm request, pass check")
		return nil
	}

	for _, message := range llmRequest.Messages {
		// message.
		if strings.ToLower(message.Role) != "user" {
			continue
		}
		// check deny rules
		for _, regex := range c.denyRegexList {
			if regex.MatchString(message.Content) {
				err := fmt.Errorf("message \"%v\" was denied by deny rule", message.Content)
				proxywasm.LogInfof("%v", err)
				return err
			}
		}

		if len(c.allowRegexList) == 0 {
			continue
		}
		matched := false
		for _, regex := range c.allowRegexList {
			if regex.MatchString(message.Content) {
				matched = true
				break
			}
		}
		if !matched {
			err := fmt.Errorf("message \"%v\" was denied because no allow rule matched", message.Content)
			proxywasm.LogInfof("%v", err)
			return err
		}
	}
	return nil
}

func (c *LLMProxyConfig) RunIntelligentGuard(llmRequest *LLMRequestBody) error {
	proxywasm.LogInfo("in RunIntelligentGuard")
	userInput := llmRequest.GetMessageString()
	body := fmt.Sprintf(`
{
	"model": "%v",
	"messages": [
		{
			"role": "system",
			"content": "You are a sensitive information inspector, responsible for helping me filter the information user input to guarantee that neither personal nor corporate private and confidential information is disclosed. If the message content contains private or classified information, please return a response in the following format: {\"result\": \"allow\" or \"deny\", \"reason\": \"why it was denied\"}"
		},
		{
			"role": "user",
			"content": %v
		}
	],

	"stream": false
}
	`, *c.IntelligentGuard.Model, strconv.Quote(userInput))
	proxywasm.LogInfof("body: %v", body)
	proxywasm.LogInfof("host: %v", *c.IntelligentGuard.Host)
	// call intelligent guard
	_, err := proxywasm.DispatchHttpCall(
		// outbound|443||dashscope.aliyuncs.com
		fmt.Sprintf("outbound|%v||%v", *c.IntelligentGuard.Port, *c.IntelligentGuard.Host),
		[][2]string{
			// path, method and authority are required. envoy will check them.
			{"content-type", "application/json"},
			{":path", *c.IntelligentGuard.Path},
			{"Authorization", "Bearer " + *c.IntelligentGuard.API_KEY},
			{":method", "POST"},
			{":authority", *c.IntelligentGuard.Host},
		},
		[]byte(body),
		nil,
		10000,
		c.IntelligentGuardCallback,
	)
	if err != nil {
		proxywasm.LogErrorf("error in DispatchHttpCall: %v", err)
		return err
	}
	return nil
}

func (c *LLMProxyConfig) IntelligentGuardCallback(numHeaders, bodySize, numTrailers int) {
	proxywasm.LogInfo("in IntelligentGuardCallback")
	// We want to always resume the intercepted request regardless of success/fail to avoid indefinitely blocking anything
	defer func() {
		if err := proxywasm.ResumeHttpRequest(); err != nil {
			proxywasm.LogCriticalf("failed to ResumeHttpRequest after calling auth: %v", err)
		}
	}()

	// Get the response headers from external service
	headers, err := proxywasm.GetHttpCallResponseHeaders()
	if err != nil {
		proxywasm.LogCriticalf("failed to GetHttpCallResponseHeaders from external response: %v", err)
		return
	}
	headersMap := headerArrayToMap(headers)
	body, err := proxywasm.GetHttpCallResponseBody(0, bodySize)
	if err != nil {
		proxywasm.LogErrorf("failed to GetHttpCallResponseBody from external response: %v", err)
		proxywasm.SendHttpResponse(403, nil, []byte("failed to GetHttpCallResponseBody from external response: "+err.Error()), -1)
		return
	}
	proxywasm.LogInfof("body: %v", string(body))

	if headersMap[":status"] != "200" {
		proxywasm.LogCritical("external service returned non-200 status")
		proxywasm.SendHttpResponse(403, nil, []byte("external service returned non-200 status : "+headersMap[":status"]), -1)
		return
	}

	llmResponseBody := &LLMResponseBody{}
	err = json.Unmarshal(body, &llmResponseBody)
	if err != nil {
		proxywasm.LogErrorf("failed to unmarshal external response: %v", err)
		proxywasm.SendHttpResponse(403, nil, []byte("failed to unmarshal external response: "+err.Error()), -1)
		return
	}
	customIntelligentGuardResponse := &CustomIntelligentGuardResponse{}
	// only sopport non-stream mode
	err = json.Unmarshal([]byte(llmResponseBody.Choices[0].Message.Content), customIntelligentGuardResponse)
	if err != nil {
		proxywasm.LogErrorf("llmResponseBody.Choices[0].Message.Content: %v", llmResponseBody.Choices[0].Message.Content)
		proxywasm.LogErrorf("failed to unmarshal CustomIntelligentGuardResponse: %v", err)
		proxywasm.SendHttpResponse(403, nil, []byte("failed to unmarshal CustomIntelligentGuardResponse: "+err.Error()), -1)
		return
	}
	if customIntelligentGuardResponse.Result != nil && *customIntelligentGuardResponse.Result != "allow" {

		proxywasm.LogInfof("external service returned deny")
		proxywasm.SendHttpResponse(403, nil, []byte("external service returned deny: "+llmResponseBody.Choices[0].Message.Content), -1)
		return
	}
}

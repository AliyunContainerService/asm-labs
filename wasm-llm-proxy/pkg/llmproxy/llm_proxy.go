package llmproxy

import (
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

type LLMProxy struct {
	// Embed the default http context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultHttpContext
	Config          *LLMProxyConfig
	requestBodySize int
	enabled         bool // enable this plugin by LLMProxyConfig.Hosts
}

// Override types.DefaultHttpContext.
func (p *LLMProxy) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	host, err := proxywasm.GetHttpRequestHeader("host")
	if err != nil {
		p.enabled = false
		proxywasm.LogInfof("cannot get host header, error: %v", err)
		return types.ActionContinue
	}
	for _, h := range p.Config.Hosts {
		if h == host {
			p.enabled = true
			break
		}
	}
	if !p.enabled {
		proxywasm.LogInfo("llm proxy plugin disabled")
		return types.ActionContinue
	}
	return p.addAuhthorizationHeader()
}

// Override types.DefaultHttpContext.
func (p *LLMProxy) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
	if !p.enabled {
		proxywasm.LogInfo("llm proxy plugin disabled")
		return types.ActionContinue
	}

	// cache entire request body
	p.requestBodySize += bodySize
	if !endOfStream {
		return types.ActionPause
	}
	if p.Config == nil {
		proxywasm.LogWarn("llm proxy: do nothing cause empty config")
	}

	requestBytes, err := proxywasm.GetHttpRequestBody(0, p.requestBodySize)
	if err != nil {
		proxywasm.LogWarnf("error in GetHttpRequestBody: %v", err)
		return types.ActionContinue
	}
	proxywasm.LogInfo(string(requestBytes))
	llmRequestBody, err := NewLLMRequestBody(requestBytes)
	if err != nil {
		proxywasm.LogWarnf("error in GetLLMRequest: %v", err)
		return types.ActionContinue
	}

	err = p.Config.RunMessageGuard(llmRequestBody)
	if err != nil {
		proxywasm.LogWarnf("error in RunMessageGuard: %v, send local reply", err)
		err := proxywasm.SendHttpResponse(403, nil, []byte("request was denied by asm llm proxy"), -1)
		if err != nil {
			proxywasm.LogErrorf("error in send local reply, %v", err)
		}
		return types.ActionPause
	}

	if p.Config.IntelligentGuard != nil {
		err = p.Config.RunIntelligentGuard(llmRequestBody)
		if err != nil {
			proxywasm.LogWarnf("error in RunIntelligentGuard: %v, send local reply", err)
			err := proxywasm.SendHttpResponse(403, nil, []byte(err.Error()), -1)
			if err != nil {
				proxywasm.LogErrorf("error in send local reply, %v", err)
			}
			return types.ActionPause
		}
		// external http call always need pause action
		return types.ActionPause
	}

	return types.ActionContinue
}

func (p *LLMProxy) addAuhthorizationHeader() types.Action {
	_, err := proxywasm.GetHttpRequestHeader("authorization")
	switch {
	case err != nil && strings.Contains(err.Error(), "not found"):
		{
			// authz header doesn't exist
			// add it header
			proxywasm.AddHttpRequestHeader("authorization", "Bearer "+p.Config.API_KEY)
		}
	case err != nil:
		{
			// other error, direct response
			proxywasm.LogInfof("failed to get authorization header: %v, do nothing", err)
		}
	default:
		{
			// header exists,replace it
			proxywasm.ReplaceHttpRequestHeader("authorization", "Bearer "+p.Config.API_KEY)
		}
	}
	return types.ActionContinue
}

// Override types.DefaultHttpContext.
// func (ctx *LLMProxy) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
// 	if bodySize > 0 {
// 		ctx.responseChunks++
// 		ctx.totalResponseBodySize += bodySize
// 		chunk, err := proxywasm.GetHttpResponseBody(0, bodySize)
// 		if err != nil {
// 			proxywasm.LogCriticalf("failed to get response body: %v", err)
// 			return types.ActionContinue
// 		}
// 		proxywasm.LogInfof("read chunk size: %d", len(chunk))
// 		proxywasm.LogInfof("current chunk: %v", string(chunk))
// 		if len(chunk) != bodySize {
// 			proxywasm.LogErrorf("read data does not match the expected size: %d != %d", len(chunk), bodySize)
// 		}
// 	}
// 	if endOfStream {
// 		proxywasm.LogInfof("stream end, totol response body size %d, total chunks: %d", ctx.totalResponseBodySize, ctx.responseChunks)
// 	}
// 	return types.ActionContinue
// }

// Override types.DefaultHttpContext.
// func (p *LLMProxy) OnHttpRequestBody(bodySize int, endOfStream bool) types.Action {
// 	proxywasm.LogInfof("OnHttpRequestBody called. BodySize: %d, totalRequestBodyReadSize: %d, endOfStream: %v", bodySize, p.totalRequestBodyReadSize, endOfStream)
// 	if bodySize > 0 {
// 		p.totalRequestBodyReadSize += bodySize
// 		p.receivedChunks++
// 		chunk, err := proxywasm.GetHttpRequestBody(0, bodySize)
// 		if err != nil {
// 			proxywasm.LogCriticalf("failed to get request body: %v", err)
// 			return types.ActionContinue
// 		}
// 		proxywasm.LogInfof("read chunk size: %d", len(chunk))
// 		proxywasm.LogInfof("current chunk: %v", string(chunk))
// 		if len(chunk) != bodySize {
// 			proxywasm.LogErrorf("read data does not match the expected size: %d != %d", len(chunk), bodySize)
// 		}
// 	}
// 	// When endOfStream is true, we have received the entire body. We expect the total size is equal to the sum of the sizes of the chunks.
// 	if endOfStream {
// 		proxywasm.LogInfof("stream end, totol request body size %d", p.totalRequestBodyReadSize)
// 	}
// 	return types.ActionContinue
// }

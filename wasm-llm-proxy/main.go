// Copyright 2020-2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"wasm-llm-proxy/pkg/llmproxy"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	_ "github.com/wasilibs/nottinygc"
)

func main() {
	proxywasm.SetVMContext(&vmContext{})
}

type vmContext struct {
	// Embed the default VM context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultVMContext
}

// Override types.DefaultVMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

type pluginContext struct {
	// Embed the default plugin context here,
	// so that we don't need to reimplement all the methods.
	types.DefaultPluginContext
	llmProxyConfig *llmproxy.LLMProxyConfig
}

// Override types.DefaultPluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	proxywasm.LogDebug("loading llm-proxy plugin config")
	data, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		proxywasm.LogErrorf("error in GetPluginConfiguration: %v", err)
		return types.OnPluginStartStatusFailed
	}
	if data == nil {
		proxywasm.LogError("empty config, pluginStartFailed")
		return types.OnPluginStartStatusFailed
	}
	llmProxyConfig, err := llmproxy.NewLLMProxyConfig(data)
	if err != nil {
		proxywasm.LogErrorf("error in NewLLMProxyConfig: %v", err)
		return types.OnPluginStartStatusFailed
	}
	ctx.llmProxyConfig = llmProxyConfig
	return types.OnPluginStartStatusOK
}

// Override types.DefaultPluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &llmproxy.LLMProxy{
		Config: ctx.llmProxyConfig,
	}
}

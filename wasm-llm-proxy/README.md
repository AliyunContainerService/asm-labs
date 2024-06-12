# LLM Proxy
This pulgin based on [aliyun dashscope](https://help.aliyun.com/zh/dashscope/create-a-chat-foundation-model?spm=a2c4g.11186623.0.0.7afefa70LkdpYO). Please refer to the official documentation for details.

Before start, you need to get a dashscope api_key first. [How to get](https://help.aliyun.com/zh/dashscope/developer-reference/activate-dashscope-and-create-an-api-key?spm=a2c4g.11186623.0.i4).

### how to build and debug

```sh
cd wasm-llm-proxy
./build/build.sh
./test/run.sh
```
Customize your api_key in `test/envoy-demo.yaml`.

Test it with:
```sh
./test/test.sh
```

build a docker image
```sh
docker build -f build/Dockerfile -t asm-llm-plugin:latest . 
```
You can push the image to your registry and then you can use it in ASM sidecar or gateway.
### how to use it in ASM
First you need to create a ServiceEntry to expose the llm service in ASM.  
You can learn how to use wasm image in ASM from [here](https://help.aliyun.com/zh/asm/user-guide/use-coraza-wasm-plug-in-to-implement-waf-capability-on-asm-gateway?spm=a2c4g.11186623.0.i60).  
This plugin's related WasmPlugin CR is in `./asm/wasm-plugin.yaml`.  
### Plugin Reference
```go
type LLMProxyConfig struct {
	Hosts            []string          `json:"hosts"`               // match request's host header. if not matched, llm proxy will not be enabled.
	API_KEY          string            `json:"api_key"`             // api_key for dashscope
	AllowPatterns    []string          `json:"allow_patterns"`      // regex list for allow.
	DenyPatterns     []string          `json:"deny_patterns"`       // regex list for deny.
	IntelligentGuard *IntelligentGuard `json:"intelligent_guard"`   // intelligent guard: use llm to check whether the request should be blocked.
}

type IntelligentGuard struct {
	Host    *string `json:"host"` // host header in check request, example: dashscope.aliyuncs.com. Must be defined in ServiceEntry.
	Port    *uint32 `json:"port"`    // ServiceEntry's http port.
	Path    *string `json:"path"`    // default "/compatible-mode/v1/chat/completions"
	Model   *string `json:"model"`   // default qwen-turbo
	API_KEY *string `json:"api_key"` // api_key for dashscope, can not be empty
}
```
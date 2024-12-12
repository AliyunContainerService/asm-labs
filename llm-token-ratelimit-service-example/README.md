# 限流服务示例
本示例需要与ASM提供的LLM token限流插件配合使用。
该服务主要对外提供了两个API：
* /ratelimit
* /update_ratelimit_record
## /ratelimit
该接口用于查询是否对指定请求进行限流，返回限流结果。
请求方法为GET。
请求路径式为：
```
# ratelimit_keys的值是一个json格式的字符串list
/ratelimit?ratelimit_keys=["test-key-1","test-key-2"]
```
限流服务需要根据这些key来判断是否进行限流。
响应状态码为非200时，表示限流服务异常。限流插件会根据配置的fallOpen参数决定放行请求还是拒绝请求。
响应状态为200时，表示限流服务正常。
此时请求应该携带如下格式的body，用于判定是否限流：
```json
{
    "allow": true,
    "description": "why not allow"
}
```
## /update_ratelimit_record
该接口用于更新限流记录。
请求方法为POST。
请求格式为：
```json
{
    "ratelimit_keys": ["test-key-1","test-key-2"],
    "prompt_tokens": 100,
    "completion_tokens": 100,
    "total_tokens": 200
}
```
限流服务需要根据这些key来更新限流记录，用于下次限流判断。

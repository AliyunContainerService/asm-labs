# 缓存服务示例
本示例需要与ASM提供的LLM缓存插件配合使用。
该服务主要提供了两个API：
* /lookup
* /update
## /lookup
该接口用于查询缓存，返回缓存结果。
请求方法为POST。
请求体格式为：
```json
{
    "request": {
        "headers": {
            "key": "value"
        },
        "body": "llm-request-body"
    }
}
```
响应格式为：
```json
{
    "response":     {
        "headers": {
            "key": "value"
        },
        "body": "llm-response-body"
    }
}
```
## /update
该接口用于更新缓存。
请求方法为POST。
请求格式为：
```json
{
    "request": {
        "headers": {
            "key": "value"
        },
        "body": "llm-request-body"
    },
    "response":     {
        "headers": {
            "key": "value"
        },
        "body": "llm-response-body"
    }
}
```

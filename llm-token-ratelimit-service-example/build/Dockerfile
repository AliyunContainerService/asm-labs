FROM golang:1.22 AS builder

RUN mkdir llm-token-ratelimit-service-example
WORKDIR /llm-token-ratelimit-service-example
COPY ./ .
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
RUN go build -o llm-token-ratelimit-service-example main.go

FROM alpine:3.18
WORKDIR /
COPY --from=builder /llm-token-ratelimit-service-example/llm-token-ratelimit-service-example /llm-token-ratelimit-service-example
ENTRYPOINT [ "/llm-token-ratelimit-service-example" ]

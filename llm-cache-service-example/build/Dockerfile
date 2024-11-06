FROM golang:1.22 AS builder

RUN mkdir llm-cache-service-example
WORKDIR /llm-cache-service-example
COPY ./ .
RUN go build -o llm-cache-service-example main.go

FROM alpine:3.18
WORKDIR /
COPY --from=builder /llm-cache-service-example/llm-cache-service-example /llm-cache-service-example
ENTRYPOINT [ "/llm-cache-service-example" ]

ARG GO_VERSION=1.14.13

# BASE_DISTRIBUTION is used to switch between the old base distribution and distroless base images
ARG BASE_DISTRIBUTION=default

# Version is the base image version from the TLD Makefile
ARG BASE_VERSION=latest

# Build the manager binary
FROM registry.cn-hangzhou.aliyuncs.com/acs/asm-se-syncer:buildbase.v1 AS builder

##RUN apk add --update --no-cache ca-certificates make git curl mercurial

ARG PACKAGE=gitlab.alibaba-inc.com/cos/asm-se-syncer

RUN mkdir -p /go/src/${PACKAGE}
WORKDIR /go/src/${PACKAGE}

COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor  vendor/
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o asm-se-syncer gitlab.alibaba-inc.com/cos/asm-se-syncer/cmd/...


# The following section is used as base image if BASE_DISTRIBUTION=default
#FROM registry.cn-hangzhou.aliyuncs.com/acs/istio_base:1.6-dev.2 as default

# This will build the final image based on either default or distroless from above
# hadolint ignore=DL3006
FROM registry.cn-hangzhou.aliyuncs.com/acs/istio-operator:runtimebase.v2

#FROM alpine:3.11.6
#RUN apk add --update ca-certificates \
# && apk add --update -t deps curl \
# && curl -L https://storage.googleapis.com/kubernetes-release/release/v1.14.9/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
# && chmod +x /usr/local/bin/kubectl \
WORKDIR /

COPY --from=builder /go/src/gitlab.alibaba-inc.com/cos/asm-se-syncer/asm-se-syncer /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/asm-se-syncer"]

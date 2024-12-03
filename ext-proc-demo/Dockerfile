FROM golang:1.23-alpine as build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /ext-proc-demo
COPY . .
RUN go mod download
RUN go build -o /ext-proc

FROM alpine:latest
WORKDIR /
COPY --from=build /ext-proc /ext-proc

ENTRYPOINT ["/ext-proc"]
FROM golang:1.16-alpine

WORKDIR /app


COPY *.go ./

RUN go build -o /main main.go

EXPOSE 8000

CMD [ "/main" ]

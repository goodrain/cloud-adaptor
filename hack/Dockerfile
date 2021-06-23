FROM golang:1.13 as builder
ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV GOPROXY=https://goproxy.io

WORKDIR /go/src/github.com/goodrain/cloud-adaptor
COPY . .

ARG GOARCH
ARG LDFLAGS
RUN go build -ldflags "$LDFLAGS" -o /cloud-adaptor $PWD/cmd/cloud-adaptor

FROM alpine:3.11.2
RUN apk add --update tzdata \
    && mkdir /app \
    && apk add --update apache2-utils \
    && rm -rf /var/cache/apk/*
ENV TZ=Asia/Shanghai
COPY --from=builder cloud-adaptor /app

ENTRYPOINT ["/app/cloud-adaptor"]
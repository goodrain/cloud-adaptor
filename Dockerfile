FROM golang:1.15 as builder
ENV CGO_ENABLED=0
ENV GOPATH=/go

WORKDIR /go/src/goodrain.com/cloud-adaptor/
COPY . .

ARG LDFLAGS
RUN go build -ldflags "$LDFLAGS" -o /cloud-adaptor

FROM alpine:3.8
ADD chart /app/chart
RUN apk add --no-cache tzdata bash && \
    wget https://goodrain-pkg.oss-cn-shanghai.aliyuncs.com/pkg/helm && chmod +x helm && mv helm /usr/local/bin/
ENV HELM_PATH=/usr/local/bin/helm
ENV CHART_PATH=/app/chart
ENV CONFIG_DIR=/data
ENV DB_PATH=/data/db
ENV TZ=Asia/Shanghai
COPY --from=builder cloud-adaptor /app
VOLUME [ "/data" ]

ENTRYPOINT ["/app/cloud-adaptor"]
CMD [ "server" ]
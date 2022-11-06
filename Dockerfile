FROM rainbond/golang-gcc-buildstack:1.17-alpine3.16 as builder
ENV CGO_ENABLED=1
ENV GOPATH=/go
ADD . /go/src/goodrain.com/cloud-adaptor/
WORKDIR /go/src/goodrain.com/cloud-adaptor/
#ENV GOPROXY=https://goproxy.cn
RUN go build -ldflags "-w -s -linkmode external -extldflags '-static'" -o cloudadaptor ./cmd/cloud-adaptor

FROM goodrainapps/alpine:3.16
WORKDIR /run
ARG RELEASE_DESC

RUN mkdir -p /app && wget "https://pkg.goodrain.com/pkg/helm" && chmod +x helm && mv helm /app/helm
COPY --from=builder /go/src/goodrain.com/cloud-adaptor/cloudadaptor /run/cloudadaptor
COPY ./chart /app/chart

ENV RELEASE_DESC=${RELEASE_DESC}
CMD ["/run/cloudadaptor", "daemon"]

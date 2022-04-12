FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build/zero

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
COPY service/userbehavior/rpc/etc /app/etc
RUN go build -ldflags="-s -w" -o /app/rpcuserbehavior service/userbehavior/rpc/rpcuserbehavior.go


FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/rpcuserbehavior /app/rpcuserbehavior
COPY --from=builder /app/etc /app/etc

CMD ["./rpcuserbehavior", "-f", "etc/rpcuserbehavior.dev.yaml"]

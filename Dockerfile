FROM golang:1.14-alpine AS builder1

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on GOPROXY=https://goproxy.cn,direct go build -o baidu-su .

FROM starudream/alpine:latest AS builder2

COPY --from=builder1 /build/baidu-su /baidu-su

RUN apk add --no-cache upx && upx -9 -q /baidu-su

FROM starudream/alpine-glibc:latest

WORKDIR /

COPY config.json config.json
COPY --from=builder2 /baidu-su /baidu-su

CMD /baidu-su

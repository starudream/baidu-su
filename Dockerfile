FROM golang:1.15-alpine AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build -o baidu-su .

RUN apk add --no-cache upx && upx baidu-su

FROM starudream/alpine-glibc:latest

COPY config.json config.json
COPY --from=builder /build/baidu-su /baidu-su

WORKDIR /

CMD /baidu-su

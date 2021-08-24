FROM starudream/golang AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build -o baidu-su .

FROM starudream/upx AS upx

COPY --from=builder /build/baidu-su /build/baidu-su

RUN upx /build/baidu-su

FROM starudream/alpine-glibc:latest

COPY config.json config.json
COPY --from=upx /build/baidu-su /baidu-su

WORKDIR /

CMD /baidu-su

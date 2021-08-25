FROM starudream/golang AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build -o baidu-su . ; upx -V && upx /build/baidu-su || exit 0

FROM starudream/alpine-glibc:latest

COPY config.json config.json

COPY --from=builder /build/baidu-su /baidu-su

WORKDIR /

CMD /baidu-su

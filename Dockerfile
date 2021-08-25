FROM starudream/golang AS builder

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GO111MODULE=on go build -o baidu-su . && if type upx >/dev/null 2>&1; then upx baidu-su; fi

FROM starudream/alpine-glibc:latest

COPY config.json config.json

COPY --from=builder /build/baidu-su /baidu-su

WORKDIR /

CMD /baidu-su

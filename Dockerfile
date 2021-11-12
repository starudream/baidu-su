FROM starudream/golang AS builder

WORKDIR /build

COPY . .

RUN apk add --no-cache alpine-sdk && make bin && make upx

FROM starudream/alpine-glibc:latest

WORKDIR /

COPY config.json config.json

COPY --from=builder /build/bin/app /app

CMD /app

# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /iptv ./cmd/iptv/main.go

# 运行阶段
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /iptv /app/iptv
COPY config.yml /app/
COPY logos /app/logos/

# 可配置的环境变量
ENV INTERVAL="24h" \
    PORT="8088" \
    UPSTREAM_URL="http://192.168.3.1:4022"

EXPOSE ${PORT}
ENTRYPOINT ["./iptv", "serve"]
CMD ["-i", "${INTERVAL}", "-p", "${PORT}", "-u", "${UPSTREAM_URL}"]

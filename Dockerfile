# 使用多阶段构建
# 第一阶段：构建应用
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
ARG TARGETOS TARGETARCH TARGETVARIANT
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build -ldflags="-w -s" -o /iptv ./cmd/iptv/main.go

# 第二阶段：运行环境
FROM alpine:3.19

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 从构建阶段复制二进制文件和资源
COPY --from=builder /iptv /iptv
COPY config.yml /config.yml
COPY logos /logos

# 设置默认环境变量
ENV INTERVAL="24h" \
    PORT="8088" \
    URL="http://192.168.3.1:4022"

# 暴露端口
EXPOSE ${PORT}

# 设置入口点
ENTRYPOINT ["/iptv", "serve"]
CMD ["-i", "${INTERVAL}", "-p", "${PORT}", "-u", "${URL}"]

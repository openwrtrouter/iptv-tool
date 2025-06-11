# 构建阶段
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o iptv .

# 最终镜像
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/iptv .
# 设置默认环境变量
ENV INTERVAL=24h
ENV PORT=8088
ENV INNER_URL=http://192.168.3.1:4022
# 暴露端口
EXPOSE $PORT
# 设置入口点
ENTRYPOINT ["./iptv"]
CMD ["serve", "-i", "${INTERVAL}", "-p", "${PORT}", "-u", "inner=${INNER_URL}"]

FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o iptv .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/iptv .
ENTRYPOINT ["./iptv"]
CMD ["serve"]

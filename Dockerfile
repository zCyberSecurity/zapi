FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o zapi ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/zapi .

ENV ADDR=:8080
ENV DB_PATH=/data/zapi.db
ENV ADMIN_TOKEN=change-me

VOLUME ["/data"]
EXPOSE 8080

CMD ["./zapi"]

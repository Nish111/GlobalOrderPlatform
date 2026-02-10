FROM golang:1.21-alpine AS builder

WORKDIR /app

# Initialize module inside container
RUN go mod init github.com/Nish111/GlobalOrderPlatform

# Install system dependencies for confluent-kafka-go (needs librdkafka)
RUN apk add --no-cache build-base pkgconf

# Download dependencies (we add the imports to a dummy file to force download, 
# or just rely on 'go get' in the next step. Since we copied main.go, go mod tidy works)
COPY cmd/order-service/main.go ./cmd/order-service/
COPY cmd/kitchen-service/main.go ./cmd/kitchen-service/

# Fetch deps
RUN go get github.com/confluentinc/confluent-kafka-go/kafka
RUN go get github.com/gin-gonic/gin
RUN go get github.com/hamba/avro
RUN go get github.com/google/uuid
RUN go get github.com/golang-jwt/jwt/v5

# Build the binaries
RUN go build -tags musl -o order-service ./cmd/order-service/main.go
RUN go build -tags musl -o kitchen-service ./cmd/kitchen-service/main.go

FROM alpine:latest

WORKDIR /root/

# Install runtime libs for Kafka
RUN apk add --no-cache libstdc++

COPY --from=builder /app/order-service .
COPY --from=builder /app/kitchen-service .

EXPOSE 8080

# Default command (can be overridden in docker-compose)
CMD ["./order-service"]

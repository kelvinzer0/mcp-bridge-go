FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o mcp-bridge ./cmd/mcp-bridge

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/mcp-bridge .

EXPOSE 8080
ENV HOST=127.0.0.1
ENV PORT=8080

ENTRYPOINT ["/app/mcp-bridge"]

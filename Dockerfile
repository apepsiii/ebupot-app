# ============================================================
#  Dockerfile — e-Bupot Portal
#  Multi-stage build: compile di golang image, jalankan di alpine
# ============================================================

# --- Stage 1: Build ---
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Cache dependency download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (CGO disabled — pure-Go SQLite driver)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ebupot-app .

# --- Stage 2: Runtime ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary
COPY --from=builder /build/ebupot-app .

# Copy runtime assets (dibaca dari disk saat aplikasi berjalan)
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/public ./public
COPY --from=builder /build/config.yaml ./config.yaml

# Buat folder data & uploads
RUN mkdir -p data uploads/ebupots

# Volume untuk persistensi database & file upload
VOLUME ["/app/data", "/app/uploads"]

EXPOSE 8080

CMD ["./ebupot-app"]

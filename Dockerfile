FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" -o /bin/chalkboard ./cmd/chalkboard/
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=builder /bin/chalkboard /usr/local/bin/chalkboard
ENV PORT="8960" DATA_DIR="/data" CHALKBOARD_LICENSE_KEY=""
EXPOSE 8960
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD curl -sf http://localhost:8960/health || exit 1
ENTRYPOINT ["chalkboard"]

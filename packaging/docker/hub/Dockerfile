# Build the embedded frontend with the repository's locked pnpm dependency graph.
FROM node:22-alpine AS web-builder
WORKDIR /src/web
RUN npm install --global pnpm@10.6.5
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm run build

# Build the Hub with the frontend dist available for go:embed.
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web-builder /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/cadentra-hub ./cmd/hub

# Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/cadentra-hub /usr/local/bin/cadentra-hub
RUN mkdir -p /var/lib/cadentra-hub /var/lib/cadentra-hub/artifacts
EXPOSE 8080 8443
VOLUME ["/var/lib/cadentra-hub"]
ENTRYPOINT ["/usr/local/bin/cadentra-hub"]

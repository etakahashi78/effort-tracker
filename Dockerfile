# Build stage
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /app

# Install certificates and git
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Support multi-platform builds (amd64, arm64, etc.)
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o server ./cmd/server

# Production stage
# Distroless static is highly secure and lightweight
FROM gcr.io/distroless/static-debian12:latest

# Use nonroot user for security (UID 65532)
USER nonroot:nonroot

# Copy compiled binary from builder
COPY --from=builder /app/server /server

EXPOSE 8080

ENV ADDR=:8080

ENTRYPOINT ["/server"]

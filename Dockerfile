FROM --platform=$BUILDPLATFORM golang:1.24.5-alpine AS builder

# Install build dependencies
# RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build arguments for LDFLAGS
ARG LDFLAGS=""
ARG TARGETOS
ARG TARGETARCH

# Build the application
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on \
    go build -ldflags="$LDFLAGS" -a -o mindmap-server cmd/*.go

FROM gcr.io/distroless/static-debian12:nonroot

# Use non-root user (65532:65532)
USER 65532:65532
COPY --from=builder --chown=65532:65532 /app/mindmap-server /mindmap-server

EXPOSE 8080

ENTRYPOINT ["/mindmap-server"]
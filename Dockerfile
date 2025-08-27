# --- Builder stage ---
FROM golang:1.24 AS builder

WORKDIR /app

# Install buf CLI
ENV BUF_VERSION=1.28.1
RUN curl -sSL -o /usr/local/bin/buf "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
    && chmod +x /usr/local/bin/buf

# Install wire
RUN go install github.com/google/wire/cmd/wire@v0.6.0

# Install protoc-gen-go and protoc-gen-go-grpc for buf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Copy the rest of the source
COPY . .

# Generate protobuf code
RUN cd proto && buf generate

# Generate wire code
RUN cat -n internal/di/wire.go | head -25
RUN cd internal/di && wire

# Build the binary using vendor mode
RUN GOOS=linux CGO_ENABLED=0 go build -mod=vendor -o /app/bin/chain-xrpl ./cmd/chain-xrpl

# --- Final stage ---
FROM alpine:3.22 AS final

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/bin/chain-xrpl /app/chain-xrpl
COPY config.yaml /app/config.yaml

ENTRYPOINT ["/app/chain-xrpl"] 
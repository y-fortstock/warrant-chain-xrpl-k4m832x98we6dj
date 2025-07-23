# --- Builder stage ---
FROM golang:1.24

# Install buf CLI
ENV BUF_VERSION=1.28.1
RUN curl -sSL -o /usr/local/bin/buf "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
    && chmod +x /usr/local/bin/buf

# Install wire
RUN go install github.com/google/wire/cmd/wire@v0.6.0

# Install protoc-gen-go and protoc-gen-go-grpc for buf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

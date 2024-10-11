# Use the official Alpine image as a base
FROM python:3.10.9-alpine3.17

# Set the maintainer
LABEL maintainer="Caio Cominato"
LABEL version="0.1"
LABEL description="HARPY container using Python3.10.9 and Go1.23.0"

# Set the Go version
ENV GO_VERSION=1.23.0

# Install required dependencies and specific Go version
RUN apk update && \
    apk add --no-cache \
    bash \
    build-base \
    libffi-dev \
    openssl-dev \
    curl \
    git && \
    cd /tmp && \
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt && \
    rm -rf /tmp/* /var/cache/apk/*

# Set environment variables
ENV PATH="/opt/venv/bin:$PATH"
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir /app
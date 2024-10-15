# Use the official Debian-based Python image as a base
FROM python:3.10.9-slim-bullseye

# Set the maintainer
LABEL maintainer="Caio Cominato"
LABEL version="0.1"
LABEL description="HARPY container using Python3.10.9, Go1.23.0, and Rust"

# Set the Go and Rust versions
ENV GO_VERSION=1.23.0

# Install required dependencies
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    bash \
    build-essential \
    libffi-dev \
    libssl-dev \
    curl \
    git \
    cmake \
    gcc \
    g++ \
    make \
    libboost-all-dev \
    zlib1g-dev \
    libbz2-dev \
    libsnappy-dev \
    libbrotli-dev \
    liblz4-dev \
    libzstd-dev \
    libgflags-dev \
    libjemalloc-dev \
    libutf8proc-dev \
    rapidjson-dev \
    wget && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Go
RUN cd /tmp && \
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt && \
    rm -rf /tmp/*

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y && \
    . "$HOME/.cargo/env" && \
    rm -rf /var/lib/apt/lists/*

# Add Apache Arrow repository
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    lsb-release \
    wget \
    gnupg && \
    wget https://apache.jfrog.io/artifactory/arrow/$(lsb_release --id --short | tr 'A-Z' 'a-z')/apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt-get install -y ./apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    libarrow-dev \
    libarrow-glib-dev \
    libarrow-dataset-dev \
    libarrow-dataset-glib-dev \
    libarrow-acero-dev \
    libarrow-flight-dev \
    libarrow-flight-glib-dev \
    libarrow-flight-sql-dev \
    libarrow-flight-sql-glib-dev \
    libgandiva-dev \
    libgandiva-glib-dev \
    libparquet-dev \
    libparquet-glib-dev && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* ./apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb

# Set environment variables
ENV PATH="/opt/venv/bin:$PATH"
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
ENV PATH="/root/.cargo/bin:$PATH"
ENV CMAKE_PREFIX_PATH="/usr/local/lib/cmake/arrow"

RUN mkdir /app
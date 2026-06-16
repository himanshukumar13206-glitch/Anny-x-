# Stage 1: Build ntgcalls (if needed)
FROM ubuntu:22.04 AS builder

RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    cmake \
    pkg-config \
    libssl-dev \
    libopus-dev \
    libvpx-dev \
    libx264-dev \
    libavcodec-dev \
    libavformat-dev \
    libavutil-dev \
    libswresample-dev \
    libsrtp2-dev \
    && rm -rf /var/lib/apt/lists/*

# Clone and build ntgcalls (skip if you have prebuilt)
WORKDIR /build
RUN git clone --depth 1 https://github.com/Laky-64/ntgcalls.git \
    && cd ntgcalls \
    && mkdir build && cd build \
    && cmake .. \
    && make -j$(nproc) \
    && make install

# Stage 2: Build the bot
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ffmpeg \
    libopus0 \
    libvpx7 \
    libx264-164 \
    libavcodec-extra \
    libavformat58 \
    libavutil56 \
    libswresample4 \
    libsrtp2-1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy compiled ntgcalls library from builder
COPY --from=builder /usr/local/lib/libntgcalls.so /usr/local/lib/
COPY --from=builder /usr/local/include/ntgcalls /usr/local/include/ntgcalls

# Set LD_LIBRARY_PATH
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the app with CGO
RUN CGO_ENABLED=1 go build -v -trimpath -ldflags="-w -s" -o app ./cmd/app

# Expose port (Render sets PORT env)
EXPOSE 8080

# Run the bot
CMD ["./app"]

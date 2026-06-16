# Stage 1: Build ntgcalls (only needed if you don't have a prebuilt)
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

WORKDIR /build
RUN git clone --depth 1 https://github.com/Laky-64/ntgcalls.git \
    && cd ntgcalls \
    && mkdir build && cd build \
    && cmake .. \
    && make -j$(nproc) \
    && make install

# Stage 2: Runtime image
FROM ubuntu:22.04

# Install ONLY the runtime dependencies that are actually needed
RUN apt-get update && apt-get install -y \
    ffmpeg \
    libopus0 \
    libvpx7 \
    libsrtp2-1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the ntgcalls library from builder
COPY --from=builder /usr/local/lib/libntgcalls.so /usr/local/lib/
COPY --from=builder /usr/local/include/ntgcalls /usr/local/include/ntgcalls

ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -v -trimpath -ldflags="-w -s" -o app ./cmd/app

EXPOSE 8080
CMD ["./app"]

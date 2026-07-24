# ============ 构建阶段 ============
FROM rust:1-slim-bookworm AS builder

RUN apt-get update && apt-get install -y pkg-config libssl-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY Cargo.toml Cargo.lock* ./
COPY src/ ./src/

RUN cargo build --release && \
    strip target/release/copymanga-web && \
    cp target/release/copymanga-web /copymanga-web

# ============ 运行阶段 ============
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates libssl3 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /copymanga-web /usr/local/bin/copymanga-web

ENV DOWNLOAD_DIR=/downloads
ENV PORT=3000

EXPOSE 3000

VOLUME ["/downloads"]

CMD ["copymanga-web"]

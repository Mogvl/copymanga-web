# ============ 构建阶段 ============
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates

# 复制 Go 模块文件
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端代码
COPY backend/ ./

# 构建后端
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ============ 运行阶段 ============
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 复制前端静态文件（如果 CI 没有预先构建）
COPY backend/static/ ./static/

# 创建下载目录
RUN mkdir -p /downloads

ENV DOWNLOAD_DIR=/downloads
ENV STATIC_DIR=/app/static
ENV PORT=8080

EXPOSE 8080

VOLUME ["/downloads"]

CMD ["./server"]

#!/bin/bash
# ============ 构建 copymanga-web ============
set -e

echo "🚀 构建 copymanga-web..."

# 检查是否安装了必要的工具
check_command() {
  if ! command -v "$1" &> /dev/null; then
    echo "❌ 未找到 $1，请先安装"
    exit 1
  fi
}

check_command go
check_command node
check_command npm

# 构建前端
echo "📦 构建前端..."
cd frontend
npm install
npm run build
cd ..

# 复制前端到后端
echo "📋 复制前端静态文件..."
mkdir -p backend/static
cp -r frontend/dist/* backend/static/

# 构建后端
echo "🔨 构建后端..."
cd backend
go build -o server ./cmd/server
cd ..

echo "✅ 构建完成!"
echo ""
echo "运行方式："
echo "  cd backend && DOWNLOAD_DIR=./downloads STATIC_DIR=./static ./server"
echo ""
echo "Docker 构建："
echo "  docker build -t copymanga-web ."
echo "  docker run -p 8080:8080 -v /path/to/downloads:/downloads copymanga-web"

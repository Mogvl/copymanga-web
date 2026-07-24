#!/bin/bash
# ============ 构建 copymanga-web Docker 镜像 ============
set -e

echo "🚀 构建 copymanga-web Docker 镜像..."

# 构建 Linux amd64 镜像
docker build -t copymanga-web:latest .

echo "✅ 构建成功!"
echo ""
echo "导出镜像文件:"
echo "  docker save copymanga-web:latest -o copymanga-web.tar"
echo ""
echo "在绿联云上部署:"
echo "  1. 将 copymanga-web.tar 上传到绿联云"
echo "  2. SSH 连接绿联云，执行:"
echo "     docker load -i copymanga-web.tar"
echo "     mkdir -p /volume1/漫画"
echo "     docker compose up -d"
echo ""
echo "  或者在绿联云 Docker 项目里直接复制 docker-compose.yml 的内容"

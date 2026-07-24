# =============================================
#  copymanga-web 一键部署工具
#  用法: 在 PowerShell 中执行此脚本
# =============================================

$ErrorActionPreference = "Stop"
$PROJECT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$OUTPUT_TAR = "$PROJECT_DIR\copymanga-web.tar"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  📚 copymanga-web 一键部署" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Docker
$dockerCheck = Get-Command "docker" -ErrorAction SilentlyContinue
if (-not $dockerCheck) {
    Write-Host "❌ 没有找到 Docker！" -ForegroundColor Red
    Write-Host ""
    Write-Host "请先安装 Docker Desktop：https://www.docker.com/products/docker-desktop/"
    Write-Host "安装完成后重启终端再试"
    Read-Host "按回车键退出"
    exit 1
}

Write-Host "✅ Docker 已安装" -ForegroundColor Green
Write-Host ""

# 选择部署方式
Write-Host "请选择部署方式：" -ForegroundColor Yellow
Write-Host "  1) GitHub Actions 自动构建（推荐）" -ForegroundColor White
Write-Host "  2) 本地 Docker Desktop 构建" -ForegroundColor White
$choice = Read-Host "请输入 1 或 2"

if ($choice -eq "1") {
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "  GitHub Actions 自动构建" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "步骤：" -ForegroundColor Yellow
    Write-Host "1. 在 https://github.com/new 创建一个新仓库" -ForegroundColor White
    Write-Host "2. 在项目目录下执行：" -ForegroundColor White
    Write-Host ""
    Write-Host "   git init" -ForegroundColor Green
    Write-Host "   git add ." -ForegroundColor Green
    Write-Host "   git commit -m '初始提交'" -ForegroundColor Green
    Write-Host "   git branch -M main" -ForegroundColor Green
    Write-Host "   git remote add origin https://github.com/你的用户名/你的仓库名.git" -ForegroundColor Green
    Write-Host "   git push -u origin main" -ForegroundColor Green
    Write-Host ""
    Write-Host "3. 推送后 GitHub Actions 会自动构建镜像并推送到 ghcr.io" -ForegroundColor White
    Write-Host "4. 在绿联云上修改 docker-compose.yml：用 ghcr.io 的镜像" -ForegroundColor White
    Write-Host ""
    Write-Host '   将 image: ghcr.io/你的用户名/copymanga-web:latest 取消注释' -ForegroundColor Green
    Write-Host ""
    Write-Host "5. 然后在绿联云 Docker → 项目 → 粘贴 docker-compose.yml 启动" -ForegroundColor White

} elseif ($choice -eq "2") {
    Write-Host ""
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host "  本地 Docker Desktop 构建" -ForegroundColor Cyan
    Write-Host "======================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Step 1: 构建 Docker 镜像..." -ForegroundColor Yellow
    docker build -t copymanga-web:latest "$PROJECT_DIR"
    Write-Host "✅ 构建完成" -ForegroundColor Green

    Write-Host ""
    Write-Host "Step 2: 导出镜像文件..." -ForegroundColor Yellow
    docker save copymanga-web:latest -o $OUTPUT_TAR
    Write-Host "✅ 导出完成: $OUTPUT_TAR" -ForegroundColor Green

    Write-Host ""
    Write-Host "Step 3: 部署到绿联云" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "请将以下两个文件上传到绿联云（通过文件管理器或 SMB）：" -ForegroundColor White
    Write-Host "  1) $OUTPUT_TAR" -ForegroundColor Green
    Write-Host "  2) $PROJECT_DIR\docker-compose.yml" -ForegroundColor Green
    Write-Host ""
    Write-Host "上传后，在绿联云 Docker → 项目 → 新建项目：" -ForegroundColor White
    Write-Host "  项目名：copymanga-web" -ForegroundColor White
    Write-Host "  粘贴 docker-compose.yml 的内容" -ForegroundColor White
    Write-Host "  注意修改 volumes 里的路径！" -ForegroundColor Red
    Write-Host ""
    Write-Host "或者 SSH 执行：" -ForegroundColor White
    Write-Host "  docker load -i /上传路径/copymanga-web.tar" -ForegroundColor Green
    Write-Host "  docker compose up -d" -ForegroundColor Green
} else {
    Write-Host "输入无效，已退出" -ForegroundColor Red
    Read-Host "按回车键退出"
    exit 1
}

Read-Host "按回车键退出"

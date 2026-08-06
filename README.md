# 拷贝漫画 Web 下载器 (copymanga-web)

从拷贝漫画 (copymanga) 下载漫画的 **Web 服务版**，无需桌面环境，适合部署在 **NAS（绿联云、群晖等）、VPS、树莓派** 上。

## 功能

- ✅ **搜索漫画** — 按关键词搜索
- ✅ **查看详情** — 查看漫画信息、章节列表
- ✅ **多选下载** — 勾选章节，一键下载
- ✅ **并发下载** — 多章节、多图片并发下载，速度飞快
- ✅ **下载进度** — 实时查看下载进度
- ✅ **下载任务** — 任务页实时刷新，失败自动重试并提示原因
- ✅ **高清图** — 自动下载高清版图片 (`.c1500x.`)
- ✅ **已下载列表** — 查看已下载的漫画
- ✅ **Web 界面** — 浏览器访问即可操作
- ✅ **Docker 部署** — 一键部署

## 快速开始

### 1. 在你的电脑上构建

需要安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)。

```bash
# 进入项目目录
cd copymanga-web

# 构建镜像
docker build -t copymanga-web:latest .

# 导出镜像文件
docker save copymanga-web:latest -o copymanga-web.tar
```

### 2. 部署到绿联云

**方式一：SSH + Docker Compose（推荐）**

```bash
# 1. 先上传 copymanga-web.tar 和 docker-compose.yml 到绿联云
# 例如上传到 /volume1/docker/copymanga/

# 2. SSH 登录绿联云
ssh root@192.168.x.x

# 3. 加载镜像
cd /volume1/docker/copymanga
docker load -i copymanga-web.tar

# 4. 修改 docker-compose.yml，把 /volume1/漫画 换成你的真实路径
#    然后启动
docker compose up -d
```

**方式二：绿联云 Docker 项目**
1. 打开绿联云 → **Docker** → **项目** → **新建项目**
2. 项目名：`copymanga-web`
3. 粘贴 `docker-compose.yml` 的内容
4. 点击部署

### 3. 使用

1. 浏览器打开 `http://绿联云IP:3000`
2. 搜索漫画
3. 选择章节 → 下载
4. 在绿联云文件管理器里查看下载的漫画

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DOWNLOAD_DIR` | 漫画下载目录 | `/downloads` |
| `STATIC_DIR` | 前端静态文件目录 | `/app/static` |
| `PORT` | Web 服务端口（容器内） | `8080` |
| `TOKEN` | 登录 Token（可加快捷访问，可选） | 空 |
| `IMAGE_CONCURRENCY` | 单章节内图片并发数 | `5` |
| `IMAGE_RETRY` | 图片下载失败重试次数 | `3` |
| `CHAPTER_INTERVAL_SEC` | 章节下载完成后的间隔秒数（缓解风控） | `0` |

### 下载目录结构

```
{下载目录}/{漫画名}/{分组名}/{排序} {章节名}/{真实页码}.{格式}
```

例如：`/downloads/海贼王学院/默认/031 第31话/00031.jpg`

- 图片文件名按**真实页码**编号，与网站展示顺序一致
- 每章一个文件夹，自动按排序号 `001`、`002`… 排列
- 图片格式可选 **WebP / JPG**（章节详情页选择）

## 拉取镜像

```bash
docker pull ghcr.io/mogvl/copymanga-web:latest
```

## docker-compose.yml 配置

```yaml
version: '3.8'
services:
  copymanga-web:
    image: ghcr.io/mogvl/copymanga-web:latest
    container_name: copymanga-web
    ports:
      - "3000:8080"
    environment:
      - DOWNLOAD_DIR=/downloads
      - STATIC_DIR=/app/static
      - PORT=8080
    volumes:
      - /volume1/漫画:/downloads
    restart: unless-stopped
```

> ⚠️ **记得把 `/volume1/漫画` 改成你存漫画的真实路径！**

## 从源码构建

需要安装 Go 1.22+、Node.js 20+ 和 Docker：

```bash
# 构建前端
cd frontend
npm install
npm run build
cp -r dist/* ../backend/static/

# 构建后端
cd ../backend
go build -o server ./cmd/server

# 或者直接构建 Docker 镜像
docker build -t copymanga-web .
```

## 技术栈

- **后端**: Go + Gin
- **前端**: Vue3 + Vite + TypeScript
- **UI 风格**: MIUI 设计语言（渐变背景 + 浮动卡片）
- **容器**: Docker 多阶段构建

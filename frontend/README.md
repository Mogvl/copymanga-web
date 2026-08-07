# 前端说明 (frontend)

基于 **Vue 3 + TypeScript + Vite** 的单页应用，提供 `copymanga-web` 的 Web 界面。

## 本地开发运行

```bash
npm install
npm run dev   # 启动 Vite 开发服务器，监听 http://localhost:3000（热更新）
```

> 开发服务器默认监听 **3000**，并把 `/api` 代理到本地 Go 后端 **3001**（见 `vite.config.ts`）。

## 构建

```bash
npm run build   # 产物输出到 frontend/dist/
```

构建后把 `dist/*` 复制到 `backend/static/` 由 Go 服务托管，或交给 CI（`.github/workflows/docker-build.yml`）在构建镜像时自动完成。

## 目录结构

- `src/views/` — 页面视图：`HomeView`(首页)、`SearchView`(搜索)、`ComicView`(漫画详情/下载/PDF)、`TasksView`(下载任务)、`DownloadedView`(已下载)
- `src/api/client.ts` — 后端 API 封装
- `src/types/index.ts` — TypeScript 类型定义
- `src/router/index.ts` — 前端路由

use std::sync::Arc;

use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::{Html, IntoResponse, Json},
    routing::{get, post},
    Router,
};
use serde::Deserialize;
use tokio::sync::Mutex;
use tower_http::cors::CorsLayer;

mod client;
mod downloader;
mod types;

use client::CopyMangaClient;
use downloader::DownloadManager;
use types::*;

/// 应用状态
struct AppState {
    client: CopyMangaClient,
    manager: Arc<DownloadManager>,
}

// ============ 查询参数 ============

#[derive(Deserialize)]
struct SearchQuery {
    q: String,
    #[serde(default = "default_page")]
    page: i64,
}

fn default_page() -> i64 { 1 }

#[derive(Deserialize)]
struct ComicQuery {
    path: String,
}

#[derive(Deserialize)]
struct DownloadReq {
    comic_name: String,
    comic_path_word: String,
    group_path_word: String,
    chapter_uuids: Vec<String>,
}

// ============ 页面路由 ============

/// 首页 - 搜索页面
async fn index() -> Html<String> {
    Html(render_template("index", &serde_json::json!({
        "title": "拷贝漫画下载器 - Web版"
    })))
}

/// 搜索
async fn search(
    State(state): State<Arc<Mutex<AppState>>>,
    Query(q): Query<SearchQuery>,
) -> Result<Html<String>, AppError> {
    let state = state.lock().await;
    let results = state.client.search(&q.q, q.page).await?;

    let comics: Vec<serde_json::Value> = results
        .iter()
        .map(|c| {
            serde_json::json!({
                "name": c.name,
                "path_word": c.path_word,
                "cover": c.cover,
                "author": c.author.iter().map(|a| a.name.clone()).collect::<Vec<_>>().join(", "),
                "popular": c.popular,
            })
        })
        .collect();

    Ok(Html(render_template("search_results", &serde_json::json!({
        "keyword": q.q,
        "comics": comics,
    }))))
}

/// 漫画详情页
async fn comic_detail(
    State(state): State<Arc<Mutex<AppState>>>,
    Query(query): Query<ComicQuery>,
) -> Result<Html<String>, AppError> {
    let path_word = &query.path;
    let state = state.lock().await;

    // 获取漫画信息
    let comic_data = state.client.get_comic(&path_word).await?;

    // 获取所有分组的章节
    let mut groups = Vec::new();
    for (group_key, group_info) in &comic_data.groups {
        let chapters = state
            .client
            .get_group_chapters(&path_word, &group_info.path_word)
            .await?;

        let chapters_view: Vec<ChapterView> = chapters
            .iter()
            .map(|c| ChapterView {
                uuid: c.uuid.clone(),
                title: c.name.clone(),
                order: format!("{:.1}", c.ordered as f64 / 10.0),
                size: c.size,
                count: c.count,
                created_at: c.datetime_created.clone(),
            })
            .collect();

        groups.push(ChapterGroup {
            group_name: group_info.name.clone(),
            group_path_word: group_key.clone(),
            chapters: chapters_view,
        });
    }

    let detail = ComicDetailInfo {
        comic: ComicInfo {
            uuid: comic_data.comic.uuid.clone(),
            name: comic_data.comic.name.clone(),
            path_word: comic_data.comic.path_word.clone(),
            cover: comic_data.comic.cover.clone(),
            author: comic_data
                .comic
                .author
                .iter()
                .map(|a| a.name.clone())
                .collect::<Vec<_>>()
                .join(", "),
            brief: comic_data.comic.brief.clone(),
            status: comic_data.comic.status.display.clone(),
            groups: comic_data
                .groups
                .values()
                .map(|g| GroupInfo {
                    path_word: g.path_word.clone(),
                    count: g.count,
                    name: g.name.clone(),
                })
                .collect(),
        },
        chapters: groups,
    };

    Ok(Html(render_template("comic", &serde_json::to_value(detail).unwrap())))
}

/// 提交下载任务
async fn start_download(
    State(state): State<Arc<Mutex<AppState>>>,
    Json(req): Json<DownloadReq>,
) -> Result<Json<serde_json::Value>, AppError> {
    let state = state.lock().await;

    let mut task_ids = Vec::new();
    let comic_name = req.comic_name.clone();
    let comic_path_word = req.comic_path_word.clone();
    let group_path_word = req.group_path_word.clone();

    for chapter_uuid in &req.chapter_uuids {
        // 获取章节图片 URL
        let images = state
            .client
            .get_chapter_images(&req.comic_path_word, chapter_uuid)
            .await?;

        if images.is_empty() {
            continue;
        }

        // 计算下载路径
        let comic_dir = state.manager.download_dir.join(&req.comic_name);
        let chapter_dir = comic_dir.join(&req.group_path_word).join(chapter_uuid);

        let id = state
            .manager
            .create_task(
                &req.comic_name,
                &req.comic_path_word,
                chapter_uuid,
                chapter_uuid,
                images.clone(),
            )
            .await;

        // 后台执行下载
        let manager = state.manager.clone();
        let id_clone = id.clone();
        let cu = chapter_uuid.clone();
        let cpw = req.comic_path_word.clone();
        tokio::spawn(async move {
            manager
                .download_chapter(
                    id_clone,
                    cpw,
                    cu,
                    images,
                    chapter_dir,
                )
                .await;
        });

        task_ids.push(id);
    }

    Ok(Json(serde_json::json!({
        "success": true,
        "task_ids": task_ids,
        "message": format!("已创建 {} 个下载任务", task_ids.len())
    })))
}

/// 获取下载任务列表
async fn get_tasks(
    State(state): State<Arc<Mutex<AppState>>>,
) -> Json<Vec<DownloadTask>> {
    let state = state.lock().await;
    let tasks = state.manager.get_tasks().await;
    Json(tasks)
}

/// 已下载列表
async fn downloaded_list(
    State(state): State<Arc<Mutex<AppState>>>,
) -> Html<String> {
    let state = state.lock().await;
    let download_dir = &state.manager.download_dir;

    let mut comics = Vec::new();
    if let Ok(entries) = std::fs::read_dir(download_dir) {
        for entry in entries.flatten() {
            if entry.file_type().map(|t| t.is_dir()).unwrap_or(false) {
                let comic_dir = entry.path();
                let chapter_count = count_subdirs(&comic_dir);
                let total_pages = count_files(&comic_dir);
                comics.push(DownloadedComic {
                    name: entry.file_name().to_string_lossy().to_string(),
                    path_word: String::new(),
                    chapter_count,
                    total_pages,
                });
            }
        }
    }

    Html(render_template("downloaded", &serde_json::json!({
        "comics": comics
    })))
}

fn count_subdirs(dir: &std::path::Path) -> usize {
    std::fs::read_dir(dir)
        .map(|e| e.filter(|e| e.as_ref().ok().map(|e| e.file_type().ok().map(|t| t.is_dir()).unwrap_or(false)).unwrap_or(false)).count())
        .unwrap_or(0)
}

fn count_files(dir: &std::path::Path) -> u64 {
    walkdir::WalkDir::new(dir)
        .into_iter()
        .filter_map(|e| e.ok())
        .filter(|e| e.file_type().is_file())
        .count() as u64
}

/// 渲染 MiniJinja 模板
fn render_template(name: &str, ctx: &serde_json::Value) -> String {
    let template_str = match name {
        "index" => TEMPLATE_INDEX,
        "search_results" => TEMPLATE_SEARCH,
        "comic" => TEMPLATE_COMIC,
        "downloaded" => TEMPLATE_DOWNLOADED,
        _ => TEMPLATE_INDEX,
    };

    let mut env = minijinja::Environment::new();
    env.add_template(name, template_str).unwrap();
    let tmpl = env.get_template(name).unwrap();
    tmpl.render(ctx).unwrap()
}

// ============ 全局错误处理 ============

struct AppError(eyre::Report);
impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("错误: {}", self.0),
        )
            .into_response()
    }
}
impl<E: Into<eyre::Report>> From<E> for AppError {
    fn from(e: E) -> Self {
        Self(e.into())
    }
}

// ============ 主函数 ============

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let download_dir = std::env::var("DOWNLOAD_DIR")
        .map(Into::into)
        .unwrap_or_else(|_| std::path::PathBuf::from("/downloads"));

    let token = std::env::var("TOKEN").unwrap_or_default();
    let port = std::env::var("PORT").unwrap_or_else(|_| "3000".to_string());

    let client = CopyMangaClient::new();
    if !token.is_empty() {
        // 可以通过 env TOKEN 直接设置
    }

    let manager = Arc::new(DownloadManager::new(
        Arc::new(client),
        download_dir,
    ));

    let state = Arc::new(Mutex::new(AppState {
        manager,
        client: CopyMangaClient::new(),
    }));

    let app = Router::new()
        .route("/", get(index))
        .route("/search", get(search))
        .route("/comic", get(comic_detail))
        .route("/api/download", post(start_download))
        .route("/api/tasks", get(get_tasks))
        .route("/downloaded", get(downloaded_list))
        .layer(CorsLayer::permissive())
        .with_state(state);

    let addr = format!("0.0.0.0:{port}");
    tracing::info!("Web 服务启动: http://{}", addr);
    tracing::info!("下载目录: {:?}", &*std::env::var("DOWNLOAD_DIR").unwrap_or_else(|_| "/downloads".to_string()));

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

// ============ 内联 HTML 模板 ============

const TEMPLATE_INDEX: &str = r#"<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{ title }}</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #333; max-width: 960px; margin: 0 auto; padding: 20px; }
.header { text-align: center; padding: 30px 0; }
.header h1 { font-size: 24px; margin-bottom: 10px; color: #e74c3c; }
.search-box { display: flex; gap: 10px; max-width: 500px; margin: 0 auto; }
.search-box input { flex: 1; padding: 12px 16px; border: 2px solid #ddd; border-radius: 8px; font-size: 16px; }
.search-box button { padding: 12px 24px; background: #e74c3c; color: white; border: none; border-radius: 8px; font-size: 16px; cursor: pointer; }
.search-box button:hover { background: #c0392b; }
.nav { display: flex; justify-content: center; gap: 20px; margin: 20px 0; }
.nav a { color: #e74c3c; text-decoration: none; font-size: 14px; }
.nav a:hover { text-decoration: underline; }
.footer { text-align: center; color: #999; font-size: 12px; padding: 30px 0; }
</style>
</head>
<body>
<div class="header">
<h1>📚 拷贝漫画下载器</h1>
<p>搜索漫画并下载到本地</p>
</div>
<div class="nav">
<a href="/">🔍 搜索</a>
<a href="/downloaded">📂 已下载</a>
<span id="taskLink"><a href="javascript:void(0)" onclick="toggleTasks()">⏳ 下载中</a></span>
</div>
<form class="search-box" action="/search" method="GET">
<input type="text" name="q" placeholder="输入漫画名称搜索..." required autofocus>
<button type="submit">搜索</button>
</form>
<div id="taskModal" style="display:none;position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.5);z-index:1000">
<div style="background:#fff;margin:60px auto;max-width:600px;padding:20px;border-radius:12px;max-height:70vh;overflow-y:auto">
<h2 style="margin-bottom:15px">下载任务</h2>
<div id="taskList"></div>
<button onclick="toggleTasks()" style="margin-top:15px;padding:8px 16px;background:#999;color:white;border:none;border-radius:6px;cursor:pointer">关闭</button>
</div></div>
<script>
let taskInterval = null;
async function showTasks() {
    const r = await fetch('/api/tasks');
    const tasks = await r.json();
    const list = document.getElementById('taskList');
    if (tasks.length === 0) { list.innerHTML = '<p style="color:#999">暂无下载任务</p>'; return; }
    list.innerHTML = tasks.map(t => {
        const statusEmoji = {Pending:'⏳',Downloading:'⬇️',Completed:'✅',Failed:'❌',Paused:'⏸️'}[t.status] || '⏳';
        return `<div style="padding:8px;border-bottom:1px solid #eee">
            <div>${statusEmoji} <b>${t.comic_name}</b> - ${t.chapter_title}</div>
            <div style="font-size:13px;color:#666;margin-top:4px">
                进度: ${t.progress} | 状态: ${t.status}
            </div>
        </div>`;
    }).join('');
}
function toggleTasks() {
    const modal = document.getElementById('taskModal');
    if (modal.style.display === 'block') {
        modal.style.display = 'none';
        if (taskInterval) { clearInterval(taskInterval); taskInterval = null; }
    } else {
        modal.style.display = 'block';
        showTasks();
        taskInterval = setInterval(showTasks, 3000);
    }
}
</script>
<div class="footer">
<p>copymanga-web v0.1.0</p>
</div>
</body>
</html>"#;

const TEMPLATE_SEARCH: &str = r#"<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>搜索结果 - {{ keyword }}</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #333; max-width: 960px; margin: 0 auto; padding: 20px; }
.header { text-align: center; padding: 20px 0; }
.header h1 { font-size: 20px; }
.search-box { display: flex; gap: 10px; max-width: 500px; margin: 20px auto; }
.search-box input { flex: 1; padding: 10px 14px; border: 2px solid #ddd; border-radius: 8px; font-size: 15px; }
.search-box button { padding: 10px 20px; background: #e74c3c; color: white; border: none; border-radius: 8px; cursor: pointer; }
a { text-decoration: none; color: inherit; }
.comic-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 15px; }
.comic-card { background: #fff; border-radius: 12px; padding: 15px; display: flex; gap: 15px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); transition: transform .15s; }
.comic-card:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,0.12); }
.comic-card img { width: 80px; height: 110px; object-fit: cover; border-radius: 6px; background: #eee; }
.comic-card .info { flex: 1; }
.comic-card .info .name { font-size: 16px; font-weight: 600; margin-bottom: 5px; color: #333; }
.comic-card .info .author { font-size: 13px; color: #888; margin-bottom: 3px; }
.comic-card .info .popular { font-size: 12px; color: #e74c3c; }
.back { display: inline-block; margin: 15px 0; color: #e74c3c; text-decoration: none; }
.back:hover { text-decoration: underline; }
</style>
</head>
<body>
<a class="back" href="/">← 返回搜索</a>
<div class="search-box">
<form action="/search" method="GET" style="display:flex;gap:10px;width:100%">
<input type="text" name="q" value="{{ keyword }}" placeholder="输入漫画名称搜索...">
<button type="submit">搜索</button>
</form>
</div>
<h2 style="margin:15px 0">搜索结果: {{ keyword }}</h2>
{% if comics|length == 0 %}
<p style="color:#999;text-align:center;padding:40px">没有找到相关漫画</p>
{% else %}
<div class="comic-list">
{% for comic in comics %}
<a href="/comic?path={{ comic.path_word }}" class="comic-card">
<img src="{{ comic.cover }}" alt="" loading="lazy" onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2280%22 height=%22110%22><rect fill=%22%23eee%22 width=%2280%22 height=%22110%22/></svg>'">
<div class="info">
<div class="name">{{ comic.name }}</div>
<div class="author">{{ comic.author }}</div>
<div class="popular">🔥 {{ comic.popular }}</div>
</div>
</a>
{% endfor %}
</div>
{% endif %}
</body>
</html>"#;

const TEMPLATE_COMIC: &str = r#"<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{ comic.name }} - 拷贝漫画下载</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #333; max-width: 960px; margin: 0 auto; padding: 20px; }
.back { display: inline-block; margin-bottom: 15px; color: #e74c3c; text-decoration: none; }
.back:hover { text-decoration: underline; }
.comic-header { display: flex; gap: 20px; background: #fff; border-radius: 12px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
.comic-header img { width: 160px; height: 220px; object-fit: cover; border-radius: 8px; background: #eee; }
.comic-header .info { flex: 1; }
.comic-header .info h1 { font-size: 22px; margin-bottom: 8px; }
.comic-header .info .meta { color: #888; font-size: 14px; margin-bottom: 5px; }
.comic-header .info .brief { color: #555; font-size: 14px; line-height: 1.6; margin-top: 10px; }
.group-section { background: #fff; border-radius: 12px; padding: 20px; margin-bottom: 15px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
.group-section h3 { font-size: 16px; color: #e74c3c; margin-bottom: 10px; padding-bottom: 8px; border-bottom: 2px solid #fde8e8; }
.toolbar { margin-bottom: 12px; display: flex; gap: 10px; align-items: center; }
.toolbar label { font-size: 14px; cursor: pointer; }
.toolbar button { padding: 8px 20px; background: #e74c3c; color: white; border: none; border-radius: 6px; cursor: pointer; font-size: 14px; }
.toolbar button:hover { background: #c0392b; }
.toolbar button:disabled { background: #ccc; cursor: not-allowed; }
.chapter-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 8px; }
.chapter-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: #fafafa; border-radius: 6px; border: 1px solid #eee; font-size: 14px; }
.chapter-item:hover { background: #fef0ef; }
.chapter-item input[type=checkbox] { cursor: pointer; }
.chapter-item label { cursor: pointer; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.chapter-item .size { font-size: 12px; color: #999; }
.toast { position: fixed; top: 20px; right: 20px; background: #27ae60; color: white; padding: 12px 24px; border-radius: 8px; display: none; z-index: 999; }
</style>
</head>
<body>
<div id="toast" class="toast"></div>
<a class="back" href="javascript:history.back()">← 返回搜索结果</a>
<div class="comic-header">
<img src="{{ comic.cover }}" alt="" onerror="this.src='data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%22160%22 height=%22220%22><rect fill=%22%23eee%22 width=%22160%22 height=%22220%22/></svg>'">
<div class="info">
<h1>{{ comic.name }}</h1>
<div class="meta">✍️ {{ comic.author }}</div>
<div class="meta">📌 {{ comic.status }}</div>
<div class="brief">{{ comic.brief }}</div>
</div>
</div>
{% for group in chapters %}
<div class="group-section">
<h3>📖 {{ group.group_name }}</h3>
<div class="toolbar">
<label><input type="checkbox" onchange="toggleGroup(this, '{{ group.group_path_word }}')"> 全选本组</label>
<button onclick="downloadSelected('{{ comic.name }}', '{{ comic.path_word }}', '{{ group.group_path_word }}')" id="dlBtn{{ loop.index }}">下载选中章节</button>
</div>
<div class="chapter-list" id="group-{{ group.group_path_word }}">
{% for chapter in group.chapters %}
<div class="chapter-item">
<input type="checkbox" class="chk-{{ group.group_path_word }}" value="{{ chapter.uuid }}" data-title="{{ chapter.title }}" data-idx="{{ loop.index }}">
<label onclick="this.previousElementSibling.click()">
第{{ chapter.order }}话 {{ chapter.title }}
</label>
<span class="size">{{ chapter.count }}页</span>
</div>
{% endfor %}
</div>
</div>
{% endfor %}
<script>
function toggleGroup(el, gid) {
    document.querySelectorAll('.chk-' + gid).forEach(c => c.checked = el.checked);
}
function toast(msg) {
    const t = document.getElementById('toast');
    t.textContent = msg; t.style.display = 'block';
    setTimeout(() => t.style.display = 'none', 3000);
}
async function downloadSelected(comicName, pathWord, groupPw) {
    const checks = document.querySelectorAll('.chk-' + groupPw + ':checked');
    if (checks.length === 0) { toast('请先选择要下载的章节'); return; }
    const uuids = Array.from(checks).map(c => c.value);
    try {
        const r = await fetch('/api/download', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                comic_name: comicName,
                comic_path_word: pathWord,
                group_path_word: groupPw,
                chapter_uuids: uuids
            })
        });
        const data = await r.json();
        toast(data.message || '下载任务已创建');
    } catch(e) { toast('创建下载任务失败: ' + e); }
}
</script>
</body>
</html>"#;

const TEMPLATE_DOWNLOADED: &str = r#"<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>已下载 - 拷贝漫画下载器</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #333; max-width: 960px; margin: 0 auto; padding: 20px; }
.header { padding: 20px 0; }
.header h1 { font-size: 22px; }
.header a { color: #e74c3c; text-decoration: none; font-size: 14px; }
.comic-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 15px; }
.comic-card { background: #fff; border-radius: 12px; padding: 20px; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
.comic-card .name { font-size: 16px; font-weight: 600; margin-bottom: 8px; }
.comic-card .stat { font-size: 13px; color: #888; }
.empty { text-align: center; padding: 60px; color: #999; }
</style>
</head>
<body>
<div class="header">
<a href="/">← 返回搜索</a>
<h1>📂 已下载的漫画</h1>
</div>
{% if comics|length == 0 %}
<div class="empty">还没有下载任何漫画</div>
{% else %}
<div class="comic-list">
{% for comic in comics %}
<div class="comic-card">
<div class="name">{{ comic.name }}</div>
<div class="stat">📖 {{ comic.chapter_count }} 个章节 | 🖼️ {{ comic.total_pages }} 张图片</div>
</div>
{% endfor %}
</div>
{% endif %}
</body>
</html>"#;

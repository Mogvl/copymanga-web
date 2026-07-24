use std::{
    collections::HashMap,
    path::{Path, PathBuf},
    sync::Arc,
};

use chrono::Utc;
use tokio::sync::{Mutex, Semaphore};
use tracing::{error, info};
use uuid::Uuid;

use crate::{
    client::CopyMangaClient,
    types::*,
};

/// 下载管理器
pub struct DownloadManager {
    pub client: Arc<CopyMangaClient>,
    pub download_dir: PathBuf,
    pub tasks: Arc<Mutex<HashMap<String, DownloadTask>>>,
    pub chapter_sem: Arc<Semaphore>,
    pub img_sem: Arc<Semaphore>,
}

impl DownloadManager {
    pub fn new(client: Arc<CopyMangaClient>, download_dir: PathBuf) -> Self {
        std::fs::create_dir_all(&download_dir).ok();
        Self {
            client,
            download_dir,
            tasks: Arc::new(Mutex::new(HashMap::new())),
            chapter_sem: Arc::new(Semaphore::new(3)),
            img_sem: Arc::new(Semaphore::new(30)),
        }
    }

    /// 创建下载任务
    pub async fn create_task(
        &self,
        comic_name: &str,
        comic_path_word: &str,
        chapter_uuid: &str,
        chapter_title: &str,
        image_urls: Vec<(String, i64)>,
    ) -> String {
        let id = Uuid::new_v4().to_string();
        let task = DownloadTask {
            id: id.clone(),
            comic_name: comic_name.to_string(),
            chapter_title: chapter_title.to_string(),
            chapter_uuid: chapter_uuid.to_string(),
            comic_path_word: comic_path_word.to_string(),
            status: TaskStatus::Pending,
            progress: format!("0/{}", image_urls.len()),
            total_pages: image_urls.len() as u32,
            downloaded_pages: 0,
            created_at: Utc::now().format("%Y-%m-%d %H:%M:%S").to_string(),
        };
        self.tasks.lock().await.insert(id.clone(), task);
        id
    }

    /// 更新任务状态
    pub async fn update_task(
        &self,
        id: &str,
        downloaded: u32,
        total: u32,
        status: TaskStatus,
    ) {
        let mut tasks = self.tasks.lock().await;
        if let Some(task) = tasks.get_mut(id) {
            task.downloaded_pages = downloaded;
            task.progress = format!("{downloaded}/{total}");
            task.status = status;
        }
    }

    /// 获取所有任务
    pub async fn get_tasks(&self) -> Vec<DownloadTask> {
        let tasks = self.tasks.lock().await;
        let mut list: Vec<DownloadTask> = tasks.values().cloned().collect();
        list.sort_by(|a, b| b.created_at.cmp(&a.created_at));
        list
    }

    /// 执行下载：下载整个章节的所有图片
    pub async fn download_chapter(
        self: Arc<Self>,
        id: String,
        _comic_path_word: String,
        _chapter_uuid: String,
        image_urls: Vec<(String, i64)>,
        chapter_dir: PathBuf,
    ) {
        // 获取信号量，控制并发章节数
        let _permit = self.chapter_sem.acquire().await.unwrap();

        self.update_task(&id, 0, image_urls.len() as u32, TaskStatus::Downloading)
            .await;

        // 创建临时下载目录
        let temp_dir = chapter_dir.with_file_name(format!(".downloading-{}", chapter_dir.file_name().unwrap().to_str().unwrap_or("tmp")));
        std::fs::create_dir_all(&temp_dir).ok();

        let total = image_urls.len() as u32;
        let mut success = 0u32;

        // 并发下载图片
        let mut handles = vec![];
        for (idx, (url, _)) in image_urls.iter().enumerate() {
            let url = url.clone();
            let client = self.client.clone();
            let temp_dir = temp_dir.clone();
            let sem = self.img_sem.clone();
            let self_arc = self.clone();
            let task_id = id.clone();

            let handle = tokio::spawn(async move {
                let _permit = sem.acquire().await.unwrap();
                match client.download_image(&url).await {
                    Ok((data, content_type)) => {
                        let ext = if content_type.contains("webp") {
                            "webp"
                        } else {
                            "jpg"
                        };
                        let file_name = format!("{:04}.{ext}", idx + 1);
                        let file_path = temp_dir.join(&file_name);
                        if let Err(e) = std::fs::write(&file_path, &data) {
                            error!("保存图片失败 {}: {}", file_path.display(), e);
                        } else {
                            self_arc
                                .update_task(&task_id, idx as u32 + 1, total, TaskStatus::Downloading)
                                .await;
                        }
                    }
                    Err(e) => {
                        error!("下载图片失败 [{}] {}: {}", idx + 1, url, e);
                    }
                }
            });
            handles.push(handle);
        }

        // 等待所有图片下载完成
        for h in handles {
            h.await.ok();
            success += 1;
        }

        if success == 0 {
            error!("章节下载失败: 没有成功下载任何图片");
            self.update_task(&id, 0, total, TaskStatus::Failed).await;
            return;
        }

        // 重命名临时目录为正式目录
        if chapter_dir.exists() {
            std::fs::remove_dir_all(&chapter_dir).ok();
        }
        if let Err(e) = std::fs::rename(&temp_dir, &chapter_dir) {
            error!("重命名目录失败: {}", e);
            self.update_task(&id, success, total, TaskStatus::Failed).await;
            return;
        }

        self.update_task(&id, success, total, TaskStatus::Completed).await;
        info!("章节下载完成: {}", chapter_title_from_path(&chapter_dir));
    }
}

fn chapter_title_from_path(path: &Path) -> String {
    path.file_name()
        .and_then(|n| n.to_str())
        .unwrap_or("unknown")
        .to_string()
}

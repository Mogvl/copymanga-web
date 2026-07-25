use std::time::Duration;

use base64::{engine::general_purpose, Engine as _};
use eyre::{eyre, Result};
use reqwest::Client;
use reqwest_middleware::ClientWithMiddleware;
use reqwest_retry::{policies::ExponentialBackoff, Jitter, RetryTransientMiddleware};
use serde_json::json;

use crate::types::*;

/// 拷贝漫画 API 客户端
pub struct CopyMangaClient {
    api_client: ClientWithMiddleware,
    img_client: ClientWithMiddleware,
    api_domain: String,
    token: String,
}

impl CopyMangaClient {
    pub fn new() -> Self {
        let mut headers = reqwest::header::HeaderMap::new();
        headers.insert("User-Agent", "COPY/3.0.0".parse().unwrap());
        headers.insert("Accept", "application/json".parse().unwrap());
        headers.insert("version", "2025.08.15".parse().unwrap());
        headers.insert("platform", "1".parse().unwrap());
        headers.insert("webp", "1".parse().unwrap());
        headers.insert("region", "1".parse().unwrap());

        let retry_policy = ExponentialBackoff::builder()
            .base(1)
            .jitter(Jitter::Bounded)
            .build_with_total_retry_duration(Duration::from_secs(5));

        let api_client = reqwest::Client::builder()
            .default_headers(headers)
            .timeout(Duration::from_secs(10))
            .build()
            .expect("创建 HTTP 客户端失败");

        let api_client = reqwest_middleware::ClientBuilder::new(api_client)
            .with(RetryTransientMiddleware::new_with_policy(retry_policy))
            .build();

        let img_client = reqwest::Client::builder()
            .timeout(Duration::from_secs(30))
            .build()
            .expect("创建图片下载客户端失败");

        let img_client = reqwest_middleware::ClientBuilder::new(img_client)
            .with(RetryTransientMiddleware::new_with_policy(
                ExponentialBackoff::builder().build_with_max_retries(3),
            ))
            .build();

        Self {
            api_client,
            img_client,
            api_domain: "api.copy202601.com".to_string(),
            token: String::new(),
        }
    }

    /// 设置 API 域名（用于自定义域名）
    pub fn set_api_domain(&mut self, domain: &str) {
        self.api_domain = domain.to_string();
    }

    /// 设置登录 token（用于风控）
    pub fn set_token(&mut self, token: &str) {
        self.token = token.to_string();
    }

    /// 登录获取 token
    pub async fn login(&self, username: &str, password: &str) -> Result<String> {
        const SALT: i32 = 1729;
        let password = general_purpose::STANDARD.encode(format!("{password}-{SALT}"));

        let resp = self
            .api_client
            .post(format!("https://{}/api/v3/login", self.api_domain))
            .form(&json!({"username": username, "password": password, "salt": SALT}))
            .send()
            .await?
            .json::<CopyResp>()
            .await?;

        if resp.code != 200 {
            return Err(eyre!("登录失败: {}", resp.message));
        }

        let data: serde_json::Value = resp.results;
        let token = data["token"]
            .as_str()
            .ok_or_else(|| eyre!("登录返回没有 token"))?
            .to_string();
        Ok(token)
    }

    /// 搜索漫画
    pub async fn search(&self, keyword: &str, page: i64) -> Result<(Vec<ComicInSearch>, i64)> {
        let offset = (page - 1) * 20;
        let resp = self
            .api_client
            .get(format!("https://{}/api/v3/search/comic", self.api_domain))
            .query(&json!({"limit": 20, "offset": offset, "q": keyword, "q_type": "", "platform": 1}))
            .send()
            .await?
            .json::<CopyResp>()
            .await?;

        if resp.code != 200 {
            return Err(eyre!("搜索失败: {}", resp.message));
        }

        let list: SearchRespData = serde_json::from_value(resp.results)?;
        let total = list.0.total;
        Ok((list.0.list, total))
    }

    /// 获取漫画详情
    pub async fn get_comic(&self, path_word: &str) -> Result<GetComicRespData> {
        let resp = self
            .api_client
            .get(format!(
                "https://{}/api/v3/comic2/{path_word}",
                self.api_domain
            ))
            .query(&json!({"platform": 1}))
            .send()
            .await?
            .json::<CopyResp>()
            .await?;

        if resp.code != 200 {
            return Err(eyre!("获取漫画失败: {}", resp.message));
        }

        Ok(serde_json::from_value(resp.results)?)
    }

    /// 获取某个分组的全部章节
    pub async fn get_group_chapters(
        &self,
        comic_path_word: &str,
        group_path_word: &str,
    ) -> Result<Vec<ChapterItem>> {
        let mut all = vec![];
        let limit = 100;
        let first = self
            .get_chapters(comic_path_word, group_path_word, limit, 0)
            .await?;
        all.extend(first.0.list);

        let total_pages = first.0.total / limit + 1;
        if total_pages <= 1 {
            return Ok(all);
        }

        for page in 2..=total_pages {
            let offset = (page - 1) * limit;
            let data = self
                .get_chapters(comic_path_word, group_path_word, limit, offset)
                .await?;
            all.extend(data.0.list);
        }

        all.sort_by(|a, b| a.ordered.cmp(&b.ordered));
        Ok(all)
    }

    async fn get_chapters(
        &self,
        comic_path_word: &str,
        group_path_word: &str,
        limit: i64,
        offset: i64,
    ) -> Result<GetChaptersRespData> {
        let resp = self
            .api_client
            .get(format!(
                "https://{}/api/v3/comic/{comic_path_word}/group/{group_path_word}/chapters",
                self.api_domain
            ))
            .query(&json!({"limit": limit, "offset": offset}))
            .send()
            .await?
            .json::<CopyResp>()
            .await?;

        if resp.code != 200 {
            return Err(eyre!("获取章节失败: {}", resp.message));
        }

        Ok(serde_json::from_value(resp.results)?)
    }

    /// 获取章节图片 URL 列表
    pub async fn get_chapter_images(&self, comic_path_word: &str, chapter_uuid: &str) -> Result<Vec<(String, i64)>> {
        let authorization = if self.token.is_empty() {
            String::new()
        } else {
            format!("Token {}", self.token)
        };

        let mut req = self
            .api_client
            .get(format!(
                "https://{}/api/v3/comic/{comic_path_word}/chapter2/{chapter_uuid}",
                self.api_domain
            ))
            .query(&json!({"platform": 1}));

        if !authorization.is_empty() {
            req = req.header("authorization", &authorization);
        }

        let resp = req.send().await?.json::<CopyResp>().await?;

        if resp.code != 200 {
            return Err(eyre!("获取章节图片失败: {}", resp.message));
        }

        let data: GetChapterRespData = serde_json::from_value(resp.results)?;

        let urls: Vec<(String, i64)> = data
            .chapter
            .contents
            .into_iter()
            .enumerate()
            .map(|(i, content)| {
                let url = content.url.replace(".c800x.", ".c1500x.");
                let idx = data.chapter.words.get(i).copied().unwrap_or(i as i64);
                (url, idx)
            })
            .collect();

        Ok(urls)
    }

    /// 下载单张图片
    pub async fn download_image(&self, url: &str) -> Result<(bytes::Bytes, String)> {
        let resp = self.img_client.get(url).send().await?;

        if !resp.status().is_success() {
            return Err(eyre!("下载图片失败: HTTP {}", resp.status()));
        }

        let content_type = resp
            .headers()
            .get("content-type")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("image/jpeg")
            .to_string();

        let data = resp.bytes().await?;
        Ok((data, content_type))
    }
}

use serde::{Deserialize, Serialize};

/// 拷贝漫画 API 通用响应结构
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CopyResp {
    pub code: i64,
    pub message: String,
    pub results: serde_json::Value,
}

/// 搜索响应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchRespData(pub Pagination<ComicInSearch>);

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Pagination<T> {
    pub list: Vec<T>,
    pub total: i64,
    pub limit: i64,
    pub offset: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ComicInSearch {
    pub name: String,
    pub alias: Option<String>,
    #[serde(rename = "path_word")]
    pub path_word: String,
    pub cover: String,
    pub ban: i64,
    pub author: Vec<AuthorInfo>,
    pub popular: i64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct AuthorInfo {
    pub name: String,
    pub alias: Option<String>,
    #[serde(rename = "path_word")]
    pub path_word: String,
}

/// 漫画详情响应
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GetComicRespData {
    pub is_banned: bool,
    pub comic: ComicDetail,
    pub popular: i64,
    pub groups: std::collections::HashMap<String, GroupInfo>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ComicDetail {
    pub uuid: String,
    pub name: String,
    #[serde(rename = "path_word")]
    pub path_word: String,
    pub author: Vec<AuthorInfo>,
    pub cover: String,
    pub brief: String,
    pub status: LabeledValue,
    pub theme: Vec<ThemeInfo>,
    pub datetime_updated: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct LabeledValue {
    pub value: i64,
    pub display: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ThemeInfo {
    pub name: String,
    #[serde(rename = "path_word")]
    pub path_word: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GroupInfo {
    #[serde(rename = "path_word")]
    pub path_word: String,
    pub count: u32,
    pub name: String,
}

/// 章节列表响应
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GetChaptersRespData(pub Pagination<ChapterItem>);

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChapterItem {
    pub index: i64,
    pub uuid: String,
    pub count: i64,
    pub ordered: i64,
    pub size: i64,
    pub name: String,
    #[serde(rename = "comic_path_word")]
    pub comic_path_word: String,
    #[serde(rename = "group_path_word")]
    pub group_path_word: String,
    pub datetime_created: String,
}

/// 单个章节详情（含图片 URL）
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct GetChapterRespData {
    pub chapter: ChapterContent,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ChapterContent {
    pub uuid: String,
    pub name: String,
    #[serde(rename = "comic_path_word")]
    pub comic_path_word: String,
    #[serde(rename = "group_path_word")]
    pub group_path_word: String,
    pub contents: Vec<ContentUrl>,
    pub words: Vec<i64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ContentUrl {
    pub url: String,
}

// ============ 前端展示用类型 ============

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComicInfo {
    pub uuid: String,
    pub name: String,
    pub path_word: String,
    pub cover: String,
    pub author: String,
    pub brief: String,
    pub status: String,
    pub groups: Vec<GroupInfo>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ComicDetailInfo {
    pub comic: ComicInfo,
    pub chapters: Vec<ChapterGroup>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChapterGroup {
    pub group_name: String,
    pub group_path_word: String,
    pub chapters: Vec<ChapterView>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChapterView {
    pub uuid: String,
    pub title: String,
    pub order: String,
    pub size: i64,
    pub count: i64,
    pub created_at: String,
}

/// 下载任务状态
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum TaskStatus {
    Pending,
    Downloading,
    Completed,
    Failed,
    Paused,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DownloadTask {
    pub id: String,
    pub comic_name: String,
    pub chapter_title: String,
    pub chapter_uuid: String,
    pub comic_path_word: String,
    pub status: TaskStatus,
    pub progress: String,  // "3/20"
    pub total_pages: u32,
    pub downloaded_pages: u32,
    pub created_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DownloadedComic {
    pub name: String,
    pub path_word: String,
    pub chapter_count: usize,
    pub total_pages: u64,
}

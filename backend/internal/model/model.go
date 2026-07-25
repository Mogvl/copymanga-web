package model

import "encoding/json"

// CopyResp 拷贝漫画 API 通用响应
type CopyResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Results json.RawMessage `json:"results"`
}

// ComicInSearch 搜索结果中的漫画
type ComicInSearch struct {
	Name     string       `json:"name"`
	Alias    string       `json:"alias,omitempty"`
	PathWord string       `json:"path_word"`
	Cover    string       `json:"cover"`
	Ban      int          `json:"ban"`
	Author   []AuthorInfo `json:"author"`
	Popular  int64        `json:"popular"`
}

// AuthorInfo 作者信息
type AuthorInfo struct {
	Name     string `json:"name"`
	Alias    string `json:"alias,omitempty"`
	PathWord string `json:"path_word"`
}

// SearchRespData 搜索响应数据
type SearchRespData struct {
	List   []ComicInSearch `json:"list"`
	Total  int64           `json:"total"`
	Limit  int64           `json:"limit"`
	Offset int64           `json:"offset"`
}

// GetComicRespData 获取漫画详情响应
type GetComicRespData struct {
	IsBanned bool              `json:"is_banned"`
	Comic    ComicDetail       `json:"comic"`
	Popular  int64             `json:"popular"`
	Groups   map[string]Group  `json:"groups"`
}

// ComicDetail 漫画详情
type ComicDetail struct {
	UUID           string        `json:"uuid"`
	Name           string        `json:"name"`
	PathWord       string        `json:"path_word"`
	Author         []AuthorInfo  `json:"author"`
	Cover          string        `json:"cover"`
	Brief          string        `json:"brief"`
	Status         LabeledValue  `json:"status"`
	Theme          []ThemeInfo   `json:"theme"`
	DateTimeUpdated string       `json:"datetime_updated"`
}

// LabeledValue 标签值
type LabeledValue struct {
	Value   int    `json:"value"`
	Display string `json:"display"`
}

// ThemeInfo 主题信息
type ThemeInfo struct {
	Name     string `json:"name"`
	PathWord string `json:"path_word"`
}

// Group 分组信息
type Group struct {
	PathWord string `json:"path_word"`
	Count    int    `json:"count"`
	Name     string `json:"name"`
}

// ChapterItem 章节项
type ChapterItem struct {
	Index           int    `json:"index"`
	UUID            string `json:"uuid"`
	Count           int    `json:"count"`
	Ordered         int    `json:"ordered"`
	Size            int    `json:"size"`
	Name            string `json:"name"`
	ComicPathWord   string `json:"comic_path_word"`
	GroupPathWord   string `json:"group_path_word"`
	DateTimeCreated string `json:"datetime_created"`
}

// GetChaptersRespData 获取章节列表响应
type GetChaptersRespData struct {
	List   []ChapterItem `json:"list"`
	Total  int64         `json:"total"`
	Limit  int64         `json:"limit"`
	Offset int64         `json:"offset"`
}

// GetChapterRespData 获取章节详情响应
type GetChapterRespData struct {
	Chapter ChapterContent `json:"chapter"`
}

// ChapterContent 章节内容
type ChapterContent struct {
	UUID          string        `json:"uuid"`
	Name          string        `json:"name"`
	ComicPathWord string        `json:"comic_path_word"`
	GroupPathWord string        `json:"group_path_word"`
	Contents      []ContentURL  `json:"contents"`
	Words         []int64       `json:"words"`
}

// ContentURL 内容URL
type ContentURL struct {
	URL string `json:"url"`
}

// LoginRespData 登录响应
type LoginRespData struct {
	Token string `json:"token"`
}

// UserProfile 用户信息
type UserProfile struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID              string `json:"id"`
	ComicName       string `json:"comic_name"`
	ComicPathWord   string `json:"comic_path_word"`
	ChapterUUID     string `json:"chapter_uuid"`
	ChapterTitle    string `json:"chapter_title"`
	Status          string `json:"status"` // pending, downloading, completed, failed
	Progress        string `json:"progress"`
	TotalPages      int    `json:"total_pages"`
	DownloadedPages int    `json:"downloaded_pages"`
	CreatedAt       string `json:"created_at"`
}

// DownloadedComic 已下载漫画
type DownloadedComic struct {
	Name         string `json:"name"`
	PathWord     string `json:"path_word"`
	ChapterCount int    `json:"chapter_count"`
	TotalPages   int64  `json:"total_pages"`
}

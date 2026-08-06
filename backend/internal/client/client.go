package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"copymanga-backend/internal/model"
	"go.uber.org/zap"
)

const (
	defaultAPIDomain = "api.copy202601.com"
	defaultUserAgent = "COPY/3.0.0"
)

// Client 拷贝漫画 API 客户端
type Client struct {
	httpClient *http.Client
	apiDomain  string
	token      string
	mu         sync.RWMutex
	logger     *zap.Logger
}

// New 创建新的客户端
func New(logger *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		apiDomain: defaultAPIDomain,
		logger:    logger,
	}
}

// SetToken 设置 token
func (c *Client) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

// GetToken 获取 token
func (c *Client) GetToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.token
}

// SetAPIDomain 设置 API 域名
func (c *Client) SetAPIDomain(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiDomain = domain
}

// GetAPIDomain 获取 API 域名
func (c *Client) GetAPIDomain() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiDomain
}

// doRequest 执行 HTTP 请求
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("version", "2025.08.15")
	req.Header.Set("platform", "1")
	req.Header.Set("webp", "1")
	req.Header.Set("region", "1")

	if token != "" {
		req.Header.Set("authorization", fmt.Sprintf("Token %s", token))
	}

	return c.httpClient.Do(req)
}

// Login 登录
func (c *Client) Login(username, password string) (*model.LoginRespData, error) {
	// 对密码进行编码
	encodedPassword := encodePassword(password)

	form := url.Values{}
	form.Set("username", username)
	form.Set("password", encodedPassword)
	form.Set("salt", "1729")

	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/login", apiDomain)

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = form.Encode()

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("登录失败，状态码(%d): %s", resp.StatusCode, string(body))
	}

	var copyResp model.CopyResp
	if err := json.Unmarshal(body, &copyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %s", string(body))
	}

	if copyResp.Code != 200 {
		return nil, fmt.Errorf("登录失败: %s", copyResp.Message)
	}

	var loginResp model.LoginRespData
	if err := json.Unmarshal(copyResp.Results, &loginResp); err != nil {
		return nil, fmt.Errorf("解析登录数据失败")
	}

	c.SetToken(loginResp.Token)
	return &loginResp, nil
}

// Register 注册
func (c *Client) Register(username, password string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("source", "freeSite")

	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/register", apiDomain)

	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = form.Encode()

	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == 210 {
		return fmt.Errorf("风控限制: %s", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("注册失败，状态码(%d): %s", resp.StatusCode, string(body))
	}

	var copyResp model.CopyResp
	if err := json.Unmarshal(body, &copyResp); err != nil {
		return fmt.Errorf("解析响应失败: %s", string(body))
	}

	if copyResp.Code != 200 {
		return fmt.Errorf("注册失败: %s", copyResp.Message)
	}

	return nil
}

// Search 搜索漫画
func (c *Client) Search(keyword string, page int64) (*model.SearchRespData, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * 20

	params := url.Values{}
	params.Set("limit", "20")
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("q", keyword)
	params.Set("q_type", "")
	params.Set("platform", "1")

	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/search/comic?%s", apiDomain, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索失败，状态码(%d): %s", resp.StatusCode, string(body))
	}

	var copyResp model.CopyResp
	if err := json.Unmarshal(body, &copyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %s", string(body))
	}

	if copyResp.Code != 200 {
		return nil, fmt.Errorf("搜索失败: %s", copyResp.Message)
	}

	var searchResp model.SearchRespData
	if err := json.Unmarshal(copyResp.Results, &searchResp); err != nil {
		return nil, fmt.Errorf("解析搜索数据失败")
	}

	return &searchResp, nil
}

// GetComic 获取漫画详情
func (c *Client) GetComic(comicPathWord string) (*model.GetComicRespData, error) {
	params := url.Values{}
	params.Set("platform", "1")

	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/comic2/%s?%s", apiDomain, comicPathWord, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取漫画失败，状态码(%d): %s", resp.StatusCode, string(body))
	}

	var copyResp model.CopyResp
	if err := json.Unmarshal(body, &copyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %s", string(body))
	}

	if copyResp.Code != 200 {
		return nil, fmt.Errorf("获取漫画失败: %s", copyResp.Message)
	}

	var comicResp model.GetComicRespData
	if err := json.Unmarshal(copyResp.Results, &comicResp); err != nil {
		return nil, fmt.Errorf("解析漫画数据失败")
	}

	return &comicResp, nil
}

// GetGroupChapters 获取分组章节列表
func (c *Client) GetGroupChapters(comicPathWord, groupPathWord string) ([]model.ChapterItem, error) {
	var allChapters []model.ChapterItem
	limit := int64(100)

	// 获取第一页
	firstResp, err := c.getChapters(comicPathWord, groupPathWord, limit, 0)
	if err != nil {
		return nil, err
	}
	allChapters = append(allChapters, firstResp.List...)

	// 计算总页数
	totalPages := firstResp.Total/limit + 1
	if totalPages <= 1 {
		return allChapters, nil
	}

	// 获取剩余页
	for page := int64(2); page <= totalPages; page++ {
		offset := (page - 1) * limit
		resp, err := c.getChapters(comicPathWord, groupPathWord, limit, offset)
		if err != nil {
			return nil, err
		}
		allChapters = append(allChapters, resp.List...)
	}

	return allChapters, nil
}

// getChapters 获取章节分页
func (c *Client) getChapters(comicPathWord, groupPathWord string, limit, offset int64) (*model.GetChaptersRespData, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/comic/%s/group/%s/chapters?%s",
		apiDomain, comicPathWord, groupPathWord, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取章节失败，状态码(%d): %s", resp.StatusCode, string(body))
	}

	var copyResp model.CopyResp
	if err := json.Unmarshal(body, &copyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %s", string(body))
	}

	if copyResp.Code != 200 {
		return nil, fmt.Errorf("获取章节失败: %s", copyResp.Message)
	}

	var chaptersResp model.GetChaptersRespData
	if err := json.Unmarshal(copyResp.Results, &chaptersResp); err != nil {
		return nil, fmt.Errorf("解析章节数据失败")
	}

	return &chaptersResp, nil
}

// GetChapterImages 获取章节图片
func (c *Client) GetChapterImages(comicPathWord, chapterUUID string) (*model.GetChapterRespData, error) {
	params := url.Values{}
	params.Set("platform", "1")

	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/comic/%s/chapter2/%s?%s",
		apiDomain, comicPathWord, chapterUUID, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 处理风控
	if resp.StatusCode == 210 {
		return nil, fmt.Errorf("账号被风控，请等待1小时后再试")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取章节图片失败，状态码(%d): %s", resp.StatusCode, string(body))
	}

	var copyResp model.CopyResp
	if err := json.Unmarshal(body, &copyResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %s", string(body))
	}

	if copyResp.Code != 200 {
		return nil, fmt.Errorf("获取章节图片失败: %s", copyResp.Message)
	}

	var chapterResp model.GetChapterRespData
	if err := json.Unmarshal(copyResp.Results, &chapterResp); err != nil {
		return nil, fmt.Errorf("解析章节图片数据失败")
	}

	// 替换为高清图片 URL（与原版一致：.c800x. -> .c1500x.）
	for i := range chapterResp.Chapter.Contents {
		chapterResp.Chapter.Contents[i].URL = strings.ReplaceAll(
			chapterResp.Chapter.Contents[i].URL, ".c800x.", ".c1500x.",
		)
	}

	return &chapterResp, nil
}

// TestAPI 测试 API 连通性
func (c *Client) TestAPI() (int, string, error) {
	apiDomain := c.GetAPIDomain()
	reqURL := fmt.Sprintf("https://%s/api/v3/search/comic?limit=1&offset=0&q=test&q_type=&platform=1", apiDomain)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return 0, "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", nil
	}

	return resp.StatusCode, string(body[:min(len(body), 200)]), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

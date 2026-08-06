package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"copymanga-backend/internal/client"
	"copymanga-backend/internal/model"
	"go.uber.org/zap"
)

// DownloadOptions 下载相关可配置参数（通过环境变量注入）
type DownloadOptions struct {
	// ImageConcurrency 单章节内图片并发数，默认 5
	ImageConcurrency int
	// ImageRetry 图片下载失败重试次数，默认 3
	ImageRetry int
	// ChapterIntervalSec 章节下载完成后休眠秒数，缓解风控，默认 0
	ChapterIntervalSec int
}

// DownloadManager 下载管理器
type DownloadManager struct {
	client      *client.Client
	downloadDir string
	options     DownloadOptions
	tasks       map[string]*model.DownloadTask
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewDownloadManager 创建下载管理器
func NewDownloadManager(c *client.Client, downloadDir string, options DownloadOptions, logger *zap.Logger) *DownloadManager {
	if options.ImageConcurrency <= 0 {
		options.ImageConcurrency = 5
	}
	if options.ImageRetry <= 0 {
		options.ImageRetry = 3
	}
	return &DownloadManager{
		client:      c,
		downloadDir: downloadDir,
		options:     options,
		tasks:       make(map[string]*model.DownloadTask),
		logger:      logger,
	}
}

// GenerateTaskID 生成下载任务 ID
func GenerateTaskID(comicPathWord, chapterUUID string) string {
	return fmt.Sprintf("%s_%s_%d", comicPathWord, chapterUUID, time.Now().UnixNano())
}

// GetDownloadDir 获取下载目录
func (dm *DownloadManager) GetDownloadDir() string {
	return dm.downloadDir
}

// CreateTask 创建下载任务
func (dm *DownloadManager) CreateTask(comicName, comicPathWord, chapterUUID, groupTitle string, order float64, chapterTitle, imageFormat string, images []model.ContentURL) string {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	taskID := GenerateTaskID(comicPathWord, chapterUUID)
	task := &model.DownloadTask{
		ID:            taskID,
		ComicName:     comicName,
		ComicPathWord: comicPathWord,
		ChapterUUID:   chapterUUID,
		GroupTitle:    groupTitle,
		Order:         order,
		ChapterTitle:  chapterTitle,
		ImageFormat:   imageFormat,
		Status:        "pending",
		Progress:      "等待中",
		TotalPages:    len(images),
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	dm.tasks[taskID] = task
	return taskID
}

// StartDownload 开始下载任务
// words 为每张图片对应的真实页码，用于生成文件名
func (dm *DownloadManager) StartDownload(taskID, comicPathWord, chapterUUID, chapterTitle, groupTitle string, order float64, imageFormat string, images []model.ContentURL, words []int64) {
	imgConcurrency := dm.options.ImageConcurrency

	dm.mu.Lock()
	task, ok := dm.tasks[taskID]
	if !ok {
		dm.mu.Unlock()
		return
	}
	task.Status = "downloading"
	task.Progress = "下载中"
	dm.mu.Unlock()

	// 创建下载目录：{下载目录}/{漫画名}/{分组名}/{order} {章节名}
	comicDir := sanitizeFilename(task.ComicName)
	groupDir := sanitizeFilename(groupTitle)
	if groupDir == "" {
		groupDir = "未分组"
	}
	chapterName := sanitizeFilename(chapterTitle)
	if chapterName == "" {
		chapterName = chapterUUID
	}
	orderLabel := formatOrder(order)
	chapterDir := filepath.Join(dm.downloadDir, comicDir, groupDir, fmt.Sprintf("%s %s", orderLabel, chapterName))
	if err := os.MkdirAll(chapterDir, 0755); err != nil {
		dm.logger.Error("创建目录失败", zap.Error(err))
		dm.updateTaskStatus(taskID, "failed", "创建目录失败", err.Error())
		return
	}

	// 并发下载图片
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, imgConcurrency)
	var mu sync.Mutex
	var failedError error

	for i, img := range images {
		wg.Add(1)
		go func(index int, imgURL string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 使用真实页码作为文件名（若缺失则退回序号）
			pageNum := index + 1
			if index < len(words) && words[index] > 0 {
				pageNum = int(words[index])
			}

			// 确定文件扩展名（gif 保留原样，否则用所选格式）
			ext := ".jpg"
			lowerURL := strings.ToLower(imgURL)
			if strings.HasSuffix(lowerURL, ".gif") {
				ext = ".gif"
			} else if imageFormat == "webp" {
				ext = ".webp"
			} else if imageFormat == "jpg" {
				ext = ".jpg"
			}

			filename := filepath.Join(chapterDir, fmt.Sprintf("%s%s", padNumber(pageNum), ext))

			if err := dm.downloadImage(imgURL, filename); err != nil {
				mu.Lock()
				if failedError == nil {
					failedError = err
				}
				mu.Unlock()
				dm.logger.Error("下载图片失败",
					zap.String("url", imgURL),
					zap.String("filename", filename),
					zap.Error(err),
				)
				return
			}

			// 更新进度
			dm.mu.Lock()
			if t, ok := dm.tasks[taskID]; ok {
				t.DownloadedPages++
				t.Progress = fmt.Sprintf("%d/%d", t.DownloadedPages, t.TotalPages)
			}
			dm.mu.Unlock()
		}(i, img.URL)
	}

	wg.Wait()

	// 校验下载张数：下载不完整则标记失败（与下载任务文件数按真实页码去重有关）
	dm.mu.Lock()
	if t, ok := dm.tasks[taskID]; ok {
		if t.DownloadedPages != t.TotalPages {
			// 可能是重复页码导致文件覆盖，此时按实际文件数判断
			actual := int(countFiles(chapterDir))
			if actual >= t.TotalPages {
				t.DownloadedPages = t.TotalPages
				t.Progress = fmt.Sprintf("%d/%d", t.TotalPages, t.TotalPages)
			} else if failedError != nil {
				t.Error = failedError.Error()
			}
		}
	}
	dm.mu.Unlock()

	dm.updateTaskStatus(taskID, "completed", "下载完成")

	// 章节间隔，缓解风控
	if dm.options.ChapterIntervalSec > 0 && len(images) > 0 {
		time.Sleep(time.Duration(dm.options.ChapterIntervalSec) * time.Second)
	}
}

// downloadImage 下载图片（带重试）
func (dm *DownloadManager) downloadImage(url, filename string) error {
	lastErr := fmt.Errorf("未执行下载")
	for attempt := 1; attempt <= dm.options.ImageRetry; attempt++ {
		err := dm.downloadImageOnce(url, filename)
		if err == nil {
			return nil
		}
		lastErr = err
		if attempt < dm.options.ImageRetry {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return fmt.Errorf("重试 %d 次后仍失败: %w", dm.options.ImageRetry, lastErr)
}

func (dm *DownloadManager) downloadImageOnce(url, filename string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("请求图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}

// updateTaskStatus 更新任务状态
func (dm *DownloadManager) updateTaskStatus(taskID, status, progress string, errMsg ...string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if task, ok := dm.tasks[taskID]; ok {
		task.Status = status
		task.Progress = progress
		if len(errMsg) > 0 && errMsg[0] != "" {
			task.Error = errMsg[0]
		}
	}
}

// GetTasks 获取所有任务（按创建时间排序）
func (dm *DownloadManager) GetTasks() []model.DownloadTask {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	tasks := make([]model.DownloadTask, 0, len(dm.tasks))
	for _, task := range dm.tasks {
		tasks = append(tasks, *task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt < tasks[j].CreatedAt
	})
	return tasks
}

// GetTask 获取单个任务
func (dm *DownloadManager) GetTask(taskID string) *model.DownloadTask {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.tasks[taskID]
}

// GetDownloadedComics 获取已下载漫画列表（按漫画归并子目录统计）
func (dm *DownloadManager) GetDownloadedComics() ([]model.DownloadedComic, error) {
	comics := make([]model.DownloadedComic, 0)

	entries, err := os.ReadDir(dm.downloadDir)
	if err != nil {
		if os.IsNotExist(err) {
			return comics, nil
		}
		return nil, fmt.Errorf("读取下载目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		comicDir := filepath.Join(dm.downloadDir, entry.Name())
		chapterCount := countSubdirLayer(comicDir)
		totalPages := countFiles(comicDir)

		comics = append(comics, model.DownloadedComic{
			Name:         entry.Name(),
			ChapterCount: chapterCount,
			TotalPages:   totalPages,
		})
	}

	return comics, nil
}

// countSubdirLayer 统计分组下所有章节目录数量（递归统计最深层的章节目录）
func countSubdirLayer(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sub := filepath.Join(dir, entry.Name())
		// 若子目录下还有子目录，则其为分组层，继续递归
		if hasSubdir(sub) {
			count += countSubdirLayer(sub)
		} else {
			count++
		}
	}
	return count
}

func hasSubdir(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return true
		}
	}
	return false
}

// countFiles 计算文件数量（递归）
func countFiles(dir string) int64 {
	var count int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

// sanitizeFilename 清理文件名中的非法字符
func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}

// padNumber 将页码补零为 5 位
func padNumber(n int) string {
	return fmt.Sprintf("%05d", n)
}

// formatOrder 格式化章节排序（order 已归一化为「第N话」的 N）
func formatOrder(order float64) string {
	if order == 0 {
		return "000"
	}
	if order == float64(int64(order)) {
		return fmt.Sprintf("%03d", int64(order))
	}
	return fmt.Sprintf("%.1f", order)
}

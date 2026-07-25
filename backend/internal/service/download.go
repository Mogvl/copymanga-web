package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"copymanga-backend/internal/client"
	"copymanga-backend/internal/model"
	"go.uber.org/zap"
)

// DownloadManager 下载管理器
type DownloadManager struct {
	client      *client.Client
	downloadDir string
	tasks       map[string]*model.DownloadTask
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewDownloadManager 创建下载管理器
func NewDownloadManager(c *client.Client, downloadDir string, logger *zap.Logger) *DownloadManager {
	return &DownloadManager{
		client:      c,
		downloadDir: downloadDir,
		tasks:       make(map[string]*model.DownloadTask),
		logger:      logger,
	}
}

// GetDownloadDir 获取下载目录
func (dm *DownloadManager) GetDownloadDir() string {
	return dm.downloadDir
}

// CreateTask 创建下载任务
func (dm *DownloadManager) CreateTask(comicName, comicPathWord, chapterUUID, chapterTitle string, images []model.ContentURL) string {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	taskID := fmt.Sprintf("%s_%s_%d", comicPathWord, chapterUUID, time.Now().UnixNano())
	task := &model.DownloadTask{
		ID:            taskID,
		ComicName:     comicName,
		ComicPathWord: comicPathWord,
		ChapterUUID:   chapterUUID,
		ChapterTitle:  chapterTitle,
		Status:        "pending",
		Progress:      "等待中",
		TotalPages:    len(images),
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	dm.tasks[taskID] = task
	return taskID
}

// StartDownload 开始下载任务
func (dm *DownloadManager) StartDownload(taskID, comicPathWord, chapterUUID string, images []model.ContentURL, words []int64) {
	dm.mu.Lock()
	task, ok := dm.tasks[taskID]
	if !ok {
		dm.mu.Unlock()
		return
	}
	task.Status = "downloading"
	task.Progress = "下载中"
	dm.mu.Unlock()

	// 创建下载目录
	chapterDir := filepath.Join(dm.downloadDir, task.ComicName, chapterUUID)
	if err := os.MkdirAll(chapterDir, 0755); err != nil {
		dm.logger.Error("创建目录失败", zap.Error(err))
		dm.updateTaskStatus(taskID, "failed", "创建目录失败")
		return
	}

	// 并发下载图片
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数

	for i, img := range images {
		wg.Add(1)
		go func(index int, imgURL string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 确定文件扩展名
			ext := ".jpg"
			if len(words) > index && words[index] == int64(index) {
				ext = ".webp"
			}

			filename := filepath.Join(chapterDir, fmt.Sprintf("%04d%s", index+1, ext))

			// 下载图片
			if err := dm.downloadImage(imgURL, filename); err != nil {
				dm.logger.Error("下载图片失败",
					zap.String("url", imgURL),
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

	// 更新任务状态
	dm.updateTaskStatus(taskID, "completed", "下载完成")
}

// downloadImage 下载图片
func (dm *DownloadManager) downloadImage(url, filename string) error {
	resp, err := http.Get(url)
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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}

// updateTaskStatus 更新任务状态
func (dm *DownloadManager) updateTaskStatus(taskID, status, progress string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if task, ok := dm.tasks[taskID]; ok {
		task.Status = status
		task.Progress = progress
	}
}

// GetTasks 获取所有任务
func (dm *DownloadManager) GetTasks() []model.DownloadTask {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	tasks := make([]model.DownloadTask, 0, len(dm.tasks))
	for _, task := range dm.tasks {
		tasks = append(tasks, *task)
	}
	return tasks
}

// GetTask 获取单个任务
func (dm *DownloadManager) GetTask(taskID string) *model.DownloadTask {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	return dm.tasks[taskID]
}

// GetDownloadedComics 获取已下载漫画列表
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
		chapterCount := countSubdirs(comicDir)
		totalPages := countFiles(comicDir)

		comics = append(comics, model.DownloadedComic{
			Name:         entry.Name(),
			ChapterCount: chapterCount,
			TotalPages:   totalPages,
		})
	}

	return comics, nil
}

// countSubdirs 计算子目录数量
func countSubdirs(dir string) int {
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}
	return count
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

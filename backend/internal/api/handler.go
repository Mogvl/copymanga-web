package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"copymanga-backend/internal/client"
	"copymanga-backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler API 处理器
type Handler struct {
	client  *client.Client
	manager *service.DownloadManager
	logger  *zap.Logger
}

// NewHandler 创建处理器
func NewHandler(c *client.Client, m *service.DownloadManager, logger *zap.Logger) *Handler {
	return &Handler{
		client:  c,
		manager: m,
		logger:  logger,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// API 路由
	api := r.Group("/api")
	{
		api.GET("/ping", h.Ping)
		api.POST("/login", h.Login)
		api.POST("/register", h.Register)
		api.GET("/search", h.Search)
		api.GET("/comic/:pathWord", h.GetComic)
		api.GET("/comic/:pathWord/group/:groupPathWord/chapters", h.GetGroupChapters)
		api.GET("/comic/:pathWord/chapter/:chapterUUID", h.GetChapterImages)
		api.POST("/download", h.StartDownload)
		api.GET("/tasks", h.GetTasks)
		api.GET("/downloaded", h.GetDownloadedComics)
	}
}

// Ping 测试连通性
func (h *Handler) Ping(c *gin.Context) {
	status, body, err := h.client.TestAPI()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       status,
		"ok":           status == 200,
		"body_preview": body,
	})
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	loginResp, err := h.client.Login(req.Username, req.Password)
	if err != nil {
		h.logger.Error("登录失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "登录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   loginResp.Token,
		"message": "登录成功",
	})
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	err := h.client.Register(req.Username, req.Password)
	if err != nil {
		h.logger.Error("注册失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "注册失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "注册成功",
	})
}

// Search 搜索漫画
func (h *Handler) Search(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "搜索关键词不能为空",
		})
		return
	}

	page := int64(1)
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}

	searchResp, err := h.client.Search(keyword, page)
	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "搜索失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    searchResp,
	})
}

// GetComic 获取漫画详情
func (h *Handler) GetComic(c *gin.Context) {
	pathWord := c.Param("pathWord")
	if pathWord == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "漫画路径不能为空",
		})
		return
	}

	comicResp, err := h.client.GetComic(pathWord)
	if err != nil {
		h.logger.Error("获取漫画失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取漫画失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    comicResp,
	})
}

// GetGroupChapters 获取分组章节
func (h *Handler) GetGroupChapters(c *gin.Context) {
	comicPathWord := c.Param("pathWord")
	groupPathWord := c.Param("groupPathWord")

	if comicPathWord == "" || groupPathWord == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数不完整",
		})
		return
	}

	chapters, err := h.client.GetGroupChapters(comicPathWord, groupPathWord)
	if err != nil {
		h.logger.Error("获取章节失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取章节失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chapters,
	})
}

// GetChapterImages 获取章节图片
func (h *Handler) GetChapterImages(c *gin.Context) {
	comicPathWord := c.Param("pathWord")
	chapterUUID := c.Param("chapterUUID")

	if comicPathWord == "" || chapterUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数不完整",
		})
		return
	}

	chapterResp, err := h.client.GetChapterImages(comicPathWord, chapterUUID)
	if err != nil {
		h.logger.Error("获取章节图片失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取章节图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chapterResp,
	})
}

// DownloadRequest 下载请求
type DownloadRequest struct {
	ComicName     string   `json:"comic_name" binding:"required"`
	ComicPathWord string   `json:"comic_path_word" binding:"required"`
	ChapterUUID   string   `json:"chapter_uuid" binding:"required"`
	ChapterTitle  string   `json:"chapter_title"`
	ImageFormat   string   `json:"image_format"`
}

// StartDownload 开始下载
func (h *Handler) StartDownload(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取章节图片
	chapterResp, err := h.client.GetChapterImages(req.ComicPathWord, req.ChapterUUID)
	if err != nil {
		h.logger.Error("获取章节图片失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取章节图片失败: " + err.Error(),
		})
		return
	}

	// 创建下载任务
	taskID := h.manager.CreateTask(
		req.ComicName,
		req.ComicPathWord,
		req.ChapterUUID,
		req.ChapterTitle,
		req.ImageFormat,
		chapterResp.Chapter.Contents,
	)

	// 后台执行下载
	go h.manager.StartDownload(
		taskID,
		req.ComicPathWord,
		req.ChapterUUID,
		req.ImageFormat,
		chapterResp.Chapter.Contents,
		chapterResp.Chapter.Words,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"task_id": taskID,
		"message": "下载任务已创建",
	})
}

// GetTasks 获取下载任务列表
func (h *Handler) GetTasks(c *gin.Context) {
	tasks := h.manager.GetTasks()
	c.JSON(http.StatusOK, tasks)
}

// GetDownloadedComics 获取已下载漫画
func (h *Handler) GetDownloadedComics(c *gin.Context) {
	comics, err := h.manager.GetDownloadedComics()
	if err != nil {
		h.logger.Error("获取已下载漫画失败", zap.Error(err))
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取已下载漫画失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    comics,
	})
}

// ServeStatic 提供静态文件服务
func (h *Handler) ServeStatic(r *gin.Engine, staticDir string) {
	// 提供前端静态文件
	r.Static("/assets", filepath.Join(staticDir, "assets"))
	r.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))
	r.StaticFile("/vite.svg", filepath.Join(staticDir, "vite.svg"))

	// 所有其他路由返回 index.html (SPA)
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "index.html"))
	})
}

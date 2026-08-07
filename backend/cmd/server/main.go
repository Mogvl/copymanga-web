package main

import (
	"fmt"
	"os"

	"copymanga-backend/internal/api"
	"copymanga-backend/internal/client"
	"copymanga-backend/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 初始化日志
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, err := logConfig.Build()
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 获取配置
	downloadDir := getEnv("DOWNLOAD_DIR", "./downloads")
	port := getEnv("PORT", "8080")
	staticDir := getEnv("STATIC_DIR", "./static")
	token := getEnv("TOKEN", "")

	// 下载相关可配置项（环境变量，默认值在 NewDownloadManager 内兜底）
	options := service.DownloadOptions{
		ImageConcurrency:   getEnvInt("IMAGE_CONCURRENCY", 5),
		ImageRetry:         getEnvInt("IMAGE_RETRY", 3),
		ChapterIntervalSec: getEnvInt("CHAPTER_INTERVAL_SEC", 0),
	}

	// 创建客户端
	c := client.New(logger)
	if token != "" {
		c.SetToken(token)
	}

	// 创建下载管理器
	manager := service.NewDownloadManager(c, downloadDir, options, logger)

	// 创建 PDF 导出器
	exporter := service.NewPdfExporter(c, downloadDir, options, logger)

	// 创建处理器
	handler := api.NewHandler(c, manager, exporter, logger)

	// 配置 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 配置 CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// 注册路由
	handler.RegisterRoutes(r)

	// 提供静态文件服务（如果存在）
	if _, err := os.Stat(staticDir); err == nil {
		handler.ServeStatic(r, staticDir)
	}

	// 启动服务
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	logger.Info("Web 服务启动",
		zap.String("addr", addr),
		zap.String("download_dir", downloadDir),
	)

	if err := r.Run(addr); err != nil {
		logger.Fatal("启动服务失败", zap.Error(err))
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var n int
		if _, err := fmt.Sscanf(value, "%d", &n); err == nil {
			return n
		}
	}
	return defaultValue
}

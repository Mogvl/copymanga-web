package service

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestExportChapterPdfOffline(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ex := NewPdfExporter(nil, "C:/Users/a9797/copymanga-web/downloads", DownloadOptions{ImageConcurrency: 5, ImageRetry: 3}, logger)

	// 最终话：本地 21 张（.jpg 实为 WebP），use_local_only 避免触发网络
	data, err := ex.ExportChapterPdf(ExportChapterParams{
		ComicName:    "咒術回戰≡",
		GroupTitle:   "",
		Order:        0,
		ChapterTitle: "最终话",
		UseLocalOnly: true,
	})
	if err != nil {
		t.Fatalf("导出失败: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("PDF 为空")
	}
	if err := os.WriteFile("test_chapter.pdf", data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("章节 PDF 生成成功，%d 字节", len(data))

	// 整本：扫 咒術回戰≡ 下所有章节
	data2, err := ex.ExportComicPdf(ExportComicParams{
		ComicName:    "咒術回戰≡",
		UseLocalOnly: true,
	})
	if err != nil {
		t.Fatalf("整本导出失败: %v", err)
	}
	if err := os.WriteFile("test_comic.pdf", data2, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("整本 PDF 生成成功，%d 字节", len(data2))
}
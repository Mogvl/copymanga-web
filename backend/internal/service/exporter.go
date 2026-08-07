package service

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"copymanga-backend/internal/client"
	"github.com/go-pdf/fpdf"
	"golang.org/x/image/webp"
	"go.uber.org/zap"
)

// PdfExporter PDF 导出器
type PdfExporter struct {
	client      *client.Client
	downloadDir string
	options     DownloadOptions
	logger      *zap.Logger
}

// NewPdfExporter 创建 PDF 导出器
func NewPdfExporter(c *client.Client, downloadDir string, options DownloadOptions, logger *zap.Logger) *PdfExporter {
	return &PdfExporter{
		client:      c,
		downloadDir: downloadDir,
		options:     options,
		logger:      logger,
	}
}

// PageSource 一页图片的数据来源（本地文件或网络 URL）
type PageSource struct {
	LocalPath string  // 本地已下载文件绝对路径（存在时非空）
	RemoteURL string  // 高清图 URL（本地缺失时非空）
	W         int     // 图片宽度
	H         int     // 图片高度
	OrderKey  float64 // 全局排序键（整本导出用）
	pageNum   int     // 页内真实页码（本地文件名解析）
}

// ExportChapterParams 单章导出参数
type ExportChapterParams struct {
	ComicName     string
	ComicPathWord string
	ChapterUUID   string
	GroupTitle    string
	Order         float64
	ChapterTitle  string
	UseLocalOnly  bool
}

// ExportComicParams 整本导出参数
type ExportComicParams struct {
	ComicName     string
	ComicPathWord string
	UseLocalOnly  bool
}

// ExportChapterPdf 导出单章 PDF
func (e *PdfExporter) ExportChapterPdf(p ExportChapterParams) ([]byte, error) {
	chapterDir := e.findChapterDir(p.ComicName, p.GroupTitle, p.ChapterTitle)
	if chapterDir == "" {
		return nil, fmt.Errorf("未找到章节目录: %s", p.ChapterTitle)
	}
	pages, err := e.collectFromChapter(p, chapterDir)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("没有可导出的图片")
	}
	return e.renderPDF(pages)
}

// ExportComicPdf 导出整本 PDF（扫描漫画目录下所有已下载图片）
func (e *PdfExporter) ExportComicPdf(p ExportComicParams) ([]byte, error) {
	comicDir := filepath.Join(e.downloadDir, sanitizeFilename(p.ComicName))
	if _, err := os.Stat(comicDir); err != nil {
		return nil, fmt.Errorf("未找到漫画目录: %s", p.ComicName)
	}

	// 递归收集该漫画下所有叶子图片，按相对路径自然排序（comic/组/章/页码）
	pages, err := e.collectLocalImages(comicDir)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("没有可导出的图片")
	}
	return e.renderPDF(pages)
}

// collectLocalImages 递归收集目录下所有图片文件，按文件路径自然升序
func (e *PdfExporter) collectLocalImages(root string) ([]PageSource, error) {
	var pages []PageSource
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".png" && ext != ".gif" {
			return nil
		}
		w, h, derr := imageDimensions(path)
		if derr != nil {
			return nil
		}
		pages = append(pages, PageSource{
			LocalPath: path,
			W:         w,
			H:         h,
			OrderKey:  fileOrderKey(root, path),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].OrderKey < pages[j].OrderKey
	})
	return pages, nil
}

// fileOrderKey 生成整本的稳定排序键：先按相对目录顺序，再按页码数字
func fileOrderKey(root, path string) float64 {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0
	}
	dir := filepath.Dir(rel)
	base := filepath.Base(rel)
	// 目录名里的首个数字（order）优先，再用页码
	dirNum := leadingNumber(dir)
	page := leadingNumber(base)
	return float64(dirNum)*100000 + float64(page)
}

// leadingNumber 取字符串里最靠前的数字（用于目录/文件的排序），找不到返回 0
func leadingNumber(s string) int {
	// 兼容纯数字
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	// 从字符串中提取首个数字序列
	var b strings.Builder
	started := false
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			started = true
		} else if started {
			break
		}
	}
	if started {
		if n, err := strconv.Atoi(b.String()); err == nil {
			return n
		}
	}
	return 0
}

// findChapterDir 在漫画目录下定位某章节的图片目录（兼容 flat 与 group 两种结构）
func (e *PdfExporter) findChapterDir(comicName, groupTitle, chapterTitle string) string {
	comicDir := filepath.Join(e.downloadDir, sanitizeFilename(comicName))
	target := sanitizeFilename(chapterTitle)
	if target == "" {
		return ""
	}
	var match string
	_ = filepath.Walk(comicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		if path == comicDir {
			return nil
		}
		name := sanitizeFilename(info.Name())
		if name == target || strings.HasSuffix(name, target) {
			// 优先选最精确匹配
			match = path
			return filepath.SkipDir // 不深入，避免匹配到同名子目录
		}
		return nil
	})
	return match
}

// collectFromChapter 从指定章节目录收集页面（本地为主 + 网络补足）
func (e *PdfExporter) collectFromChapter(p ExportChapterParams, chapterDir string) ([]PageSource, error) {
	localPages, err := scanChapterLocal(chapterDir)
	if err != nil {
		return nil, err
	}
	if p.UseLocalOnly || e.client == nil || len(localPages) == 0 {
		return localPages, nil
	}

	// 需要网络补足缺失页：按章节 API 补齐
	chapterResp, cerr := e.client.GetChapterImages(p.ComicPathWord, p.ChapterUUID)
	if cerr != nil || chapterResp == nil || len(chapterResp.Chapter.Contents) == 0 {
		// 网络失败但本地有图：退本地
		return localPages, nil
	}
	remoteCount := len(chapterResp.Chapter.Contents)
	result := make([]PageSource, remoteCount)
	for _, lp := range localPages {
		if lp.pageNum > 0 && lp.pageNum <= remoteCount {
			result[lp.pageNum-1] = lp
		}
	}
	for i := 0; i < remoteCount; i++ {
		if result[i].LocalPath == "" {
			result[i] = PageSource{RemoteURL: chapterResp.Chapter.Contents[i].URL}
		}
	}
	return result, nil
}

// scanChapterLocal 扫描章节目录本地图片（按页码升序），并读尺寸
func scanChapterLocal(chapterDir string) ([]PageSource, error) {
	entries, err := os.ReadDir(chapterDir)
	if err != nil {
		return []PageSource{}, nil
	}
	var pages []PageSource
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".png" && ext != ".gif" {
			continue
		}
		full := filepath.Join(chapterDir, entry.Name())
		w, h, derr := imageDimensions(full)
		if derr != nil {
			continue
		}
		pages = append(pages, PageSource{
			LocalPath: full,
			W:         w,
			H:         h,
			pageNum:   parsePageNum(entry.Name()),
		})
	}
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].pageNum < pages[j].pageNum
	})
	return pages, nil
}

// parsePageNum 从文件名解析页码（去扩展名转数字），失败返回 0
func parsePageNum(name string) int {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	n, err := strconv.Atoi(base)
	if err != nil {
		return 0
	}
	return n
}

// imageDimensions 读取图片尺寸（只读文件头，按魔数判断格式）
func imageDimensions(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	head := make([]byte, 12)
	if _, err := io.ReadFull(f, head); err != nil {
		return 0, 0, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return 0, 0, err
	}
	if len(head) >= 12 && string(head[0:4]) == "RIFF" && string(head[8:12]) == "WEBP" {
		cfg, err := webp.DecodeConfig(f)
		if err != nil {
			return 0, 0, err
		}
		return cfg.Width, cfg.Height, nil
	}
	if cfg, _, err := image.DecodeConfig(f); err == nil {
		return cfg.Width, cfg.Height, nil
	}
	return 0, 0, fmt.Errorf("无法识别图片格式: %s", path)
}

// renderPDF 渲染 PDF（逐页独立尺寸，1px≈1pt，img2pdf 思路）
func (e *PdfExporter) renderPDF(pages []PageSource) ([]byte, error) {
	pdf := fpdf.New("P", "pt", "", "")
	pdf.SetTitle("Copymanga PDF", true)
	for i, p := range pages {
		imgReader, w, h, err := e.loadPageImage(p)
		if err != nil {
			return nil, fmt.Errorf("第 %d 张图片加载失败: %w", i+1, err)
		}
		pdf.AddPageFormat("P", fpdf.SizeType{Wd: float64(w), Ht: float64(h)})
		name := fmt.Sprintf("page_%d", i)
		pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: "PNG"}, imgReader)
		pdf.ImageOptions(name, 0, 0, float64(w), float64(h), false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("生成 PDF 失败: %w", err)
	}
	return buf.Bytes(), nil
}

// loadPageImage 加载一页图片并转成 PNG 字节流，返回 (reader, w, h)
func (e *PdfExporter) loadPageImage(p PageSource) (io.Reader, int, int, error) {
	var raw []byte
	var err error
	switch {
	case p.LocalPath != "":
		raw, err = os.ReadFile(p.LocalPath)
	case p.RemoteURL != "":
		raw, err = e.downloadWithRetry(p.RemoteURL)
	default:
		return nil, 0, 0, fmt.Errorf("图片无数据源")
	}
	if err != nil {
		return nil, 0, 0, err
	}

	var img image.Image
	if len(raw) >= 12 && string(raw[0:4]) == "RIFF" && string(raw[8:12]) == "WEBP" {
		img, err = webp.Decode(bytes.NewReader(raw))
	} else {
		img, _, err = image.Decode(bytes.NewReader(raw))
	}
	if err != nil {
		return nil, 0, 0, fmt.Errorf("解码图片失败: %w", err)
	}
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, 0, 0, fmt.Errorf("编码 PNG 失败: %w", err)
	}
	return bytes.NewReader(buf.Bytes()), dx, dy, nil
}

// downloadWithRetry 下载远程图片（带重试）
func (e *PdfExporter) downloadWithRetry(url string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= e.options.ImageRetry; attempt++ {
		data, derr := e.downloadOnce(url)
		if derr == nil {
			return data, nil
		}
		lastErr = derr
		if attempt < e.options.ImageRetry {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return nil, fmt.Errorf("下载图片失败(重试%dd次): %w", e.options.ImageRetry, lastErr)
}

func (e *PdfExporter) downloadOnce(url string) ([]byte, error) {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
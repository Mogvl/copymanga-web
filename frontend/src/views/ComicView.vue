<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getComic, getGroupChapters, startDownload, exportChapterPdf } from '../api/client'
import type { GetComicRespData, ChapterItem } from '../types'

const route = useRoute()
const router = useRouter()
const comic = ref<GetComicRespData | null>(null)
const chapters = ref<Record<string, ChapterItem[]>>({})
const loading = ref(true)
const error = ref('')
const selectedChapters = ref<Record<string, Set<string>>>({})
const downloading = ref(false)
const exporting = ref<string | null>(null)
const imageFormat = ref('webp')
const toast = ref('')
const toastTimeout = ref<number | null>(null)

async function loadComic() {
  const pathWord = route.params.pathWord as string
  if (!pathWord) return

  loading.value = true
  error.value = ''

  try {
    comic.value = await getComic(pathWord)

    for (const groupKey of Object.keys(comic.value.groups)) {
      const group = comic.value.groups[groupKey]
      chapters.value[groupKey] = await getGroupChapters(pathWord, group.path_word)
      selectedChapters.value[groupKey] = new Set()
    }
  } catch (e: any) {
    error.value = e.message || '获取漫画失败'
  } finally {
    loading.value = false
  }
}

function toggleSelectAll(groupKey: string) {
  if (!comic.value || !chapters.value[groupKey]) return

  const allSelected = chapters.value[groupKey].every(c => selectedChapters.value[groupKey]?.has(c.uuid))

  if (allSelected) {
    selectedChapters.value[groupKey] = new Set()
  } else {
    selectedChapters.value[groupKey] = new Set(chapters.value[groupKey].map(c => c.uuid))
  }
}

function toggleChapter(groupKey: string, chapterUUID: string) {
  if (!selectedChapters.value[groupKey]) {
    selectedChapters.value[groupKey] = new Set()
  }

  if (selectedChapters.value[groupKey].has(chapterUUID)) {
    selectedChapters.value[groupKey].delete(chapterUUID)
  } else {
    selectedChapters.value[groupKey].add(chapterUUID)
  }
}

function isSelected(groupKey: string, chapterUUID: string): boolean {
  return selectedChapters.value[groupKey]?.has(chapterUUID) || false
}

function getSelectedCount(groupKey: string): number {
  return selectedChapters.value[groupKey]?.size || 0
}

async function downloadSelected(groupKey: string) {
  if (!comic.value || !chapters.value[groupKey]) return

  const selected = selectedChapters.value[groupKey]
  if (!selected || selected.size === 0) {
    showToast('请先选择要下载的章节')
    return
  }

  downloading.value = true

  try {
    for (const chapterUUID of selected) {
      const chapter = chapters.value[groupKey].find(c => c.uuid === chapterUUID)
      if (chapter) {
        await startDownload(
          comic.value.comic.name,
          comic.value.comic.path_word,
          chapterUUID,
          chapter.name,
          comic.value.groups[groupKey].name,
          chapter.ordered / 10,
          imageFormat.value
        )
      }
    }
    showToast(`已创建 ${selected.size} 个下载任务`)
  } catch (e: any) {
    showToast('创建下载任务失败: ' + e.message)
  } finally {
    downloading.value = false
  }
}

async function exportChapter(groupKey: string, chapter: ChapterItem) {
  if (!comic.value) return
  exporting.value = chapter.uuid
  try {
    const filename = `${comic.value.comic.name} ${chapter.name}.pdf`
    await exportChapterPdf(
      {
        comicName: comic.value.comic.name,
        comicPathWord: comic.value.comic.path_word,
        chapterUUID: chapter.uuid,
        groupTitle: comic.value.groups[groupKey].name,
        order: chapter.ordered / 10,
        chapterTitle: chapter.name
      },
      filename
    )
    showToast('PDF 已开始下载')
  } catch (e: any) {
    showToast('导出失败: ' + (e.message || '未知错误'))
  } finally {
    exporting.value = null
  }
}

function showToast(message: string) {
  toast.value = message
  if (toastTimeout.value) clearTimeout(toastTimeout.value)
  toastTimeout.value = window.setTimeout(() => { toast.value = '' }, 3000)
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

onMounted(loadComic)
</script>

<template>
  <div class="comic-view">
    <!-- 弥散粉紫背景光环 -->
    <div class="aura aura-1"></div>
    <div class="aura aura-2"></div>

    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <button class="back-btn" @click="goBack">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
          <span>返回</span>
        </button>
        <div class="header-title">漫画详情</div>
        <div class="header-right">
          <router-link to="/" class="nav-btn">首页</router-link>
          <router-link to="/downloaded" class="nav-btn">下载</router-link>
          <router-link to="/tasks" class="nav-btn">任务</router-link>
        </div>
      </div>
    </header>

    <!-- Toast -->
    <Transition name="toast">
      <div v-if="toast" class="toast">{{ toast }}</div>
    </Transition>

    <!-- 主内容 -->
    <main class="main">
      <!-- 加载状态 -->
      <div v-if="loading" class="state-card">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="state-card error">
        <p>{{ error }}</p>
        <button class="btn" @click="loadComic">重试</button>
      </div>

      <!-- 漫画详情 -->
      <template v-else-if="comic">
        <!-- 漫画信息卡片 -->
        <div class="info-card glass">
          <div class="comic-cover">
            <img :src="comic.comic.cover" :alt="comic.comic.name" />
          </div>
          <div class="comic-detail">
            <h1 class="comic-name">{{ comic.comic.name }}</h1>
            <div class="comic-meta">
              <span class="meta-item">
                <span class="meta-label">作者</span>
                <span class="meta-value">{{ comic.comic.author.map(a => a.name).join(', ') }}</span>
              </span>
              <span class="meta-item">
                <span class="meta-label">状态</span>
                <span class="meta-value">{{ comic.comic.status.display }}</span>
              </span>
            </div>
            <div v-if="comic.comic.theme.length > 0" class="comic-tags">
              <span v-for="theme in comic.comic.theme" :key="theme.path_word" class="tag">
                {{ theme.name }}
              </span>
            </div>
          </div>
        </div>

        <!-- 简介 -->
        <div v-if="comic.comic.brief" class="section-card glass">
          <h2 class="section-title">简介</h2>
          <p class="brief-text">{{ comic.comic.brief }}</p>
        </div>

        <!-- 章节列表 -->
        <div v-for="(group, groupKey) in comic.groups" :key="groupKey" class="section-card glass">
          <div class="section-header">
            <h2 class="section-title">{{ group.name }}</h2>
            <span class="chapter-count">{{ chapters[groupKey as string]?.length || 0 }} 章</span>
          </div>

          <!-- 工具栏 -->
          <div class="toolbar">
            <button class="tool-btn" @click="toggleSelectAll(groupKey as string)">
              {{ chapters[groupKey as string]?.every(c => isSelected(groupKey as string, c.uuid)) ? '取消全选' : '全选' }}
            </button>
            <button
              class="tool-btn primary"
              :disabled="downloading || getSelectedCount(groupKey as string) === 0"
              @click="downloadSelected(groupKey as string)"
            >
              {{ downloading ? '下载中...' : `下载选中 (${getSelectedCount(groupKey as string)})` }}
            </button>
            <span class="format-group">
              <label class="format-radio" :class="{ active: imageFormat === 'webp' }">
                <input type="radio" value="webp" v-model="imageFormat"> WebP
              </label>
              <label class="format-radio" :class="{ active: imageFormat === 'jpg' }">
                <input type="radio" value="jpg" v-model="imageFormat"> JPG
              </label>
            </span>
          </div>

          <!-- 章节网格 -->
          <div class="chapter-grid">
            <div
              v-for="chapter in chapters[groupKey as string]"
              :key="chapter.uuid"
              class="chapter-item"
              :class="{ selected: isSelected(groupKey as string, chapter.uuid) }"
              @click="toggleChapter(groupKey as string, chapter.uuid)"
            >
              <span class="chapter-check">
                <svg v-if="isSelected(groupKey as string, chapter.uuid)" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M9.55 15.16 5.39 11 4 12.39l5.55 5.55L20 7.39 18.61 6z"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <circle cx="12" cy="12" r="9"/>
                </svg>
              </span>
              <span class="chapter-order">{{ (chapter.ordered / 10).toFixed(1) }}</span>
              <span class="chapter-name">{{ chapter.name }}</span>
              <span class="chapter-pages">{{ chapter.count }}P</span>
              <button
                class="pdf-btn"
                :disabled="exporting === chapter.uuid"
                @click.stop="exportChapter(groupKey as string, chapter)"
                :title="'导出 ' + chapter.name + ' 为 PDF'"
              >
                {{ exporting === chapter.uuid ? '导出中' : 'PDF' }}
              </button>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.comic-view {
  min-height: 100dvh;
  position: relative;
}

/* 背景光环 */
.aura {
  position: fixed;
  border-radius: 50%;
  filter: blur(90px);
  pointer-events: none;
  z-index: 0;
}

.aura-1 {
  width: 520px;
  height: 520px;
  top: -160px;
  right: -120px;
  background: radial-gradient(circle, rgba(216, 180, 254, 0.42) 0%, transparent 70%);
}

.aura-2 {
  width: 460px;
  height: 460px;
  bottom: -140px;
  left: -120px;
  background: radial-gradient(circle, rgba(240, 171, 252, 0.30) 0%, transparent 70%);
}

/* 顶部导航 */
.header {
  position: sticky;
  top: 0;
  z-index: 100;
  padding: 14px 24px;
  background: rgba(247, 243, 255, 0.6);
  -webkit-backdrop-filter: blur(20px);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--divider);
}

.header-content {
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: 999px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.back-btn:hover {
  color: var(--primary);
  transform: translateX(-2px);
}

.back-btn svg {
  width: 16px;
  height: 16px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-right {
  display: flex;
  gap: 10px;
}

.nav-btn {
  padding: 10px 20px;
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: 999px;
  font-size: 13px;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.nav-btn:hover {
  color: var(--primary);
  border-color: rgba(139, 92, 246, 0.25);
}

/* Toast */
.toast {
  position: fixed;
  top: 84px;
  left: 50%;
  transform: translateX(-50%);
  padding: 14px 28px;
  background: var(--card-solid);
  color: var(--text-primary);
  border-radius: 999px;
  font-size: 14px;
  z-index: 1000;
  box-shadow: var(--shadow-hover);
  border: 1px solid var(--card-brd);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-10px);
}

/* 主内容 */
.main {
  position: relative;
  z-index: 1;
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 24px 72px;
}

/* 状态卡片 */
.state-card {
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: var(--radius);
  padding: 96px 20px;
  text-align: center;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(20px) saturate(1.4);
  backdrop-filter: blur(20px) saturate(1.4);
}

.state-card.error {
  color: #C26D7A;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 2px solid var(--divider);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 20px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn {
  margin-top: 20px;
  padding: 12px 32px;
  background: linear-gradient(135deg, var(--primary), #A78BFA);
  color: white;
  border: none;
  border-radius: 999px;
  font-size: 14px;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.2s;
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.28);
}

.btn:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

/* 漫画信息卡片 */
.info-card {
  display: flex;
  gap: 32px;
  padding: 32px;
  margin-bottom: 20px;
}

.comic-cover {
  width: 168px;
  height: 224px;
  flex-shrink: 0;
  border-radius: var(--radius-sm);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.comic-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.comic-detail {
  flex: 1;
  min-width: 0;
}

.comic-name {
  font-size: 26px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 24px;
  line-height: 1.3;
  letter-spacing: -0.3px;
}

.comic-meta {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 20px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.meta-label {
  font-size: 13px;
  color: var(--text-hint);
  min-width: 36px;
}

.meta-value {
  font-size: 14px;
  color: var(--text-primary);
}

.comic-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.tag {
  padding: 7px 18px;
  background: var(--primary-soft);
  border: 1px solid var(--divider);
  border-radius: 999px;
  font-size: 12px;
  color: var(--primary);
}

/* 区块卡片 */
.section-card {
  padding: 28px 32px;
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.chapter-count {
  font-size: 13px;
  color: var(--text-hint);
  background: var(--primary-soft);
  padding: 5px 14px;
  border-radius: 999px;
}

.brief-text {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.9;
}

/* 图片格式选择 */
.format-group {
  display: flex;
  gap: 2px;
  background: var(--glass-hi);
  border: 1px solid var(--divider);
  border-radius: 999px;
  padding: 3px;
  margin-left: auto;
}

.format-radio {
  padding: 6px 14px;
  font-size: 12px;
  border-radius: 999px;
  cursor: pointer;
  color: var(--text-hint);
  transition: all 0.2s;
}

.format-radio input { display: none; }
.format-radio.active { background: linear-gradient(135deg, var(--primary), #A78BFA); color: #fff; box-shadow: 0 4px 12px rgba(139, 92, 246, 0.25); }
.format-radio:hover:not(.active) { color: var(--text-primary); }

/* 工具栏 */
.toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--divider);
}

.tool-btn {
  padding: 10px 24px;
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: 999px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.tool-btn:hover {
  color: var(--primary);
  background: #fff;
  border-color: rgba(139, 92, 246, 0.25);
}

.tool-btn.primary {
  background: linear-gradient(135deg, var(--primary), #A78BFA);
  color: white;
  border-color: transparent;
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.28);
}

.tool-btn.primary:hover:not(:disabled) {
  opacity: 0.92;
  color: white;
}

.tool-btn.primary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  box-shadow: none;
}

/* 章节网格 */
.chapter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.chapter-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  background: var(--glass-hi);
  border: 1px solid var(--divider);
  border-radius: var(--radius-xs);
  cursor: pointer;
  transition: all 0.2s;
  -webkit-backdrop-filter: blur(12px) saturate(1.3);
  backdrop-filter: blur(12px) saturate(1.3);
}

.chapter-item:hover {
  background: #fff;
  border-color: rgba(139, 92, 246, 0.22);
}

.chapter-item.selected {
  background: var(--primary-soft);
  border-color: rgba(139, 92, 246, 0.35);
}

.chapter-check {
  flex-shrink: 0;
  width: 18px;
  height: 18px;
  color: var(--text-hint);
}

.chapter-check svg {
  width: 100%;
  height: 100%;
}

.chapter-item.selected .chapter-check {
  color: var(--primary);
}

.chapter-order {
  flex-shrink: 0;
  width: 32px;
  font-size: 12px;
  color: var(--text-hint);
  text-align: right;
}

.chapter-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chapter-pages {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--text-hint);
}

.pdf-btn {
  flex-shrink: 0;
  margin-left: 4px;
  padding: 5px 14px;
  background: var(--primary-soft);
  border: 1px solid var(--divider);
  border-radius: 999px;
  font-size: 11px;
  color: var(--primary);
  cursor: pointer;
  transition: all 0.2s;
}

.pdf-btn:hover:not(:disabled) {
  background: var(--primary);
  color: white;
  border-color: transparent;
}

.pdf-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 移动端适配 */
@media (max-width: 640px) {
  .header {
    padding: 12px 16px;
  }

  .back-btn {
    padding: 8px 14px;
    font-size: 13px;
  }

  .nav-btn {
    padding: 8px 14px;
    font-size: 12px;
  }

  .main {
    padding: 20px 16px 56px;
  }

  .info-card {
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 24px;
    gap: 20px;
  }

  .comic-cover {
    width: 132px;
    height: 176px;
  }

  .comic-name {
    font-size: 21px;
  }

  .comic-meta {
    align-items: center;
  }

  .comic-tags {
    justify-content: center;
  }

  .section-card {
    padding: 20px;
  }

  .chapter-grid {
    grid-template-columns: 1fr;
  }
}
</style>

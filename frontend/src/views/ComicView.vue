<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getComic, getGroupChapters, startDownload } from '../api/client'
import type { GetComicRespData, ChapterItem } from '../types'

const route = useRoute()
const router = useRouter()
const comic = ref<GetComicRespData | null>(null)
const chapters = ref<Record<string, ChapterItem[]>>({})
const loading = ref(true)
const error = ref('')
const selectedChapters = ref<Record<string, Set<string>>>({})
const downloading = ref(false)
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
    <!-- 背景 -->
    <div class="bg">
      <div class="bg-circle bg-circle-1"></div>
      <div class="bg-circle bg-circle-2"></div>
    </div>

    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <button class="back-btn" @click="goBack">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
        <div class="info-card">
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
        <div v-if="comic.comic.brief" class="section-card">
          <h2 class="section-title">📖 简介</h2>
          <p class="brief-text">{{ comic.comic.brief }}</p>
        </div>

        <!-- 章节列表 -->
        <div v-for="(group, groupKey) in comic.groups" :key="groupKey" class="section-card">
          <div class="section-header">
            <h2 class="section-title">📚 {{ group.name }}</h2>
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
                <svg v-if="isSelected(groupKey as string, chapter.uuid)" viewBox="0 0 24 24" fill="var(--primary)">
                  <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                </svg>
                <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                </svg>
              </span>
              <span class="chapter-order">{{ (chapter.ordered / 10).toFixed(1) }}</span>
              <span class="chapter-name">{{ chapter.name }}</span>
              <span class="chapter-pages">{{ chapter.count }}P</span>
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.comic-view {
  min-height: 100vh;
  position: relative;
}

/* 背景 */
.bg {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #FFF5EB 0%, #FFF0E0 50%, #FFE8CC 100%);
}

.bg-circle {
  position: absolute;
  border-radius: 50%;
  opacity: 0.3;
}

.bg-circle-1 {
  width: 500px;
  height: 500px;
  top: -150px;
  right: -100px;
  background: radial-gradient(circle, rgba(255, 105, 0, 0.3) 0%, transparent 70%);
}

.bg-circle-2 {
  width: 400px;
  height: 400px;
  bottom: -100px;
  left: -100px;
  background: radial-gradient(circle, rgba(255, 150, 50, 0.25) 0%, transparent 70%);
}

/* 顶部导航 */
.header {
  position: sticky;
  top: 0;
  z-index: 100;
  padding: 12px 20px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.5);
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
  gap: 6px;
  padding: 8px 16px;
  background: var(--card);
  border: none;
  border-radius: 20px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
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
  gap: 8px;
}

.nav-btn {
  padding: 8px 16px;
  background: var(--card);
  border-radius: 20px;
  font-size: 13px;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.nav-btn:hover {
  color: var(--primary);
}

/* Toast */
.toast {
  position: fixed;
  top: 72px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  background: #333;
  color: white;
  border-radius: 24px;
  font-size: 14px;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
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
  padding: 20px;
}

/* 状态卡片 */
.state-card {
  background: var(--card);
  border-radius: var(--radius);
  padding: 80px 20px;
  text-align: center;
  box-shadow: var(--shadow);
}

.state-card.error {
  color: #D32F2F;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--divider);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn {
  margin-top: 20px;
  padding: 10px 24px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 20px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:hover {
  background: var(--primary-hover);
}

/* 漫画信息卡片 */
.info-card {
  display: flex;
  gap: 24px;
  background: var(--card);
  border-radius: var(--radius);
  padding: 24px;
  box-shadow: var(--shadow);
  margin-bottom: 16px;
}

.comic-cover {
  width: 160px;
  height: 213px;
  flex-shrink: 0;
  border-radius: var(--radius-sm);
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
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
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 16px;
  line-height: 1.3;
}

.comic-meta {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
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
  gap: 8px;
}

.tag {
  padding: 5px 14px;
  background: var(--primary-light);
  border-radius: 16px;
  font-size: 12px;
  color: var(--primary);
}

/* 区块卡片 */
.section-card {
  background: var(--card);
  border-radius: var(--radius);
  padding: 20px;
  box-shadow: var(--shadow);
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.chapter-count {
  font-size: 13px;
  color: var(--text-hint);
  background: var(--bg);
  padding: 4px 12px;
  border-radius: 12px;
}

/* 图片格式选择 */
.format-group {
  display: flex; gap: 2px; background: var(--bg); border-radius: 6px; padding: 2px; margin-left: auto; border: 1px solid var(--divider);
}
.format-radio {
  padding: 4px 10px; font-size: 12px; border-radius: 4px; cursor: pointer; color: var(--text-hint); transition: all .2s;
}
.format-radio input { display: none; }
.format-radio.active { background: var(--primary); color: #fff; }
.format-radio:hover:not(.active) { color: var(--text-primary); }

.brief-text {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.8;
}

/* 工具栏 */
.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--divider);
}

.tool-btn {
  padding: 8px 18px;
  background: var(--bg);
  border: none;
  border-radius: 20px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.tool-btn:hover {
  color: var(--primary);
  background: var(--primary-light);
}

.tool-btn.primary {
  background: var(--primary);
  color: white;
}

.tool-btn.primary:hover:not(:disabled) {
  background: var(--primary-hover);
}

.tool-btn.primary:disabled {
  background: var(--divider);
  color: var(--text-hint);
  cursor: not-allowed;
}

/* 章节网格 */
.chapter-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.chapter-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--bg);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
}

.chapter-item:hover {
  background: var(--primary-light);
}

.chapter-item.selected {
  background: var(--primary-light);
  box-shadow: inset 0 0 0 2px var(--primary);
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

/* 移动端适配 */
@media (max-width: 640px) {
  .header {
    padding: 10px 16px;
  }

  .back-btn {
    padding: 6px 12px;
    font-size: 13px;
  }

  .nav-btn {
    padding: 6px 12px;
    font-size: 12px;
  }

  .main {
    padding: 16px;
  }

  .info-card {
    flex-direction: column;
    align-items: center;
    text-align: center;
    padding: 20px;
  }

  .comic-cover {
    width: 120px;
    height: 160px;
  }

  .comic-name {
    font-size: 20px;
  }

  .comic-meta {
    align-items: center;
  }

  .comic-tags {
    justify-content: center;
  }

  .section-card {
    padding: 16px;
  }

  .chapter-grid {
    grid-template-columns: 1fr;
  }
}
</style>

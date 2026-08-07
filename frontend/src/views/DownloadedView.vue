<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDownloadedComics, exportComicPdf } from '../api/client'
import type { DownloadedComic } from '../types'

const router = useRouter()
const comics = ref<DownloadedComic[]>([])
const loading = ref(true)
const error = ref('')
const exporting = ref('')

onMounted(async () => {
  try {
    comics.value = await getDownloadedComics()
  } catch (e: any) {
    error.value = e.message || '获取已下载漫画失败'
  } finally {
    loading.value = false
  }
})

async function exportPdf(comic: DownloadedComic) {
  exporting.value = comic.name
  try {
    const filename = `${comic.name}.pdf`
    await exportComicPdf(
      {
        comicName: comic.name,
        comicPathWord: comic.path_word
      },
      filename
    )
    alert('PDF 已开始下载')
  } catch (e: any) {
    alert('导出失败: ' + (e.message || '未知错误'))
  } finally {
    exporting.value = ''
  }
}

function goBack() {
  router.push('/')
}
</script>

<template>
  <div class="downloaded-view">
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
          <span>返回首页</span>
        </button>
        <div class="header-title">📥 已下载</div>
        <div class="header-right">
          <router-link to="/tasks" class="nav-btn">任务</router-link>
        </div>
      </div>
    </header>

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
      </div>

      <!-- 空状态 -->
      <div v-else-if="comics.length === 0" class="state-card">
        <div class="empty-icon">📭</div>
        <h3>还没有下载</h3>
        <p>搜索并下载你喜欢的漫画吧</p>
        <button class="btn" @click="goBack">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <circle cx="11" cy="11" r="8"/>
            <path d="m21 21-4.35-4.35"/>
          </svg>
          去搜索
        </button>
      </div>

      <!-- 已下载列表 -->
      <template v-else>
        <div class="list-info">
          <span>共 {{ comics.length }} 部漫画</span>
        </div>

        <div class="comic-list">
          <div v-for="comic in comics" :key="comic.name" class="comic-card">
            <div class="card-icon">📖</div>
            <div class="card-content">
              <h3 class="comic-name">{{ comic.name }}</h3>
              <div class="comic-stats">
                <span class="stat">
                  <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
                    <path d="M18 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zM6 4h5v8l-2.5-1.5L6 12V4z"/>
                  </svg>
                  {{ comic.chapter_count }} 章
                </span>
                <span class="stat-divider">·</span>
                <span class="stat">
                  <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
                    <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
                  </svg>
                  {{ comic.total_pages }} 图
                </span>
              </div>
            </div>
            <button
              class="pdf-btn"
              :disabled="exporting === comic.name"
              @click="exportPdf(comic)"
            >
              {{ exporting === comic.name ? '导出中...' : '导出整本 PDF' }}
            </button>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.downloaded-view {
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

.empty-icon {
  font-size: 56px;
  margin-bottom: 20px;
}

.state-card h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.state-card p {
  font-size: 14px;
  color: var(--text-hint);
  margin-bottom: 24px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 28px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 15px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 4px 12px rgba(255, 105, 0, 0.3);
}

.btn:hover {
  background: var(--primary-hover);
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 105, 0, 0.4);
}

/* 列表信息 */
.list-info {
  margin-bottom: 16px;
  padding: 0 4px;
}

.list-info span {
  font-size: 13px;
  color: var(--text-hint);
}

/* 漫画列表 */
.comic-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.comic-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 20px;
  background: var(--card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  transition: all 0.3s ease;
}

.comic-card:hover {
  transform: translateX(6px);
  box-shadow: var(--shadow-hover);
}

.card-icon {
  font-size: 36px;
  flex-shrink: 0;
}

.card-content {
  flex: 1;
  min-width: 0;
}

.comic-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.comic-stats {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--text-secondary);
}

.stat svg {
  color: var(--text-hint);
}

.stat-divider {
  color: var(--divider);
}

.pdf-btn {
  flex-shrink: 0;
  padding: 8px 16px;
  background: var(--primary-light);
  border: none;
  border-radius: 18px;
  font-size: 13px;
  color: var(--primary);
  cursor: pointer;
  transition: all 0.2s;
}

.pdf-btn:hover:not(:disabled) {
  background: var(--primary);
  color: white;
}

.pdf-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

  .main {
    padding: 16px;
  }

  .comic-card {
    padding: 14px 16px;
  }

  .card-icon {
    font-size: 28px;
  }

  .comic-name {
    font-size: 15px;
  }
}
</style>

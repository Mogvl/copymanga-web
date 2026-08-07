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
          <span>返回首页</span>
        </button>
        <div class="header-title">已下载</div>
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
        <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
          <rect x="3" y="3" width="18" height="18" rx="3"/>
          <path d="M3 9h18"/>
          <path d="M9 21V9"/>
        </svg>
        <h3>还没有下载</h3>
        <p>搜索并下载你喜欢的漫画吧</p>
        <button class="btn" @click="goBack">去搜索</button>
      </div>

      <!-- 已下载列表 -->
      <template v-else>
        <div class="list-info">
          <span>共 {{ comics.length }} 部漫画</span>
        </div>

        <div class="comic-list">
          <div v-for="comic in comics" :key="comic.name" class="comic-card">
            <svg class="card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="4" y="3" width="16" height="18" rx="3"/>
              <path d="M8 8h8M8 12h8M8 16h4"/>
            </svg>
            <div class="card-content">
              <h3 class="comic-name">{{ comic.name }}</h3>
              <div class="comic-stats">
                <span class="stat">{{ comic.chapter_count }} 章</span>
                <span class="stat-divider">·</span>
                <span class="stat">{{ comic.total_pages }} 图</span>
              </div>
            </div>
            <button
              class="pdf-btn"
              :disabled="exporting === comic.name"
              @click="exportPdf(comic)"
            >
              {{ exporting === comic.name ? '导出中' : '导出整本 PDF' }}
            </button>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.downloaded-view {
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

.empty-icon {
  width: 52px;
  height: 52px;
  color: var(--text-hint);
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
  transform: translateY(-2px);
}

/* 列表信息 */
.list-info {
  margin-bottom: 24px;
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
  gap: 16px;
}

.comic-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px 28px;
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(20px) saturate(1.4);
  backdrop-filter: blur(20px) saturate(1.4);
  transition: all 0.3s ease;
}

.comic-card:hover {
  transform: translateX(4px);
  box-shadow: var(--shadow-hover), var(--shadow-inset);
  border-color: rgba(139, 92, 246, 0.22);
}

.card-icon {
  width: 34px;
  height: 34px;
  color: var(--primary);
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
  margin-bottom: 8px;
}

.comic-stats {
  display: flex;
  align-items: center;
  gap: 10px;
}

.stat {
  font-size: 13px;
  color: var(--text-hint);
}

.stat-divider {
  color: var(--divider);
}

.pdf-btn {
  flex-shrink: 0;
  padding: 10px 22px;
  background: var(--primary-soft);
  border: 1px solid var(--divider);
  border-radius: 999px;
  font-size: 13px;
  color: var(--primary);
  cursor: pointer;
  transition: all 0.2s;
}

.pdf-btn:hover:not(:disabled) {
  background: var(--primary);
  color: white;
  border-color: transparent;
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.28);
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

  .comic-card {
    padding: 18px 20px;
    gap: 14px;
  }

  .card-icon {
    width: 28px;
    height: 28px;
  }

  .comic-name {
    font-size: 15px;
  }
}
</style>

<script lang="ts">
export default {
  name: 'SearchView'
}
</script>

<script setup lang="ts">
import { ref, onActivated } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { search } from '../api/client'
import type { ComicInSearch } from '../types'

const route = useRoute()
const router = useRouter()
const keyword = ref('')
const comics = ref<ComicInSearch[]>([])
const page = ref(1)
const totalPages = ref(0)
const loading = ref(false)
const error = ref('')
const lastKeyword = ref('')
const lastPage = ref(0)

async function doSearch() {
  const q = keyword.value.trim()
  if (!q) return

  if (q === lastKeyword.value && page.value === lastPage.value && comics.value.length > 0) {
    return
  }

  loading.value = true
  error.value = ''
  lastKeyword.value = q
  lastPage.value = page.value

  try {
    const data = await search(q, page.value)
    comics.value = data.list
    totalPages.value = Math.ceil(data.total / 20)
  } catch (e: any) {
    error.value = e.message || '搜索失败'
    comics.value = []
  } finally {
    loading.value = false
  }
}

async function goToPage(p: number) {
  page.value = p
  router.push({ name: 'search', query: { q: keyword.value, page: p } })
  await doSearch()
}

function handleSubmit() {
  lastKeyword.value = ''
  page.value = 1
  router.push({ name: 'search', query: { q: keyword.value, page: 1 } })
  doSearch()
}

function goBack() {
  router.push('/')
}

onActivated(() => {
  const newKeyword = (route.query.q as string) || ''
  const newPage = parseInt(route.query.page as string) || 1

  if (newKeyword !== lastKeyword.value || newPage !== lastPage.value) {
    keyword.value = newKeyword
    page.value = newPage
    if (keyword.value) {
      doSearch()
    }
  }
})
</script>

<template>
  <div class="search-view">
    <!-- 弥散粉紫背景光环 -->
    <div class="aura aura-1"></div>
    <div class="aura aura-2"></div>

    <!-- 顶部搜索栏 -->
    <header class="header">
      <div class="header-content">
        <button class="back-btn" @click="goBack">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
        </button>

        <div class="search-card glass">
          <form class="search-form" @submit.prevent="handleSubmit">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="11" cy="11" r="8"/>
              <path d="m21 21-4.35-4.35"/>
            </svg>
            <input
              v-model="keyword"
              type="text"
              placeholder="搜索漫画..."
              autofocus
            />
            <button v-if="keyword" type="button" class="clear-btn" @click="keyword = ''">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M18.3 5.71 12 12l6.3 6.29a1 1 0 1 1-1.42 1.42L10.6 13.4l-6.3 6.3a1 1 0 0 1-1.42-1.42L9.17 12 2.88 5.71A1 1 0 0 1 4.3 4.3L10.6 10.6l6.29-6.3a1 1 0 1 1 1.42 1.42z"/>
              </svg>
            </button>
          </form>
        </div>

        <router-link to="/downloaded" class="icon-btn" title="已下载">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="3" y="3" width="18" height="18" rx="3"/>
            <path d="M3 9h18"/>
            <path d="M9 21V9"/>
          </svg>
        </router-link>
      </div>
    </header>

    <!-- 主内容 -->
    <main class="main">
      <!-- 加载状态 -->
      <div v-if="loading" class="state-card">
        <div class="spinner"></div>
        <span>搜索中...</span>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="state-card error">
        <p>{{ error }}</p>
        <button class="btn" @click="doSearch">重试</button>
      </div>

      <!-- 空状态 -->
      <div v-else-if="comics.length === 0 && keyword" class="state-card">
        <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
          <circle cx="11" cy="11" r="8"/>
          <path d="m21 21-4.35-4.35"/>
          <path d="M8 11h6M11 8v6"/>
        </svg>
        <h3>没有找到相关漫画</h3>
        <p>换个关键词试试</p>
      </div>

      <!-- 搜索结果 -->
      <template v-else>
        <div class="result-info">
          <span>找到 {{ comics.length }} 个结果</span>
          <span class="page-badge">第 {{ page }} 页</span>
        </div>

        <div class="comic-grid">
          <router-link
            v-for="comic in comics"
            :key="comic.path_word"
            :to="{ name: 'comic', params: { pathWord: comic.path_word } }"
            class="comic-card"
          >
            <div class="comic-cover">
              <img
                :src="comic.cover"
                :alt="comic.name"
                loading="lazy"
                @error="($event.target as HTMLImageElement).style.display='none'"
              />
            </div>
            <div class="comic-info">
              <h3 class="comic-name">{{ comic.name }}</h3>
              <p class="comic-author">{{ comic.author.map(a => a.name).join(', ') }}</p>
              <p class="comic-popular">{{ comic.popular }} 热度</p>
            </div>
          </router-link>
        </div>

        <!-- 分页 -->
        <div v-if="totalPages > 1" class="pagination">
          <button
            :disabled="page <= 1"
            class="page-btn"
            @click="goToPage(page - 1)"
          >
            上一页
          </button>

          <div class="page-numbers">
            <button
              v-for="p in totalPages"
              :key="p"
              v-show="p >= page - 2 && p <= page + 2"
              :class="['page-num', { active: p === page }]"
              @click="goToPage(p)"
            >
              {{ p }}
            </button>
          </div>

          <button
            :disabled="page >= totalPages"
            class="page-btn"
            @click="goToPage(page + 1)"
          >
            下一页
          </button>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.search-view {
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

/* 顶部搜索栏 */
.header {
  position: sticky;
  top: 0;
  z-index: 100;
  padding: 20px 32px;
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
  gap: 16px;
}

.back-btn,
.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border: 1px solid var(--card-brd);
  background: var(--glass-hi);
  border-radius: 50%;
  cursor: pointer;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  flex-shrink: 0;
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.back-btn:hover,
.icon-btn:hover {
  color: var(--primary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-hover);
}

.back-btn svg,
.icon-btn svg {
  width: 20px;
  height: 20px;
}

.search-card {
  flex: 1;
  border-radius: 24px;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
}

.search-form {
  display: flex;
  align-items: center;
  padding: 4px 6px 4px 18px;
}

.search-icon {
  width: 19px;
  height: 19px;
  color: var(--text-hint);
  flex-shrink: 0;
}

.search-form input {
  flex: 1;
  padding: 12px 10px;
  border: none;
  outline: none;
  font-size: 15px;
  background: transparent;
  color: var(--text-primary);
}

.search-form input::placeholder {
  color: var(--text-hint);
}

.clear-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--text-hint);
  border-radius: 50%;
  transition: all 0.2s;
  margin-right: 4px;
}

.clear-btn:hover {
  background: var(--primary-soft);
  color: var(--primary);
}

.clear-btn svg {
  width: 16px;
  height: 16px;
}

/* 主内容 */
.main {
  position: relative;
  z-index: 1;
  max-width: 960px;
  margin: 0 auto;
  padding: 40px 32px 64px;
}

/* 状态卡片 */
.state-card {
  background: var(--glass-hi);
  -webkit-backdrop-filter: blur(20px) saturate(1.4);
  backdrop-filter: blur(20px) saturate(1.4);
  border: 1px solid var(--card-brd);
  border-radius: var(--radius);
  padding: 96px 20px;
  text-align: center;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
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
}

.btn {
  margin-top: 24px;
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

/* 结果信息 */
.result-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  padding: 0 4px;
}

.result-info span:first-child {
  font-size: 13px;
  color: var(--text-hint);
}

.page-badge {
  font-size: 12px;
  color: var(--primary);
  background: var(--primary-soft);
  padding: 6px 16px;
  border-radius: 999px;
}

/* 网格布局 */
.comic-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(168px, 1fr));
  gap: 24px;
  margin-bottom: 40px;
}

.comic-card {
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: var(--radius-sm);
  overflow: hidden;
  text-decoration: none;
  color: inherit;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(16px) saturate(1.3);
  backdrop-filter: blur(16px) saturate(1.3);
}

.comic-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover), var(--shadow-inset);
  border-color: rgba(139, 92, 246, 0.22);
}

.comic-cover {
  position: relative;
  width: 100%;
  padding-top: 133%;
  background: var(--primary-soft);
  overflow: hidden;
}

.comic-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.4s ease;
}

.comic-card:hover .comic-cover img {
  transform: scale(1.04);
}

.comic-info {
  padding: 16px 18px 20px;
}

.comic-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.comic-author {
  font-size: 12px;
  color: var(--text-hint);
  margin-bottom: 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.comic-popular {
  font-size: 11px;
  color: var(--accent-2);
  font-weight: 500;
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 24px 0;
}

.page-btn {
  padding: 12px 24px;
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

.page-btn:hover:not(:disabled) {
  color: var(--primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-hover);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-numbers {
  display: flex;
  gap: 6px;
}

.page-num {
  min-width: 38px;
  height: 38px;
  padding: 0 10px;
  background: var(--glass-hi);
  border: 1px solid var(--card-brd);
  border-radius: 50%;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
}

.page-num:hover {
  color: var(--primary);
}

.page-num.active {
  background: linear-gradient(135deg, var(--primary), #A78BFA);
  color: white;
  border-color: transparent;
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.30);
}

/* 移动端适配 */
@media (max-width: 640px) {
  .header {
    padding: 14px 16px;
  }

  .header-content {
    gap: 10px;
  }

  .back-btn,
  .icon-btn {
    width: 38px;
    height: 38px;
  }

  .search-form {
    padding: 3px 4px 3px 14px;
  }

  .search-form input {
    padding: 10px 8px;
    font-size: 14px;
  }

  .main {
    padding: 24px 16px 48px;
  }

  .comic-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 14px;
  }

  .comic-info {
    padding: 12px 14px 16px;
  }

  .comic-name {
    font-size: 13px;
  }

  .pagination {
    gap: 8px;
  }

  .page-btn {
    padding: 10px 16px;
    font-size: 13px;
  }

  .page-num {
    min-width: 34px;
    height: 34px;
    font-size: 13px;
  }
}

@media (max-width: 400px) {
  .comic-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
}
</style>

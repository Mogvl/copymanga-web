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
    <!-- 背景 -->
    <div class="bg">
      <div class="bg-circle bg-circle-1"></div>
      <div class="bg-circle bg-circle-2"></div>
    </div>

    <!-- 顶部搜索栏 -->
    <header class="header">
      <div class="header-content">
        <button class="back-btn" @click="goBack">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 12H5M12 19l-7-7 7-7"/>
          </svg>
        </button>

        <div class="search-card">
          <form class="search-form" @submit.prevent="handleSubmit">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
                <path d="M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm5 13.59L15.59 17 12 13.41 8.41 17 7 15.59 10.59 12 7 8.41 8.41 7 12 10.59 15.59 7 17 8.41 13.41 12 17 15.59z"/>
              </svg>
            </button>
          </form>
        </div>

        <router-link to="/downloaded" class="download-btn">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7,10 12,15 17,10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
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
        <div class="empty-icon">🔍</div>
        <p>没有找到相关漫画</p>
        <span>换个关键词试试</span>
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
              <div class="cover-placeholder">📚</div>
            </div>
            <div class="comic-info">
              <h3 class="comic-name">{{ comic.name }}</h3>
              <p class="comic-author">{{ comic.author.map(a => a.name).join(', ') }}</p>
              <p class="comic-popular">🔥 {{ comic.popular }}</p>
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

/* 顶部搜索栏 */
.header {
  position: sticky;
  top: 0;
  z-index: 100;
  padding: 16px 20px;
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
  gap: 12px;
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  background: var(--card);
  border-radius: 50%;
  cursor: pointer;
  color: var(--text-primary);
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  flex-shrink: 0;
}

.back-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.back-btn svg {
  width: 20px;
  height: 20px;
}

.search-card {
  flex: 1;
  background: var(--card);
  border-radius: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  transition: all 0.3s;
}

.search-card:focus-within {
  box-shadow: 0 4px 20px rgba(255, 105, 0, 0.15);
}

.search-form {
  display: flex;
  align-items: center;
  padding: 4px;
}

.search-icon {
  width: 20px;
  height: 20px;
  margin-left: 14px;
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
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--text-hint);
  border-radius: 50%;
  transition: all 0.2s;
  margin-right: 4px;
}

.clear-btn:hover {
  background: var(--bg);
  color: var(--text-secondary);
}

.clear-btn svg {
  width: 18px;
  height: 18px;
}

.download-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--card);
  border-radius: 50%;
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  flex-shrink: 0;
}

.download-btn:hover {
  color: var(--primary);
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.download-btn svg {
  width: 20px;
  height: 20px;
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
  padding: 60px 20px;
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
  font-size: 48px;
  margin-bottom: 16px;
}

.state-card p {
  font-size: 16px;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.state-card span {
  font-size: 14px;
  color: var(--text-hint);
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

/* 结果信息 */
.result-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 0 4px;
}

.result-info span:first-child {
  font-size: 13px;
  color: var(--text-hint);
}

.page-badge {
  font-size: 12px;
  color: var(--primary);
  background: var(--primary-light);
  padding: 4px 12px;
  border-radius: 12px;
}

/* 网格布局 */
.comic-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.comic-card {
  background: var(--card);
  border-radius: var(--radius);
  overflow: hidden;
  text-decoration: none;
  color: inherit;
  transition: all 0.3s ease;
  box-shadow: var(--shadow);
}

.comic-card:hover {
  transform: translateY(-6px);
  box-shadow: var(--shadow-hover);
}

.comic-cover {
  position: relative;
  width: 100%;
  padding-top: 133%;
  background: var(--bg);
  overflow: hidden;
}

.comic-cover img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.comic-card:hover .comic-cover img {
  transform: scale(1.05);
}

.cover-placeholder {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 36px;
  opacity: 0.5;
}

.comic-info {
  padding: 12px;
}

.comic-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.comic-author {
  font-size: 12px;
  color: var(--text-hint);
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.comic-popular {
  font-size: 11px;
  color: var(--primary);
}

/* 分页 */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px 0;
}

.page-btn {
  padding: 10px 20px;
  background: var(--card);
  border: none;
  border-radius: 20px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: var(--shadow);
}

.page-btn:hover:not(:disabled) {
  color: var(--primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-hover);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-numbers {
  display: flex;
  gap: 4px;
}

.page-num {
  min-width: 36px;
  height: 36px;
  padding: 0 8px;
  background: var(--card);
  border: none;
  border-radius: 18px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.page-num:hover {
  color: var(--primary);
}

.page-num.active {
  background: var(--primary);
  color: white;
  box-shadow: 0 2px 8px rgba(255, 105, 0, 0.3);
}

/* 移动端适配 */
@media (max-width: 640px) {
  .header {
    padding: 12px 16px;
  }

  .header-content {
    gap: 10px;
  }

  .back-btn,
  .download-btn {
    width: 36px;
    height: 36px;
  }

  .search-form input {
    padding: 10px 8px;
    font-size: 14px;
  }

  .main {
    padding: 16px;
  }

  .comic-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  .comic-info {
    padding: 8px 10px;
  }

  .comic-name {
    font-size: 13px;
  }

  .pagination {
    gap: 6px;
  }

  .page-btn {
    padding: 8px 14px;
    font-size: 13px;
  }

  .page-num {
    min-width: 32px;
    height: 32px;
    font-size: 13px;
  }
}

@media (max-width: 400px) {
  .comic-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }
}
</style>

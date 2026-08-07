<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { getTasks } from '../api/client'
import type { DownloadTask } from '../types'

const router = useRouter()
const tasks = ref<DownloadTask[]>([])
const loading = ref(true)
const error = ref('')
let pollTimer: number | null = null

const statusMap: Record<DownloadTask['status'], { label: string; cls: string }> = {
  pending: { label: '等待中', cls: 'pending' },
  downloading: { label: '下载中', cls: 'downloading' },
  completed: { label: '已完成', cls: 'completed' },
  failed: { label: '失败', cls: 'failed' }
}

async function loadTasks() {
  try {
    tasks.value = await getTasks()
    error.value = ''
  } catch (e: any) {
    error.value = e.message || '获取任务失败'
  } finally {
    loading.value = false
  }
}

function hasActive(tasks: DownloadTask[]): boolean {
  return tasks.some(t => t.status === 'pending' || t.status === 'downloading')
}

onMounted(() => {
  loadTasks()
  // 有任务下载中时，每 2 秒轮询刷新；无活动任务则停止轮询
  pollTimer = window.setInterval(() => {
    if (hasActive(tasks.value)) {
      loadTasks()
    }
  }, 2000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

function goBack() {
  router.push('/')
}
</script>

<template>
  <div class="tasks-view">
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
        <div class="header-title">下载任务</div>
        <div class="header-right">
          <router-link to="/" class="nav-btn">首页</router-link>
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
        <button class="btn" @click="loadTasks">重试</button>
      </div>

      <!-- 空状态 -->
      <div v-else-if="tasks.length === 0" class="state-card">
        <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
          <circle cx="12" cy="12" r="9"/>
          <path d="M12 7v5l3 3"/>
        </svg>
        <h3>暂无下载任务</h3>
        <p>去漫画详情页勾选章节即可下载</p>
        <button class="btn" @click="goBack">去搜索</button>
      </div>

      <!-- 任务列表 -->
      <template v-else>
        <div class="list-info">
          <span>共 {{ tasks.length }} 个任务</span>
          <span v-if="hasActive(tasks)" class="live-badge">实时刷新</span>
        </div>

        <div class="task-list">
          <div
            v-for="task in tasks"
            :key="task.id"
            class="task-card glass"
            :class="task.status"
          >
            <div class="task-main">
              <div class="task-title">
                <span class="task-comic">{{ task.comic_name }}</span>
                <span class="task-group">{{ task.group_title || '未分组' }}</span>
              </div>
              <div class="task-chapter">
                <span class="order">{{ task.order ? task.order.toFixed(1) : '' }}</span>
                <span class="chapter-name">{{ task.chapter_title }}</span>
                <span class="format-badge">{{ task.image_format?.toUpperCase() }}</span>
              </div>
            </div>

            <div class="task-side">
              <span class="status-badge" :class="statusMap[task.status]?.cls || ''">
                {{ statusMap[task.status]?.label || task.status }}
              </span>
              <span class="task-time">{{ new Date(task.created_at).toLocaleString() }}</span>
            </div>

            <!-- 进度条 -->
            <div v-if="task.status === 'downloading' || task.status === 'pending'" class="task-progress">
              <div
                class="progress-bar"
                :style="{ width: task.total_pages ? (task.downloaded_pages / task.total_pages * 100) + '%' : '0%' }"
              ></div>
              <span class="progress-text">{{ task.progress }}</span>
            </div>

            <!-- 失败原因 -->
            <div v-if="task.status === 'failed' && task.error" class="task-error">
              <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
              </svg>
              {{ task.error }}
            </div>
          </div>
        </div>
      </template>
    </main>
  </div>
</template>

<style scoped>
.tasks-view {
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding: 0 4px;
}

.list-info span:first-child {
  font-size: 13px;
  color: var(--text-hint);
}

.live-badge {
  font-size: 12px;
  color: var(--primary);
  background: var(--primary-soft);
  padding: 5px 14px;
  border-radius: 999px;
  animation: pulse 1.8s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.55; }
}

/* 任务列表 */
.task-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-card {
  position: relative;
  padding: 20px 28px;
  transition: all 0.3s ease;
  border-left: 3px solid var(--divider);
}

.task-card:hover {
  box-shadow: var(--shadow-hover), var(--shadow-inset);
}

.task-card.downloading {
  border-left-color: var(--primary);
}

.task-card.completed {
  border-left-color: #7CC38F;
}

.task-card.failed {
  border-left-color: #D29AA3;
}

.task-card.pending {
  border-left-color: #D6BCF5;
}

.task-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.task-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.task-comic {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.task-group {
  font-size: 12px;
  color: var(--text-hint);
  background: var(--primary-soft);
  padding: 3px 12px;
  border-radius: 999px;
}

.task-chapter {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}

.order {
  font-size: 12px;
  color: var(--text-hint);
  min-width: 28px;
}

.chapter-name {
  font-size: 14px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.format-badge {
  font-size: 11px;
  color: var(--primary);
  background: var(--primary-soft);
  padding: 3px 10px;
  border-radius: 999px;
  flex-shrink: 0;
}

.task-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  flex-shrink: 0;
}

.status-badge {
  font-size: 12px;
  padding: 4px 14px;
  border-radius: 999px;
  font-weight: 500;
}

.status-badge.pending { background: #F3EDFD; color: #9A6DE0; }
.status-badge.downloading { background: var(--primary-soft); color: var(--primary); }
.status-badge.completed { background: #EAF6EE; color: #4C9E63; }
.status-badge.failed { background: #FBEFF1; color: #C26D7A; }

.task-time {
  font-size: 11px;
  color: var(--text-hint);
}

/* 进度条 */
.task-progress {
  margin-top: 16px;
  position: relative;
  height: 6px;
  background: var(--glass-hi);
  border-radius: 999px;
  overflow: visible;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--primary), var(--accent));
  border-radius: 999px;
  transition: width 0.4s ease;
  min-width: 2px;
}

.progress-text {
  position: absolute;
  top: 10px;
  right: 0;
  font-size: 11px;
  color: var(--text-hint);
}

/* 失败原因 */
.task-error {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #C26D7A;
  background: #FBEFF1;
  padding: 10px 16px;
  border-radius: 12px;
  word-break: break-all;
}

.task-error svg {
  flex-shrink: 0;
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

  .task-card {
    padding: 16px 18px;
  }

  .task-side {
    align-items: flex-start;
  }
}
</style>

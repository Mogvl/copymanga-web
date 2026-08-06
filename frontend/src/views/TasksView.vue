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
        <div class="header-title">⬇️ 下载任务</div>
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
        <div class="empty-icon">📭</div>
        <h3>暂无下载任务</h3>
        <p>去漫画详情页勾选章节即可下载</p>
        <button class="btn" @click="goBack">去搜索</button>
      </div>

      <!-- 任务列表 -->
      <template v-else>
        <div class="list-info">
          <span>共 {{ tasks.length }} 个任务</span>
          <span v-if="hasActive(tasks)" class="live-badge">● 实时刷新</span>
        </div>

        <div class="task-list">
          <div
            v-for="task in tasks"
            :key="task.id"
            class="task-card"
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
}

/* 列表信息 */
.list-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding: 0 4px;
}

.list-info span:first-child {
  font-size: 13px;
  color: var(--text-hint);
}

.live-badge {
  font-size: 12px;
  color: var(--primary);
  background: var(--primary-light);
  padding: 4px 12px;
  border-radius: 12px;
  animation: pulse 1.6s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* 任务列表 */
.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-card {
  position: relative;
  background: var(--card);
  border-radius: var(--radius);
  padding: 16px 20px;
  box-shadow: var(--shadow);
  transition: all 0.3s ease;
  border-left: 4px solid var(--divider);
}

.task-card.downloading {
  border-left-color: var(--primary);
}

.task-card.completed {
  border-left-color: #34A853;
}

.task-card.failed {
  border-left-color: #D32F2F;
}

.task-card.pending {
  border-left-color: #F9A825;
}

.task-main {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.task-title {
  display: flex;
  align-items: center;
  gap: 8px;
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
  background: var(--bg);
  padding: 2px 10px;
  border-radius: 10px;
}

.task-chapter {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
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
  background: var(--primary-light);
  padding: 2px 8px;
  border-radius: 10px;
  flex-shrink: 0;
}

.task-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
  flex-shrink: 0;
}

.status-badge {
  font-size: 12px;
  padding: 3px 12px;
  border-radius: 12px;
  font-weight: 500;
}

.status-badge.pending { background: #FFF8E1; color: #F9A825; }
.status-badge.downloading { background: var(--primary-light); color: var(--primary); }
.status-badge.completed { background: #E6F4EA; color: #34A853; }
.status-badge.failed { background: #FDECEA; color: #D32F2F; }

.task-time {
  font-size: 11px;
  color: var(--text-hint);
}

/* 进度条 */
.task-progress {
  margin-top: 12px;
  position: relative;
  height: 8px;
  background: var(--bg);
  border-radius: 4px;
  overflow: visible;
}

.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, var(--primary), var(--primary-hover));
  border-radius: 4px;
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
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #D32F2F;
  background: #FDECEA;
  padding: 8px 12px;
  border-radius: 8px;
  word-break: break-all;
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

  .task-card {
    padding: 14px 16px;
  }

  .task-side {
    align-items: flex-start;
  }
}
</style>

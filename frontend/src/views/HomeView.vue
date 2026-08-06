<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const keyword = ref('')
const searchFocused = ref(false)

function search() {
  if (keyword.value.trim()) {
    router.push({ name: 'search', query: { q: keyword.value.trim() } })
  }
}
</script>

<template>
  <div class="home">
    <!-- 背景 -->
    <div class="bg">
      <div class="bg-circle bg-circle-1"></div>
      <div class="bg-circle bg-circle-2"></div>
      <div class="bg-circle bg-circle-3"></div>
    </div>

    <!-- 主内容 -->
    <div class="content">
      <!-- Logo 区域 -->
      <div class="logo-section">
        <div class="logo-icon">📚</div>
        <h1 class="logo-title">拷贝漫画下载器</h1>
        <p class="logo-desc">搜索你喜欢的漫画，一键下载到本地</p>
      </div>

      <!-- 搜索框 -->
      <div class="search-card" :class="{ focused: searchFocused }">
        <form class="search-form" @submit.prevent="search">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <path d="m21 21-4.35-4.35"/>
          </svg>
          <input
            v-model="keyword"
            type="text"
            placeholder="输入漫画名称搜索..."
            @focus="searchFocused = true"
            @blur="searchFocused = false"
          />
          <button type="submit" class="search-btn">搜索</button>
        </form>
      </div>

      <!-- 热门搜索 -->
      <div class="hot-section">
        <span class="hot-label">热门搜索</span>
        <div class="hot-tags">
          <span class="tag" @click="keyword = '海贼王'; search()">海贼王</span>
          <span class="tag" @click="keyword = '咒术回战'; search()">咒术回战</span>
          <span class="tag" @click="keyword = '间谍过家家'; search()">间谍过家家</span>
          <span class="tag" @click="keyword = '进击的巨人'; search()">进击的巨人</span>
        </div>
      </div>

      <!-- 功能入口 -->
      <div class="features">
        <router-link to="/tasks" class="feature-card">
          <div class="feature-icon">⬇️</div>
          <div class="feature-info">
            <h3>下载任务</h3>
            <p>查看下载进度和任务状态</p>
          </div>
          <svg class="feature-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 18l6-6-6-6"/>
          </svg>
        </router-link>
        <router-link to="/downloaded" class="feature-card">
          <div class="feature-icon">📥</div>
          <div class="feature-info">
            <h3>已下载</h3>
            <p>查看已下载的漫画</p>
          </div>
          <svg class="feature-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 18l6-6-6-6"/>
          </svg>
        </router-link>
      </div>

      <!-- 底部 -->
      <footer class="footer">
        <p>copymanga-web v2.0.0 · Go + Vue3</p>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.home {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
}

/* 背景渐变 */
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
  opacity: 0.4;
}

.bg-circle-1 {
  width: 600px;
  height: 600px;
  top: -200px;
  right: -100px;
  background: radial-gradient(circle, rgba(255, 105, 0, 0.3) 0%, transparent 70%);
  animation: float1 20s ease-in-out infinite;
}

.bg-circle-2 {
  width: 400px;
  height: 400px;
  bottom: -100px;
  left: -100px;
  background: radial-gradient(circle, rgba(255, 150, 50, 0.25) 0%, transparent 70%);
  animation: float2 25s ease-in-out infinite;
}

.bg-circle-3 {
  width: 300px;
  height: 300px;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: radial-gradient(circle, rgba(255, 200, 100, 0.2) 0%, transparent 70%);
  animation: float3 18s ease-in-out infinite;
}

@keyframes float1 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-30px, 30px); }
}

@keyframes float2 {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, -20px); }
}

@keyframes float3 {
  0%, 100% { transform: translate(-50%, -50%) scale(1); }
  50% { transform: translate(-50%, -50%) scale(1.1); }
}

/* 主内容 */
.content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

/* Logo 区域 */
.logo-section {
  text-align: center;
  margin-bottom: 32px;
}

.logo-icon {
  font-size: 64px;
  margin-bottom: 16px;
  animation: bounce 2s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.logo-title {
  font-size: 36px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
  letter-spacing: -0.5px;
}

.logo-desc {
  font-size: 16px;
  color: var(--text-secondary);
}

/* 搜索框 */
.search-card {
  width: 100%;
  max-width: 560px;
  background: var(--card);
  border-radius: 28px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  margin-bottom: 24px;
}

.search-card.focused {
  transform: scale(1.02);
  box-shadow: 0 8px 32px rgba(255, 105, 0, 0.15);
}

.search-form {
  display: flex;
  align-items: center;
  padding: 6px;
}

.search-icon {
  width: 22px;
  height: 22px;
  margin-left: 18px;
  color: var(--text-hint);
  flex-shrink: 0;
}

.search-form input {
  flex: 1;
  padding: 16px 14px;
  border: none;
  outline: none;
  font-size: 16px;
  background: transparent;
  color: var(--text-primary);
}

.search-form input::placeholder {
  color: var(--text-hint);
}

.search-btn {
  padding: 14px 32px;
  background: var(--primary);
  color: white;
  border: none;
  border-radius: 22px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.search-btn:hover {
  background: var(--primary-hover);
  transform: scale(1.02);
}

.search-btn:active {
  transform: scale(0.98);
}

/* 热门搜索 */
.hot-section {
  text-align: center;
  margin-bottom: 40px;
}

.hot-label {
  font-size: 13px;
  color: var(--text-hint);
  margin-bottom: 12px;
  display: block;
}

.hot-tags {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 10px;
}

.tag {
  padding: 8px 18px;
  background: var(--card);
  border-radius: 20px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.tag:hover {
  color: var(--primary);
  background: white;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 功能入口 */
.features {
  width: 100%;
  max-width: 560px;
  margin-bottom: 40px;
}

.feature-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  background: var(--card);
  border-radius: var(--radius);
  text-decoration: none;
  color: inherit;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.feature-icon {
  font-size: 32px;
  flex-shrink: 0;
}

.feature-info {
  flex: 1;
}

.feature-info h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.feature-info p {
  font-size: 13px;
  color: var(--text-hint);
}

.feature-arrow {
  width: 20px;
  height: 20px;
  color: var(--text-hint);
  flex-shrink: 0;
}

/* 底部 */
.footer {
  text-align: center;
  font-size: 12px;
  color: var(--text-hint);
  opacity: 0.7;
}

/* 移动端适配 */
@media (max-width: 640px) {
  .content {
    padding: 30px 16px;
  }

  .logo-icon {
    font-size: 48px;
  }

  .logo-title {
    font-size: 28px;
  }

  .logo-desc {
    font-size: 14px;
  }

  .search-form input {
    padding: 14px 12px;
    font-size: 15px;
  }

  .search-btn {
    padding: 12px 24px;
    font-size: 15px;
  }

  .hot-tags {
    gap: 8px;
  }

  .tag {
    padding: 6px 14px;
    font-size: 13px;
  }
}
</style>

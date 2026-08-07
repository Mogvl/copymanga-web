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
    <!-- 弥散粉紫背景光环 -->
    <div class="aura aura-1"></div>
    <div class="aura aura-2"></div>
    <div class="aura aura-3"></div>

    <main class="content">
      <!-- Logo 区域 -->
      <div class="logo-section">
        <div class="logo-mark" aria-hidden="true">
          <svg viewBox="0 0 120 120" fill="none">
            <defs>
              <!-- 星球本体渐变 -->
              <radialGradient id="planetCore" cx="35%" cy="32%" r="75%">
                <stop offset="0%" stop-color="#D8B4FE"/>
                <stop offset="38%" stop-color="#A78BFA"/>
                <stop offset="72%" stop-color="#7C4BE0"/>
                <stop offset="100%" stop-color="#5B21B6"/>
              </radialGradient>
              <!-- 大气层弥散光 -->
              <radialGradient id="planetAtmo" cx="50%" cy="50%" r="50%">
                <stop offset="55%" stop-color="#C4B5FD" stop-opacity="0"/>
                <stop offset="100%" stop-color="#A78BFA" stop-opacity="0.35"/>
              </radialGradient>
              <!-- 表面亮纹 -->
              <linearGradient id="planetBand" x1="0" y1="0" x2="0.3" y2="1">
                <stop offset="0%" stop-color="#FFFFFF" stop-opacity="0.28"/>
                <stop offset="50%" stop-color="#FFFFFF" stop-opacity="0.05"/>
                <stop offset="100%" stop-color="#FFFFFF" stop-opacity="0.18"/>
              </linearGradient>
              <!-- 轨道线 -->
              <linearGradient id="ringGlow" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0%" stop-color="#E9D5FF"/>
                <stop offset="100%" stop-color="#C4B5FD"/>
              </linearGradient>
            </defs>

            <!-- 内圈旋转光环 -->
            <circle cx="60" cy="60" r="48" stroke="url(#ringGlow)" stroke-width="1" stroke-dasharray="2 9" stroke-linecap="round" opacity="0.7">
              <animateTransform attributeName="transform" type="rotate" from="0 60 60" to="360 60 60" dur="30s" repeatCount="indefinite"/>
            </circle>

            <!-- 轨道圆点 -->
            <circle cx="102" cy="60" r="3" fill="#E9D5FF">
              <animateMotion dur="14s" repeatCount="indefinite" path="M 60 18 a 42 42 0 1 1 -0.01 0 Z"/>
            </circle>

            <!-- 大气层弥散 -->
            <circle cx="60" cy="60" r="34" fill="url(#planetAtmo)">
              <animate attributeName="opacity" values="0.9;1;0.9" dur="4s" repeatCount="indefinite"/>
            </circle>

            <!-- 星球本体 -->
            <circle cx="60" cy="60" r="30" fill="url(#planetCore)"/>

            <!-- 表面高光弧 -->
            <path d="M42 48 a26 26 0 0 1 22 -12" stroke="#FFFFFF" stroke-opacity="0.5" stroke-width="2.4" stroke-linecap="round" fill="none"/>
            <path d="M40 66 a26 26 0 0 0 22 12" stroke="#5B21B6" stroke-opacity="0.35" stroke-width="1.6" stroke-linecap="round" fill="none"/>

            <!-- 表面云带 -->
            <ellipse cx="60" cy="64" rx="16" ry="4.5" fill="url(#planetBand)" opacity="0.7"/>
            <ellipse cx="60" cy="52" rx="11" ry="3" fill="url(#planetBand)" opacity="0.55"/>

            <!-- 离轴光斑 -->
            <circle cx="50" cy="50" r="6" fill="#FFFFFF" opacity="0.30"/>
            <circle cx="47" cy="47" r="2.2" fill="#FFFFFF" opacity="0.9"/>
          </svg>
        </div>
        <h1 class="logo-title">拷贝漫画下载器</h1>
        <p class="logo-desc">搜索你喜欢的漫画，一键下载到本地</p>
      </div>

      <!-- 搜索框 -->
      <div class="search-card glass" :class="{ focused: searchFocused }">
        <form class="search-form" @submit.prevent="search">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
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
          <button class="tag" @click="keyword = '海贼王'; search()">海贼王</button>
          <button class="tag" @click="keyword = '咒术回战'; search()">咒术回战</button>
          <button class="tag" @click="keyword = '间谍过家家'; search()">间谍过家家</button>
          <button class="tag" @click="keyword = '进击的巨人'; search()">进击的巨人</button>
        </div>
      </div>

      <!-- 功能入口 -->
      <div class="features">
        <router-link to="/tasks" class="feature-card">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7,10 12,15 17,10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          <div class="feature-info">
            <h3>下载任务</h3>
            <p>查看下载进度和任务状态</p>
          </div>
          <svg class="feature-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M9 18l6-6-6-6"/>
          </svg>
        </router-link>
        <router-link to="/downloaded" class="feature-card">
          <svg class="feature-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="3" y="3" width="18" height="18" rx="3"/>
            <path d="M3 9h18"/>
            <path d="M9 21V9"/>
          </svg>
          <div class="feature-info">
            <h3>已下载</h3>
            <p>查看已下载的漫画</p>
          </div>
          <svg class="feature-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M9 18l6-6-6-6"/>
          </svg>
        </router-link>
      </div>

      <!-- 底部 -->
      <footer class="footer">
        <p>copymanga-web v2.0.0</p>
      </footer>
    </main>
  </div>
</template>

<style scoped>
.home {
  min-height: 100dvh;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 背景光环：弥散粉紫，微弱珠光 */
.aura {
  position: fixed;
  border-radius: 50%;
  filter: blur(90px);
  pointer-events: none;
  z-index: 0;
}

.aura-1 {
  width: 560px;
  height: 560px;
  top: -180px;
  right: -120px;
  background: radial-gradient(circle, rgba(216, 180, 254, 0.50) 0%, transparent 70%);
  animation: drift 26s ease-in-out infinite;
}

.aura-2 {
  width: 480px;
  height: 480px;
  bottom: -160px;
  left: -140px;
  background: radial-gradient(circle, rgba(240, 171, 252, 0.38) 0%, transparent 70%);
  animation: drift 32s ease-in-out infinite reverse;
}

.aura-3 {
  width: 360px;
  height: 360px;
  top: 40%;
  left: 55%;
  background: radial-gradient(circle, rgba(196, 181, 253, 0.30) 0%, transparent 70%);
  animation: drift 22s ease-in-out infinite;
}

@keyframes drift {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(-40px, 36px); }
}

.content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 720px;
  padding: 96px 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Logo 区域 */
.logo-section {
  text-align: center;
  margin-bottom: 56px;
}

.logo-mark {
  width: 100px;
  height: 100px;
  margin: 0 auto 32px;
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  animation: floatMark 5s ease-in-out infinite;
  filter: drop-shadow(0 16px 32px rgba(91, 33, 182, 0.30));
}

@keyframes floatMark {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.logo-mark svg {
  width: 100%;
  height: 100%;
  overflow: visible;
}

.logo-title {
  font-size: 34px;
  font-weight: 600;
  letter-spacing: -0.5px;
  color: var(--text-primary);
  margin-bottom: 14px;
}

.logo-desc {
  font-size: 16px;
  color: var(--text-hint);
  font-weight: 400;
}

/* 搜索框：柔雾玻璃 */
.search-card {
  width: 100%;
  max-width: 560px;
  border-radius: 28px;
  margin-bottom: 56px;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.search-card.focused {
  transform: translateY(-1px);
  box-shadow: var(--shadow-hover), var(--shadow-inset);
}

.search-form {
  display: flex;
  align-items: center;
  padding: 8px 8px 8px 20px;
}

.search-icon {
  width: 20px;
  height: 20px;
  color: var(--text-hint);
  flex-shrink: 0;
}

.search-form input {
  flex: 1;
  padding: 16px 16px;
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
  padding: 14px 36px;
  background: linear-gradient(135deg, var(--primary), #A78BFA);
  color: #fff;
  border: none;
  border-radius: 22px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: opacity 0.2s ease, transform 0.2s ease;
  white-space: nowrap;
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.28);
}

.search-btn:hover {
  opacity: 0.92;
  transform: translateY(-1px);
}

.search-btn:active {
  transform: scale(0.98);
}

/* 热门搜索 */
.hot-section {
  text-align: center;
  margin-bottom: 56px;
}

.hot-label {
  font-size: 13px;
  color: var(--text-hint);
  margin-bottom: 20px;
  display: block;
  letter-spacing: 1px;
}

.hot-tags {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 12px;
}

.tag {
  padding: 10px 24px;
  background: var(--glass-hi);
  -webkit-backdrop-filter: blur(12px);
  backdrop-filter: blur(12px);
  border: 1px solid var(--card-brd);
  border-radius: 999px;
  font-size: 14px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
}

.tag:hover {
  color: var(--primary);
  background: #fff;
  border-color: rgba(139, 92, 246, 0.25);
  transform: translateY(-2px);
  box-shadow: var(--shadow-hover);
}

/* 功能入口 */
.features {
  width: 100%;
  max-width: 560px;
  margin-bottom: 64px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.feature-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px 28px;
  background: var(--card);
  border: 1px solid var(--card-brd);
  border-radius: var(--radius);
  -webkit-backdrop-filter: blur(20px) saturate(1.4);
  backdrop-filter: blur(20px) saturate(1.4);
  text-decoration: none;
  color: inherit;
  box-shadow: var(--shadow-sm), var(--shadow-inset);
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-hover), var(--shadow-inset);
  border-color: rgba(139, 92, 246, 0.25);
}

.feature-icon {
  width: 26px;
  height: 26px;
  color: var(--primary);
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
  width: 18px;
  height: 18px;
  color: var(--text-hint);
  flex-shrink: 0;
}

/* 底部 */
.footer {
  text-align: center;
  font-size: 12px;
  color: var(--text-hint);
  opacity: 0.75;
  letter-spacing: 0.5px;
}

/* 移动端适配 */
@media (max-width: 640px) {
  .content {
    padding: 72px 20px;
  }

  .logo-title {
    font-size: 28px;
  }

  .logo-desc {
    font-size: 14px;
  }

  .search-form {
    padding: 6px 6px 6px 16px;
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
    padding: 8px 18px;
    font-size: 13px;
  }
}
</style>

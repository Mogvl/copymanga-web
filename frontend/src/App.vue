<script setup lang="ts">
</script>

<template>
  <router-view v-slot="{ Component }">
    <keep-alive include="SearchView">
      <component :is="Component" />
    </keep-alive>
  </router-view>
</template>

<style>
/* ============ 设计令牌：粉紫柔雾主题 ============ */
:root {
  /* 主色：低饱和粉紫渐变（禁高饱和亮色） */
  --primary: #8B5CF6;
  --primary-hover: #7C4BE0;
  --primary-soft: rgba(139, 92, 246, 0.08);
  --primary-glow: rgba(167, 139, 250, 0.35);
  --accent: #D8B4FE;
  --accent-2: #F0ABFC;

  /* 背景：柔雾粉紫弥散 */
  --bg-start: #F7F3FF;
  --bg-mid: #F5EEFF;
  --bg-end: #FAF4FE;
  --card: rgba(255, 255, 255, 0.62);
  --card-solid: #FFFFFF;
  --card-brd: rgba(255, 255, 255, 0.85);
  --glass-hi: rgba(255, 255, 255, 0.7);
  --glass-lo: rgba(255, 255, 255, 0.25);

  /* 文本：克制层级 */
  --text-primary: #2B2238;
  --text-secondary: #6E6382;
  --text-hint: #A79CB8;
  --divider: rgba(139, 92, 246, 0.10);
  --border: rgba(139, 92, 246, 0.14);

  /* 阴影：轻薄弥散 + 微弱珠光，禁刺眼强阴影 */
  --shadow-sm: 0 1px 2px rgba(87, 62, 132, 0.04), 0 4px 16px rgba(139, 92, 246, 0.06);
  --shadow: 0 2px 8px rgba(87, 62, 132, 0.05), 0 16px 40px rgba(139, 92, 246, 0.10);
  --shadow-hover: 0 4px 12px rgba(87, 62, 132, 0.06), 0 24px 56px rgba(139, 92, 246, 0.16);
  --shadow-inset: inset 0 1px 0 rgba(255, 255, 255, 0.7);

  /* 圆角：规整大圆角 */
  --radius: 24px;
  --radius-sm: 16px;
  --radius-xs: 12px;

  /* 间距：宽松留白 */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 24px;
  --space-6: 32px;
  --space-7: 48px;
  --space-8: 64px;

  /* 字体 */
  --font-sans: -apple-system, BlinkMacSystemFont, "SF Pro Display", "Segoe UI", "PingFang SC", "Microsoft YaHei", Roboto, sans-serif;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: var(--font-sans);
  background: linear-gradient(135deg, var(--bg-start) 0%, var(--bg-mid) 45%, var(--bg-end) 100%);
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.6;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  min-height: 100dvh;
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(139, 92, 246, 0.20);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(139, 92, 246, 0.35);
}

/* 柔雾玻璃拟态卡片 */
.glass {
  background: linear-gradient(135deg, var(--glass-hi), var(--glass-lo));
  -webkit-backdrop-filter: blur(20px) saturate(1.4);
  backdrop-filter: blur(20px) saturate(1.4);
  border: 1px solid var(--card-brd);
  box-shadow: var(--shadow), var(--shadow-inset);
  border-radius: var(--radius);
}

/* 弥散粉紫背景光环 */
.aura {
  position: fixed;
  border-radius: 50%;
  filter: blur(90px);
  pointer-events: none;
  z-index: 0;
}
</style>

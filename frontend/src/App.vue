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
:root {
  --primary: #FF6900;
  --primary-hover: #E55D00;
  --primary-light: rgba(255, 105, 0, 0.1);
  --bg: #F8F9FA;
  --card: #FFFFFF;
  --text-primary: #1A1A1A;
  --text-secondary: #555555;
  --text-hint: #999999;
  --divider: #EEEEEE;
  --border: #E8E8E8;
  --shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  --shadow-hover: 0 8px 24px rgba(0, 0, 0, 0.12);
  --radius: 16px;
  --radius-sm: 12px;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, "SF Pro Display", "Segoe UI", Roboto, sans-serif;
  background: var(--bg);
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.5;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #D0D0D0;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #C0C0C0;
}
</style>

<template>
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup>
import { onMounted } from 'vue'

// 应用初始化时恢复主题设置
onMounted(() => {
  const savedTheme = localStorage.getItem('aurora-theme')
  const followSystem = localStorage.getItem('aurora-theme-follow-system')
  const root = document.documentElement

  // 恢复主题
  if (followSystem === 'true') {
    const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
    if (systemDark) {
      root.setAttribute('data-theme', 'dark')
    } else {
      root.removeAttribute('data-theme')
    }
  } else if (savedTheme === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else {
    root.removeAttribute('data-theme')
  }

  // 恢复强调色
  const savedAccent = localStorage.getItem('aurora-accent-color')
  if (savedAccent) {
    root.style.setProperty('--color-primary', savedAccent)
  }

  console.log('[App] 主题已恢复:', savedTheme || 'light', '跟随系统:', followSystem)
})
</script>

<style>
/* 引入极客风设计系统变量 */
@import './styles/variables.css';
/* 引入极客风组件样式 */
@import './styles/components.css';
/* 引入极客风动画效果 */
@import './styles/geek-animations.css';
/* 引入极客风管理页面样式 */
@import './styles/geek-admin.css';

/* 全局过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1));
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 基础样式重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html,
body,
#app {
  height: 100%;
  font-family: var(--font-sans, 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif);
  background: var(--bg-deep, #0F172A);
  color: var(--text-primary, #F8FAFC);
}

/* 滚动条样式 - 极客风 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-thumb {
  background: var(--border-default, #475569);
  border-radius: 4px;
  transition: background 0.15s ease;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--primary, #3B82F6);
}

::-webkit-scrollbar-track {
  background: var(--bg-deep, #0F172A);
  border-radius: 4px;
}

/* 选中文字样式 */
::selection {
  background: var(--primary, #3B82F6);
  color: #fff;
}
</style>

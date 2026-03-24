<template>
  <a href="#main-content" class="skip-link">跳到主要内容</a>
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

/* 跳过导航链接 - 可访问性 */
.skip-link {
  position: absolute;
  top: -100%;
  left: 16px;
  z-index: 10000;
  padding: 8px 16px;
  background: var(--primary, #3B82F6);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  border-radius: 0 0 8px 8px;
  text-decoration: none;
  transition: top 0.2s ease;
}

.skip-link:focus {
  top: 0;
  outline: 2px solid var(--primary);
  outline-offset: 2px;
}

/* 屏幕阅读器专用 - 视觉隐藏但可被辅助技术读取 */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

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
  touch-action: manipulation;
}

/* 消除移动端 300ms 点击延迟 */
a,
button,
[role="button"],
.el-button,
.el-menu-item,
.el-sub-menu__title,
.el-dropdown,
.el-tabs__item,
.el-tag,
.el-input__wrapper,
.el-select,
.el-date-editor,
.el-switch,
.el-checkbox,
.el-radio {
  touch-action: manipulation;
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

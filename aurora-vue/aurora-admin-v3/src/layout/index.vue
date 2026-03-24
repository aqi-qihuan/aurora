<template>
  <div class="layout-wrapper">
    <SideBar ref="sidebar" @menu-clicked="handleMenuClicked" />
    <div class="main-container" :class="{ 'hideSideBar': appStore.collapse }">
      <div class="header-wrapper">
        <NavBar @toggle-mobile-sidebar="toggleMobileSidebar" />
      </div>
      <main id="main-content" class="main-content">
        <div class="fade-transform-box">
          <router-view v-slot="{ Component, route: currentRoute }">
            <keep-alive :include="cachedViews">
              <component :is="Component" :key="currentRoute.path" v-if="currentRoute.meta?.keepAlive !== false" />
            </keep-alive>
            <component :is="Component" :key="currentRoute.path" v-if="currentRoute.meta?.keepAlive === false" />
          </router-view>
        </div>
      </main>
    </div>
    <div v-if="isMobile && mobileSidebarVisible" class="sidebar-overlay" role="presentation" @click="toggleMobileSidebar" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import NavBar from '@/layout/components/NavBar.vue'
import SideBar from '@/layout/components/SideBar.vue'
import { useAppStore } from '@/stores/app'

const route = useRoute()
const appStore = useAppStore()

const isMobile = ref(false)
const mobileSidebarVisible = ref(false)
const sidebar = ref(null)

// 缓存已打开的标签页组件名
const cachedViews = computed(() => {
  return appStore.tabList
    .filter(tab => tab.name)
    .map(tab => tab.name)
})

const checkMobile = () => {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) {
    mobileSidebarVisible.value = false
  }
}

const toggleMobileSidebar = () => {
  if (isMobile.value) {
    mobileSidebarVisible.value = !mobileSidebarVisible.value
    const sidebarEl = sidebar.value?.$el
    if (sidebarEl) {
      if (mobileSidebarVisible.value) {
        sidebarEl.classList.add('mobile-visible')
      } else {
        sidebarEl.classList.remove('mobile-visible')
      }
    }
  }
}

const handleMenuClicked = () => {
  if (isMobile.value) {
    mobileSidebarVisible.value = false
    const sidebarEl = sidebar.value?.$el
    if (sidebarEl) {
      sidebarEl.classList.remove('mobile-visible')
    }
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped>
/* 布局容器 */
.layout-wrapper {
  min-height: 100vh;
  background: var(--bg-deep, #0F172A);
}

/* 主容器 - 添加左边距为侧边栏留空间 */
.main-container {
  margin-left: 210px;
  min-height: 100vh;
  transition: margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 折叠状态 */
.hideSideBar {
  margin-left: 64px;
}

/* 头部包装器 */
.header-wrapper {
  position: sticky;
  top: 0;
  z-index: 999;
  background: var(--bg-base, #1B2336);
}

/* 主内容区域 */
.main-content {
  background: var(--bg-deep, #0F172A);
  min-height: calc(100vh - 96px);
}

/* 移动端适配 */
@media (max-width: 768px) {
  .main-container,
  .hideSideBar {
    margin-left: 0 !important;
  }
}

/* 移动端侧边栏遮罩 */
.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  animation: fadeIn 0.3s ease;
  backdrop-filter: blur(2px);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* 优化页面切换动画 - 使用更流畅的渐变效果 */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-transform-enter {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.98);
}

.fade-transform-box {
  position: relative;
  top: 0px;
  bottom: 0px;
  width: 100%;
  overflow: hidden;
  padding: 20px;
  z-index: 1;
}

@media (max-width: 768px) {
  .fade-transform-box {
    padding: 12px;
  }
}
</style>

<template>
  <div>
    <div class="nav-bar">
      <div class="left-menu">
        <div class="hambuger-container" @click="trigger">
          <el-icon :size="20">
            <component :is="appStore.collapse ? 'Expand' : 'Fold'" />
          </el-icon>
        </div>
        <el-breadcrumb separator="/">
          <el-breadcrumb-item v-for="item of breadcrumbs" :key="item.path">
            <span v-if="item.redirect">{{ item.name }}</span>
            <router-link v-else :to="item.path">{{ item.name }}</router-link>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <div class="right-menu">
        <div class="screen-full" @click="fullScreen" title="全屏">
          <el-icon :size="18"><FullScreen /></el-icon>
        </div>
        <div class="theme-toggle" @click="showThemeSettings" title="主题设置">
          <el-icon :size="18"><Operation /></el-icon>
        </div>
        <el-dropdown @command="handleCommand">
          <div class="user-dropdown">
            <el-avatar :size="36" :src="userStore.userInfo.avatar" />
            <span class="user-name">{{ userStore.userInfo?.nickname || 'Admin' }}</span>
            <el-icon :size="12"><CaretBottom /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="setting">
                <el-icon><User /></el-icon>
                <span>个人中心</span>
              </el-dropdown-item>
              <el-dropdown-item command="logout" divided>
                <el-icon><SwitchButton /></el-icon>
                <span>退出登录</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>
    <div class="tabs-view-container">
      <div class="tabs-wrapper">
        <span 
          :class="isActive(item)" 
          v-for="item of appStore.tabList" 
          :key="item.path" 
          @click="goTo(item)">
          {{ item.name }}
          <el-icon 
            v-if="item.path != '/home'" 
            class="el-icon-close" 
            @click.stop="removeTab(item)">
            <Close />
          </el-icon>
        </span>
      </div>
      <div class="tabs-close-item" @click="closeAllTab">全部关闭</div>
    </div>
    <ThemeSettings ref="themeSettingsRef" v-model:visible="themeSettingsVisible" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Fold, Expand, FullScreen, Operation, CaretBottom,
  User, SwitchButton, Close
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import ThemeSettings from '@/components/ThemeSettings.vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const appStore = useAppStore()

const themeSettingsVisible = ref(false)
const themeSettingsRef = ref(null)
const fullscreen = ref(false)

// 面包屑导航
const breadcrumbs = computed(() => {
  let matched = route.matched.filter((item) => item.name)
  const first = matched[0]
  if (first && first.name !== '首页') {
    matched = [{ path: '/home', name: '首页' }].concat(matched)
  }
  return matched
})

// 监听路由变化，添加标签
watch(
  () => route.path,
  (newPath) => {
    if (newPath && newPath !== '/login') {
      appStore.saveTab(route)
    }
  },
  { immediate: true }
)

const goTo = (tab) => {
  router.push({ path: tab.path })
}

const removeTab = (tab) => {
  // 如果关闭的是当前页面，需要先跳转再移除
  if (tab.path === route.path) {
    const tabList = appStore.tabList
    const currentIndex = tabList.findIndex(t => t.path === tab.path)
    
    // 移除标签
    appStore.removeTab(tab)
    
    // 获取更新后的列表
    const newTabList = appStore.tabList
    
    // 如果还有标签，跳转到前一个或后一个
    if (newTabList.length > 0) {
      // 优先跳转到前一个标签
      const targetIndex = Math.min(currentIndex, newTabList.length - 1)
      const targetTab = newTabList[Math.max(0, targetIndex - 1)]
      router.push({ path: targetTab.path })
    } else {
      // 如果没有标签了，跳转到首页
      router.push({ path: '/home' })
    }
  } else {
    // 关闭的不是当前页面，直接移除
    appStore.removeTab(tab)
  }
}

const trigger = () => {
  // 检测是否为移动端
  if (window.innerWidth < 768) {
    // 移动端触发侧边栏切换事件
    emit('toggle-mobile-sidebar')
  } else {
    // 桌面端切换折叠状态
    appStore.toggleCollapse()
  }
}

const handleCommand = (command) => {
  if (command == 'setting') {
    router.push({ path: '/setting' })
  }
  if (command == 'logout') {
    request.post('/users/logout').then(({ data }) => {
      userStore.logout()
      appStore.resetTab()
      router.push({ path: '/login' })
    })
  }
}

const closeAllTab = () => {
  appStore.resetTab()
  router.push({ path: '/home' })
}

const fullScreen = () => {
  let element = document.documentElement
  if (fullscreen.value) {
    if (document.exitFullscreen) {
      document.exitFullscreen()
    } else if (document.webkitCancelFullScreen) {
      document.webkitCancelFullScreen()
    } else if (document.mozCancelFullScreen) {
      document.mozCancelFullScreen()
    } else if (document.msExitFullscreen) {
      document.msExitFullscreen()
    }
  } else {
    if (element.requestFullscreen) {
      element.requestFullscreen()
    } else if (element.webkitRequestFullScreen) {
      element.webkitRequestFullScreen()
    } else if (element.mozRequestFullScreen) {
      element.webkitRequestFullScreen()
    } else if (element.msRequestFullscreen) {
      element.msRequestFullscreen()
    }
  }
  fullscreen.value = !fullscreen.value
}

const showThemeSettings = () => {
  themeSettingsVisible.value = true
  nextTick(() => {
    if (themeSettingsRef.value && themeSettingsRef.value.open) {
      themeSettingsRef.value.open()
    }
  })
}

const isActive = (tab) => {
  return tab.path == route.path ? 'tabs-view-item-active' : 'tabs-view-item'
}

const emit = defineEmits(['toggle-mobile-sidebar'])
</script>

<style scoped>
/* ===== 极客风导航栏 ===== */
.nav-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 56px;
  background: var(--bg-elevated);
  border-bottom: 2px solid var(--primary);
  box-shadow: 0 2px 12px rgba(59, 130, 246, 0.1);
  transition: all 0.3s ease;
}

[data-theme="dark"] .nav-bar {
  box-shadow: 0 2px 12px rgba(59, 130, 246, 0.2),
              0 0 20px rgba(0, 212, 255, 0.1);
}

.left-menu {
  display: flex;
  align-items: center;
  gap: 20px;
  flex: 1;
}

.hambuger-container {
  font-size: 20px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.25s ease;
  padding: 8px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hambuger-container:hover {
  background-color: var(--primary-light);
  color: var(--primary);
  transform: rotate(90deg);
}

[data-theme="dark"] .hambuger-container:hover {
  background-color: rgba(0, 212, 255, 0.15);
  color: var(--neon-blue);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.el-breadcrumb {
  flex: 1;
}

::deep(.el-breadcrumb__inner),
::deep(.el-breadcrumb__separator) {
  color: var(--text-secondary);
  font-weight: 400;
}

::deep(.el-breadcrumb__inner a),
::deep(.el-breadcrumb__inner.is-link) {
  color: var(--primary);
  transition: all 0.2s ease;
}

::deep(.el-breadcrumb__inner a:hover),
::deep(.el-breadcrumb__inner.is-link:hover) {
  color: var(--primary-hover);
}

[data-theme="dark"] ::deep(.el-breadcrumb__inner a:hover),
[data-theme="dark"] ::deep(.el-breadcrumb__inner.is-link:hover) {
  color: var(--neon-blue);
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.4);
}

.right-menu {
  display: flex;
  align-items: center;
  gap: 8px;
}

.screen-full,
.theme-toggle {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-secondary);
  border-radius: 8px;
  transition: all 0.25s ease;
  background: transparent;
  border: none;
}

.screen-full:hover,
.theme-toggle:hover {
  background-color: var(--primary-light);
  color: var(--primary);
  transform: scale(1.1);
}

[data-theme="dark"] .screen-full:hover,
[data-theme="dark"] .theme-toggle:hover {
  background-color: rgba(0, 212, 255, 0.15);
  color: var(--neon-blue);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 8px;
  transition: all 0.25s ease;
}

.user-dropdown:hover {
  background-color: var(--primary-light);
}

[data-theme="dark"] .user-dropdown:hover {
  background-color: rgba(0, 212, 255, 0.15);
}

.user-name {
  font-size: 14px;
  color: var(--text-primary);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tabs-view-container {
  display: flex;
  align-items: center;
  position: relative;
  padding: 0 20px;
  height: 40px;
  background: var(--bg-deep);
  border-bottom: 1px solid var(--border-light);
}

[data-theme="dark"] .tabs-view-container {
  border-bottom: 1px solid var(--primary);
}

.tabs-wrapper {
  flex: 1;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tabs-view-item {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  height: 28px;
  line-height: 28px;
  border: 1px solid var(--border-light);
  border-radius: 6px;
  color: var(--text-secondary);
  background: var(--bg-base);
  padding: 0 12px;
  font-size: 12px;
  transition: all 0.25s ease;
}

.tabs-view-item:hover {
  color: var(--primary);
  border-color: var(--primary);
  transform: translateY(-1px);
}

[data-theme="dark"] .tabs-view-item:hover {
  border-color: var(--neon-blue);
  color: var(--neon-blue);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.3);
}

.tabs-view-item-active {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  height: 30px;
  line-height: 30px;
  padding: 0 12px;
  font-size: 12px;
  border-radius: 6px;
  background: var(--primary-light);
  color: var(--primary);
  border: 1px solid var(--primary);
  font-weight: 600;
}

[data-theme="dark"] .tabs-view-item-active {
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.2) 0%, rgba(191, 90, 242, 0.2) 100%);
  color: var(--neon-blue);
  border: 1px solid var(--neon-blue);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.3),
              inset 0 0 10px rgba(0, 212, 255, 0.1);
}

.tabs-view-item-active:hover {
  transform: translateY(-1px);
}

[data-theme="dark"] .tabs-view-item-active:hover {
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.5),
              inset 0 0 15px rgba(0, 212, 255, 0.15);
}

[data-theme="dark"] .tabs-view-item-active:before {
  content: '';
  background: var(--neon-blue);
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
  box-shadow: 0 0 6px var(--neon-blue);
  animation: pulse-glow 2s ease-in-out infinite;
}

@keyframes pulse-glow {
  0%, 100% { box-shadow: 0 0 4px var(--neon-blue); }
  50% { box-shadow: 0 0 10px var(--neon-blue); }
}

.tabs-close-item {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  height: 28px;
  line-height: 28px;
  border: 1px solid var(--border-light);
  border-radius: 6px;
  color: var(--text-secondary);
  background: var(--bg-base);
  padding: 0 12px;
  font-size: 12px;
  transition: all 0.25s ease;
}

.tabs-close-item:hover {
  color: var(--danger);
  border-color: var(--danger);
  transform: translateY(-1px);
}

[data-theme="dark"] .tabs-close-item:hover {
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.3);
}

.el-icon-close {
  margin-left: 6px;
  padding: 2px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.el-icon-close:hover {
  background: rgba(239, 68, 68, 0.2);
  color: var(--danger);
}

*::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

*::-webkit-scrollbar-thumb {
  border-radius: 3px;
  background-color: var(--border-default);
  transition: background 0.15s ease;
}

*::-webkit-scrollbar-thumb:hover {
  background-color: var(--primary);
}

[data-theme="dark"] *::-webkit-scrollbar-thumb:hover {
  background-color: var(--neon-blue);
}

@media (max-width: 768px) {
  .nav-bar {
    padding: 0 12px;
    height: 52px;
  }

  .el-breadcrumb {
    display: none;
  }

  .user-name {
    display: none;
  }

  .tabs-view-container {
    padding: 0 12px;
    height: 36px;
  }

  .tabs-view-item,
  .tabs-view-item-active {
    padding: 0 8px;
    font-size: 11px;
  }
}
</style>

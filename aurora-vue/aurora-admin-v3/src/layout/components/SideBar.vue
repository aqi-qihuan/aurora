<template>
  <div class="sidebar-container" :class="{ 'is-collapsed': appStore.collapse }">
    <div class="sidebar-header">
      <div class="logo">
        <el-icon :size="24"><Platform /></el-icon>
        <span>Aurora Admin</span>
      </div>
    </div>
    <el-menu
      class="side-nav-bar"
      router
      :collapse="appStore.collapse"
      :default-active="route.path"
      :collapse-transition="false"
      background-color="transparent"
      text-color="var(--text-primary)"
      active-text-color="var(--primary)"
      @select="handleMenuSelect"
      :default-openeds="defaultOpeneds"
      :key="menuKey">
      <template v-for="routeItem of userStore.userMenus" :key="routeItem.path">
        <template v-if="routeItem.name && routeItem.children && !routeItem.hidden">
          <el-sub-menu :index="routeItem.path" popper-class="sidebar-popper">
            <template #title>
              <el-icon class="menu-icon">
                <component :is="getIcon(routeItem.icon)" />
              </el-icon>
              <span>{{ routeItem.name }}</span>
            </template>
            <template v-for="item in routeItem.children" :key="item.path || item.name">
              <el-menu-item v-if="!item.hidden" :index="item.path">
                <el-icon class="menu-icon">
                  <component :is="getIcon(item.icon)" />
                </el-icon>
                <span>{{ item.name }}</span>
              </el-menu-item>
            </template>
          </el-sub-menu>
        </template>
        <template v-else-if="!routeItem.hidden && routeItem.children">
          <el-menu-item :index="routeItem.path" :key="routeItem.path">
            <el-icon class="menu-icon">
              <component :is="getIcon(routeItem.children[0].icon)" />
            </el-icon>
            <span>{{ routeItem.children[0].name }}</span>
          </el-menu-item>
        </template>
      </template>
    </el-menu>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAppStore } from '@/stores/app'
import { Platform, House, Document, User, Setting, ChatDotRound, Picture, Timer, Folder, Collection } from '@element-plus/icons-vue'

const emit = defineEmits(['menu-clicked'])
const route = useRoute()
const userStore = useUserStore()
const appStore = useAppStore()

const defaultOpeneds = ref([])

// 菜单 key，当 userMenus 变化时强制 el-menu 重新渲染
const menuKey = computed(() => userStore.userMenus.length)

const iconMap = {
  'el-icon-s-platform': Platform,
  'el-icon-s-home': House,
  'el-icon-document': Document,
  'el-icon-user': User,
  'el-icon-setting': Setting,
  'el-icon-chat-dot-round': ChatDotRound,
  'el-icon-picture': Picture,
  'el-icon-timer': Timer,
  'el-icon-folder': Folder,
  'el-icon-collection': Collection,
  'el-icon-house': House,
  'el-icon-s-home': House,
  'el-icon-s-tools': Setting,
  'el-icon-s-goods': Collection,
  'el-icon-s-order': Document,
  'el-icon-s-management': Folder,
  'el-icon-s-custom': User,
  'el-icon-s-operation': Setting,
  'el-icon-s-marketing': ChatDotRound,
  'el-icon-s-data': Document,
  'el-icon-menu': Document
}

const getIcon = (iconName) => {
  return iconMap[iconName] || Document
}

const handleMenuSelect = () => {
  // 移动端点击菜单项后关闭侧边栏
  if (window.innerWidth < 768) {
    setTimeout(() => {
      emit('menu-clicked')
    }, 100)
  }
}
</script>

<style scoped>
/* ===== 侧边栏 ===== */
.sidebar-container {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 210px;
  background: var(--bg-base);
  border-right: 1px solid var(--border-default);
  box-shadow: var(--shadow-lg);
  transition: width 0.3s cubic-bezier(0.16, 1, 0.3, 1), background 0.3s ease, border-color 0.3s ease;
  z-index: 1001;
  overflow-x: hidden;
  overflow-y: auto;
}

/* 深色主题特殊样式 */
[data-theme="dark"] .sidebar-container {
  border-right: 1px solid rgba(59, 130, 246, 0.4);
  box-shadow: 2px 0 15px rgba(59, 130, 246, 0.15);
}

/* 折叠状态 */
.sidebar-container.is-collapsed {
  width: 64px;
}

/* ===== Logo 区域 ===== */
.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border-bottom: 1px solid var(--border-default);
  background: var(--bg-elevated);
  position: relative;
  overflow: hidden;
  transition: background 0.3s ease, border-color 0.3s ease;
}

/* 深色主题特殊样式 */
[data-theme="dark"] .sidebar-header {
  background: linear-gradient(135deg, #1E293B 0%, #334155 50%, #1E293B 100%);
  background-size: 200% 200%;
  animation: gradient-shift 8s ease infinite;
}

/* Logo区域发光效果（仅深色主题） */
[data-theme="dark"] .sidebar-header::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--primary), transparent);
  opacity: 0.5;
}

@keyframes gradient-shift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

/* 未折叠时显示文字 */
.sidebar-container:not(.is-collapsed) .sidebar-header {
  justify-content: flex-start;
  padding: 0 20px;
}

.sidebar-container:not(.is-collapsed) .logo {
  gap: 12px;
}

.logo {
  display: flex;
  align-items: center;
  color: var(--text-primary);
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.5px;
  transition: all 0.3s ease;
  width: 100%;
  position: relative;
}

/* Logo图标 */
.logo .el-icon {
  flex-shrink: 0;
  color: var(--primary);
  transition: color 0.3s ease;
}

/* 深色主题Logo图标发光效果 */
[data-theme="dark"] .logo .el-icon {
  color: var(--neon-blue);
  filter: drop-shadow(0 0 6px rgba(0, 212, 255, 0.4));
  will-change: filter;
}

[data-theme="dark"] .sidebar-container:hover .logo .el-icon {
  filter: drop-shadow(0 0 10px rgba(0, 212, 255, 0.5));
}

.logo span {
  white-space: nowrap;
  transition: opacity 0.3s ease, width 0.3s ease;
  overflow: hidden;
  color: var(--text-primary);
}

/* 深色主题Logo文字渐变 */
[data-theme="dark"] .logo span {
  background: linear-gradient(90deg, var(--text-primary), var(--neon-blue));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 折叠时隐藏文字 */
.sidebar-container.is-collapsed .logo span {
  opacity: 0;
  width: 0;
}

/* ===== 菜单样式 ===== */
.side-nav-bar:not(.el-menu--collapse) {
  width: 210px;
  border: none;
}

.side-nav-bar {
  border: none;
  padding: 12px 8px;
}

.side-nav-bar .menu-icon {
  margin-right: 12px;
  font-size: 18px;
  color: var(--primary);
  transition: all 0.25s ease;
}

/* 深色主题菜单图标 */
[data-theme="dark"] .side-nav-bar .menu-icon {
  color: var(--neon-blue, #00D4FF);
}

/* ===== 菜单项 - 基础样式 ===== */
::deep(.el-menu-item),
::deep(.el-sub-menu__title) {
  border-radius: 8px;
  margin: 4px 0;
  transition: all 0.25s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1));
  height: 48px;
  line-height: 48px;
  -webkit-tap-highlight-color: transparent;
  cursor: pointer;
  color: var(--text-primary);
}

::deep(.el-menu-item:hover),
::deep(.el-sub-menu__title:hover) {
  background-color: var(--primary-light) !important;
  transform: translateX(4px);
}

::deep(.el-menu-item:hover) .menu-icon,
::deep(.el-sub-menu__title:hover) .menu-icon {
  color: var(--primary, #3B82F6);
}

/* ===== 激活状态 ===== */
::deep(.el-menu-item.is-active) {
  background: var(--primary-light);
  color: var(--primary);
  border: 1px solid var(--border-glow);
  box-shadow: var(--shadow-glow);
  position: relative;
}

/* 激活项左侧发光条 */
::deep(.el-menu-item.is-active)::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: var(--primary);
  border-radius: 0 2px 2px 0;
}

::deep(.el-menu-item.is-active) .menu-icon {
  color: var(--primary);
}

/* ===== 深色主题激活状态 - 霓虹发光 ===== */
[data-theme="dark"] ::deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.12) 0%, rgba(139, 92, 246, 0.12) 100%);
  color: var(--neon-blue, #00D4FF);
  border: 1px solid rgba(59, 130, 246, 0.2);
  box-shadow: 0 0 10px rgba(59, 130, 246, 0.2);
}

[data-theme="dark"] ::deep(.el-menu-item.is-active)::before {
  background: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 8px var(--neon-blue, #00D4FF);
}

[data-theme="dark"] ::deep(.el-menu-item.is-active) .menu-icon {
  color: var(--neon-blue, #00D4FF);
  filter: drop-shadow(0 0 6px rgba(0, 212, 255, 0.5));
}

/* ===== 深色主题悬停效果 ===== */
[data-theme="dark"] ::deep(.el-menu-item:hover) .menu-icon,
[data-theme="dark"] ::deep(.el-sub-menu__title:hover) .menu-icon {
  filter: drop-shadow(0 0 4px rgba(59, 130, 246, 0.4));
}

::deep(.el-sub-menu .el-menu-item) {
  padding-left: 56px !important;
}

/* ===== 折叠状态 ===== */
.side-nav-bar.el-menu--collapse {
  width: 64px;
}

.side-nav-bar.el-menu--collapse :deep(.el-menu-item),
.side-nav-bar.el-menu--collapse :deep(.el-sub-menu__title) {
  padding: 0 20px;
  justify-content: center;
}

.side-nav-bar.el-menu--collapse :deep(.el-menu-item) .menu-icon,
.side-nav-bar.el-menu--collapse :deep(.el-sub-menu__title) .menu-icon {
  margin-right: 0;
}

.side-nav-bar.el-menu--collapse :deep(.el-menu-item) span,
.side-nav-bar.el-menu--collapse :deep(.el-sub-menu__title) span {
  display: none;
}

/* ===== 滚动条 ===== */
*::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

*::-webkit-scrollbar-thumb {
  border-radius: 3px;
  background-color: var(--border-default, #475569);
  transition: background 0.15s ease;
}

*::-webkit-scrollbar-thumb:hover {
  background-color: var(--primary, #3B82F6);
}

/* ===== 移动端适配 ===== */
@media (max-width: 768px) {
  .sidebar-container {
    transform: translateX(-100%);
    width: 280px !important;
  }

  .sidebar-container.mobile-visible {
    transform: translateX(0);
  }

  .sidebar-container.is-collapsed {
    width: 280px !important;
  }

  .sidebar-header {
    height: 56px;
  }

  .logo {
    font-size: 16px;
  }

  .logo .el-icon {
    font-size: 22px;
  }

  .side-nav-bar {
    padding: 10px 6px;
  }

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    height: 44px;
    line-height: 44px;
  }
}
</style>

<style>
/* ===== 下拉菜单 - 基础样式 ===== */
.sidebar-popper {
  background: var(--bg-elevated) !important;
  border: 1px solid var(--border-default) !important;
  border-radius: 12px !important;
  box-shadow: var(--shadow-xl) !important;
  padding: 8px !important;
}

.sidebar-popper.el-menu--popup {
  min-width: 160px;
}

.sidebar-popper .el-menu-item {
  border-radius: 8px;
  height: 40px;
  line-height: 40px;
  margin: 4px 0;
  color: var(--text-primary);
  transition: all 0.25s ease;
}

.sidebar-popper .el-menu-item:hover {
  background-color: var(--primary-light) !important;
  color: var(--primary, #3B82F6);
}

.sidebar-popper .el-menu-item.is-active {
  background: var(--primary-light) !important;
  color: var(--primary) !important;
  border: 1px solid var(--border-glow);
}

/* 深色主题下拉菜单 */
[data-theme="dark"] .sidebar-popper {
  background: var(--bg-elevated, #272F42) !important;
  border: 1px solid var(--border-default, #475569) !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5) !important;
}

[data-theme="dark"] .sidebar-popper .el-menu-item.is-active {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(139, 92, 246, 0.15) 100%) !important;
  color: var(--neon-blue, #00D4FF) !important;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

/* 移动端下拉菜单优化 */
@media (max-width: 768px) {
  .sidebar-popper.el-menu--popup {
    min-width: 140px;
  }

  .sidebar-popper .el-menu-item {
    height: 36px;
    line-height: 36px;
    font-size: 13px;
  }
}
</style>

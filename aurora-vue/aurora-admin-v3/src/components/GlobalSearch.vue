<template>
  <div class="global-search">
    <el-input
      v-model="searchQuery"
      placeholder="全局搜索..."
      :prefix-icon="Search"
      clearable
      @focus="showResults = true"
      @blur="handleBlur"
      @input="handleSearch"
      @keyup.enter="handleEnter"
      ref="searchInput"
    >
      <template #append>
        <kbd class="keyboard-shortcut">Ctrl K</kbd>
      </template>
    </el-input>

    <transition name="fade-in-down">
      <div v-show="showResults && searchQuery" class="search-results" @mousedown.prevent>
        <div class="search-results-header">
          <span class="results-count">找到 {{ filteredResults.length }} 个结果</span>
          <el-button type="primary" link size="small" @click="clearSearch">
            <el-icon><Close /></el-icon> 清除
          </el-button>
        </div>

        <div v-if="filteredResults.length === 0" class="no-results">
          <el-icon :size="32"><Search /></el-icon>
          <p>未找到相关结果</p>
          <span>试试其他关键词</span>
        </div>

        <div v-else class="results-list">
          <div
            v-for="(item, index) in filteredResults"
            :key="item.path"
            class="result-item"
            :class="{ active: selectedIndex === index }"
            @click="navigateTo(item)"
            @mouseenter="selectedIndex = index"
          >
            <div class="result-icon" :style="{ backgroundColor: item.color + '20', color: item.color }">
              <el-icon :size="20"><component :is="item.iconComponent" /></el-icon>
            </div>
            <div class="result-content">
              <div class="result-title" v-html="highlightText(item.title)"></div>
              <div class="result-description" v-html="highlightText(item.description)"></div>
              <div class="result-path">
                <el-breadcrumb separator="/">
                  <el-breadcrumb-item v-for="(crumb, i) in item.breadcrumbs" :key="i">
                    {{ crumb }}
                  </el-breadcrumb-item>
                </el-breadcrumb>
              </div>
            </div>
            <div class="result-action">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>
        </div>

        <div class="search-results-footer">
          <div class="shortcut-hint">
            <kbd>↑</kbd>
            <kbd>↓</kbd>
            <span>选择</span>
          </div>
          <div class="shortcut-hint">
            <kbd>Enter</kbd>
            <span>打开</span>
          </div>
          <div class="shortcut-hint">
            <kbd>Esc</kbd>
            <span>关闭</span>
          </div>
        </div>
      </div>
    </transition>

    <transition name="fade-in-down">
      <div v-show="showResults && !searchQuery" class="search-shortcuts" @mousedown.prevent>
        <div class="shortcuts-section">
          <div class="shortcuts-title">最近访问</div>
          <div class="shortcuts-list">
            <div
              v-for="(item, index) in recentItems"
              :key="item.path"
              class="shortcut-item"
              :class="{ active: selectedIndex === index }"
              @click="navigateTo(item)"
              @mouseenter="selectedIndex = index"
            >
              <el-icon><component :is="item.iconComponent" /></el-icon>
              <span>{{ item.title }}</span>
            </div>
          </div>
        </div>

        <div class="shortcuts-section">
          <div class="shortcuts-title">快捷导航</div>
          <div class="shortcuts-grid">
            <div
              v-for="item in quickNavItems"
              :key="item.path"
              class="quick-nav-item"
              @click="navigateTo(item)"
            >
              <div class="quick-nav-icon" :style="{ backgroundColor: item.color + '20', color: item.color }">
                <el-icon :size="24"><component :is="item.iconComponent" /></el-icon>
              </div>
              <span>{{ item.title }}</span>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search, Close, ArrowRight, Document, Edit, FolderOpened, CollectionTag,
  User, UserFilled, Menu, Operation, Timer, Notebook, Warning,
  ChatDotRound, ChatLineRound, Picture, Link, House, Tools, Setting as SettingIcon
} from '@element-plus/icons-vue'

const router = useRouter()

const searchQuery = ref('')
const showResults = ref(false)
const selectedIndex = ref(0)
const recentItems = ref([])
const searchInput = ref(null)

const searchItems = [
  {
    title: '文章列表',
    description: '查看和管理所有文章',
    path: '/articles',
    icon: 'el-icon-document',
    iconComponent: Document,
    color: '#2563EB',
    breadcrumbs: ['文章', '文章列表']
  },
  {
    title: '写文章',
    description: '创建新文章',
    path: '/articles/write',
    icon: 'el-icon-edit',
    iconComponent: Edit,
    color: '#10B981',
    breadcrumbs: ['文章', '写文章']
  },
  {
    title: '分类管理',
    description: '管理文章分类',
    path: '/categories',
    icon: 'el-icon-folder-opened',
    iconComponent: FolderOpened,
    color: '#F59E0B',
    breadcrumbs: ['分类', '分类管理']
  },
  {
    title: '标签管理',
    description: '管理文章标签',
    path: '/tags',
    icon: 'el-icon-collection-tag',
    iconComponent: CollectionTag,
    color: '#8B5CF6',
    breadcrumbs: ['标签', '标签管理']
  },
  {
    title: '用户列表',
    description: '查看和管理用户',
    path: '/users',
    icon: 'el-icon-user',
    iconComponent: User,
    color: '#EC4899',
    breadcrumbs: ['用户', '用户列表']
  },
  {
    title: '在线用户',
    description: '查看在线用户',
    path: '/users/online',
    icon: 'el-icon-user-solid',
    iconComponent: UserFilled,
    color: '#06B6D4',
    breadcrumbs: ['用户', '在线用户']
  },
  {
    title: '菜单管理',
    description: '配置系统菜单',
    path: '/menus',
    icon: 'el-icon-menu',
    iconComponent: Menu,
    color: '#F97316',
    breadcrumbs: ['系统', '菜单管理']
  },
  {
    title: '角色管理',
    description: '配置用户角色',
    path: '/roles',
    icon: 'el-icon-s-custom',
    iconComponent: Operation,
    color: '#6366F1',
    breadcrumbs: ['系统', '角色管理']
  },
  {
    title: '定时任务',
    description: '管理定时任务',
    path: '/quartz',
    icon: 'el-icon-time',
    iconComponent: Timer,
    color: '#14B8A6',
    breadcrumbs: ['系统', '定时任务']
  },
  {
    title: '操作日志',
    description: '查看操作记录',
    path: '/logs/operation',
    icon: 'el-icon-notebook-1',
    iconComponent: Notebook,
    color: '#64748B',
    breadcrumbs: ['日志', '操作日志']
  },
  {
    title: '异常日志',
    description: '查看异常记录',
    path: '/logs/exception',
    icon: 'el-icon-warning',
    iconComponent: Warning,
    color: '#EF4444',
    breadcrumbs: ['日志', '异常日志']
  },
  {
    title: '评论管理',
    description: '管理文章评论',
    path: '/comments',
    icon: 'el-icon-chat-dot-round',
    iconComponent: ChatDotRound,
    color: '#3B82F6',
    breadcrumbs: ['内容', '评论管理']
  },
  {
    title: '说说管理',
    description: '管理说说动态',
    path: '/talks',
    icon: 'el-icon-chat-line-round',
    iconComponent: ChatLineRound,
    color: '#8B5CF6',
    breadcrumbs: ['内容', '说说管理']
  },
  {
    title: '相册管理',
    description: '管理图片相册',
    path: '/albums',
    icon: 'el-icon-picture',
    iconComponent: Picture,
    color: '#EC4899',
    breadcrumbs: ['内容', '相册管理']
  },
  {
    title: '友链管理',
    description: '管理友情链接',
    path: '/links',
    icon: 'el-icon-link',
    iconComponent: Link,
    color: '#10B981',
    breadcrumbs: ['内容', '友链管理']
  },
  {
    title: '网站信息',
    description: '配置网站基本信息',
    path: '/website',
    icon: 'el-icon-s-home',
    iconComponent: House,
    color: '#2563EB',
    breadcrumbs: ['设置', '网站信息']
  },
  {
    title: '系统设置',
    description: '配置系统参数',
    path: '/settings',
    icon: 'el-icon-s-tools',
    iconComponent: SettingIcon,
    color: '#64748B',
    breadcrumbs: ['设置', '系统设置']
  }
]

const filteredResults = computed(() => {
  if (!searchQuery.value) return []
  const query = searchQuery.value.toLowerCase()
  return searchItems.filter(item =>
    item.title.toLowerCase().includes(query) ||
    item.description.toLowerCase().includes(query) ||
    item.breadcrumbs.some(crumb => crumb.toLowerCase().includes(query))
  )
})

const quickNavItems = computed(() => {
  return [
    searchItems[0],
    searchItems[4],
    searchItems[12],
    searchItems[15],
    searchItems[1],
    searchItems[9]
  ]
})

const handleSearch = () => {
  selectedIndex.value = 0
}

const handleBlur = () => {
  setTimeout(() => {
    showResults.value = false
    searchQuery.value = ''
    selectedIndex.value = 0
  }, 200)
}

const handleEnter = () => {
  const results = searchQuery.value ? filteredResults.value : recentItems.value
  if (results[selectedIndex.value]) {
    navigateTo(results[selectedIndex.value])
  }
}

const handleKeydown = (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    showResults.value = true
    nextTick(() => {
      searchInput.value.focus()
    })
  }

  if (e.key === 'Escape' && showResults.value) {
    showResults.value = false
    searchInput.value.blur()
  }

  if (!showResults.value) return

  const results = searchQuery.value ? filteredResults.value : recentItems.value

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = (selectedIndex.value + 1) % results.length
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = (selectedIndex.value - 1 + results.length) % results.length
  }
}

const navigateTo = (item) => {
  saveToRecent(item)
  router.push(item.path)
  showResults.value = false
  searchQuery.value = ''
  selectedIndex.value = 0
}

const clearSearch = () => {
  searchQuery.value = ''
  searchInput.value.focus()
}

const highlightText = (text) => {
  if (!searchQuery.value) return text
  // XSS 防护：转义特殊字符
  const escapeHtml = (str) => {
    const div = document.createElement('div')
    div.textContent = str
    return div.innerHTML
  }
  const escapedText = escapeHtml(text)
  const escapedQuery = escapeHtml(searchQuery.value)
  const regex = new RegExp(`(${escapedQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi')
  return escapedText.replace(regex, '<mark>$1</mark>')
}

const loadRecentItems = () => {
  const recent = localStorage.getItem('aurora-admin-recent-nav')
  if (recent) {
    const paths = JSON.parse(recent).slice(0, 5)
    recentItems.value = paths.map(path =>
      searchItems.find(item => item.path === path)
    ).filter(Boolean)
  }
}

const saveToRecent = (item) => {
  let recent = JSON.parse(localStorage.getItem('aurora-admin-recent-nav') || '[]')
  recent = recent.filter(path => path !== item.path)
  recent.unshift(item.path)
  recent = recent.slice(0, 10)
  localStorage.setItem('aurora-admin-recent-nav', JSON.stringify(recent))
}

onMounted(() => {
  loadRecentItems()
  document.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.global-search {
  position: relative;
  width: 280px;
}

.global-search :deep(.el-input__inner) {
  border-radius: 8px;
  background-color: var(--el-fill-color-light, #f5f7fa);
  border-color: transparent;
  padding-left: 40px;
  height: 40px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.global-search :deep(.el-input__inner:focus) {
  background-color: #fff;
  border-color: var(--el-color-primary, #409eff);
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.global-search :deep(.el-input__prefix) {
  left: 12px;
  font-size: 16px;
  color: var(--el-text-color-secondary, #909399);
}

.global-search :deep(.el-input-group__append) {
  background-color: transparent;
  border-color: transparent;
  padding: 0 12px;
}

.keyboard-shortcut {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 8px;
  font-size: 12px;
  font-family: monospace;
  color: var(--el-text-color-secondary, #909399);
  background-color: var(--el-fill-color-light, #f5f7fa);
  border: 1px solid var(--el-border-color, #dcdfe6);
  border-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.search-results,
.search-shortcuts {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  right: 0;
  background-color: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  border: 1px solid var(--el-border-color, #dcdfe6);
  z-index: 2000;
  max-height: 480px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color, #dcdfe6);
}

.results-count {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  font-weight: 500;
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: var(--el-text-color-secondary, #909399);
}

.no-results i {
  margin-bottom: 12px;
  color: var(--el-text-color-placeholder, #c0c4cc);
}

.no-results p {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary, #303133);
  margin: 0 0 4px;
}

.no-results span {
  font-size: 13px;
}

.results-list {
  overflow-y: auto;
  max-height: 320px;
  padding: 8px;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.result-item:hover,
.result-item.active {
  background-color: var(--el-fill-color-light, #f5f7fa);
}

.result-item.active {
  background-color: rgba(64, 158, 255, 0.1);
}

.result-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.result-content {
  flex: 1;
  min-width: 0;
}

.result-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary, #303133);
  margin-bottom: 4px;
}

.result-title :deep(mark) {
  background-color: rgba(64, 158, 255, 0.3);
  color: var(--el-color-primary, #409eff);
  padding: 0 2px;
  border-radius: 2px;
}

.result-description {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  margin-bottom: 4px;
}

.result-description :deep(mark) {
  background-color: rgba(64, 158, 255, 0.3);
  color: var(--el-color-primary, #409eff);
  padding: 0 2px;
  border-radius: 2px;
}

.result-path {
  font-size: 12px;
}

.result-path :deep(.el-breadcrumb__inner) {
  color: var(--el-text-color-secondary, #909399);
  font-weight: normal;
}

.result-action {
  color: var(--el-text-color-secondary, #909399);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.result-item:hover .result-action,
.result-item.active .result-action {
  opacity: 1;
  color: var(--el-color-primary, #409eff);
}

.search-results-footer {
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  border-top: 1px solid var(--el-border-color, #dcdfe6);
  background-color: var(--el-fill-color-blank, #ffffff);
}

.shortcut-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
}

.shortcut-hint kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  padding: 2px 4px;
  font-family: monospace;
  font-size: 10px;
  color: var(--el-text-color-regular, #606266);
  background-color: #fff;
  border: 1px solid var(--el-border-color, #dcdfe6);
  border-radius: 4px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.search-shortcuts {
  padding: 16px;
}

.shortcuts-section {
  margin-bottom: 20px;
}

.shortcuts-section:last-child {
  margin-bottom: 0;
}

.shortcuts-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary, #909399);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 12px;
}

.shortcuts-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.shortcut-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  color: var(--el-text-color-primary, #303133);
}

.shortcut-item:hover,
.shortcut-item.active {
  background-color: var(--el-fill-color-light, #f5f7fa);
}

.shortcut-item i {
  font-size: 16px;
  color: var(--el-text-color-regular, #606266);
}

.shortcuts-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.quick-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-nav-item:hover {
  background-color: var(--el-fill-color-light, #f5f7fa);
}

.quick-nav-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.quick-nav-item span {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  text-align: center;
}

.fade-in-down-enter-active,
.fade-in-down-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.fade-in-down-enter,
.fade-in-down-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

@media (max-width: 768px) {
  .global-search {
    width: 200px;
  }

  .keyboard-shortcut {
    display: none;
  }

  .search-results,
  .search-shortcuts {
    position: fixed;
    top: 60px;
    left: 16px;
    right: 16px;
    max-height: calc(100vh - 100px);
  }

  .shortcuts-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .search-results-footer {
    display: none;
  }
}
</style>

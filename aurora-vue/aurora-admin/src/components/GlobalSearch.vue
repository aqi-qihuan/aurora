<template>
  <div class="global-search">
    <el-input
      v-model="searchQuery"
      placeholder="全局搜索..."
      prefix-icon="el-icon-search"
      clearable
      @focus="showResults = true"
      @blur="handleBlur"
      @input="handleSearch"
      @keyup.enter.native="handleEnter"
      ref="searchInput"
    >
      <template slot="append">
        <kbd class="keyboard-shortcut">Ctrl K</kbd>
      </template>
    </el-input>

    <transition name="fade-in-down">
      <div v-show="showResults && searchQuery" class="search-results" @mousedown.prevent>
        <div class="search-results-header">
          <span class="results-count">找到 {{ filteredResults.length }} 个结果</span>
          <el-button type="text" size="mini" @click="clearSearch">
            <i class="el-icon-close"></i> 清除
          </el-button>
        </div>

        <div v-if="filteredResults.length === 0" class="no-results">
          <i class="el-icon-search" style="font-size: 32px;"></i>
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
              <i :class="item.icon" style="font-size: 20px;"></i>
            </div>
            <div class="result-content">
              <div class="result-title" v-html="highlightText(item.title)"></div>
              <div class="result-description" v-html="highlightText(item.description)"></div>
            </div>
            <div class="result-action">
              <i class="el-icon-arrow-right"></i>
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
              <i :class="item.icon"></i>
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
                <i :class="item.icon" style="font-size: 24px;"></i>
              </div>
              <span>{{ item.title }}</span>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  name: 'GlobalSearch',
  data() {
    return {
      searchQuery: '',
      showResults: false,
      selectedIndex: 0,
      recentItems: [],
      searchItems: [
        {
          title: '文章列表',
          description: '查看和管理所有文章',
          path: '/articles',
          icon: 'el-icon-document',
          color: '#2563EB'
        },
        {
          title: '写文章',
          description: '创建新文章',
          path: '/articles/write',
          icon: 'el-icon-edit',
          color: '#10B981'
        },
        {
          title: '分类管理',
          description: '管理文章分类',
          path: '/categories',
          icon: 'el-icon-folder-opened',
          color: '#F59E0B'
        },
        {
          title: '标签管理',
          description: '管理文章标签',
          path: '/tags',
          icon: 'el-icon-collection-tag',
          color: '#8B5CF6'
        },
        {
          title: '用户列表',
          description: '查看和管理用户',
          path: '/users',
          icon: 'el-icon-user',
          color: '#EC4899'
        },
        {
          title: '在线用户',
          description: '查看在线用户',
          path: '/users/online',
          icon: 'el-icon-user-solid',
          color: '#06B6D4'
        },
        {
          title: '菜单管理',
          description: '配置系统菜单',
          path: '/menus',
          icon: 'el-icon-menu',
          color: '#F97316'
        },
        {
          title: '角色管理',
          description: '配置用户角色',
          path: '/roles',
          icon: 'el-icon-s-custom',
          color: '#6366F1'
        },
        {
          title: '定时任务',
          description: '管理定时任务',
          path: '/quartz',
          icon: 'el-icon-time',
          color: '#14B8A6'
        },
        {
          title: '操作日志',
          description: '查看操作记录',
          path: '/logs/operation',
          icon: 'el-icon-notebook-1',
          color: '#64748B'
        },
        {
          title: '异常日志',
          description: '查看异常记录',
          path: '/logs/exception',
          icon: 'el-icon-warning',
          color: '#EF4444'
        },
        {
          title: '评论管理',
          description: '管理文章评论',
          path: '/comments',
          icon: 'el-icon-chat-dot-round',
          color: '#3B82F6'
        },
        {
          title: '说说管理',
          description: '管理说说动态',
          path: '/talks',
          icon: 'el-icon-chat-line-round',
          color: '#8B5CF6'
        },
        {
          title: '相册管理',
          description: '管理图片相册',
          path: '/albums',
          icon: 'el-icon-picture',
          color: '#EC4899'
        },
        {
          title: '友链管理',
          description: '管理友情链接',
          path: '/links',
          icon: 'el-icon-link',
          color: '#10B981'
        },
        {
          title: '网站信息',
          description: '配置网站基本信息',
          path: '/website',
          icon: 'el-icon-s-home',
          color: '#2563EB'
        },
        {
          title: '系统设置',
          description: '配置系统参数',
          path: '/settings',
          icon: 'el-icon-s-tools',
          color: '#64748B'
        }
      ]
    }
  },
  computed: {
    filteredResults() {
      if (!this.searchQuery) return []
      const query = this.searchQuery.toLowerCase()
      return this.searchItems.filter(item =>
        item.title.toLowerCase().includes(query) ||
        item.description.toLowerCase().includes(query)
      )
    },
    quickNavItems() {
      return [
        this.searchItems[0],
        this.searchItems[4],
        this.searchItems[12],
        this.searchItems[15],
        this.searchItems[1],
        this.searchItems[9]
      ]
    }
  },
  methods: {
    handleSearch() {
      this.selectedIndex = 0
    },
    handleBlur() {
      setTimeout(() => {
        this.showResults = false
        this.searchQuery = ''
        this.selectedIndex = 0
      }, 200)
    },
    handleEnter() {
      const results = this.searchQuery ? this.filteredResults : this.recentItems
      if (results[this.selectedIndex]) {
        this.navigateTo(results[this.selectedIndex])
      }
    },
    handleKeydown(e) {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault()
        this.showResults = true
        this.$nextTick(() => {
          this.$refs.searchInput.focus()
        })
      }

      if (e.key === 'Escape' && this.showResults) {
        this.showResults = false
        this.$refs.searchInput.blur()
      }

      if (!this.showResults) return

      const results = this.searchQuery ? this.filteredResults : this.recentItems

      if (e.key === 'ArrowDown') {
        e.preventDefault()
        this.selectedIndex = (this.selectedIndex + 1) % results.length
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        this.selectedIndex = (this.selectedIndex - 1 + results.length) % results.length
      }
    },
    navigateTo(item) {
      this.saveToRecent(item)
      this.$router.push(item.path)
      this.showResults = false
      this.searchQuery = ''
      this.selectedIndex = 0
    },
    clearSearch() {
      this.searchQuery = ''
      this.$refs.searchInput.focus()
    },
    highlightText(text) {
      if (!this.searchQuery) return text
      const regex = new RegExp(`(${this.searchQuery})`, 'gi')
      return text.replace(regex, '<mark>$1</mark>')
    },
    loadRecentItems() {
      const recent = localStorage.getItem('aurora-admin-recent-nav')
      if (recent) {
        const paths = JSON.parse(recent).slice(0, 5)
        this.recentItems = paths.map(path =>
          this.searchItems.find(item => item.path === path)
        ).filter(Boolean)
      }
    },
    saveToRecent(item) {
      let recent = JSON.parse(localStorage.getItem('aurora-admin-recent-nav') || '[]')
      recent = recent.filter(path => path !== item.path)
      recent.unshift(item.path)
      recent = recent.slice(0, 10)
      localStorage.setItem('aurora-admin-recent-nav', JSON.stringify(recent))
    }
  },
  mounted() {
    this.loadRecentItems()
    document.addEventListener('keydown', this.handleKeydown)
  },
  beforeDestroy() {
    document.removeEventListener('keydown', this.handleKeydown)
  }
}
</script>

<style scoped>
.global-search {
  position: relative;
  width: 280px;
}

.global-search >>> .el-input__inner {
  border-radius: 8px;
  background-color: #f5f7fa;
  border-color: transparent;
  padding-left: 40px;
  height: 40px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.global-search >>> .el-input__inner:focus {
  background-color: #fff;
  border-color: #409eff;
  box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1);
}

.global-search >>> .el-input__prefix {
  left: 12px;
  font-size: 16px;
  color: #909399;
}

.global-search >>> .el-input-group__append {
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
  color: #909399;
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
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
  border: 1px solid #dcdfe6;
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
  border-bottom: 1px solid #dcdfe6;
}

.results-count {
  font-size: 12px;
  color: #909399;
  font-weight: 500;
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #909399;
}

.no-results i {
  margin-bottom: 12px;
  color: #c0c4cc;
}

.no-results p {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
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
  background-color: #f5f7fa;
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
  color: #303133;
  margin-bottom: 4px;
}

.result-title >>> mark {
  background-color: rgba(64, 158, 255, 0.3);
  color: #409eff;
  padding: 0 2px;
  border-radius: 2px;
}

.result-description {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.result-description >>> mark {
  background-color: rgba(64, 158, 255, 0.3);
  color: #409eff;
  padding: 0 2px;
  border-radius: 2px;
}

.result-action {
  color: #909399;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.result-item:hover .result-action,
.result-item.active .result-action {
  opacity: 1;
  color: #409eff;
}

.search-results-footer {
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  border-top: 1px solid #dcdfe6;
  background-color: #ffffff;
}

.shortcut-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #909399;
}

.shortcut-hint kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  padding: 2px 4px;
  font-family: monospace;
  font-size: 10px;
  color: #606266;
  background-color: #fff;
  border: 1px solid #dcdfe6;
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
  color: #909399;
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
  color: #303133;
}

.shortcut-item:hover,
.shortcut-item.active {
  background-color: #f5f7fa;
}

.shortcut-item i {
  font-size: 16px;
  color: #606266;
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
  background-color: #f5f7fa;
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
  color: #909399;
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

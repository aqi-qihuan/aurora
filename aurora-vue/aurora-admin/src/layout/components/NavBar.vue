<template>
  <div>
    <div class="nav-bar">
      <div class="left-menu">
        <div class="hambuger-container" @click="trigger">
          <i :class="isFold" />
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
          <i class="iconfont el-icon-myicwindowzoom48px" />
        </div>
        <div class="theme-toggle" @click="showThemeSettings" title="主题设置">
          <i class="el-icon-s-operation" />
        </div>
        <el-dropdown @command="handleCommand">
          <div class="user-dropdown">
            <el-avatar :size="36" :src="this.$store.state.userInfo.avatar" />
            <span class="user-name">{{ this.$store.state.userInfo?.nickname || 'Admin' }}</span>
            <i class="el-icon-caret-bottom" />
          </div>
          <el-dropdown-menu slot="dropdown">
            <el-dropdown-item command="setting">
              <i class="el-icon-s-custom" />
              <span>个人中心</span>
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <i class="iconfont el-icon-mytuichu" />
              <span>退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </el-dropdown>
      </div>
    </div>
    <div class="tabs-view-container">
      <div class="tabs-wrapper">
        <span :class="isActive(item)" v-for="item of this.$store.state.tabList" :key="item.path" @click="goTo(item)">
          {{ item.name }}
          <i class="el-icon-close" v-if="item.path != '/'" @click.stop="removeTab(item)" />
        </span>
      </div>
      <div class="tabs-close-item" @click="closeAllTab">全部关闭</div>
    </div>
    <theme-settings ref="themeSettings" :visible.sync="themeSettingsVisible" />
  </div>
</template>

<script>
import { resetRouter } from '@/router'
import ThemeSettings from '@/components/ThemeSettings.vue'
export default {
  components: {
    ThemeSettings
  },
  created() {
    let matched = this.$route.matched.filter((item) => item.name)
    const first = matched[0]
    if (first && first.name !== '首页') {
      matched = [{ path: '/', name: '首页' }].concat(matched)
    }
    this.breadcrumbs = matched
    this.$store.commit('saveTab', this.$route)

    // 初始化快捷键监听
    this.initShortcutKeys()
  },
  data: function () {
    return {
      isSearch: false,
      fullscreen: false,
      breadcrumbs: [],
      themeSettingsVisible: false
    }
  },
  methods: {
    goTo(tab) {
      this.$router.push({ path: tab.path })
    },
    removeTab(tab) {
      this.$store.commit('removeTab', tab)
      if (tab.path == this.$route.path) {
        var tabList = this.$store.state.tabList
        this.$router.push({ path: tabList[tabList.length - 1].path })
      }
    },
    trigger() {
      // 检测是否为移动端
      if (window.innerWidth < 768) {
        // 移动端触发侧边栏切换事件
        this.$emit('toggle-mobile-sidebar')
      } else {
        // 桌面端切换折叠状态
        this.$store.commit('trigger')
      }
    },
    handleCommand(command) {
      if (command == 'setting') {
        this.$router.push({ path: '/setting' })
      }
      if (command == 'logout') {
        this.axios.post('/api/users/logout').then(({ data }) => {
          this.$store.commit('logout')
          this.$store.commit('resetTab')
          resetRouter()
          this.$router.push({ path: '/login' })
        })
      }
    },
    closeAllTab() {
      this.$store.commit('resetTab')
      this.$router.push({ path: '/' })
    },
    fullScreen() {
      let element = document.documentElement
      if (this.fullscreen) {
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
      this.fullscreen = !this.fullscreen
    },
    showThemeSettings() {
      this.themeSettingsVisible = true
      // 确保子组件能正确响应
      this.$nextTick(() => {
        const themeSettings = this.$refs.themeSettings
        if (themeSettings && themeSettings.open) {
          themeSettings.open()
        }
      })
    },
    initShortcutKeys() {
      document.addEventListener('keydown', this.handleKeyDown)
    },
    handleKeyDown(e) {
      // Ctrl/Cmd + S 保存
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault()
        this.$message.info('快捷键: Ctrl/Cmd + S - 保存功能')
      }

      // Ctrl/Cmd + Shift + T 切换主题
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'T') {
        e.preventDefault()
        this.showThemeSettings()
        this.$message.success('已打开主题设置面板')
      }

      // Ctrl/Cmd + Shift + F 全屏
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'F') {
        e.preventDefault()
        this.fullScreen()
      }

      // Esc 关闭模态框
      if (e.key === 'Escape') {
        if (this.themeSettingsVisible) {
          this.themeSettingsVisible = false
        }
      }
    }
  },
  beforeDestroy() {
    // 移除快捷键监听
    document.removeEventListener('keydown', this.handleKeyDown)
  },
  computed: {
    isActive() {
      return function (tab) {
        if (tab.path == this.$route.path) {
          return 'tabs-view-item-active'
        }
        return 'tabs-view-item'
      }
    },
    isFold() {
      return this.$store.state.collapse ? 'el-icon-s-unfold' : 'el-icon-s-fold'
    }
  }
}
</script>

<style scoped>
.nav-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 56px;
  background: #fff;
  border-bottom: 1px solid var(--border-light, #e4e7ed);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
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
  color: var(--text-regular, #606266);
  transition: all 0.3s ease;
  padding: 8px;
  border-radius: 8px;
}

.hambuger-container:hover {
  background-color: rgba(64, 158, 255, 0.08);
  color: var(--primary-color, #409EFF);
  transform: rotate(90deg);
}

.el-breadcrumb {
  flex: 1;
}

.el-breadcrumb /deep/ .el-breadcrumb__inner {
  color: var(--text-secondary, #909399);
  font-weight: 500;
}

.el-breadcrumb /deep/ .el-breadcrumb__inner:hover {
  color: var(--primary-color, #409EFF);
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
  color: var(--text-regular, #606266);
  font-size: 18px;
  border-radius: 8px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: transparent;
  border: none;
}

.screen-full:hover,
.theme-toggle:hover {
  background-color: rgba(64, 158, 255, 0.08);
  color: var(--primary-color, #409EFF);
  transform: scale(1.1);
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.user-dropdown:hover {
  background-color: rgba(64, 158, 255, 0.08);
}

.user-name {
  font-size: 14px;
  color: var(--text-regular, #606266);
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.el-icon-caret-bottom {
  font-size: 12px;
  color: var(--text-secondary, #909399);
}

.tabs-view-container {
  display: flex;
  align-items: center;
  position: relative;
  padding: 0 20px;
  height: 40px;
  background: #fff;
  border-bottom: 1px solid var(--border-lighter, #ebeef5);
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
  border: 1px solid var(--border-light, #e4e7ed);
  border-radius: 6px;
  color: var(--text-regular, #606266);
  background: #fff;
  padding: 0 12px;
  font-size: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.tabs-view-item:hover {
  color: var(--primary-color, #409EFF);
  border-color: var(--primary-color, #409EFF);
  transform: translateY(-1px);
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.tabs-view-item-active:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.tabs-view-item-active:before {
  content: '';
  background: #fff;
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
}

.tabs-close-item {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  height: 28px;
  line-height: 28px;
  border: 1px solid var(--border-light, #e4e7ed);
  border-radius: 6px;
  color: var(--text-regular, #606266);
  background: #fff;
  padding: 0 12px;
  font-size: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.tabs-close-item:hover {
  color: var(--primary-color, #409EFF);
  border-color: var(--primary-color, #409EFF);
  transform: translateY(-1px);
}

.el-icon-close {
  margin-left: 6px;
  padding: 2px;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.el-icon-close:hover {
  background: rgba(255, 255, 255, 0.3);
}

*::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

*::-webkit-scrollbar-thumb {
  border-radius: 3px;
  background-color: rgba(144, 147, 153, 0.2);
}

*::-webkit-scrollbar-thumb:hover {
  background-color: rgba(144, 147, 153, 0.4);
}

/* 暗色主题适配 */
:root[data-theme="dark"] .nav-bar {
  background: #1a1a1a;
  border-bottom-color: #2b2b2c;
}

:root[data-theme="dark"] .hambuger-container:hover,
:root[data-theme="dark"] .screen-full:hover,
:root[data-theme="dark"] .theme-toggle:hover,
:root[data-theme="dark"] .user-dropdown:hover {
  background-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme="dark"] .el-breadcrumb /deep/ .el-breadcrumb__inner {
  color: #a3a6ad;
}

:root[data-theme="dark"] .tabs-view-container {
  background: #141414;
  border-bottom-color: #2b2b2c;
}

:root[data-theme="dark"] .tabs-view-item,
:root[data-theme="dark"] .tabs-close-item {
  background: #1a1a1a;
  border-color: #363637;
  color: #cfd3dc;
}

:root[data-theme="dark"] .tabs-view-item:hover,
:root[data-theme="dark"] .tabs-close-item:hover {
  border-color: #4c4d4f;
}

:root[data-theme="dark"] .user-name {
  color: #cfd3dc;
}

/* 移动端适配 */
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

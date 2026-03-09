<template>
  <el-container>
    <el-aside width="auto" style="position: static; width: auto !important;">
      <SideBar ref="sidebar" @menu-clicked="handleMenuClicked" />
    </el-aside>
    <el-container :class="'main-container ' + isHide">
      <el-header height="96px" style="padding: 0">
        <NavBar :key="$route.fullPath" @toggle-mobile-sidebar="toggleMobileSidebar" />
      </el-header>
      <el-main style="background: #f7f9fb">
        <div class="fade-transform-box">
          <transition name="fade-transform" mode="out-in">
            <router-view :key="$route.fullPath" />
          </transition>
        </div>
      </el-main>
    </el-container>
    <div v-if="isMobile && mobileSidebarVisible" class="sidebar-overlay" @click="toggleMobileSidebar" />
  </el-container>
</template>

<script>
import NavBar from '@/layout/components/NavBar.vue'
import SideBar from '@/layout/components/SideBar.vue'
export default {
  components: {
    NavBar,
    SideBar
  },
  data() {
    return {
      isMobile: false,
      mobileSidebarVisible: false
    }
  },
  mounted() {
    this.checkMobile()
    window.addEventListener('resize', this.checkMobile)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.checkMobile)
  },
  methods: {
    checkMobile() {
      this.isMobile = window.innerWidth < 768
      if (!this.isMobile) {
        this.mobileSidebarVisible = false
      }
    },
    toggleMobileSidebar() {
      if (this.isMobile) {
        this.mobileSidebarVisible = !this.mobileSidebarVisible
        const sidebar = this.$refs.sidebar?.$el
        if (sidebar) {
          if (this.mobileSidebarVisible) {
            sidebar.classList.add('mobile-visible')
          } else {
            sidebar.classList.remove('mobile-visible')
          }
        }
      }
    },
    handleMenuClicked() {
      if (this.isMobile) {
        this.mobileSidebarVisible = false
        const sidebar = this.$refs.sidebar?.$el
        if (sidebar) {
          sidebar.classList.remove('mobile-visible')
        }
      }
    }
  },
  computed: {
    isHide() {
      return this.$store.state.collapse ? 'hideSideBar' : ''
    }
  }
}
</script>

<style scoped>
.main-container {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-left: 210px;
  min-height: 100vh;
  position: relative;
}

.hideSideBar {
  margin-left: 64px;
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

<template>
  <div class="sidebar-container" :class="{ 'is-collapsed': this.$store.state.collapse }">
    <div class="sidebar-header">
      <div class="logo">
        <i class="el-icon-s-platform" />
        <span>Aurora Admin</span>
      </div>
    </div>
    <el-menu
      class="side-nav-bar"
      router
      :collapse="this.$store.state.collapse"
      :default-active="this.$route.path"
      :collapse-transition="false"
      background-color="transparent"
      text-color="var(--text-regular, #606266)"
      active-text-color="var(--primary-color, #409EFF)"
      @select="handleMenuSelect"
      :default-openeds="defaultOpeneds">
      <template v-for="route of this.$store.state.userMenus">
        <template v-if="route.name && route.children && !route.hidden">
          <el-submenu :key="route.path" :index="route.path" popper-class="sidebar-popper">
            <template slot="title">
              <i :class="route.icon" class="menu-icon" />
              <span>{{ route.name }}</span>
            </template>
            <template v-for="(item, index) of route.children">
              <el-menu-item v-if="!item.hidden" :key="index" :index="item.path">
                <i :class="item.icon" class="menu-icon" />
                <span slot="title">{{ item.name }}</span>
              </el-menu-item>
            </template>
          </el-submenu>
        </template>
        <template v-else-if="!route.hidden">
          <el-menu-item :index="route.path" :key="route.path">
            <i :class="route.children[0].icon" class="menu-icon" />
            <span slot="title">{{ route.children[0].name }}</span>
          </el-menu-item>
        </template>
      </template>
    </el-menu>
  </div>
</template>

<script>
export default {
  data() {
    return {
      defaultOpeneds: []
    }
  },
  methods: {
    handleMenuSelect(index) {
      // 移动端点击菜单项后关闭侧边栏
      if (window.innerWidth < 768) {
        // 延迟关闭侧边栏，确保路由跳转完成
        setTimeout(() => {
          this.$emit('menu-clicked')
        }, 100)
      }
    }
  }
}
</script>

<style scoped>
.sidebar-container {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 210px;
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%);
  border-right: 1px solid var(--border-light, #e4e7ed);
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.04);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 1001;
  overflow: hidden;
}

/* 折叠状态 */
.sidebar-container.is-collapsed {
  width: 64px;
}

.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border-bottom: 1px solid var(--border-lighter, #ebeef5);
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  transition: all 0.3s ease;
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
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  transition: all 0.3s ease;
  width: 100%;
}

.logo i {
  font-size: 24px;
  flex-shrink: 0;
}

.logo span {
  white-space: nowrap;
  transition: opacity 0.3s ease, width 0.3s ease;
  overflow: hidden;
}

/* 折叠时隐藏文字 */
.sidebar-container.is-collapsed .logo span {
  opacity: 0;
  width: 0;
}

.side-nav-bar:not(.el-menu--collapse) {
  width: 210px;
  border: none;
}

.side-nav-bar {
  border: none;
  padding: 12px 8px;
}

.side-nav-bar i.menu-icon {
  margin-right: 12px;
  font-size: 18px;
  color: var(--primary-color, #409EFF);
  transition: all 0.3s ease;
}

.side-nav-bar .el-menu-item,
.side-nav-bar .el-submenu__title {
  border-radius: 8px;
  margin: 4px 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  height: 48px;
  line-height: 48px;
  -webkit-tap-highlight-color: transparent;
  cursor: pointer;
}

.side-nav-bar .el-menu-item:hover,
.side-nav-bar .el-submenu__title:hover {
  background-color: rgba(64, 158, 255, 0.08) !important;
  transform: translateX(4px);
}

.side-nav-bar .el-menu-item.is-active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.side-nav-bar .el-menu-item.is-active .menu-icon {
  color: #fff;
}

.side-nav-bar .el-submenu .el-menu-item {
  padding-left: 56px !important;
}

.side-nav-bar.el-menu--collapse {
  width: 64px;
}

.side-nav-bar.el-menu--collapse .el-menu-item,
.side-nav-bar.el-menu--collapse .el-submenu__title {
  padding: 0 20px;
  justify-content: center;
}

.side-nav-bar.el-menu--collapse .el-menu-item i.menu-icon,
.side-nav-bar.el-menu--collapse .el-submenu__title i.menu-icon {
  margin-right: 0;
}

.side-nav-bar.el-menu--collapse .el-menu-item span,
.side-nav-bar.el-menu--collapse .el-submenu__title span {
  display: none;
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

/* 移动端适配 */
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

  .logo i {
    font-size: 22px;
  }

  .side-nav-bar {
    padding: 10px 6px;
  }

  .side-nav-bar .el-menu-item,
  .side-nav-bar .el-submenu__title {
    height: 44px;
    line-height: 44px;
  }
}

/* 暗色主题适配 */
::root[data-theme="dark"] .sidebar-container {
  background: linear-gradient(180deg, #1a1a1a 0%, #141414 100%);
  border-right-color: #2b2b2c;
}

::root[data-theme="dark"] .sidebar-header {
  background: linear-gradient(135deg, #4a5568 0%, #2d3748 100%);
}

::root[data-theme="dark"] .side-nav-bar .el-menu-item:hover,
::root[data-theme="dark"] .side-nav-bar .el-submenu__title:hover {
  background-color: rgba(255, 255, 255, 0.08) !important;
}

::root[data-theme="dark"] .side-nav-bar .el-menu-item.is-active {
  background: linear-gradient(135deg, #4a5568 0%, #2d3748 100%);
}
</style>

<style>
.sidebar-popper {
  background: #fff !important;
  border: 1px solid var(--border-light, #e4e7ed);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  padding: 8px;
}

.sidebar-popper.el-menu--popup {
  min-width: 160px;
}

.sidebar-popper .el-menu-item {
  border-radius: 6px;
  height: 40px;
  line-height: 40px;
  margin: 4px 0;
}

.sidebar-popper .el-menu-item:hover {
  background-color: rgba(64, 158, 255, 0.08) !important;
}

.sidebar-popper .el-menu-item.is-active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

::root[data-theme="dark"] .sidebar-popper {
  background: #1a1a1a !important;
  border-color: #2b2b2c;
}

::root[data-theme="dark"] .sidebar-popper .el-menu-item:hover {
  background-color: rgba(255, 255, 255, 0.08) !important;
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

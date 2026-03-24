<template>
  <div class="online-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Monitor /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">在线用户</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><ChromeFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ browserStats.top }}</span>
          <span class="stat-label">Chrome 用户</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Position /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ ipSet.size }}</span>
          <span class="stat-label">不同 IP</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Refresh" @click="listOnlineUsers" class="btn-refresh" :loading="loading">
            <span>刷新</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索用户昵称..."
            clearable
            class="search-input"
            @keyup.enter="listOnlineUsers" />
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        :data="users"
        v-loading="loading"
        class="online-table">
        <el-table-column prop="avatar" label="头像" align="center" width="70">
          <template #default="{ row }">
            <el-avatar :src="row.avatar" :size="38" class="online-avatar">
              <el-icon><UserFilled /></el-icon>
            </el-avatar>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120">
          <template #default="{ row }">
            <span class="nickname-text">{{ row.nickname }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ipAddress" label="IP 地址" align="center" width="140">
          <template #default="{ row }">
            <span class="ip-text">{{ row.ipAddress || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ipSource" label="登录地址" min-width="130">
          <template #default="{ row }">
            <span class="location-text">{{ row.ipSource || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="browser" label="浏览器" align="center" min-width="120">
          <template #default="{ row }">
            <span :class="['browser-badge', getBrowserCls(row.browser)]">
              {{ row.browser || '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="os" label="操作系统" align="center" width="130">
          <template #default="{ row }">
            <span class="os-text">{{ row.os || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="lastLoginTime" label="登录时间" width="170" align="center">
          <template #default="{ row }">
            <span class="time-text">{{ formatDateTime(row.lastLoginTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" align="center" width="80" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="强制下线" placement="top">
              <button class="action-btn offline" @click="handleOffline(row)">
                <el-icon><SwitchButton /></el-icon>
              </button>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="count > 0">
        <el-pagination
          @size-change="sizeChange"
          @current-change="currentChange"
          :current-page="current"
          :page-size="size"
          :total="count"
          :page-sizes="[10, 20]"
          layout="total, sizes, prev, pager, next, jumper" />
      </div>
    </el-card>

    <!-- 下线确认对话框 -->
    <el-dialog v-model="isOffline" width="400px" custom-class="elegant-dialog" :show-close="false">
      <template #header>
        <div class="dialog-title-container">
          <div class="dialog-icon-wrapper warning">
            <el-icon><WarningFilled /></el-icon>
          </div>
          <span class="dialog-title-text">强制下线</span>
        </div>
      </template>
      <div class="dialog-content-text">
        确定要将用户 <strong>{{ offlineUser?.nickname }}</strong> 强制下线吗？
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isOffline = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="confirmOffline" class="btn-confirm-danger">
            <el-icon><SwitchButton /></el-icon>
            <span>确认下线</span>
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElNotification } from 'element-plus'
import {
  Search, Monitor, UserFilled, Position,
  WarningFilled, SwitchButton, Refresh, ChromeFilled
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()
const userStore = useUserStore()

const loading = ref(true)
const users = ref([])
const keywords = ref(null)
const current = ref(pageStateStore.online || 1)
const size = ref(10)
const count = ref(0)
const isOffline = ref(false)
const offlineUser = ref(null)

const ipSet = computed(() => new Set(users.value.map(u => u.ipAddress).filter(Boolean)))

const browserStats = computed(() => {
  const browsers = {}
  users.value.forEach(u => {
    if (u.browser) {
      browsers[u.browser] = (browsers[u.browser] || 0) + 1
    }
  })
  const sorted = Object.entries(browsers).sort((a, b) => b[1] - a[1])
  return {
    top: sorted.find(([b]) => b.includes('Chrome'))?.[1] || 0
  }
})

const getBrowserCls = (browser) => {
  if (!browser) return ''
  if (browser.includes('Chrome')) return 'browser-chrome'
  if (browser.includes('Firefox')) return 'browser-firefox'
  if (browser.includes('Safari')) return 'browser-safari'
  if (browser.includes('Edge')) return 'browser-edge'
  return 'browser-other'
}

const formatDateTime = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(() => {
  listOnlineUsers()
})

const listOnlineUsers = () => {
  loading.value = true
  request.get('/admin/users/online', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      users.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

const sizeChange = (val) => {
  size.value = val
  listOnlineUsers()
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updateOnlinePageState(val)
  listOnlineUsers()
}

const handleOffline = (user) => {
  offlineUser.value = user
  isOffline.value = true
}

const confirmOffline = () => {
  const user = offlineUser.value
  request.delete('/admin/users/' + user.userInfoId + '/online').then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      if (user.userInfoId === userStore.userInfo?.id) {
        router.push({ path: '/login' })
        sessionStorage.removeItem('token')
      }
      listOnlineUsers()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    isOffline.value = false
  })
}
</script>

<style scoped>
/* ==================== 页面容器 ==================== */
.online-page {
  padding: 4px;
}

/* ==================== 统计卡片 ==================== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  background: var(--bg-card, #fff);
  border-radius: 16px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.06));
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.04));
}

.stat-card:hover {
  transform: translateY(-3px);
  box-shadow: var(--shadow-md, 0 8px 25px rgba(0, 0, 0, 0.1));
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}

.stat-icon.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.stat-icon.success {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: #fff;
}

.stat-icon.warning {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: #fff;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary, #1a1a2e);
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: var(--text-muted, #8e8ea0);
  margin-top: 2px;
}

/* ==================== 主内容卡片 ==================== */
.main-card {
  border-radius: 16px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.06));
  box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.04));
}

.main-card :deep(.el-card__body) {
  padding: 24px;
}

/* ==================== 工具栏 ==================== */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  gap: 10px;
}

.toolbar-right {
  display: flex;
  gap: 10px;
}

.btn-refresh {
  border-radius: 10px;
  font-weight: 600;
  padding: 9px 18px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-refresh:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.search-input {
  width: 260px;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.08));
  box-shadow: none;
  transition: all 0.3s;
}

.search-input :deep(.el-input__wrapper):focus-within {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.12);
}

/* ==================== 表格 ==================== */
.online-table {
  width: 100%;
  border-radius: 12px;
}

.online-table :deep(.el-table__header th) {
  background: var(--bg-secondary, #f5f7fa) !important;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-secondary, #666);
  border-bottom: 1px solid var(--border-light, rgba(0, 0, 0, 0.06));
}

.online-table :deep(.el-table__row) {
  transition: all 0.25s ease;
}

.online-table :deep(.el-table__row:hover > td) {
  background: var(--bg-hover, rgba(102, 126, 234, 0.03)) !important;
}

/* 头像 */
.online-avatar {
  border: 2px solid var(--border-light, rgba(0, 0, 0, 0.06));
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.online-table :deep(.el-table__row:hover) .online-avatar {
  transform: scale(1.1);
  border-color: rgba(102, 126, 234, 0.3);
  box-shadow: 0 2px 10px rgba(102, 126, 234, 0.15);
}

/* 昵称 */
.nickname-text {
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
  font-size: 14px;
}

/* IP */
.ip-text {
  font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px;
  color: var(--text-secondary, #666);
  background: var(--bg-secondary, #f5f7fa);
  padding: 2px 8px;
  border-radius: 6px;
}

/* 地址 */
.location-text {
  font-size: 13px;
  color: var(--text-secondary, #666);
}

/* 浏览器徽章 */
.browser-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.browser-chrome {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.12), rgba(118, 75, 162, 0.12));
  color: #667eea;
}

.browser-firefox {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), rgba(251, 191, 36, 0.12));
  color: #d97706;
}

.browser-safari {
  background: linear-gradient(135deg, rgba(67, 233, 123, 0.12), rgba(56, 249, 215, 0.12));
  color: #10b981;
}

.browser-edge {
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.12), rgba(56, 189, 248, 0.12));
  color: #0ea5e9;
}

.browser-other {
  background: var(--bg-secondary, #f0f0f0);
  color: var(--text-muted, #999);
}

/* 操作系统 */
.os-text {
  font-size: 13px;
  color: var(--text-secondary, #666);
}

/* 时间 */
.time-text {
  font-size: 13px;
  color: var(--text-muted, #999);
  font-variant-numeric: tabular-nums;
}

/* 操作按钮 */
.action-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.08));
  background: var(--bg-card, #fff);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.action-btn.offline {
  color: #f5576c;
}

.action-btn.offline:hover {
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 87, 108, 0.3);
}

/* ==================== 分页 ==================== */
.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light, rgba(0, 0, 0, 0.04));
}

/* ==================== 对话框 ==================== */
.dialog-title-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dialog-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.dialog-icon-wrapper.warning {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), rgba(251, 191, 36, 0.12));
  color: #f59e0b;
}

.dialog-title-text {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary, #1a1a2e);
}

.dialog-content-text {
  font-size: 14px;
  color: var(--text-secondary, #666);
  line-height: 1.6;
  padding-left: 56px;
}

.dialog-content-text strong {
  color: var(--text-primary, #1a1a2e);
}

.dialog-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.btn-cancel {
  border-radius: 10px;
  font-weight: 500;
  padding: 9px 20px;
}

.btn-confirm-danger {
  border-radius: 10px;
  font-weight: 600;
  padding: 9px 20px;
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  border: none;
  transition: all 0.3s;
}

.btn-confirm-danger:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(245, 87, 108, 0.4);
}

/* ==================== 深色模式 ==================== */
[data-theme="dark"] .stat-card {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.06));
}

[data-theme="dark"] .main-card {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.06));
}

[data-theme="dark"] .online-table :deep(.el-table__header th) {
  background: var(--bg-secondary, #2a2a3e) !important;
}

[data-theme="dark"] .online-table :deep(.el-table__row:hover > td) {
  background: rgba(102, 126, 234, 0.06) !important;
}

[data-theme="dark"] .action-btn {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.08));
}

[data-theme="dark"] .browser-chrome {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.18), rgba(118, 75, 162, 0.18));
}

[data-theme="dark"] .browser-firefox {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.18), rgba(251, 191, 36, 0.18));
}

[data-theme="dark"] .browser-safari {
  background: linear-gradient(135deg, rgba(67, 233, 123, 0.18), rgba(56, 249, 215, 0.18));
}

[data-theme="dark"] .browser-edge {
  background: linear-gradient(135deg, rgba(14, 165, 233, 0.18), rgba(56, 189, 248, 0.18));
}

[data-theme="dark"] .browser-other {
  background: var(--bg-secondary, #2a2a3e);
}

[data-theme="dark"] .ip-text {
  background: var(--bg-secondary, #2a2a3e);
}

/* ==================== 响应式 ==================== */
@media (max-width: 1280px) {
  .stats-row {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
    gap: 10px;
  }

  .stat-card {
    padding: 14px;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 18px;
    border-radius: 11px;
  }

  .stat-value {
    font-size: 20px;
  }

  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left {
    width: 100%;
  }

  .btn-refresh {
    width: 100%;
  }

  .toolbar-right {
    width: 100%;
  }

  .search-input {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .stats-row {
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .stat-card {
    padding: 12px;
    gap: 10px;
  }

  .stat-icon {
    width: 36px;
    height: 36px;
    font-size: 16px;
    border-radius: 10px;
  }

  .stat-value {
    font-size: 18px;
  }
}
</style>

<template>
  <div class="user-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><UserFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">用户总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ enabledCount }}</span>
          <span class="stat-label">正常</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><CircleClose /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ disabledCount }}</span>
          <span class="stat-label">已禁用</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon info">
          <el-icon><Message /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ emailCount }}</span>
          <span class="stat-label">邮箱注册</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select
            v-model="loginType"
            clearable
            placeholder="登录方式"
            class="filter-select">
            <el-option v-for="item in typeList" :key="item.type" :label="item.desc" :value="item.type" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索用户昵称..."
            clearable
            class="search-input"
            @keyup.enter="searchUsers" />
        </div>
      </div>

      <!-- 表格 -->
      <el-table
        :data="userList"
        v-loading="loading"
        class="user-table">
        <el-table-column prop="avatar" label="头像" align="center" width="70">
          <template #default="{ row }">
            <el-avatar :src="row.avatar" :size="38" class="user-avatar" />
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120">
          <template #default="{ row }">
            <div class="nickname-cell">
              <span class="user-nickname">{{ row.nickname }}</span>
              <span class="user-id">#{{ row.userInfoId }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="loginType" label="登录方式" align="center" width="110">
          <template #default="{ row }">
            <span :class="['login-badge', getLoginType(row.loginType).cls]">
              {{ getLoginType(row.loginType).label }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="roles" label="用户角色" min-width="160">
          <template #default="{ row }">
            <div class="role-tags">
              <el-tag
                v-for="(item, index) in row.roles"
                :key="index"
                size="small"
                effect="plain"
                class="role-tag">
                {{ item.roleName }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="isDisable" label="状态" align="center" width="100">
          <template #default="{ row }">
            <span :class="['status-badge', row.isDisable ? 'status-disabled' : 'status-enabled']">
              <span class="status-dot"></span>
              {{ row.isDisable ? '已禁用' : '正常' }}
            </span>
            <el-switch
              v-model="row.isDisable"
              :active-value="1"
              :inactive-value="0"
              @change="changeDisable(row)"
              class="status-switch" />
          </template>
        </el-table-column>
        <el-table-column prop="ipAddress" label="登录IP" align="center" width="130">
          <template #default="{ row }">
            <span class="ip-text">{{ row.ipAddress || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ipSource" label="登录地址" min-width="120">
          <template #default="{ row }">
            <span class="location-text">{{ row.ipSource || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="160" align="center">
          <template #default="{ row }">
            <span class="time-text">{{ formatDate(row.createTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="lastLoginTime" label="上次登录" width="160" align="center">
          <template #default="{ row }">
            <span class="time-text">{{ formatDate(row.lastLoginTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" align="center" width="80" fixed="right">
          <template #default="{ row }">
            <el-tooltip content="编辑" placement="top">
              <button class="action-btn edit" @click="openEditModel(row)">
                <el-icon><Edit /></el-icon>
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

    <!-- 编辑用户对话框 -->
    <el-dialog v-model="isEdit" width="420px" custom-class="elegant-dialog" :show-close="false">
      <template #header>
        <div class="dialog-title-container">
          <div class="dialog-icon-wrapper primary">
            <el-icon><Edit /></el-icon>
          </div>
          <span class="dialog-title-text">修改用户</span>
        </div>
      </template>
      <el-form ref="userFormRef" label-width="70px" :model="userForm" :rules="userRules" class="user-form">
        <el-form-item label="昵称" prop="nickname">
          <el-input v-model="userForm.nickname" placeholder="请输入用户昵称" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="角色">
          <el-checkbox-group v-model="roleIds" class="role-checkbox-group">
            <el-checkbox
              v-for="item in userRoles"
              :key="item.id"
              :label="item.id"
              border
              size="small">
              {{ item.roleName }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isEdit = false" class="btn-cancel">取 消</el-button>
          <el-button type="primary" @click="editUserRole" class="btn-confirm">
            <el-icon><Check /></el-icon>
            <span>确 定</span>
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  UserFilled, Search, Edit, User, Check, Clock, Message,
  ChatDotRound, Share, QuestionFilled, CircleCheck, CircleClose
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import logger from '@/utils/logger'

const route = useRoute()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const isEdit = ref(false)
const userForm = reactive({
  userInfoId: null,
  nickname: ''
})
const userRules = {
  nickname: [{ required: true, message: '请输入用户昵称', trigger: 'blur' }]
}
const loginType = ref(null)
const userRoles = ref([])
const roleIds = ref([])
const userList = ref([])
const keywords = ref('')
const current = ref(1)
const size = ref(10)
const count = ref(0)

const typeList = [
  { type: 1, desc: '邮箱' },
  { type: 2, desc: 'QQ' },
  { type: 3, desc: '微博' }
]

const enabledCount = computed(() => userList.value.filter(u => u.isDisable === 0).length)
const disabledCount = computed(() => userList.value.filter(u => u.isDisable === 1).length)
const emailCount = computed(() => userList.value.filter(u => u.loginType === 1).length)

const getLoginType = (type) => {
  const types = {
    1: { label: '邮箱', cls: 'badge-email' },
    2: { label: 'QQ', cls: 'badge-qq' },
    3: { label: '微博', cls: 'badge-weibo' }
  }
  return types[type] || { label: '未知', cls: 'badge-unknown' }
}

const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

const searchUsers = () => {
  current.value = 1
  listUsers()
}

const sizeChange = (newSize) => {
  size.value = newSize
  listUsers()
}

const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('user', newCurrent)
  listUsers()
}

const changeDisable = (user) => {
  request.put('/admin/users/disable', {
    id: user.userInfoId,
    isDisable: user.isDisable
  }).catch(error => {
    ElMessage.error('操作失败')
    logger.error('API Error:', error)
  })
}

const openEditModel = (user) => {
  roleIds.value = []
  Object.assign(userForm, JSON.parse(JSON.stringify(user)))
  userForm.roles.forEach((item) => {
    roleIds.value.push(item.id)
  })
  isEdit.value = true
}

const editUserRole = () => {
  userForm.roleIds = roleIds.value
  request.put('/admin/users/role', userForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listUsers()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    isEdit.value = false
  }).catch(error => {
    ElMessage.error('修改失败')
    logger.error('API Error:', error)
  })
}

const listUsers = () => {
  loading.value = true
  request.get('/admin/users', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value,
      loginType: loginType.value
    }
  }).then(({ data }) => {
    userList.value = data.data.records
    count.value = data.data.count
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取用户列表失败')
    logger.error('API Error:', error)
  })
}

const listRoles = () => {
  request.get('/admin/users/role').then(({ data }) => {
    userRoles.value = data?.data || []
  }).catch(error => {
    ElMessage.error('获取角色列表失败')
    logger.error('API Error:', error)
  })
}

watch(loginType, () => {
  current.value = 1
  listUsers()
})

onMounted(() => {
  current.value = pageStateStore.pageState.user
  listUsers()
  listRoles()
})
</script>

<style scoped>
/* ==================== 页面容器 ==================== */
.user-page {
  padding: 4px;
}

/* ==================== 统计卡片 ==================== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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

.stat-icon.info {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
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

.filter-select {
  width: 160px;
}

.filter-select :deep(.el-input__wrapper) {
  border-radius: 10px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.08));
  box-shadow: none;
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
.user-table {
  width: 100%;
  border-radius: 12px;
}

.user-table :deep(.el-table__header th) {
  background: var(--bg-secondary, #f5f7fa) !important;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-secondary, #666);
  border-bottom: 1px solid var(--border-light, rgba(0, 0, 0, 0.06));
}

.user-table :deep(.el-table__row) {
  transition: all 0.25s ease;
}

.user-table :deep(.el-table__row:hover > td) {
  background: var(--bg-hover, rgba(102, 126, 234, 0.03)) !important;
}

.user-table :deep(.el-table__row) {
  cursor: default;
}

/* 头像 */
.user-avatar {
  border: 2px solid var(--border-light, rgba(0, 0, 0, 0.06));
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.user-table :deep(.el-table__row:hover) .user-avatar {
  transform: scale(1.1);
  border-color: rgba(102, 126, 234, 0.3);
  box-shadow: 0 2px 10px rgba(102, 126, 234, 0.15);
}

/* 昵称 */
.nickname-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-nickname {
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
  font-size: 14px;
}

.user-id {
  font-size: 11px;
  color: var(--text-muted, #bbb);
  font-variant-numeric: tabular-nums;
}

/* 登录方式徽章 */
.login-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.badge-email {
  background: linear-gradient(135deg, rgba(67, 233, 123, 0.12), rgba(56, 249, 215, 0.12));
  color: #10b981;
}

.badge-qq {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.12), rgba(118, 75, 162, 0.12));
  color: #667eea;
}

.badge-weibo {
  background: linear-gradient(135deg, rgba(245, 87, 108, 0.12), rgba(240, 147, 251, 0.12));
  color: #f5576c;
}

.badge-unknown {
  background: var(--bg-secondary, #f0f0f0);
  color: var(--text-muted, #999);
}

/* 角色标签 */
.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.role-tag {
  border-radius: 8px;
  font-weight: 500;
  font-size: 12px;
}

/* 状态 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-enabled {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.status-enabled .status-dot {
  background: #10b981;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}

.status-disabled {
  background: rgba(245, 87, 108, 0.1);
  color: #f5576c;
}

.status-disabled .status-dot {
  background: #f5576c;
  box-shadow: 0 0 6px rgba(245, 87, 108, 0.4);
}

.status-switch {
  margin-top: 4px;
}

/* IP / 地址 / 时间 */
.ip-text {
  font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px;
  color: var(--text-secondary, #666);
  background: var(--bg-secondary, #f5f7fa);
  padding: 2px 8px;
  border-radius: 6px;
}

.location-text {
  font-size: 13px;
  color: var(--text-secondary, #666);
}

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

.action-btn.edit {
  color: #667eea;
}

.action-btn.edit:hover {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
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

.dialog-icon-wrapper.primary {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.12), rgba(118, 75, 162, 0.12));
  color: #667eea;
}

.dialog-title-text {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary, #1a1a2e);
}

.user-form {
  padding: 16px 0 0;
}

.user-form :deep(.el-input__wrapper) {
  border-radius: 10px;
}

.user-form :deep(.el-input__wrapper):focus-within {
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.12);
}

.role-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.role-checkbox-group .el-checkbox {
  margin: 0;
  border-radius: 8px;
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

.btn-confirm {
  border-radius: 10px;
  font-weight: 600;
  padding: 9px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s;
}

.btn-confirm:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
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

[data-theme="dark"] .user-table :deep(.el-table__header th) {
  background: var(--bg-secondary, #2a2a3e) !important;
}

[data-theme="dark"] .user-table :deep(.el-table__row:hover > td) {
  background: rgba(102, 126, 234, 0.06) !important;
}

[data-theme="dark"] .action-btn {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.08));
}

[data-theme="dark"] .badge-email {
  background: linear-gradient(135deg, rgba(67, 233, 123, 0.18), rgba(56, 249, 215, 0.18));
}

[data-theme="dark"] .badge-qq {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.18), rgba(118, 75, 162, 0.18));
}

[data-theme="dark"] .badge-weibo {
  background: linear-gradient(135deg, rgba(245, 87, 108, 0.18), rgba(240, 147, 251, 0.18));
}

[data-theme="dark"] .ip-text {
  background: var(--bg-secondary, #2a2a3e);
}

[data-theme="dark"] .badge-unknown {
  background: var(--bg-secondary, #2a2a3e);
}

/* ==================== 响应式 ==================== */
@media (max-width: 1280px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
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

  .filter-select {
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

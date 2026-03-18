<template>
  <el-card class="main-card user-management">
    <div class="title">
      <el-icon><UserFilled /></el-icon>
      {{ route.name }}
    </div>
    <div class="operation-container">
      <div class="filter-section">
        <el-select 
          clearable 
          v-model="loginType" 
          placeholder="请选择登录方式" 
          size="small" 
          class="filter-select">
          <el-option v-for="item in typeList" :key="item.type" :label="item.desc" :value="item.type" />
        </el-select>
        <el-input
          v-model="keywords"
          :prefix-icon="Search"
          size="small"
          placeholder="请输入昵称"
          class="search-input"
          @keyup.enter="searchUsers" />
        <el-button type="primary" size="small" :icon="Search" @click="searchUsers">
          搜索
        </el-button>
      </div>
    </div>
    <el-table 
      border 
      :data="userList" 
      v-loading="loading"
      class="user-table"
      :header-cell-style="{ background: 'var(--bg-elevated)', color: 'var(--text-primary)', fontWeight: '600' }">
      <el-table-column prop="linkAvatar" label="头像" align="center" width="80">
        <template #default="{ row }">
          <el-avatar :src="row.avatar" :size="40" class="user-avatar" />
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="昵称" align="center" min-width="120">
        <template #default="{ row }">
          <span class="user-nickname">{{ row.nickname }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="loginType" label="登录方式" align="center" width="100">
        <template #default="{ row }">
          <el-tag 
            :type="getLoginType(row.loginType).type" 
            effect="light"
            class="login-type-tag">
            <el-icon><component :is="getLoginType(row.loginType).icon" /></el-icon>
            {{ getLoginType(row.loginType).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="roles" label="用户角色" align="center" min-width="150">
        <template #default="{ row }">
          <div class="role-tags">
            <el-tag 
              v-for="(item, index) in row.roles" 
              :key="index" 
              type="primary"
              effect="plain"
              size="small"
              class="role-tag">
              {{ item.roleName }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="isDisable" label="状态" align="center" width="90">
        <template #default="{ row }">
          <el-switch
            v-model="row.isDisable"
            :active-color="'#F56C6C'"
            :inactive-color="'#67C23A'"
            :active-value="1"
            :inactive-value="0"
            @change="changeDisable(row)"
            class="status-switch" />
          <div class="status-text" :class="row.isDisable ? 'disabled' : 'enabled'">
            {{ row.isDisable ? '已禁用' : '正常' }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="ipAddress" label="登录IP" align="center" width="130">
        <template #default="{ row }">
          <span class="ip-address">{{ row.ipAddress || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="ipSource" label="登录地址" align="center" min-width="120">
        <template #default="{ row }">
          <span class="ip-source">{{ row.ipSource || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="140" align="center">
        <template #default="{ row }">
          <div class="time-info">
            <el-icon><Clock /></el-icon>
            <span>{{ formatDate(row.createTime) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="lastLoginTime" label="上次登录" width="140" align="center">
        <template #default="{ row }">
          <div class="time-info">
            <el-icon><Clock /></el-icon>
            <span>{{ formatDate(row.lastLoginTime) }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="100" fixed="right">
        <template #default="{ row }">
          <el-button 
            type="primary" 
            size="small" 
            :icon="Edit"
            circle
            @click="openEditModel(row)">
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      class="pagination-container"
      background
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      :page-sizes="[10, 20]"
      layout="total, sizes, prev, pager, next, jumper" />
    <el-dialog v-model="isEdit" width="400px" class="user-edit-dialog">
      <template #header>
        <div class="dialog-title-container">
          <el-icon><Edit /></el-icon>
          修改用户
        </div>
      </template>
      <el-form label-width="70px" size="medium" :model="userForm" class="user-form">
        <el-form-item label="昵称">
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
          <el-button @click="isEdit = false" size="medium">取 消</el-button>
          <el-button type="primary" @click="editUserRole" size="medium" :icon="Check"> 确 定 </el-button>
        </div>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { 
  UserFilled, 
  Search, 
  Edit, 
  User, 
  Check, 
  Clock,
  Message,
  ChatDotRound,
  Share,
  QuestionFilled
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'

const route = useRoute()
const pageStateStore = usePageStateStore()

// 响应式数据
const loading = ref(true)
const isEdit = ref(false)
const userForm = reactive({
  userInfoId: null,
  nickname: ''
})
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

// 登录类型映射
const getLoginType = (type) => {
  const types = {
    1: { label: '邮箱', type: 'success', icon: 'Message' },
    2: { label: 'QQ', type: 'primary', icon: 'ChatDotRound' },
    3: { label: '微博', type: 'danger', icon: 'Share' }
  }
  return types[type] || { label: '未知', type: 'info', icon: 'QuestionFilled' }
}

// 日期格式化
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

// 搜索用户
const searchUsers = () => {
  current.value = 1
  listUsers()
}

// 分页大小变化
const sizeChange = (newSize) => {
  size.value = newSize
  listUsers()
}

// 页码变化
const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('user', newCurrent)
  listUsers()
}

// 切换禁用状态
const changeDisable = (user) => {
  request.put('/admin/users/disable', {
    id: user.userInfoId,
    isDisable: user.isDisable
  }).catch(error => {
    ElMessage.error('操作失败')
    console.error('API Error:', error)
  })
}

// 打开编辑对话框
const openEditModel = (user) => {
  roleIds.value = []
  Object.assign(userForm, JSON.parse(JSON.stringify(user)))
  userForm.roles.forEach((item) => {
    roleIds.value.push(item.id)
  })
  isEdit.value = true
}

// 编辑用户角色
const editUserRole = () => {
  userForm.roleIds = roleIds.value
  request.put('/admin/users/role', userForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listUsers()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isEdit.value = false
  }).catch(error => {
    ElMessage.error('修改失败')
    console.error('API Error:', error)
  })
}

// 获取用户列表
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
    console.error('API Error:', error)
  })
}

// 获取角色列表
const listRoles = () => {
  request.get('/admin/users/role').then(({ data }) => {
    userRoles.value = data?.data || []
  }).catch(error => {
    ElMessage.error('获取角色列表失败')
    console.error('API Error:', error)
  })
}

// 监听登录类型变化
watch(loginType, () => {
  current.value = 1
  listUsers()
})

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.user
  listUsers()
  listRoles()
})
</script>

<style scoped>
/* 用户管理页面样式 */
.user-management {
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--color-border);
}

.title .el-icon {
  color: var(--color-primary);
  font-size: 24px;
}

/* 筛选区域 */
.filter-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-select {
  width: 160px;
}

.search-input {
  width: 220px;
}

/* 用户表格 */
.user-table {
  margin-top: 20px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.user-table :deep(.el-table__header th) {
  background: var(--color-bg-hover) !important;
  font-weight: 600;
  text-transform: uppercase;
  font-size: 12px;
  letter-spacing: 0.5px;
}

.user-table :deep(.el-table__row) {
  transition: all 0.3s ease;
}

.user-table :deep(.el-table__row:hover) {
  background-color: var(--color-bg-hover) !important;
  transform: translateX(4px);
}

/* 用户头像 */
.user-avatar {
  border: 2px solid var(--color-border);
  transition: all 0.3s ease;
}

.user-avatar:hover {
  transform: scale(1.1);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 用户昵称 */
.user-nickname {
  font-weight: 500;
  color: var(--color-text);
}

/* 登录类型标签 */
.login-type-tag {
  border-radius: 20px;
  padding: 4px 12px;
  font-weight: 500;
}

.login-type-tag .el-icon {
  margin-right: 4px;
}

/* 角色标签 */
.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
}

.role-tag {
  border-radius: 4px;
  font-weight: 500;
}

/* 状态开关 */
.status-switch {
  margin-bottom: 4px;
}

.status-text {
  font-size: 12px;
  font-weight: 500;
  margin-top: 4px;
}

.status-text.enabled {
  color: var(--color-success);
}

.status-text.disabled {
  color: var(--color-error);
}

/* IP 信息 */
.ip-address {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--color-text-secondary);
  background: var(--color-bg-hover);
  padding: 2px 8px;
  border-radius: 4px;
}

.ip-source {
  font-size: 13px;
  color: var(--color-text-secondary);
}

/* 时间信息 */
.time-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.time-info .el-icon {
  color: var(--color-primary);
}

/* 分页 */
.pagination-container {
  float: right;
  margin-top: 24px;
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
}

.dialog-title-container .el-icon {
  color: var(--color-primary);
  font-size: 20px;
}

.user-form {
  padding: 20px 0;
}

.role-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.role-checkbox-group .el-checkbox {
  margin: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 响应式优化 */
@media (max-width: 768px) {
  .filter-section {
    flex-direction: column;
    width: 100%;
  }

  .filter-select,
  .search-input {
    width: 100% !important;
  }

  .user-table :deep(.el-table__header) {
    display: none;
  }

  .user-table :deep(.el-table__row) {
    display: flex;
    flex-direction: column;
    padding: 16px;
    margin-bottom: 12px;
    background: var(--color-bg-card);
    border-radius: 12px;
    border: 1px solid var(--color-border);
  }

  .user-table :deep(.el-table__row td) {
    display: flex;
    align-items: center;
    padding: 8px 0 !important;
    border: none;
  }

  .user-table :deep(.el-table__row td::before) {
    content: attr(data-label);
    font-weight: 600;
    color: var(--color-text-secondary);
    min-width: 80px;
    margin-right: 12px;
    font-size: 13px;
  }

  .user-table :deep(.el-table__row td .cell) {
    flex: 1;
    text-align: left !important;
  }

  .role-tags {
    justify-content: flex-start;
  }

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
    margin: 20px 0;
  }
}
</style>

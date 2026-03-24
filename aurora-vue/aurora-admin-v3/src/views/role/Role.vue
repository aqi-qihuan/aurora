<template>
  <div class="role-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><UserFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">角色总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Key /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ selectedCount }}</span>
          <span class="stat-label">已选中</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ roles.length }}</span>
          <span class="stat-label">本页角色</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openMenuModel(null)" class="btn-add">
            <span>新增角色</span>
          </el-button>
          <el-button
            type="danger"
            :icon="Delete"
            :disabled="roleIds.length === 0"
            @click="isDelete = true"
            class="btn-batch-delete">
            <span>批量删除 ({{ roleIds.length }})</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索角色名..."
            class="search-input"
            clearable
            @keyup.enter="searchRoles"
            @clear="searchRoles" />
          <el-button type="primary" :icon="Search" @click="searchRoles" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table
        :data="roles"
        v-loading="loading"
        @selection-change="selectionChange"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="roleName" label="角色名" min-width="160">
          <template #default="{ row }">
            <div class="role-name-cell">
              <div class="role-avatar">
                <el-icon><User /></el-icon>
              </div>
              <span class="role-name-text">{{ row.roleName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="roleLabel" label="权限标签" width="160" align="center">
          <template #default="{ row }">
            <span class="role-badge">{{ row.roleLabel }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" align="center" sortable>
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDate(row.createTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="菜单权限" placement="top" :show-after="500">
                <button class="action-btn menu" @click="openMenuModel(row)"><el-icon><Menu /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="资源权限" placement="top" :show-after="500">
                <button class="action-btn resource" @click="openResourceModel(row)"><el-icon><FolderChecked /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="删除" placement="top" :show-after="500">
                <button class="action-btn delete" @click="handleDelete(row.id)"><el-icon><Delete /></el-icon></button>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next, jumper"
          :total="count"
          :page-size="size"
          :current-page="current"
          :page-sizes="[10, 20]"
          @size-change="sizeChange"
          @current-change="currentChange" />
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="isDelete" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ roleIds.length }} 个角色吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isDelete = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteRoles(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 新增/编辑角色 - 菜单权限对话框 -->
    <el-dialog v-model="roleMenu" width="520px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><Setting /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ roleTitle }}</h3>
        <el-form ref="roleFormRef" :model="roleForm" :rules="roleRules" class="role-form" label-position="top">
          <el-form-item label="角色名" prop="roleName">
            <el-input v-model="roleForm.roleName" placeholder="请输入角色名" class="form-input" :prefix-icon="User" />
          </el-form-item>
          <el-form-item label="菜单权限">
            <div class="tree-wrapper">
              <el-tree
                :data="menus"
                :default-checked-keys="roleForm.menuIds"
                show-checkbox
                node-key="id"
                ref="menuTreeRef" />
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="roleMenu = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="saveOrUpdateRoleMenu" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 资源权限对话框 -->
    <el-dialog v-model="roleResource" width="520px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper success"><el-icon><FolderChecked /></el-icon></div>
      <div class="dialog-content">
        <h3>修改资源权限</h3>
        <el-form :model="roleForm" class="role-form" label-position="top">
          <el-form-item label="角色名">
            <el-input v-model="roleForm.roleName" placeholder="角色名" class="form-input" :prefix-icon="User" disabled />
          </el-form-item>
          <el-form-item label="资源权限">
            <div class="tree-wrapper">
              <el-tree
                :data="resources"
                :default-checked-keys="roleForm.resourceIds"
                show-checkbox
                node-key="id"
                ref="resourceTreeRef" />
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="roleResource = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="saveOrUpdateRoleResource" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, nextTick, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Plus,
  Delete,
  Search,
  User,
  UserFilled,
  Key,
  Clock,
  Warning,
  Menu,
  FolderChecked,
  Setting
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import logger from '@/utils/logger'

const route = useRoute()
const pageStateStore = usePageStateStore()

// Refs
const menuTreeRef = ref(null)
const resourceTreeRef = ref(null)

// 响应式数据
const loading = ref(true)
const isDelete = ref(false)
const roles = ref([])
const roleIds = ref([])
const keywords = ref('')
const current = ref(1)
const size = ref(10)
const count = ref(0)
const roleMenu = ref(false)
const roleResource = ref(false)
const roleTitle = ref('')
const resources = ref([])
const menus = ref([])
const roleForm = reactive({
  roleName: '',
  roleLabel: '',
  resourceIds: [],
  menuIds: []
})
const roleRules = {
  roleName: [{ required: true, message: '请输入角色名', trigger: 'blur' }]
}

const selectedCount = computed(() => roleIds.value.length)

// 日期格式化
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

// 搜索角色
const searchRoles = () => {
  current.value = 1
  listRoles()
}

// 分页大小变化
const sizeChange = (newSize) => {
  size.value = newSize
  listRoles()
}

// 页码变化
const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('role', newCurrent)
  listRoles()
}

// 选择变化
const selectionChange = (selectedRoles) => {
  roleIds.value = selectedRoles.map(item => item.id)
}

// 获取角色列表
const listRoles = () => {
  loading.value = true
  request.get('/admin/roles', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    roles.value = data.data.records
    count.value = data.data.count
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取角色列表失败')
    logger.error('API Error:', error)
  })

  request.get('/admin/role/resources').then(({ data }) => {
    resources.value = data?.data || []
  }).catch(error => {
    logger.error('API Error:', error)
  })

  request.get('/admin/role/menus').then(({ data }) => {
    menus.value = data.data
  }).catch(error => {
    logger.error('API Error:', error)
  })
}

// 删除角色
const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该角色吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteRoles(id)
  }).catch(() => {})
}

const deleteRoles = (id) => {
  const param = id == null ? { data: roleIds.value } : { data: [id] }
  request.delete('/admin/roles', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listRoles()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isDelete.value = false
  }).catch(error => {
    ElMessage.error('删除失败')
    logger.error('API Error:', error)
  })
}

// 打开菜单权限对话框
const openMenuModel = (role) => {
  nextTick(() => {
    if (menuTreeRef.value) {
      menuTreeRef.value.setCheckedKeys([])
    }
  })

  roleTitle.value = role ? '修改角色' : '新增角色'

  if (role != null) {
    Object.assign(roleForm, JSON.parse(JSON.stringify(role)))
  } else {
    Object.assign(roleForm, {
      roleName: '',
      roleLabel: '',
      resourceIds: [],
      menuIds: []
    })
  }
  roleMenu.value = true
}

// 打开资源权限对话框
const openResourceModel = (role) => {
  nextTick(() => {
    if (resourceTreeRef.value) {
      resourceTreeRef.value.setCheckedKeys([])
    }
  })
  Object.assign(roleForm, JSON.parse(JSON.stringify(role)))
  roleResource.value = true
}

// 保存资源权限
const saveOrUpdateRoleResource = () => {
  roleForm.menuIds = null
  roleForm.resourceIds = resourceTreeRef.value ? resourceTreeRef.value.getCheckedKeys() : []
  request.post('/admin/role', roleForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listRoles()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    roleResource.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    logger.error('API Error:', error)
  })
}

// 保存菜单权限
const saveOrUpdateRoleMenu = () => {
  if (roleForm.roleName.trim() == '') {
    ElMessage.error('角色名不能为空')
    return false
  }
  roleForm.resourceIds = null
  const checkedKeys = menuTreeRef.value ? menuTreeRef.value.getCheckedKeys() : []
  const halfCheckedKeys = menuTreeRef.value ? menuTreeRef.value.getHalfCheckedKeys() : []
  roleForm.menuIds = checkedKeys.concat(halfCheckedKeys)
  request.post('/admin/role', roleForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listRoles()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    roleMenu.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    logger.error('API Error:', error)
  })
}

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.role
  listRoles()
})
</script>

<style scoped>
.role-page {
  padding: 0;
}

/* 统计卡片 */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--bg-base, #fff);
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid var(--border-default, #e5e7eb);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.08);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.stat-icon.primary { background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%); color: #fff; }
.stat-icon.success { background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 100%); color: #fff; }
.stat-icon.warning { background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%); color: #fff; }

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary, #1f2937);
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary, #6b7280);
}

/* 主卡片 */
.main-card {
  border-radius: 16px;
  border: 1px solid var(--border-default, #e5e7eb);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  background: var(--bg-base, #fff);
}

.main-card :deep(.el-card__body) {
  padding: 24px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.toolbar-left {
  display: flex;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-add {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
  border-radius: 10px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.2s ease;
}

.btn-add:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-batch-delete {
  border-radius: 10px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.2s ease;
}

.btn-batch-delete:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.search-input {
  width: 280px;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  transition: all 0.2s ease;
}

.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2), 0 0 0 1px #3b82f6;
}

/* 现代化表格 */
.modern-table {
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--border-default, #e5e7eb);
}

.modern-table :deep(.el-table__header-wrapper th) {
  background: var(--bg-elevated, #f9fafb);
  color: var(--text-secondary, #6b7280);
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 16px 12px;
  border-bottom: 1px solid var(--border-default, #e5e7eb);
}

.modern-table :deep(.el-table__body tr) {
  transition: all 0.2s ease;
}

.modern-table :deep(.el-table__body tr:hover > td) {
  background: var(--bg-hover, #f3f4f6) !important;
}

.modern-table :deep(.el-table__body td) {
  padding: 16px 12px;
  border-bottom: 1px solid var(--border-light, #f3f4f6);
}

/* 角色名单元格 */
.role-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.role-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}

.role-name-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

/* 权限标签 */
.role-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 20px;
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  color: #16a34a;
  font-size: 12px;
  font-weight: 500;
}

/* 时间单元格 */
.time-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-secondary, #6b7280);
  font-size: 14px;
}

.time-icon {
  color: #3b82f6;
}

/* 操作按钮 */
.action-btns {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.action-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 16px;
}

.action-btn.menu { background: #eff6ff; color: #3b82f6; }
.action-btn.menu:hover { background: #3b82f6; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3); }
.action-btn.resource { background: #f0fdf4; color: #16a34a; }
.action-btn.resource:hover { background: #16a34a; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3); }
.action-btn.delete { background: #fef2f2; color: #ef4444; }
.action-btn.delete:hover { background: #ef4444; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3); }

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light, #f3f4f6);
}

.pagination-wrapper :deep(.el-pager li) {
  border-radius: 8px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.pagination-wrapper :deep(.el-pager li:hover) {
  background: var(--bg-hover, #f3f4f6);
}

.pagination-wrapper :deep(.el-pager li.is-active) {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

/* 优雅对话框 */
.modern-dialog :deep(.el-dialog__header) {
  display: none;
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 32px 32px 24px;
}

.modern-dialog :deep(.el-dialog__footer) {
  padding: 0 32px 32px;
}

.dialog-icon-wrapper {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  margin: 0 auto 20px;
}

.dialog-icon-wrapper.primary { background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%); color: #3b82f6; }
.dialog-icon-wrapper.success { background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%); color: #16a34a; }
.dialog-icon-wrapper.danger { background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%); color: #ef4444; }

.dialog-content {
  text-align: center;
}

.dialog-content h3 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0 0 8px;
}

.dialog-content p {
  font-size: 14px;
  color: var(--text-secondary, #6b7280);
  margin: 0;
}

.dialog-footer {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.btn-cancel {
  border-radius: 10px;
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
}

.btn-confirm {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
  border-radius: 10px;
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
}

.btn-confirm:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.btn-confirm-danger {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  border-radius: 10px;
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
}

.btn-confirm-danger:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

/* 表单样式 */
.role-form {
  margin-top: 24px;
  text-align: left;
}

.role-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  padding-bottom: 8px;
}

.form-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  height: 44px;
}

.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2), 0 0 0 1px #3b82f6;
}

/* 树形控件 */
.tree-wrapper {
  border: 1px solid var(--border-default, #e5e7eb);
  border-radius: 12px;
  padding: 12px;
  background: var(--bg-elevated, #f9fafb);
  max-height: 260px;
  overflow-y: auto;
}

/* 深色模式 - 极客风 */
[data-theme="dark"] .stat-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 15px var(--primary-glow);
}

[data-theme="dark"] .stat-value {
  color: var(--text-primary, #f9fafb);
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
}
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .main-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(51, 65, 85, 0.7) 100%);
  color: #F8FAFC;
  border-bottom: 2px solid var(--neon-blue, #00D4FF);
}

[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) {
  background: rgba(0, 212, 255, 0.08) !important;
}

[data-theme="dark"] .role-avatar {
  background: rgba(0, 212, 255, 0.12);
  color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .role-name-text { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .role-badge {
  background: rgba(0, 255, 136, 0.12);
  color: var(--neon-green, #00FF88);
  box-shadow: 0 0 8px rgba(0, 255, 136, 0.2);
}

[data-theme="dark"] .time-icon { color: var(--neon-blue, #00D4FF); }
[data-theme="dark"] .time-cell {
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  font-size: 13px;
}

[data-theme="dark"] .action-btn.menu {
  background: rgba(0, 212, 255, 0.12);
  color: var(--neon-blue, #00D4FF);
}
[data-theme="dark"] .action-btn.menu:hover {
  background: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.5);
}
[data-theme="dark"] .action-btn.resource {
  background: rgba(0, 255, 136, 0.12);
  color: var(--neon-green, #00FF88);
}
[data-theme="dark"] .action-btn.resource:hover {
  background: var(--neon-green, #00FF88);
  box-shadow: 0 0 12px rgba(0, 255, 136, 0.5);
}
[data-theme="dark"] .action-btn.delete { background: rgba(239, 68, 68, 0.12); color: #F87171; }
[data-theme="dark"] .action-btn.delete:hover { box-shadow: 0 0 12px rgba(239, 68, 68, 0.5); }

[data-theme="dark"] .btn-add {
  background: linear-gradient(135deg, var(--neon-blue, #00D4FF) 0%, var(--neon-purple, #BF5AF2) 100%);
  box-shadow: 0 4px 14px rgba(0, 212, 255, 0.4);
}

[data-theme="dark"] .pagination-wrapper :deep(.el-pager li.is-active) {
  background: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.5);
}

[data-theme="dark"] .search-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15), 0 0 12px rgba(0, 212, 255, 0.2);
}

[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .dialog-icon-wrapper.primary {
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.15) 0%, rgba(59, 130, 246, 0.25) 100%);
  color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.3);
}
[data-theme="dark"] .tree-wrapper {
  background: var(--bg-elevated, #374151);
  border-color: rgba(0, 212, 255, 0.2);
  box-shadow: inset 0 0 20px rgba(0, 212, 255, 0.03);
}

[data-theme="dark"] .form-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15), 0 0 12px rgba(0, 212, 255, 0.2);
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .stat-card:last-child { grid-column: span 2; }
}

@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; }
  .stat-card:last-child { grid-column: span 1; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; }
  .btn-add, .btn-batch-delete { width: 100%; }
  .search-input { width: 100%; }
  .pagination-wrapper { justify-content: center; }
}

@media (max-width: 480px) {
  .main-card :deep(.el-card__body) { padding: 16px; }
  .stat-card { padding: 16px; }
  .stat-icon { width: 48px; height: 48px; font-size: 20px; }
  .stat-value { font-size: 24px; }
  .modern-dialog :deep(.el-dialog) { width: 92% !important; }
}
</style>

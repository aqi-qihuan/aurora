<template>
  <div class="resource-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Link /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ moduleCount }}</span>
          <span class="stat-label">模块总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ apiCount }}</span>
          <span class="stat-label">接口总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Unlock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ anonymousCount }}</span>
          <span class="stat-label">匿名接口</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openModel(null)" class="btn-add">
            <span>新增模块</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索资源名..."
            class="search-input"
            clearable
            @keyup.enter="listResources"
            @clear="listResources" />
          <el-button type="primary" :icon="Search" @click="listResources" circle />
        </div>
      </div>

      <!-- 现代化树形表格 -->
      <el-table
        v-loading="loading"
        :data="resources"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }">
        <el-table-column prop="resourceName" label="资源名" min-width="200">
          <template #default="{ row }">
            <div class="resource-name-cell">
              <div class="resource-icon" :class="{ 'is-module': !row.url }">
                <el-icon><component :is="row.url ? 'Link' : 'FolderOpened'" /></el-icon>
              </div>
              <span class="resource-name-text">{{ row.resourceName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="资源路径" min-width="260">
          <template #default="{ row }">
            <span v-if="row.url" class="resource-path">{{ row.url }}</span>
            <span v-else class="resource-path-empty">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="requestMethod" label="请求方式" width="110" align="center">
          <template #default="{ row }">
            <span v-if="row.requestMethod" class="method-badge" :class="row.requestMethod.toLowerCase()">
              {{ row.requestMethod }}
            </span>
            <span v-else class="method-badge module-badge">MOD</span>
          </template>
        </el-table-column>
        <el-table-column prop="isAnonymous" label="匿名访问" width="110" align="center">
          <template #default="{ row }">
            <div v-if="row.url" class="switch-cell">
              <span class="switch-label" :class="{ active: row.isAnonymous === 1 }">
                {{ row.isAnonymous === 1 ? '公开' : '受限' }}
              </span>
              <el-switch
                v-model="row.isAnonymous"
                :active-value="1"
                :inactive-value="0"
                @change="changeResource(row)" />
            </div>
            <span v-else class="switch-label">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDate(row.createTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="新增子项" placement="top" :show-after="500" v-if="row.children">
                <button class="action-btn add" @click="openAddResourceModel(row)"><el-icon><Plus /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="修改" placement="top" :show-after="500">
                <button class="action-btn edit" @click="openEditResourceModel(row)"><el-icon><Edit /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="删除" placement="top" :show-after="500">
                <button class="action-btn delete" @click="handleDelete(row.id)"><el-icon><Delete /></el-icon></button>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 模块对话框 -->
    <el-dialog v-model="addModule" width="450px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><FolderOpened /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ isEditModule ? '修改模块' : '添加模块' }}</h3>
        <el-form ref="resourceFormRef" :model="resourceForm" :rules="resourceRules" class="resource-form" label-position="top">
          <el-form-item label="模块名" prop="resourceName">
            <el-input v-model="resourceForm.resourceName" placeholder="请输入模块名" class="form-input" :prefix-icon="FolderOpened" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="addModule = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="addOrEditResource" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 资源对话框 -->
    <el-dialog v-model="addResource" width="450px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper success"><el-icon><Link /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ isEditResource ? '修改资源' : '添加资源' }}</h3>
        <el-form ref="resourceEditFormRef" :model="resourceForm" :rules="resourceRules" class="resource-form" label-position="top">
          <el-form-item label="资源名" prop="resourceName">
            <el-input v-model="resourceForm.resourceName" placeholder="请输入资源名" class="form-input" :prefix-icon="Link" />
          </el-form-item>
          <el-form-item label="资源路径">
            <el-input v-model="resourceForm.url" placeholder="/api/xxx" class="form-input" />
          </el-form-item>
          <el-form-item label="请求方式">
            <div class="method-select">
              <button
                v-for="method in ['GET', 'POST', 'PUT', 'DELETE']"
                :key="method"
                class="method-option"
                :class="[method.toLowerCase(), { active: resourceForm.requestMethod === method }]"
                @click="resourceForm.requestMethod = method">
                {{ method }}
              </button>
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="addResource = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="addOrEditResource" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Plus,
  Search,
  Clock,
  Edit,
  Delete,
  Link,
  FolderOpened,
  Document,
  Unlock
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import logger from '@/utils/logger'

const route = useRoute()

// 响应式数据
const loading = ref(true)
const keywords = ref('')
const resources = ref([])
const addModule = ref(false)
const addResource = ref(false)
const isEditModule = ref(false)
const isEditResource = ref(false)
const resourceForm = ref({})

const resourceRules = {
  resourceName: [{ required: true, message: '请输入资源名称', trigger: 'blur' }]
}

// 统计数据
const flatResources = (items) => {
  let result = []
  items.forEach(item => {
    result.push(item)
    if (item.children && item.children.length) {
      result = result.concat(flatResources(item.children))
    }
  })
  return result
}

const allFlat = computed(() => flatResources(resources.value))
const moduleCount = computed(() => allFlat.value.filter(r => !r.url).length)
const apiCount = computed(() => allFlat.value.filter(r => !!r.url).length)
const anonymousCount = computed(() => allFlat.value.filter(r => r.isAnonymous === 1).length)

// 日期格式化
const formatDate = (date) => {
  if (!date) return ''
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hour = String(d.getHours()).padStart(2, '0')
  const minute = String(d.getMinutes()).padStart(2, '0')
  const second = String(d.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

// 获取资源列表
const listResources = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin/resources', {
      params: {
        keywords: keywords.value
      }
    })
    resources.value = data.data
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '获取资源列表失败'
    })
  } finally {
    loading.value = false
  }
}

// 修改资源
const changeResource = async (resource) => {
  try {
    const { data } = await request.post('/admin/resources', resource)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listResources()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '修改资源失败' })
  }
}

// 打开模块对话框
const openModel = (resource) => {
  if (resource != null) {
    resourceForm.value = JSON.parse(JSON.stringify(resource))
    isEditModule.value = true
  } else {
    resourceForm.value = {}
    isEditModule.value = false
  }
  addModule.value = true
}

// 打开编辑资源对话框
const openEditResourceModel = (resource) => {
  if (resource.url == null) {
    openModel(resource)
    return false
  }
  resourceForm.value = JSON.parse(JSON.stringify(resource))
  isEditResource.value = true
  addResource.value = true
}

// 打开添加资源对话框
const openAddResourceModel = (resource) => {
  resourceForm.value = {}
  resourceForm.value.parentId = resource.id
  isEditResource.value = false
  addResource.value = true
}

// 处理删除
const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该资源吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteResource(id)
  }).catch(() => {})
}

// 删除资源
const deleteResource = async (id) => {
  try {
    const { data } = await request.delete(`/admin/resources/${id}`)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listResources()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '删除资源失败' })
  }
}

// 添加或修改资源
const addOrEditResource = async () => {
  if (!resourceForm.value.resourceName || resourceForm.value.resourceName.trim() === '') {
    ElMessage.error('资源名不能为空')
    return false
  }

  try {
    const { data } = await request.post('/admin/resources', resourceForm.value)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listResources()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    addModule.value = false
    addResource.value = false
  } catch (error) {
    ElMessage.error('保存资源失败')
    logger.error('API Error:', error)
  }
}

// 初始化
onMounted(() => {
  listResources()
})
</script>

<style scoped>
.resource-page {
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
.stat-icon.success { background: linear-gradient(135deg, #10b981 0%, #34d399 100%); color: #fff; }
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

/* 资源名单元格 */
.resource-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.resource-icon {
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

.resource-icon.is-module {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
  color: #16a34a;
}

.resource-name-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

/* 资源路径 */
.resource-path {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  background: var(--bg-elevated, #f3f4f6);
  padding: 4px 10px;
  border-radius: 6px;
  display: inline-block;
}

.resource-path-empty {
  color: var(--text-secondary, #9ca3af);
}

/* 请求方式徽章 */
.method-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.03em;
}

.method-badge.get { background: #eff6ff; color: #3b82f6; }
.method-badge.post { background: #f0fdf4; color: #16a34a; }
.method-badge.put { background: #fffbeb; color: #d97706; }
.method-badge.delete { background: #fef2f2; color: #ef4444; }
.method-badge.module-badge { background: #f3f4f6; color: #6b7280; }

/* 开关单元格 */
.switch-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
}

.switch-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
  min-width: 28px;
}

.switch-label.active {
  color: #16a34a;
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

.action-btn.add { background: #f0fdf4; color: #16a34a; }
.action-btn.add:hover { background: #16a34a; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3); }
.action-btn.edit { background: #eff6ff; color: #3b82f6; }
.action-btn.edit:hover { background: #3b82f6; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3); }
.action-btn.delete { background: #fef2f2; color: #ef4444; }
.action-btn.delete:hover { background: #ef4444; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3); }

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

/* 表单样式 */
.resource-form {
  margin-top: 24px;
  text-align: left;
}

.resource-form :deep(.el-form-item__label) {
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

/* 请求方式选择器 */
.method-select {
  display: flex;
  gap: 10px;
  width: 100%;
}

.method-option {
  flex: 1;
  padding: 10px 12px;
  border-radius: 10px;
  border: 2px solid var(--border-default, #e5e7eb);
  background: var(--bg-base, #fff);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s ease;
  color: var(--text-secondary, #6b7280);
}

.method-option.get.active { border-color: #3b82f6; background: #eff6ff; color: #3b82f6; }
.method-option.post.active { border-color: #16a34a; background: #f0fdf4; color: #16a34a; }
.method-option.put.active { border-color: #d97706; background: #fffbeb; color: #d97706; }
.method-option.delete.active { border-color: #ef4444; background: #fef2f2; color: #ef4444; }

.method-option:hover {
  transform: translateY(-1px);
}

/* 深色模式 */
[data-theme="dark"] .stat-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .stat-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .main-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) {
  background: var(--bg-elevated, #374151);
  color: var(--text-secondary, #9ca3af);
}

[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) {
  background: var(--bg-hover, #374151) !important;
}

[data-theme="dark"] .resource-icon { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .resource-icon.is-module { background: rgba(22, 163, 74, 0.15); }
[data-theme="dark"] .resource-name-text { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .resource-path { background: var(--bg-elevated, #374151); }

[data-theme="dark"] .method-badge.get { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .method-badge.post { background: rgba(22, 163, 74, 0.15); }
[data-theme="dark"] .method-badge.put { background: rgba(217, 119, 6, 0.15); }
[data-theme="dark"] .method-badge.delete { background: rgba(239, 68, 68, 0.15); }
[data-theme="dark"] .method-badge.module-badge { background: rgba(107, 114, 128, 0.15); }

[data-theme="dark"] .action-btn.add { background: rgba(22, 163, 74, 0.15); }
[data-theme="dark"] .action-btn.edit { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239, 68, 68, 0.15); }

[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .method-option {
  border-color: var(--border-default, #374151);
  background: var(--bg-base, #1f2937);
  color: var(--text-secondary, #9ca3af);
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
  .btn-add { width: 100%; }
  .search-input { width: 100%; }
  .method-select { flex-wrap: wrap; }
  .method-option { min-width: calc(50% - 5px); }
}

@media (max-width: 480px) {
  .main-card :deep(.el-card__body) { padding: 16px; }
  .stat-card { padding: 16px; }
  .stat-icon { width: 48px; height: 48px; font-size: 20px; }
  .stat-value { font-size: 24px; }
  .modern-dialog :deep(.el-dialog) { width: 92% !important; }
}
</style>

<template>
  <div class="tag-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><PriceTag /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">标签总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ totalArticles }}</span>
          <span class="stat-label">文章总量</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><TrendCharts /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ avgArticles }}</span>
          <span class="stat-label">平均文章数</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openModel(null)" class="btn-add">
            <span>新增标签</span>
          </el-button>
          <el-button
            type="danger"
            :icon="Delete"
            :disabled="tagIds.length === 0"
            @click="isDelete = true"
            class="btn-batch-delete">
            <span>批量删除 ({{ tagIds.length }})</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索标签名..."
            class="search-input"
            clearable
            @keyup.enter="searchTags"
            @clear="searchTags" />
          <el-button type="primary" :icon="Search" @click="searchTags" circle />
        </div>
      </div>

      <!-- 标签表格 -->
      <el-table
        :data="tags"
        v-loading="loading"
        @selection-change="selectionChange"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }"
        row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="tagName" label="标签名称" min-width="200" align="left">
          <template #default="{ row }">
            <div class="tag-name-cell">
              <div class="tag-color-dot" :style="{ background: getTagColor(row.tagName) }"></div>
              <span class="tag-name-text">{{ row.tagName }}</span>
              <el-tag
                size="small"
                :type="getTagType(row.tagName)"
                effect="light"
                class="tag-badge">
                {{ row.tagName.charAt(0).toUpperCase() }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="articleCount" label="文章数量" width="140" align="center" sortable>
          <template #default="{ row }">
            <div class="article-count-cell">
              <div class="count-bar-container">
                <div
                  class="count-bar"
                  :style="{
                    width: getBarWidth(row.articleCount) + '%',
                    background: getBarColor(row.articleCount)
                  }"></div>
              </div>
              <span class="count-value">{{ row.articleCount }}</span>
              <span class="count-unit">篇</span>
            </div>
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
        <el-table-column label="操作" width="140" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="编辑" placement="top" :show-after="500">
                <button class="action-btn edit" @click="openModel(row)">
                  <el-icon><Edit /></el-icon>
                </button>
              </el-tooltip>
              <el-tooltip content="删除" placement="top" :show-after="500">
                <button class="action-btn delete" @click="handleDelete(row.id)">
                  <el-icon><Delete /></el-icon>
                </button>
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
          :page-sizes="[10, 20, 50]"
          @size-change="sizeChange"
          @current-change="currentChange" />
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="isDelete" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger">
        <el-icon><Warning /></el-icon>
      </div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ tagIds.length }} 个标签吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isDelete = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteTag(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="addOrEdit" width="450px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary">
        <el-icon><EditPen /></el-icon>
      </div>
      <div class="dialog-content">
        <h3>{{ tagTitle }}</h3>
        <el-form ref="tagFormRef" :model="tagForm" :rules="tagRules" class="tag-form" label-position="top">
          <el-form-item label="标签名称" prop="tagName">
            <el-input
              v-model="tagForm.tagName"
              placeholder="请输入标签名称"
              class="form-input"
              :prefix-icon="PriceTag" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="addOrEdit = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="addOrEditTag" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Delete, 
  Search, 
  Edit, 
  Clock,
  PriceTag,
  Warning,
  EditPen,
  Document,
  TrendCharts
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import logger from '@/utils/logger'

const route = useRoute()
const pageStateStore = usePageStateStore()

// 响应式数据
const isDelete = ref(false)
const loading = ref(true)
const addOrEdit = ref(false)
const tagTitle = ref('')
const keywords = ref('')
const tags = ref([])
const tagIds = ref([])
const tagForm = reactive({
  id: null,
  tagName: ''
})
const tagRules = {
  tagName: [{ required: true, message: '请输入标签名', trigger: 'blur' }]
}
const current = ref(1)
const size = ref(10)
const count = ref(0)

// 计算统计数据
const totalArticles = computed(() => {
  return tags.value.reduce((sum, tag) => sum + (tag.articleCount || 0), 0)
})

const avgArticles = computed(() => {
  if (tags.value.length === 0) return 0
  return Math.round(totalArticles.value / tags.value.length)
})

// 日期格式化
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

// 获取标签类型颜色
const getTagType = (name) => {
  const colors = ['primary', 'success', 'info', 'warning', 'danger']
  const index = name.length % colors.length
  return colors[index]
}

// 获取标签颜色
const getTagColor = (name) => {
  const colors = [
    '#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444',
    '#06b6d4', '#ec4899', '#84cc16', '#f97316', '#6366f1'
  ]
  const index = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length
  return colors[index]
}

// 获取进度条宽度
const getBarWidth = (count) => {
  const max = Math.max(...tags.value.map(t => t.articleCount || 0), 1)
  return Math.min((count / max) * 100, 100)
}

// 获取进度条颜色
const getBarColor = (count) => {
  if (count >= 50) return 'linear-gradient(90deg, #ef4444, #f87171)'
  if (count >= 30) return 'linear-gradient(90deg, #f59e0b, #fbbf24)'
  if (count >= 10) return 'linear-gradient(90deg, #10b981, #34d399)'
  return 'linear-gradient(90deg, #3b82f6, #60a5fa)'
}

// 选择变化
const selectionChange = (selectedTags) => {
  tagIds.value = selectedTags.map(item => item.id)
}

// 搜索标签
const searchTags = () => {
  current.value = 1
  listTags()
}

// 分页大小变化
const sizeChange = (newSize) => {
  size.value = newSize
  listTags()
}

// 页码变化
const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('tag', newCurrent)
  listTags()
}

// 处理删除（带确认框）
const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该标签吗？删除后不可恢复。', '确认删除', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteTag(id)
  }).catch(() => {})
}

// 删除标签
const deleteTag = (id) => {
  const param = id == null ? { data: tagIds.value } : { data: [id] }
  request.delete('/admin/tags', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '删除成功',
        message: data.message
      })
      listTags()
    } else {
      ElNotification.error({
        title: '删除失败',
        message: data.message
      })
    }
  }).catch(error => {
    ElMessage.error('删除失败')
    logger.error('API Error:', error)
  })
  isDelete.value = false
}

// 获取标签列表
const listTags = () => {
  loading.value = true
  request.get('/admin/tags', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      tags.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取标签列表失败')
    logger.error('API Error:', error)
  })
}

// 打开对话框
const openModel = (tag) => {
  if (tag != null) {
    Object.assign(tagForm, JSON.parse(JSON.stringify(tag)))
    tagTitle.value = '编辑标签'
  } else {
    tagForm.id = null
    tagForm.tagName = ''
    tagTitle.value = '新增标签'
  }
  addOrEdit.value = true
}

// 添加或编辑标签
const addOrEditTag = () => {
  if (tagForm.tagName.trim() == '') {
    ElMessage.error('标签名不能为空')
    return false
  }
  request.post('/admin/tags', tagForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '保存成功',
        message: data.message
      })
      listTags()
    } else {
      ElNotification.error({
        title: '保存失败',
        message: data.message
      })
    }
    addOrEdit.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    logger.error('API Error:', error)
  })
}

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.tag
  listTags()
})
</script>

<style scoped>
/* ==================== Tag Page Modern Elegant Design ==================== */

.tag-page {
  padding: 0;
}

/* ==================== 统计卡片行 ==================== */
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
}

.stat-icon.primary {
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  color: #fff;
}

.stat-icon.success {
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
  color: #fff;
}

.stat-icon.warning {
  background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
  color: #fff;
}

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

/* ==================== 主卡片 ==================== */
.main-card {
  border-radius: 16px;
  border: 1px solid var(--border-default, #e5e7eb);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  background: var(--bg-base, #fff);
}

.main-card :deep(.el-card__body) {
  padding: 24px;
}

/* ==================== 工具栏 ==================== */
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

.search-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--primary, #3b82f6);
}

.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2), 0 0 0 1px var(--primary, #3b82f6);
}

/* ==================== 现代化表格 ==================== */
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

/* 标签名称单元格 */
.tag-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tag-color-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.tag-name-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

.tag-badge {
  margin-left: auto;
  font-weight: 600;
}

/* 文章数量单元格 */
.article-count-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.count-bar-container {
  flex: 1;
  height: 6px;
  background: var(--bg-elevated, #e5e7eb);
  border-radius: 3px;
  overflow: hidden;
  min-width: 60px;
}

.count-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
}

.count-value {
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  min-width: 24px;
  text-align: right;
}

.count-unit {
  color: var(--text-secondary, #6b7280);
  font-size: 12px;
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
  color: var(--primary, #3b82f6);
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

.action-btn.edit {
  background: #eff6ff;
  color: #3b82f6;
}

.action-btn.edit:hover {
  background: #3b82f6;
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.action-btn.delete {
  background: #fef2f2;
  color: #ef4444;
}

.action-btn.delete:hover {
  background: #ef4444;
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

/* ==================== 分页 ==================== */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light, #f3f4f6);
}

.pagination-wrapper :deep(.el-pagination) {
  gap: 8px;
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

/* ==================== 现代对话框 ==================== */
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

.dialog-icon-wrapper.primary {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #3b82f6;
}

.dialog-icon-wrapper.danger {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #ef4444;
}

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
.tag-form {
  margin-top: 24px;
  text-align: left;
}

.tag-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  padding-bottom: 8px;
}

.form-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  height: 44px;
}

.form-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--primary, #3b82f6);
}

.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2), 0 0 0 1px var(--primary, #3b82f6);
}

/* ==================== 深色模式 - 极客风 ==================== */
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

[data-theme="dark"] .stat-label {
  color: var(--text-secondary, #9ca3af);
}

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

[data-theme="dark"] .tag-name-text {
  color: var(--text-primary, #f9fafb);
}

[data-theme="dark"] .count-bar-container {
  background: var(--bg-elevated, #374151);
}

[data-theme="dark"] .count-value {
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  color: var(--neon-blue, #00D4FF);
}

[data-theme="dark"] .count-bar {
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.5);
}

[data-theme="dark"] .time-icon { color: var(--neon-blue, #00D4FF); }
[data-theme="dark"] .time-cell {
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  font-size: 13px;
}

[data-theme="dark"] .action-btn.edit {
  background: rgba(0, 212, 255, 0.12);
  color: var(--neon-blue, #00D4FF);
}

[data-theme="dark"] .action-btn.edit:hover {
  background: var(--neon-blue, #00D4FF);
  color: #fff;
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.5);
}

[data-theme="dark"] .action-btn.delete {
  background: rgba(239, 68, 68, 0.12);
  color: #F87171;
}

[data-theme="dark"] .action-btn.delete:hover {
  box-shadow: 0 0 12px rgba(239, 68, 68, 0.5);
}

[data-theme="dark"] .btn-add {
  background: linear-gradient(135deg, var(--neon-blue, #00D4FF) 0%, var(--neon-purple, #BF5AF2) 100%);
  box-shadow: 0 4px 14px rgba(0, 212, 255, 0.4);
}

[data-theme="dark"] .btn-add:hover {
  box-shadow: 0 6px 20px rgba(0, 212, 255, 0.6);
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

[data-theme="dark"] .form-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15), 0 0 12px rgba(0, 212, 255, 0.2);
}

/* ==================== 响应式设计 ==================== */
@media (max-width: 1024px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }

  .stat-card:last-child {
    grid-column: span 2;
  }
}

@media (max-width: 768px) {
  .stats-row {
    grid-template-columns: 1fr;
  }

  .stat-card:last-child {
    grid-column: span 1;
  }

  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }

  .toolbar-left {
    flex-direction: column;
  }

  .btn-add,
  .btn-batch-delete {
    width: 100%;
  }

  .search-input {
    width: 100%;
  }

  .pagination-wrapper {
    justify-content: center;
  }

  .pagination-wrapper :deep(.el-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .main-card :deep(.el-card__body) {
    padding: 16px;
  }

  .stat-card {
    padding: 16px;
  }

  .stat-icon {
    width: 48px;
    height: 48px;
    font-size: 20px;
  }

  .stat-value {
    font-size: 24px;
  }

  .tag-name-cell {
    flex-wrap: wrap;
  }

  .tag-badge {
    margin-left: 0;
  }

  .article-count-cell {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .count-bar-container {
    width: 100%;
  }
}
</style>

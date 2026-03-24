<template>
  <div class="category-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><FolderOpened /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">分类总数</span>
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
            <span>新增分类</span>
          </el-button>
          <el-button
            type="danger"
            :icon="Delete"
            :disabled="categoryIds.length === 0"
            @click="isDelete = true"
            class="btn-batch-delete">
            <span>批量删除 ({{ categoryIds.length }})</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索分类名..."
            class="search-input"
            clearable
            @keyup.enter="searchCategories"
            @clear="searchCategories" />
          <el-button type="primary" :icon="Search" @click="searchCategories" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table
        :data="categories"
        v-loading="loading"
        @selection-change="selectionChange"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }"
        row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="categoryName" label="分类名称" min-width="200" align="left">
          <template #default="{ row }">
            <div class="category-name-cell">
              <div class="category-color-dot" :style="{ background: getCategoryColor(row.categoryName) }"></div>
              <el-icon :size="18" class="category-icon"><FolderOpened /></el-icon>
              <span class="category-name-text">{{ row.categoryName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="articleCount" label="文章数量" width="140" align="center" sortable>
          <template #default="{ row }">
            <div class="article-count-cell">
              <div class="count-bar-container">
                <div class="count-bar" :style="{ width: getBarWidth(row.articleCount) + '%', background: getBarColor(row.articleCount) }"></div>
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
                <button class="action-btn edit" @click="openModel(row)"><el-icon><Edit /></el-icon></button>
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
        <el-pagination background layout="total, sizes, prev, pager, next, jumper" :total="count" :page-size="size" :current-page="current" :page-sizes="[10, 20]" @size-change="sizeChange" @current-change="currentChange" />
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="isDelete" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ categoryIds.length }} 个分类吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isDelete = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteCategory(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 新增/编辑分类对话框 -->
    <el-dialog v-model="addOrEdit" width="450px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><EditPen /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ categoryTitle }}</h3>
        <el-form ref="categoryFormRef" :model="categoryForm" :rules="categoryRules" class="category-form" label-position="top">
          <el-form-item label="分类名称" prop="categoryName">
            <el-input v-model="categoryForm.categoryName" placeholder="请输入分类名称" class="form-input" :prefix-icon="FolderOpened" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="addOrEdit = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="addOrEditCategory" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageStateStore } from '@/stores/pageState'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Plus,
  Delete,
  Search,
  FolderOpened,
  Clock,
  Edit,
  EditPen,
  Warning,
  Document,
  TrendCharts
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()

const isDelete = ref(false)
const loading = ref(true)
const addOrEdit = ref(false)
const categoryTitle = ref('')
const keywords = ref(null)
const categoryIds = ref([])
const categories = ref([])
const categoryForm = ref({
  id: null,
  categoryName: ''
})
const categoryRules = {
  categoryName: [{ required: true, message: '请输入分类名', trigger: 'blur' }]
}
const current = ref(1)
const size = ref(10)
const count = ref(0)

const totalArticles = computed(() => categories.value.reduce((sum, cat) => sum + (cat.articleCount || 0), 0))
const avgArticles = computed(() => categories.value.length === 0 ? 0 : Math.round(totalArticles.value / categories.value.length))

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

const getCategoryColor = (name) => {
  const colors = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899', '#84cc16', '#f97316', '#6366f1']
  return colors[name.split('').reduce((a, ch) => a + ch.charCodeAt(0), 0) % colors.length]
}

const getBarWidth = (n) => {
  const max = Math.max(...categories.value.map(c => c.articleCount || 0), 1)
  return Math.min((n / max) * 100, 100)
}

const getBarColor = (n) => {
  if (n >= 50) return 'linear-gradient(90deg, #ef4444, #f87171)'
  if (n >= 30) return 'linear-gradient(90deg, #f59e0b, #fbbf24)'
  if (n >= 10) return 'linear-gradient(90deg, #10b981, #34d399)'
  return 'linear-gradient(90deg, #3b82f6, #60a5fa)'
}

const selectionChange = (selection) => {
  categoryIds.value = selection.map((item) => item.id)
}

const searchCategories = () => {
  current.value = 1
  listCategories()
}

const sizeChange = (val) => {
  size.value = val
  listCategories()
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updatePageState('category', val)
  listCategories()
}

const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该分类吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteCategory(id)
  }).catch(() => {})
}

const deleteCategory = async (id) => {
  const param = id ? { data: [id] } : { data: categoryIds.value }

  try {
    const { data } = await request.delete('/admin/categories', param)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listCategories()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    isDelete.value = false
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '删除失败' })
  }
}

const listCategories = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin/categories', {
      params: {
        current: current.value,
        size: size.value,
        keywords: keywords.value
      }
    })
    if (data && data.data) {
      categories.value = data.data.records || []
      count.value = data.data.count || 0
    } else {
      categories.value = []
      count.value = 0
    }
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '获取分类列表失败' })
    categories.value = []
    count.value = 0
  } finally {
    loading.value = false
  }
}

const openModel = (category) => {
  if (category != null) {
    categoryForm.value = JSON.parse(JSON.stringify(category))
    categoryTitle.value = '编辑分类'
  } else {
    categoryForm.value.id = null
    categoryForm.value.categoryName = ''
    categoryTitle.value = '添加分类'
  }
  addOrEdit.value = true
}

const addOrEditCategory = async () => {
  if (categoryForm.value.categoryName.trim() === '') {
    ElMessage.error('分类名不能为空')
    return
  }

  try {
    const { data } = await request.post('/admin/categories', categoryForm.value)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listCategories()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    addOrEdit.value = false
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '操作失败' })
  }
}

onMounted(() => {
  current.value = pageStateStore.category
  listCategories()
})
</script>

<style scoped>
.category-page {
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

/* 分类名称单元格 */
.category-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-color-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.category-icon {
  color: #8b5cf6;
}

.category-name-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

/* 文章数量进度条 */
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

.action-btn.edit { background: #eff6ff; color: #3b82f6; }
.action-btn.edit:hover { background: #3b82f6; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3); }
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
.category-form {
  margin-top: 24px;
  text-align: left;
}

.category-form :deep(.el-form-item__label) {
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

[data-theme="dark"] .category-name-text { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .count-bar-container { background: var(--bg-elevated, #374151); }

[data-theme="dark"] .action-btn.edit { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239, 68, 68, 0.15); }

[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }

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
}
</style>

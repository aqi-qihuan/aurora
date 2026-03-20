<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || '分类管理' }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" @click="openModel(null)">
        <el-icon><Plus /></el-icon>
        新增
      </el-button>
      <el-button
        type="danger"
        size="small"
        :disabled="categoryIds.length === 0"
        @click="isDelete = true">
        <el-icon><Delete /></el-icon>
        批量删除
      </el-button>
      <div class="search-container">
        <el-input
          v-model="keywords"
          size="small"
          placeholder="请输入分类名"
          style="width: 200px"
          @keyup.enter="searchCategories">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button
          type="primary"
          size="small"
          style="margin-left: 1rem"
          @click="searchCategories">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="categories"
      @selection-change="selectionChange"
      v-loading="loading"
      class="category-table"
      :header-cell-style="{ background: 'var(--bg-elevated)', color: 'var(--text-primary)', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="categoryName" label="分类名" align="center" min-width="150">
        <template #default="scope">
          <div class="category-name">
            <el-icon :size="18" style="margin-right: 8px; color: #409eff">
              <FolderOpened />
            </el-icon>
            {{ scope.row.categoryName }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="articleCount" label="文章量" align="center" width="120" sortable>
        <template #default="scope">
          <el-tag
            size="small"
            :type="getArticleCountType(scope.row.articleCount)"
            effect="plain">
            {{ scope.row.articleCount }} 篇
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center" width="160" sortable>
        <template #default="scope">
          <div class="create-time">
            <el-icon><Clock /></el-icon>
            {{ formatDate(scope.row.createTime) }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template #default="scope">
          <div class="action-buttons">
            <el-button
              type="primary"
              size="small"
              circle
              @click="openModel(scope.row)">
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-popconfirm
              title="确定删除吗？"
              @confirm="deleteCategory(scope.row.id)">
              <template #reference>
                <el-button size="small" type="danger" circle>
                  <el-icon><Delete /></el-icon>
                </el-button>
              </template>
            </el-popconfirm>
          </div>
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
    <el-dialog v-model="isDelete" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <template #footer>
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteCategory(null)">确 定</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="addOrEdit" width="30%">
      <template #header>
        <div class="dialog-title-container" ref="categoryTitleRef" />
      </template>
      <el-form label-width="80px" size="medium" :model="categoryForm">
        <el-form-item label="分类名">
          <el-input v-model="categoryForm.categoryName" style="width: 220px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditCategory">确 定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageStateStore } from '@/stores/pageState'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Plus,
  Delete,
  Search,
  FolderOpened,
  Clock,
  Edit,
  Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()

// 响应式数据
const isDelete = ref(false)
const loading = ref(true)
const addOrEdit = ref(false)
const keywords = ref(null)
const categoryIds = ref([])
const categories = ref([])
const categoryForm = ref({
  id: null,
  categoryName: ''
})
const current = ref(1)
const size = ref(10)
const count = ref(0)
const categoryTitleRef = ref(null)

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

// 获取文章数量标签类型
const getArticleCountType = (count) => {
  if (count >= 50) return 'danger'
  if (count >= 30) return 'warning'
  if (count >= 10) return 'success'
  return 'info'
}

// 选择变化
const selectionChange = (selection) => {
  categoryIds.value = selection.map((item) => item.id)
}

// 搜索分类
const searchCategories = () => {
  current.value = 1
  listCategories()
}

// 每页数量变化
const sizeChange = (val) => {
  size.value = val
  listCategories()
}

// 当前页变化
const currentChange = (val) => {
  current.value = val
  pageStateStore.updatePageState('category', val)
  listCategories()
}

// 删除分类
const deleteCategory = async (id) => {
  const param = id ? { data: [id] } : { data: categoryIds.value }
  
  try {
    const { data } = await request.delete('/admin/categories', param)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listCategories()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isDelete.value = false
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '删除失败'
    })
  }
}

// 获取分类列表
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
    ElNotification.error({
      title: '失败',
      message: error.message || '获取分类列表失败'
    })
    categories.value = []
    count.value = 0
  } finally {
    loading.value = false
  }
}

// 打开对话框
const openModel = (category) => {
  if (category != null) {
    categoryForm.value = JSON.parse(JSON.stringify(category))
    nextTick(() => {
      if (categoryTitleRef.value) {
        categoryTitleRef.value.innerHTML = '修改分类'
      }
    })
  } else {
    categoryForm.value.id = null
    categoryForm.value.categoryName = ''
    nextTick(() => {
      if (categoryTitleRef.value) {
        categoryTitleRef.value.innerHTML = '添加分类'
      }
    })
  }
  addOrEdit.value = true
}

// 添加或修改分类
const addOrEditCategory = async () => {
  if (categoryForm.value.categoryName.trim() === '') {
    ElMessage.error('分类名不能为空')
    return
  }
  
  try {
    const { data } = await request.post('/admin/categories', categoryForm.value)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listCategories()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    addOrEdit.value = false
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '操作失败'
    })
  }
}

// 初始化
onMounted(() => {
  current.value = pageStateStore.category
  listCategories()
})
</script>

<style scoped>
/* ==================== Category Page Modern Styles ==================== */

/* 页面标题 */
.title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text);
  margin-bottom: var(--space-6);
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.title::before {
  content: '';
  width: 4px;
  height: 24px;
  background: linear-gradient(180deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border-radius: var(--radius-full);
}

/* 操作区域 */
.operation-container {
  margin-top: var(--space-6);
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--color-bg-hover);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}

.operation-container .el-button {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-button:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.operation-container .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

.operation-container .el-button--danger {
  background: linear-gradient(135deg, var(--color-error) 0%, #f87171 100%);
  border: none;
}

/* 搜索区域 */
.search-container {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-left: auto;
}

.search-container .el-input {
  width: 200px;
}

.search-container .el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.search-container .el-input :deep(.el-input__inner:focus) {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 分类表格 */
.category-table {
  margin-top: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.category-table :deep(.el-table__header-wrapper) {
  background: var(--color-bg-hover);
}

.category-table :deep(.el-table__header th) {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.category-table :deep(.el-table__body td) {
  padding: var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.category-table :deep(.el-table__body tr) {
  transition: all var(--duration-fast) var(--ease-out);
}

.category-table :deep(.el-table__body tr:hover > td) {
  background-color: var(--color-primary-50) !important;
}

.category-table :deep(.el-table__row) {
  animation: slideIn var(--duration-base) var(--ease-out);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 分类名称 */
.category-name {
  font-size: var(--text-sm);
  color: var(--color-text);
  font-weight: var(--font-semibold);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
}

.category-name .el-icon {
  color: var(--color-primary);
}

.category-name:hover {
  color: var(--color-primary);
}

/* 文章数量标签 */
.category-table :deep(.el-tag) {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
}

.category-table :deep(.el-tag:hover) {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 创建时间 */
.create-time {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
}

.create-time .el-icon {
  color: var(--color-secondary);
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2);
}

.action-buttons .el-button {
  transition: all var(--duration-fast) var(--ease-out);
  border-radius: var(--radius-base);
}

.action-buttons .el-button:hover {
  transform: translateY(-2px) scale(1.05);
  box-shadow: var(--shadow-md);
}

.action-buttons .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

.action-buttons .el-button--danger {
  background: linear-gradient(135deg, var(--color-error) 0%, #f87171 100%);
  border: none;
}

/* 分页 */
.pagination-container {
  float: right;
  margin-top: var(--space-6);
  margin-bottom: var(--space-4);
}

.pagination-container :deep(.el-pagination) {
  font-weight: var(--font-medium);
}

.pagination-container :deep(.el-pagination .el-pager li) {
  border-radius: var(--radius-base);
  transition: all var(--duration-fast) var(--ease-out);
}

.pagination-container :deep(.el-pagination .el-pager li.active) {
  background: var(--color-primary);
}

.pagination-container :deep(.el-pagination .el-pager li:hover) {
  transform: translateY(-1px);
}

.pagination-container :deep(.el-pagination button) {
  border-radius: var(--radius-base);
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  font-weight: var(--font-bold);
  font-size: var(--text-lg);
  color: var(--color-text);
}

.dialog-title-container .el-icon {
  font-size: var(--text-2xl);
  margin-right: var(--space-2);
  color: var(--color-warning);
}

/* 表单样式 */
:deep(.el-form-item__label) {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

:deep(.el-input .el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

:deep(.el-input .el-input__inner:focus) {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 加载动画 */
.category-table :deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .category-table :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.9);
}

/* ==================== Responsive ==================== */
@media (max-width: 768px) {
  .title {
    font-size: var(--text-xl);
  }

  .operation-container {
    flex-direction: column;
    align-items: stretch;
  }

  .search-container {
    margin-left: 0;
    width: 100%;
  }

  .search-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .category-table :deep(.el-table__header) {
    display: none;
  }

  .category-table :deep(.el-table__row) {
    display: flex;
    flex-direction: column;
    padding: var(--space-4);
    margin-bottom: var(--space-3);
    background: var(--color-bg-card);
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border);
    box-shadow: var(--shadow-sm);
  }

  .category-table :deep(.el-table__row td) {
    border: none;
    padding: var(--space-2) 0 !important;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .category-table :deep(.el-table__row td::before) {
    content: attr(data-label);
    font-weight: var(--font-semibold);
    color: var(--color-text-secondary);
    font-size: var(--text-xs);
  }
}
</style>

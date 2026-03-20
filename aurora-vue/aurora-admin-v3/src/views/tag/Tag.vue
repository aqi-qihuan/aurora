<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" :icon="Plus" @click="openModel(null)"> 新增 </el-button>
      <el-button
        type="danger"
        size="small"
        :icon="Delete"
        :disabled="tagIds.length == 0"
        @click="isDelete = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          :prefix-icon="Search"
          size="small"
          placeholder="请输入标签名"
          style="width: 200px"
          @keyup.enter="searchTags" />
        <el-button type="primary" size="small" :icon="Search" style="margin-left: 1rem" @click="searchTags">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="tags"
      v-loading="loading"
      @selection-change="selectionChange"
      class="tag-table"
      :header-cell-style="{ background: 'var(--bg-elevated)', color: 'var(--text-primary)', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="tagName" label="标签名" align="center" min-width="150">
        <template #default="{ row }">
          <el-tag
            size="medium"
            :type="getTagType(row.tagName)"
            effect="plain"
            class="tag-item">
            <el-icon><PriceTag /></el-icon>
            {{ row.tagName }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="articleCount" label="文章量" align="center" width="120" sortable>
        <template #default="{ row }">
          <el-tag size="small" :type="getArticleCountType(row.articleCount)" effect="plain">
            {{ row.articleCount }} 篇
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center" width="160" sortable>
        <template #default="{ row }">
          <div class="create-time">
            <el-icon><Clock /></el-icon>
            {{ formatDate(row.createTime) }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button
              type="primary"
              size="small"
              :icon="Edit"
              @click="openModel(row)"
              circle />
            <el-popconfirm
              title="确定删除吗？"
              @confirm="deleteTag(row.id)">
              <template #reference>
                <el-button
                  size="small"
                  type="danger"
                  :icon="Delete"
                  circle />
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
          <el-icon style="color: #ff9900; font-size: 1.5rem; margin-right: 8px"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <template #footer>
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteTag(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="addOrEdit" width="30%">
      <template #header>
        <div class="dialog-title-container">{{ tagTitle }}</div>
      </template>
      <el-form label-width="80px" size="medium" :model="tagForm">
        <el-form-item label="标签名">
          <el-input style="width: 220px" v-model="tagForm.tagName" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditTag"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { 
  Plus, 
  Delete, 
  Search, 
  Edit, 
  Clock,
  PriceTag,
  Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'

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
const current = ref(1)
const size = ref(10)
const count = ref(0)

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

// 获取文章数量类型
const getArticleCountType = (count) => {
  if (count >= 50) return 'danger'
  if (count >= 30) return 'warning'
  if (count >= 10) return 'success'
  return 'info'
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

// 删除标签
const deleteTag = (id) => {
  const param = id == null ? { data: tagIds.value } : { data: [id] }
  request.delete('/admin/tags', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listTags()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
  }).catch(error => {
    ElMessage.error('删除失败')
    console.error('API Error:', error)
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
    console.error('API Error:', error)
  })
}

// 打开对话框
const openModel = (tag) => {
  if (tag != null) {
    Object.assign(tagForm, JSON.parse(JSON.stringify(tag)))
    tagTitle.value = '修改标签'
  } else {
    tagForm.id = null
    tagForm.tagName = ''
    tagTitle.value = '添加标签'
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
        title: '成功',
        message: data.message
      })
      listTags()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    addOrEdit.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    console.error('API Error:', error)
  })
}

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.tag
  listTags()
})
</script>

<style scoped>
/* ==================== Tag Page Modern Styles ====================
 * 基于 UI/UX Pro Max 设计系统
 * 配色: Primary #2563EB, CTA #F97316
 */

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

/* 操作区域 - 现代化工具栏 */
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
.operation-container > div:last-child {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-left: auto;
}

.operation-container .el-input {
  width: 200px;
}

.operation-container .el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-input :deep(.el-input__inner):focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 标签表格 - 现代化数据表格 */
.tag-table {
  margin-top: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.tag-table :deep(.el-table__header-wrapper) {
  background: var(--color-bg-hover);
}

.tag-table :deep(.el-table__header th) {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.tag-table :deep(.el-table__body td) {
  padding: var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.tag-table :deep(.el-table__body tr) {
  transition: all var(--duration-fast) var(--ease-out);
}

.tag-table :deep(.el-table__body tr:hover > td) {
  background-color: var(--color-primary-50) !important;
}

.tag-table :deep(.el-table__row) {
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

/* 标签样式 */
.tag-item {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  transition: all var(--duration-base) var(--ease-out);
  cursor: default;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-base);
}

.tag-item:hover {
  transform: scale(1.05) translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 文章数量标签 */
.tag-table :deep(.el-tag) {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
}

.tag-table :deep(.el-tag):hover {
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

/* 分页 - 现代化样式 */
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

.pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background: var(--color-primary);
}

.pagination-container :deep(.el-pagination .el-pager li):hover {
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

.el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-input :deep(.el-input__inner):focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 加载动画 */
.tag-table :deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .tag-table :deep(.el-loading-mask) {
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

  .operation-container > div:last-child {
    margin-left: 0;
    width: 100%;
  }

  .operation-container .el-input {
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
  .tag-table :deep(.el-table__header) {
    display: none;
  }

  .tag-table :deep(.el-table__row) {
    display: flex;
    flex-direction: column;
    padding: var(--space-4);
    margin-bottom: var(--space-3);
    background: var(--color-bg-card);
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border);
    box-shadow: var(--shadow-sm);
  }

  .tag-table :deep(.el-table__row td) {
    border: none;
    padding: var(--space-2) 0 !important;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .tag-table :deep(.el-table__row td::before) {
    content: attr(data-label);
    font-weight: var(--font-semibold);
    color: var(--color-text-secondary);
    font-size: var(--text-xs);
  }

  .tag-item {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>

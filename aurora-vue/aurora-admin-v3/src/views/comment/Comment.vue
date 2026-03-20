<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="review-menu">
      <span>状态</span>
      <span @click="changeReview(null)" :class="isReview == null ? 'active-review' : 'review'"> 全部 </span>
      <span @click="changeReview(1)" :class="isReview == 1 ? 'active-review' : 'review'"> 正常 </span>
      <span @click="changeReview(0)" :class="isReview == 0 ? 'active-review' : 'review'"> 审核中 </span>
    </div>
    <div class="operation-container">
      <el-button
        type="danger"
        size="small"
        :icon="Delete"
        :disabled="commentIds.length == 0"
        @click="remove = true">
        批量删除
      </el-button>
      <el-button
        type="success"
        size="small"
        :icon="SuccessFilled"
        :disabled="commentIds.length == 0"
        @click="updateCommentReview(null)">
        批量通过
      </el-button>
      <div style="margin-left: auto">
        <el-select clearable v-model="type" placeholder="请选择来源" size="small" style="margin-right: 1rem">
          <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-input
          v-model="keywords"
          :prefix-icon="Search"
          size="small"
          placeholder="请输入用户昵称"
          style="width: 200px"
          @keyup.enter="searchComments" />
        <el-button type="primary" size="small" :icon="Search" style="margin-left: 1rem" @click="searchComments">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="comments"
      @selection-change="selectionChange"
      v-loading="loading"
      class="comment-table"
      :header-cell-style="{ background: 'var(--bg-elevated)', color: 'var(--text-primary)', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="avatar" label="头像" align="center" width="100">
        <template #default="{ row }">
          <el-avatar :size="40" :src="row.avatar" />
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="评论人" align="center" width="120">
        <template #default="{ row }">
          <div class="nickname">
            <el-icon style="margin-right: 5px; color: #409eff"><User /></el-icon>
            {{ row.nickname }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="replyNickname" label="回复人" align="center" width="120">
        <template #default="{ row }">
          <div class="reply-nickname">
            <el-icon v-if="row.replyNickname"><ChatLineRound /></el-icon>
            {{ row.replyNickname || '无' }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="articleTitle" label="文章标题" align="center" min-width="180">
        <template #default="{ row }">
          <el-tooltip :content="row.articleTitle" placement="top" :disabled="!row.articleTitle || row.articleTitle.length <= 20">
            <div class="article-title">
              <el-icon><Document /></el-icon>
              {{ row.articleTitle || '无' }}
            </div>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column prop="commentContent" label="评论内容" align="center" min-width="200">
        <template #default="{ row }">
          <div class="comment-content" v-html="sanitizeHtml(row.commentContent)" />
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="评论时间" width="160" align="center" sortable>
        <template #default="{ row }">
          <div class="create-time">
            <el-icon><Clock /></el-icon>
            {{ formatDate(row.createTime) }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="isReview" label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag
            v-if="row.isReview == 0"
            type="warning"
            size="small"
            effect="plain">
            <el-icon><Loading /></el-icon> 审核中
          </el-tag>
          <el-tag
            v-if="row.isReview == 1"
            type="success"
            size="small"
            effect="plain">
            <el-icon><Check /></el-icon> 正常
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="来源" align="center" width="100">
        <template #default="{ row }">
          <el-tag
            :type="getSourceType(row.type).tagType"
            size="small"
            effect="plain">
            <el-icon><component :is="getSourceType(row.type).icon" /></el-icon>
            {{ getSourceType(row.type).name }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="180" fixed="right">
        <template #default="{ row }">
          <div class="action-buttons">
            <el-button
              v-if="row.isReview == 0"
              type="success"
              size="small"
              :icon="Check"
              circle
              @click="updateCommentReview(row.id)" />
            <el-popconfirm
              title="确定删除吗？"
              @confirm="deleteComments(row.id)">
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
    <el-dialog v-model="remove" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; font-size: 1.5rem; margin-right: 8px"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否彻底删除选中项？</div>
      <template #footer>
        <el-button @click="remove = false">取 消</el-button>
        <el-button type="primary" @click="deleteComments(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { 
  Delete, 
  Search, 
  SuccessFilled, 
  User,
  ChatLineRound,
  Document,
  Clock,
  Loading,
  Check,
  Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import DOMPurify from 'dompurify'

const route = useRoute()
const pageStateStore = usePageStateStore()

// XSS 防护 - HTML 消毒
const sanitizeHtml = (html) => {
  if (!html) return ''
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'br', 'p', 'span'],
    ALLOWED_ATTR: ['href', 'title', 'target', 'class']
  })
}

// 响应式数据
const loading = ref(true)
const remove = ref(false)
const options = [
  { value: 1, label: '文章' },
  { value: 2, label: '留言' },
  { value: 3, label: '关于我' },
  { value: 4, label: '友链' },
  { value: 5, label: '说说' }
]
const comments = ref([])
const commentIds = ref([])
const type = ref(null)
const keywords = ref('')
const isReview = ref(null)
const current = ref(1)
const size = ref(10)
const count = ref(0)

// 日期格式化
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

// 获取来源类型
const getSourceType = (type) => {
  const types = {
    1: { name: '文章', tagType: 'primary', icon: 'Document' },
    2: { name: '留言', tagType: 'danger', icon: 'ChatLineRound' },
    3: { name: '关于我', tagType: 'success', icon: 'User' },
    4: { name: '友链', tagType: 'warning', icon: 'Link' },
    5: { name: '说说', tagType: 'info', icon: 'Edit' }
  }
  return types[type] || { name: '未知', tagType: 'info', icon: 'QuestionFilled' }
}

// 选择变化
const selectionChange = (selectedComments) => {
  commentIds.value = selectedComments.map(item => item.id)
}

// 搜索评论
const searchComments = () => {
  current.value = 1
  listComments()
}

// 分页大小变化
const sizeChange = (newSize) => {
  size.value = newSize
  listComments()
}

// 页码变化
const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('comment', newCurrent)
  listComments()
}

// 切换审核状态筛选
const changeReview = (review) => {
  current.value = 1
  isReview.value = review
}

// 更新评论审核状态
const updateCommentReview = (id) => {
  const param = { isReview: 1 }
  param.ids = id != null ? [id] : commentIds.value
  request.put('/admin/comments/review', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listComments()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
  }).catch(error => {
    ElMessage.error('审核失败')
    console.error('API Error:', error)
  })
}

// 删除评论
const deleteComments = (id) => {
  const param = id == null ? { data: commentIds.value } : { data: [id] }
  request.delete('/admin/comments', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listComments()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    remove.value = false
  }).catch(error => {
    ElMessage.error('删除评论失败')
    console.error('API Error:', error)
  })
}

// 获取评论列表
const listComments = () => {
  loading.value = true
  request.get('/admin/comments', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value,
      type: type.value,
      isReview: isReview.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      comments.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取评论列表失败')
    console.error('API Error:', error)
  })
}

// 监听筛选变化
watch([isReview, type], () => {
  current.value = 1
  listComments()
})

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.comment
  listComments()
})
</script>

<style scoped>
/* ==================== Comment Page Modern Styles ====================
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

/* 审核菜单 - 现代化标签页样式 */
.review-menu {
  font-size: var(--text-sm);
  margin-top: var(--space-4);
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  padding: var(--space-3) 0;
  border-bottom: 2px solid var(--color-border);
  gap: var(--space-2);
}

.review-menu > span:first-child {
  font-weight: var(--font-semibold);
  color: var(--color-text);
  margin-right: var(--space-4);
}

.review-menu span {
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  transition: all var(--duration-base) var(--ease-out);
  position: relative;
  cursor: pointer;
}

.review {
  color: var(--color-text-secondary);
  background: transparent;
}

.review:hover {
  color: var(--color-primary);
  background: var(--color-primary-50);
}

.active-review {
  color: var(--color-primary);
  font-weight: var(--font-semibold);
  background: var(--color-primary-50);
}

.active-review::after {
  content: '';
  position: absolute;
  bottom: -11px;
  left: 50%;
  transform: translateX(-50%);
  width: 24px;
  height: 3px;
  background: var(--color-primary);
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

.operation-container .el-button--danger {
  background: linear-gradient(135deg, var(--color-error) 0%, #f87171 100%);
  border: none;
}

.operation-container .el-button--success {
  background: linear-gradient(135deg, var(--color-success) 0%, #34d399 100%);
  border: none;
}

.operation-container .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

/* 搜索区域 */
.operation-container > div:last-child {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-left: auto;
}

.operation-container .el-select,
.operation-container .el-input {
  width: 180px;
}

.operation-container .el-input :deep(.el-input__inner),
.operation-container .el-select :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-input :deep(.el-input__inner):focus,
.operation-container .el-select :deep(.el-input__inner):focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 评论表格 - 现代化数据表格 */
.comment-table {
  margin-top: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.comment-table :deep(.el-table__header-wrapper) {
  background: var(--color-bg-hover);
}

.comment-table :deep(.el-table__header th) {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.comment-table :deep(.el-table__body td) {
  padding: var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.comment-table :deep(.el-table__body tr) {
  transition: all var(--duration-fast) var(--ease-out);
}

.comment-table :deep(.el-table__body tr:hover > td) {
  background-color: var(--color-primary-50) !important;
}

.comment-table :deep(.el-table__row) {
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

/* 头像 */
.comment-table :deep(.el-avatar) {
  border: 2px solid var(--color-border);
  box-shadow: var(--shadow-sm);
  transition: all var(--duration-base) var(--ease-out);
}

.comment-table :deep(.el-avatar):hover {
  transform: scale(1.15);
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}

/* 昵称 */
.nickname {
  font-size: var(--text-sm);
  color: var(--color-text);
  font-weight: var(--font-semibold);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
}

.nickname .el-icon {
  color: var(--color-primary);
}

.reply-nickname {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
}

.reply-nickname .el-icon {
  color: var(--color-secondary);
}

/* 文章标题 */
.article-title {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  transition: color var(--duration-fast) var(--ease-out);
}

.article-title .el-icon {
  color: var(--color-primary);
  flex-shrink: 0;
}

.article-title:hover {
  color: var(--color-primary);
}

/* 评论内容 */
.comment-content {
  display: inline-block;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--text-sm);
  color: var(--color-text);
  line-height: var(--leading-relaxed);
  padding: var(--space-2);
  background: var(--color-bg-hover);
  border-radius: var(--radius-md);
  transition: all var(--duration-base) var(--ease-out);
}

.comment-content:hover {
  white-space: normal;
  overflow: visible;
  max-width: 400px;
  background: var(--color-bg-card);
  box-shadow: var(--shadow-lg);
  z-index: 10;
  position: relative;
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

/* 状态标签 */
.comment-table :deep(.el-tag) {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
}

.comment-table :deep(.el-tag):hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.comment-table :deep(.el-tag--warning) {
  background: var(--color-warning-light);
  border-color: var(--color-warning);
  color: var(--color-warning);
}

.comment-table :deep(.el-tag--success) {
  background: var(--color-success-light);
  border-color: var(--color-success);
  color: var(--color-success);
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

.action-buttons .el-button--success {
  background: linear-gradient(135deg, var(--color-success) 0%, #34d399 100%);
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

/* 加载动画 */
.comment-table :deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .comment-content {
  background: var(--color-bg-active);
}

[data-theme="dark"] .comment-content:hover {
  background: var(--color-bg-card);
}

[data-theme="dark"] .comment-table :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.9);
}

/* ==================== Responsive ==================== */
@media (max-width: 768px) {
  .title {
    font-size: var(--text-xl);
  }

  .review-menu {
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .review-menu span {
    padding: var(--space-1) var(--space-3);
    font-size: var(--text-xs);
  }

  .operation-container {
    flex-direction: column;
    align-items: stretch;
  }

  .operation-container > div:last-child {
    margin-left: 0;
    flex-direction: column;
    width: 100%;
  }

  .operation-container .el-select,
  .operation-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .comment-content {
    max-width: 150px;
  }

  .article-title {
    max-width: 120px;
  }

  .action-buttons {
    flex-direction: row;
  }

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .review-menu > span:first-child {
    width: 100%;
    margin-bottom: var(--space-2);
  }

  .comment-table :deep(.el-table__header) {
    display: none;
  }

  .comment-table :deep(.el-table__row) {
    display: flex;
    flex-direction: column;
    padding: var(--space-4);
    margin-bottom: var(--space-3);
    background: var(--color-bg-card);
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border);
    box-shadow: var(--shadow-sm);
  }

  .comment-table :deep(.el-table__row td) {
    border: none;
    padding: var(--space-2) 0 !important;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .comment-table :deep(.el-table__row td::before) {
    content: attr(data-label);
    font-weight: var(--font-semibold);
    color: var(--color-text-secondary);
    font-size: var(--text-xs);
  }

  .comment-content {
    max-width: none;
    white-space: normal;
  }

  .article-title {
    max-width: none;
  }
}
</style>

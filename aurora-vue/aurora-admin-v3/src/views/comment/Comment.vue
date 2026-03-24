<template>
  <div class="comment-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><ChatDotRound /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">评论总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ approvedCount }}</span>
          <span class="stat-label">已通过</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ pendingCount }}</span>
          <span class="stat-label">待审核</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon info">
          <el-icon><User /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ uniqueUsers }}</span>
          <span class="stat-label">评论用户</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 胶囊状态筛选 -->
      <div class="status-capsules">
        <span @click="changeReview(null)" :class="['capsule', { active: isReview == null }]">全部</span>
        <span @click="changeReview(1)" :class="['capsule', { active: isReview == 1 }]">正常</span>
        <span @click="changeReview(0)" :class="['capsule warning-capsule', { active: isReview == 0 }]">审核中</span>
      </div>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="danger" :icon="Delete" :disabled="commentIds.length === 0" @click="remove = true" class="btn-danger">
            <span>批量删除 ({{ commentIds.length }})</span>
          </el-button>
          <el-button type="success" :icon="SuccessFilled" :disabled="commentIds.length === 0" @click="updateCommentReview(null)" class="btn-success">
            <span>批量通过</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-select clearable v-model="type" placeholder="评论来源" size="default" class="filter-select">
            <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input clearable v-model="keywords" :prefix-icon="Search" placeholder="搜索用户昵称..." class="search-input" @keyup.enter="searchComments" />
          <el-button type="primary" :icon="Search" @click="searchComments" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table :data="comments" @selection-change="selectionChange" v-loading="loading" class="modern-table" :header-cell-style="{ background: 'transparent' }" row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="avatar" label="头像" width="80" align="center">
          <template #default="{ row }">
            <el-avatar :size="40" :src="row.avatar" class="comment-avatar" />
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="评论人" width="130" align="left">
          <template #default="{ row }">
            <div class="nickname-cell">
              <el-icon class="user-icon"><User /></el-icon>
              <span class="nickname-text">{{ row.nickname }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="replyNickname" label="回复人" width="120" align="center">
          <template #default="{ row }">
            <div class="reply-cell">
              <el-icon v-if="row.replyNickname" class="reply-icon"><ChatLineRound /></el-icon>
              <span>{{ row.replyNickname || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="articleTitle" label="文章标题" min-width="160" align="left">
          <template #default="{ row }">
            <el-tooltip :content="row.articleTitle" placement="top" :disabled="!row.articleTitle || row.articleTitle.length <= 20">
              <div class="article-title-cell">
                <el-icon class="article-icon"><Document /></el-icon>
                <span class="article-title-text">{{ row.articleTitle || '无' }}</span>
              </div>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="commentContent" label="评论内容" min-width="200" align="left">
          <template #default="{ row }">
            <div class="comment-content-wrapper">
              <div class="comment-content-text" v-html="sanitizeHtml(row.commentContent)" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="评论时间" width="160" align="center" sortable>
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDate(row.createTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="isReview" label="状态" width="100" align="center">
          <template #default="{ row }">
            <span v-if="row.isReview == 0" class="status-badge pending"><el-icon><Loading /></el-icon> 审核中</span>
            <span v-if="row.isReview == 1" class="status-badge approved"><el-icon><Check /></el-icon> 正常</span>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="90" align="center">
          <template #default="{ row }">
            <span :class="['source-badge', getSourceType(row.type).tagType]">
              {{ getSourceType(row.type).name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip v-if="row.isReview == 0" content="通过审核" placement="top" :show-after="500">
                <button class="action-btn approve" @click="updateCommentReview(row.id)"><el-icon><Check /></el-icon></button>
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
    <el-dialog v-model="remove" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ commentIds.length }} 条评论吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="remove = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteComments(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Delete,
  Search,
  SuccessFilled,
  User,
  ChatLineRound,
  ChatDotRound,
  Document,
  Clock,
  Loading,
  Check,
  Warning,
  CircleCheck
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import DOMPurify from 'dompurify'
import logger from '@/utils/logger'

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

// 计算属性
const approvedCount = computed(() => comments.value.filter(c => c.isReview === 1).length)
const pendingCount = computed(() => comments.value.filter(c => c.isReview === 0).length)
const uniqueUsers = computed(() => new Set(comments.value.map(c => c.nickname)).size)

// 日期格式化
const formatDate = (date) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'

// 获取来源类型
const getSourceType = (type) => {
  const types = {
    1: { name: '文章', tagType: 'primary' },
    2: { name: '留言', tagType: 'danger' },
    3: { name: '关于我', tagType: 'success' },
    4: { name: '友链', tagType: 'warning' },
    5: { name: '说说', tagType: 'info' }
  }
  return types[type] || { name: '未知', tagType: 'info' }
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
      ElNotification.success({ title: '成功', message: data.message })
      listComments()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  }).catch(error => {
    ElMessage.error('审核失败')
    logger.error('API Error:', error)
  })
}

// 处理删除
const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该评论吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteComments(id)
  }).catch(() => {})
}

const deleteComments = (id) => {
  const param = id == null ? { data: commentIds.value } : { data: [id] }
  request.delete('/admin/comments', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listComments()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    remove.value = false
  }).catch(error => {
    ElMessage.error('删除评论失败')
    logger.error('API Error:', error)
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
    logger.error('API Error:', error)
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
.comment-page {
  padding: 0;
}

/* 统计卡片 */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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
.stat-icon.info { background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 100%); color: #fff; }

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

/* 胶囊状态筛选 */
.status-capsules {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-light, #f3f4f6);
}

.capsule {
  padding: 8px 20px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  color: var(--text-secondary, #6b7280);
  background: transparent;
  border: 1px solid transparent;
  user-select: none;
}

.capsule:hover {
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.06);
  border-color: rgba(59, 130, 246, 0.15);
}

.capsule.active {
  color: #fff;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}

.capsule.warning-capsule:hover {
  color: #f59e0b;
  background: rgba(245, 158, 11, 0.06);
  border-color: rgba(245, 158, 11, 0.15);
}

.capsule.warning-capsule.active {
  color: #fff;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.3);
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
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-danger {
  border-radius: 10px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.2s ease;
}

.btn-danger:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
}

.btn-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  border: none;
  border-radius: 10px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.2s ease;
}

.btn-success:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.filter-select {
  width: 140px;
}

.filter-select :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  transition: all 0.2s ease;
}

.search-input {
  width: 220px;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  transition: all 0.2s ease;
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
  padding: 14px 12px;
  border-bottom: 1px solid var(--border-light, #f3f4f6);
}

/* 头像 */
.comment-avatar {
  border: 2px solid var(--border-light, #e5e7eb);
  transition: all 0.2s ease;
}

.comment-avatar:hover {
  transform: scale(1.1);
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
}

/* 昵称 */
.nickname-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-icon {
  color: #3b82f6;
  font-size: 16px;
}

.nickname-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

/* 回复人 */
.reply-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: var(--text-secondary, #6b7280);
  font-size: 13px;
}

.reply-icon {
  color: #8b5cf6;
}

/* 文章标题 */
.article-title-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.article-icon {
  color: #3b82f6;
  flex-shrink: 0;
  font-size: 16px;
}

.article-title-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary, #6b7280);
  font-size: 13px;
  transition: color 0.2s ease;
}

.article-title-cell:hover .article-title-text {
  color: #3b82f6;
}

/* 评论内容 */
.comment-content-wrapper {
  max-width: 300px;
}

.comment-content-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  line-height: 1.5;
  padding: 6px 12px;
  background: var(--bg-elevated, #f9fafb);
  border-radius: 8px;
  transition: all 0.2s ease;
}

.comment-content-wrapper:hover .comment-content-text {
  white-space: normal;
  overflow: visible;
  background: var(--bg-base, #fff);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  position: relative;
  z-index: 10;
}

/* 时间 */
.time-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-secondary, #6b7280);
  font-size: 13px;
}

.time-icon {
  color: #3b82f6;
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.status-badge.pending {
  background: #fffbeb;
  color: #d97706;
}

.status-badge.approved {
  background: #ecfdf5;
  color: #059669;
}

/* 来源徽章 */
.source-badge {
  display: inline-flex;
  align-items: center;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
}

.source-badge.primary { background: #eff6ff; color: #2563eb; }
.source-badge.danger { background: #fef2f2; color: #dc2626; }
.source-badge.success { background: #ecfdf5; color: #059669; }
.source-badge.warning { background: #fffbeb; color: #d97706; }
.source-badge.info { background: #f5f3ff; color: #7c3aed; }

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

.action-btn.approve { background: #ecfdf5; color: #10b981; }
.action-btn.approve:hover { background: #10b981; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); }
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

/* 深色模式 */
[data-theme="dark"] .stat-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .stat-card:hover {
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.3);
}

[data-theme="dark"] .stat-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .main-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .capsule { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .capsule:hover {
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.1);
  border-color: rgba(59, 130, 246, 0.2);
}

[data-theme="dark"] .capsule.warning-capsule:hover {
  color: #fbbf24;
  background: rgba(245, 158, 11, 0.1);
  border-color: rgba(245, 158, 11, 0.2);
}

[data-theme="dark"] .status-capsules { border-bottom-color: var(--border-default, #374151); }

[data-theme="dark"] .modern-table { border-color: var(--border-default, #374151); }

[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) {
  background: var(--bg-elevated, #374151);
  color: var(--text-secondary, #9ca3af);
  border-bottom-color: var(--border-default, #374151);
}

[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) {
  background: var(--bg-hover, #374151) !important;
}

[data-theme="dark"] .modern-table :deep(.el-table__body td) {
  border-bottom-color: var(--border-default, #374151);
}

[data-theme="dark"] .comment-avatar { border-color: var(--border-default, #374151); }
[data-theme="dark"] .nickname-text { color: var(--text-primary, #f9fafb); }

[data-theme="dark"] .comment-content-text {
  background: var(--bg-elevated, #374151);
}

[data-theme="dark"] .comment-content-wrapper:hover .comment-content-text {
  background: var(--bg-base, #1f2937);
}

[data-theme="dark"] .status-badge.pending { background: rgba(245, 158, 11, 0.15); }
[data-theme="dark"] .status-badge.approved { background: rgba(16, 185, 129, 0.15); }

[data-theme="dark"] .source-badge.primary { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .source-badge.danger { background: rgba(239, 68, 68, 0.15); }
[data-theme="dark"] .source-badge.success { background: rgba(16, 185, 129, 0.15); }
[data-theme="dark"] .source-badge.warning { background: rgba(245, 158, 11, 0.15); }
[data-theme="dark"] .source-badge.info { background: rgba(139, 92, 246, 0.15); }

[data-theme="dark"] .action-btn.approve { background: rgba(16, 185, 129, 0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239, 68, 68, 0.15); }

[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .dialog-icon-wrapper.danger { background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(239, 68, 68, 0.25) 100%); }

[data-theme="dark"] .pagination-wrapper { border-top-color: var(--border-default, #374151); }

/* 响应式 */
@media (max-width: 1280px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; }
  .pagination-wrapper { justify-content: center; }
}

@media (max-width: 480px) {
  .stats-row { grid-template-columns: 1fr; }
  .comment-content-wrapper { max-width: 180px; }
}
</style>

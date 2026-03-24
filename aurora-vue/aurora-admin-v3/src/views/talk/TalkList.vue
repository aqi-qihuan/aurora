<template>
  <div class="talk-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><ChatDotRound /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">说说总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><View /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ publicCount }}</span>
          <span class="stat-label">公开</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Lock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ privateCount }}</span>
          <span class="stat-label">私密</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon info">
          <el-icon><Top /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ topCount }}</span>
          <span class="stat-label">置顶</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="EditPen" @click="goToWrite" class="btn-add">
            <span>发布说说</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索说说内容..."
            :prefix-icon="Search"
            clearable
            class="search-input"
            @keyup.enter="handleSearch" />
        </div>
      </div>

      <!-- 胶囊状态筛选 -->
      <div class="status-capsules">
        <span @click="changeStatus(null)" :class="['capsule', { active: status === null }]">全部</span>
        <span @click="changeStatus(1)" :class="['capsule', { active: status === 1 }]">公开</span>
        <span @click="changeStatus(2)" :class="['capsule warning-capsule', { active: status === 2 }]">私密</span>
      </div>

      <!-- 说说列表 -->
      <div v-if="talks.length > 0" class="talk-list">
        <div class="talk-item" v-for="item of talks" :key="item.id">
          <div class="talk-header">
            <div class="talk-user">
              <el-avatar class="talk-avatar" :src="item.avatar" :size="42" />
              <div class="talk-meta">
                <div class="talk-nickname">{{ item.nickname }}</div>
                <div class="talk-time">
                  <el-icon><Clock /></el-icon>
                  <span>{{ formatDateTime(item.createTime) }}</span>
                </div>
              </div>
            </div>
            <div class="talk-badges">
              <span v-if="item.isTop === 1" class="badge badge-top">
                <el-icon><Top /></el-icon>置顶
              </span>
              <span v-if="item.status === 2" class="badge badge-private">
                <el-icon><Lock /></el-icon>私密
              </span>
            </div>
          </div>

          <div class="talk-content" v-html="sanitizeHtml(item.content)" />

          <!-- 图片网格 -->
          <el-row :gutter="8" class="talk-images" v-if="item.imgs && item.imgs.length > 0">
            <el-col
              :xs="12" :sm="8" :md="8"
              v-for="(img, index) of item.imgs"
              :key="index">
              <el-image
                class="talk-image-item"
                lazy
                :src="img"
                :preview-src-list="previews"
                fit="cover" />
            </el-col>
          </el-row>

          <!-- 操作按钮 -->
          <div class="talk-actions">
            <el-tooltip content="编辑" placement="top">
              <button class="action-btn edit" @click="goToEdit(item.id)">
                <el-icon><Edit /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <button class="action-btn delete" @click="confirmDelete(item.id)">
                <el-icon><Delete /></el-icon>
              </button>
            </el-tooltip>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <el-empty v-else description="暂无说说，快去发布第一条吧" class="empty-state" />

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="count > 0">
        <el-pagination
          :hide-on-single-page="false"
          @size-change="sizeChange"
          @current-change="currentChange"
          :current-page="current"
          :page-size="size"
          :page-sizes="[5, 10, 20]"
          :total="count"
          layout="total, sizes, prev, pager, next" />
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="isdelete" width="400px" custom-class="elegant-dialog" :show-close="false">
      <template #header>
        <div class="dialog-title-container">
          <div class="dialog-icon-wrapper danger">
            <el-icon><WarningFilled /></el-icon>
          </div>
          <span class="dialog-title-text">删除确认</span>
        </div>
      </template>
      <div class="dialog-content-text">确定要删除这条说说吗？此操作不可恢复。</div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isdelete = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteTalk" class="btn-confirm-danger">
            <el-icon><Delete /></el-icon>
            <span>确认删除</span>
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElNotification } from 'element-plus'
import {
  MoreFilled, Edit, Delete, WarningFilled,
  ChatDotRound, View, Lock, Top, Clock, Search, EditPen
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import DOMPurify from 'dompurify'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()

const sanitizeHtml = (html) => {
  if (!html) return ''
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'br', 'p', 'span', 'img'],
    ALLOWED_ATTR: ['href', 'title', 'target', 'class', 'src', 'alt']
  })
}

const current = ref(pageStateStore.talkList || 1)
const size = ref(5)
const count = ref(0)
const status = ref(null)
const isdelete = ref(false)
const talks = ref([])
const previews = ref([])
const talkId = ref(null)
const searchKeyword = ref('')

const publicCount = computed(() => talks.value.filter(t => t.status === 1).length)
const privateCount = computed(() => talks.value.filter(t => t.status === 2).length)
const topCount = computed(() => talks.value.filter(t => t.isTop === 1).length)

/**
 * 导航到发布说说页面
 * Talk 组件路由为 /talk/:talkId，用 "write" 作为 talkId 标识发布模式
 */
const goToWrite = () => {
  const allRoutes = router.getRoutes()
  const talkRoute = allRoutes.find(r => {
    const comp = r.components?.default
    if (!comp) return false
    const name = comp.__file || comp.name || ''
    return name.includes('Talk.vue') && !name.includes('TalkList.vue')
  })
  if (talkRoute) {
    router.push({ path: talkRoute.path.replace(':talkId', 'write') })
  } else {
    router.push('/talk/write')
  }
}

/**
 * 导航到编辑说说页面
 */
const goToEdit = (id) => {
  const allRoutes = router.getRoutes()
  const talkRoute = allRoutes.find(r => {
    const comp = r.components?.default
    if (!comp) return false
    const name = comp.__file || comp.name || ''
    return name.includes('Talk.vue') && !name.includes('TalkList.vue')
  })
  if (talkRoute) {
    router.push({ path: talkRoute.path.replace(':talkId', id) })
  } else {
    router.push('/talk/' + id)
  }
}

const formatDateTime = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(() => {
  listTalks()
})

const handleSearch = () => {
  current.value = 1
  listTalks()
}

const listTalks = () => {
  request.get('/admin/talks', {
    params: {
      current: current.value,
      size: size.value,
      status: status.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      talks.value = data.data.records || []
      count.value = data.data.count || 0
      const allImgs = []
      talks.value.forEach(item => {
        if (item.imgs) {
          allImgs.push(...item.imgs)
        }
      })
      previews.value = allImgs
    }
  }).catch(() => {})
}

const sizeChange = (val) => {
  previews.value = []
  size.value = val
  listTalks()
}

const currentChange = (val) => {
  previews.value = []
  current.value = val
  pageStateStore.updatePageState('talkList', val)
  listTalks()
}

const changeStatus = (newStatus) => {
  current.value = 1
  previews.value = []
  status.value = newStatus
  listTalks()
}

const confirmDelete = (id) => {
  talkId.value = id
  isdelete.value = true
}

const deleteTalk = () => {
  request.delete('/admin/talks', { data: [talkId.value] }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listTalks()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    isdelete.value = false
  }).catch(() => {})
}
</script>

<style scoped>
/* ==================== 页面容器 ==================== */
.talk-page {
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
  align-items: center;
}

.btn-add {
  border-radius: 10px;
  font-weight: 600;
  padding: 10px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.btn-add:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
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

/* ==================== 胶囊状态筛选 ==================== */
.status-capsules {
  display: flex;
  gap: 10px;
  margin-bottom: 24px;
}

.capsule {
  padding: 7px 20px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  background: var(--bg-secondary, #f0f2f5);
  color: var(--text-secondary, #666);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  user-select: none;
}

.capsule:hover {
  background: var(--bg-hover, #e8eaed);
  color: var(--text-primary, #333);
}

.capsule.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  box-shadow: 0 3px 12px rgba(102, 126, 234, 0.3);
}

.capsule.warning-capsule.active {
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  box-shadow: 0 3px 12px rgba(245, 87, 108, 0.3);
}

/* ==================== 说说列表 ==================== */
.talk-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.talk-item {
  background: var(--bg-elevated, rgba(255, 255, 255, 0.6));
  border-radius: 14px;
  padding: 20px 24px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.04));
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.talk-item:hover {
  box-shadow: var(--shadow-md, 0 8px 25px rgba(0, 0, 0, 0.08));
  transform: translateY(-2px);
  border-color: var(--border-hover, rgba(102, 126, 234, 0.15));
}

/* ==================== 说说头部 ==================== */
.talk-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.talk-user {
  display: flex;
  align-items: center;
  gap: 12px;
}

.talk-avatar {
  border: 2px solid var(--border-light, rgba(0, 0, 0, 0.06));
  transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.talk-item:hover .talk-avatar {
  transform: scale(1.08);
}

.talk-meta {
  display: flex;
  flex-direction: column;
}

.talk-nickname {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
}

.talk-time {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--text-muted, #999);
  margin-top: 2px;
}

.talk-time .el-icon {
  font-size: 13px;
}

/* ==================== 徽章 ==================== */
.talk-badges {
  display: flex;
  gap: 8px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.badge-top {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), rgba(251, 191, 36, 0.12));
  color: #d97706;
}

.badge-private {
  background: linear-gradient(135deg, rgba(156, 163, 175, 0.12), rgba(209, 213, 219, 0.12));
  color: #6b7280;
}

/* ==================== 说说内容 ==================== */
.talk-content {
  font-size: 14px;
  line-height: 1.75;
  color: var(--text-primary, #333);
  white-space: pre-line;
  word-wrap: break-word;
  word-break: break-all;
  padding: 0 4px;
}

.talk-content :deep(img) {
  max-width: 100%;
  border-radius: 8px;
}

.talk-content :deep(a) {
  color: #667eea;
  text-decoration: none;
}

.talk-content :deep(a:hover) {
  text-decoration: underline;
}

/* ==================== 图片网格 ==================== */
.talk-images {
  margin-top: 14px;
}

.talk-image-item {
  cursor: pointer;
  object-fit: cover;
  height: 180px;
  width: 100%;
  border-radius: 10px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid transparent;
}

.talk-image-item:hover {
  transform: scale(1.03);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.12);
}

/* ==================== 操作按钮 ==================== */
.talk-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid var(--border-light, rgba(0, 0, 0, 0.04));
  opacity: 0;
  transition: opacity 0.3s;
}

.talk-item:hover .talk-actions {
  opacity: 1;
}

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

.action-btn.delete {
  color: #f5576c;
}

.action-btn.delete:hover {
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 87, 108, 0.3);
}

/* ==================== 空状态 ==================== */
.empty-state {
  padding: 60px 0;
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

.dialog-icon-wrapper.danger {
  background: linear-gradient(135deg, rgba(245, 87, 108, 0.12), rgba(240, 147, 251, 0.12));
  color: #f5576c;
}

.dialog-title-text {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary, #1a1a2e);
}

.dialog-content-text {
  font-size: 14px;
  color: var(--text-secondary, #666);
  line-height: 1.6;
  padding-left: 56px;
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

.btn-confirm-danger {
  border-radius: 10px;
  font-weight: 600;
  padding: 9px 20px;
  background: linear-gradient(135deg, #f5576c 0%, #f093fb 100%);
  border: none;
  transition: all 0.3s;
}

.btn-confirm-danger:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(245, 87, 108, 0.4);
}

/* ==================== 深色模式 - 极客风 ==================== */
[data-theme="dark"] .stat-card {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.06));
}

[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 15px var(--primary-glow);
}

[data-theme="dark"] .stat-value {
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  color: var(--neon-blue, #00D4FF);
}

[data-theme="dark"] .talk-item {
  background: var(--bg-elevated, rgba(30, 30, 46, 0.8));
  border-color: var(--border-light, rgba(255, 255, 255, 0.06));
  transition: all 0.25s ease;
}

[data-theme="dark"] .talk-item:hover {
  border-color: rgba(0, 212, 255, 0.3);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.1);
}

[data-theme="dark"] .action-btn {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.08));
  transition: all 0.2s ease;
}

[data-theme="dark"] .capsule {
  background: var(--bg-secondary, #2a2a3e);
  color: var(--text-secondary, #aaa);
  transition: all 0.2s ease;
}

[data-theme="dark"] .capsule:hover {
  color: var(--neon-blue, #00D4FF);
  background: rgba(0, 212, 255, 0.1);
  border-color: rgba(0, 212, 255, 0.3);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.15);
}

[data-theme="dark"] .badge-top {
  background: linear-gradient(135deg, rgba(255, 159, 10, 0.18), rgba(251, 191, 36, 0.18));
  color: var(--neon-orange, #FF9F0A);
  box-shadow: 0 0 8px rgba(255, 159, 10, 0.15);
}

[data-theme="dark"] .badge-private {
  background: linear-gradient(135deg, rgba(156, 163, 175, 0.18), rgba(209, 213, 219, 0.18));
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

  .toolbar-right {
    width: 100%;
  }

  .search-input {
    width: 100%;
  }

  .talk-item {
    padding: 16px;
  }

  .talk-actions {
    opacity: 1;
  }

  .talk-image-item {
    height: 140px;
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

  .stat-label {
    font-size: 11px;
  }

  .capsule {
    padding: 5px 14px;
    font-size: 12px;
  }
}
</style>

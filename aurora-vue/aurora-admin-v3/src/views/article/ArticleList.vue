<template>
  <div class="article-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Notebook /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">文章总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Reading /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ publicCount }}</span>
          <span class="stat-label">公开文章</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Lock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ privateCount }}</span>
          <span class="stat-label">私密文章</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon info">
          <el-icon><EditPen /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ draftCount }}</span>
          <span class="stat-label">草稿数量</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 胶囊状态筛选 -->
      <div class="status-capsules">
        <span @click="changeStatus('all')" :class="['capsule', { active: isActive('all') === 'active-status' }]">全部</span>
        <span @click="changeStatus('public')" :class="['capsule', { active: isActive('public') === 'active-status' }]">公开</span>
        <span @click="changeStatus('private')" :class="['capsule', { active: isActive('private') === 'active-status' }]">私密</span>
        <span @click="changeStatus('draft')" :class="['capsule', { active: isActive('draft') === 'active-status' }]">草稿箱</span>
        <span @click="changeStatus('delete')" :class="['capsule danger-capsule', { active: isActive('delete') === 'active-status' }]">回收站</span>
      </div>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button v-if="isDelete === 0" type="danger" :icon="Delete" :disabled="articleIds.length === 0" @click="updateIsDeleteDialog = true" class="btn-danger">
            <span>批量删除</span>
          </el-button>
          <el-button v-else type="danger" :icon="Delete" :disabled="articleIds.length === 0" @click="removeDialog = true" class="btn-danger">
            <span>批量删除</span>
          </el-button>
          <el-button type="success" :icon="Download" :disabled="articleIds.length === 0" @click="exportDialog = true" class="btn-success">
            <span>批量导出</span>
          </el-button>
          <el-upload action="/api/admin/articles/import" multiple :limit="9" :show-file-list="false" :headers="uploadHeaders" :on-success="uploadArticle">
            <el-button type="primary" :icon="Upload" class="btn-primary">
              <span>批量导入</span>
            </el-button>
          </el-upload>
        </div>
        <div class="toolbar-right">
          <el-select clearable v-model="type" placeholder="文章类型" size="default" class="filter-select">
            <el-option label="全部" value="" />
            <el-option v-for="item in types" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-select clearable size="default" v-model="categoryId" filterable placeholder="选择分类" class="filter-select">
            <el-option label="全部" value="" />
            <el-option v-for="item in categories" :key="item.id" :label="item.categoryName" :value="item.id" />
          </el-select>
          <el-select clearable size="default" v-model="tagId" filterable placeholder="选择标签" class="filter-select">
            <el-option label="全部" value="" />
            <el-option v-for="item in tags" :key="item.id" :label="item.tagName" :value="item.id" />
          </el-select>
          <el-input clearable v-model="keywords" :prefix-icon="Search" placeholder="搜索文章名..." class="search-input" @keyup.enter="searchArticles" />
          <el-button type="primary" :icon="Search" @click="searchArticles" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table :data="articles" @selection-change="selectionChange" v-loading="loading" class="modern-table" :header-cell-style="{ background: 'transparent' }" row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="articleCover" label="文章封面" width="180" align="center">
          <template #default="scope">
            <div class="article-cover-wrapper">
              <el-image class="article-cover" lazy :src="scope.row.articleCover || 'https://static.talkxj.com/articles/c5cc2b2561bd0e3060a500198a4ad37d.png'" :preview-src-list="[scope.row.articleCover]" fit="cover" />
              <div class="article-status-badge">
                <el-tag v-if="scope.row.status === 1" size="small" type="success" effect="dark"><el-icon><View /></el-icon> 公开</el-tag>
                <el-tag v-if="scope.row.status === 2" size="small" type="info" effect="dark"><el-icon><Lock /></el-icon> 私密</el-tag>
                <el-tag v-if="scope.row.status === 3" size="small" type="warning" effect="dark"><el-icon><Document /></el-icon> 草稿</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="articleTitle" label="标题" min-width="180" align="left">
          <template #default="scope">
            <span class="article-title-text">{{ scope.row.articleTitle }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="categoryName" label="分类" width="110" align="center">
          <template #default="scope">
            <span class="category-text">{{ scope.row.categoryName }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="tagDTOs" label="标签" width="170" align="center">
          <template #default="scope">
            <div class="tag-cell">
              <el-tag v-for="item in scope.row.tagDTOs" :key="item.tagId" size="small" effect="light" class="article-tag">{{ item.tagName }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="viewsCount" label="浏览量" width="140" align="center">
          <template #default="scope">
            <div class="views-cell">
              <div class="views-bar-container">
                <div class="views-bar" :style="{ width: getViewsBarWidth(scope.row.viewsCount) + '%', background: getViewsBarColor(scope.row.viewsCount) }"></div>
              </div>
              <span class="views-value">{{ scope.row.viewsCount || 0 }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="80" align="center">
          <template #default="scope">
            <el-tag :type="articleType(scope.row.type).tagType" effect="light" class="article-type-tag">{{ articleType(scope.row.type).name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="发表时间" width="160" align="center">
          <template #default="scope">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDate(scope.row.createTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="isTop" label="置顶" width="80" align="center">
          <template #default="scope">
            <el-switch v-model="scope.row.isTop" :active-value="1" :inactive-value="0" :disabled="scope.row.isDelete === 1" @change="changeTopAndFeatured(scope.row)" />
          </template>
        </el-table-column>
        <el-table-column prop="isFeatured" label="推荐" width="80" align="center">
          <template #default="scope">
            <el-switch v-model="scope.row.isFeatured" :active-value="1" :inactive-value="0" :disabled="scope.row.isDelete === 1" @change="changeTopAndFeatured(scope.row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" align="center" width="140" fixed="right">
          <template #default="scope">
            <div class="action-btns">
              <template v-if="scope.row.isDelete === 0">
                <el-tooltip content="编辑" placement="top" :show-after="500">
                  <button class="action-btn edit" @click="editArticle(scope.row.id)"><el-icon><Edit /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top" :show-after="500">
                  <button class="action-btn delete" @click="handleDelete(scope.row.id)"><el-icon><Delete /></el-icon></button>
                </el-tooltip>
              </template>
              <template v-if="scope.row.isDelete === 1">
                <el-tooltip content="恢复" placement="top" :show-after="500">
                  <button class="action-btn restore" @click="handleRestore(scope.row.id)"><el-icon><RefreshRight /></el-icon></button>
                </el-tooltip>
                <el-tooltip content="彻底删除" placement="top" :show-after="500">
                  <button class="action-btn delete" @click="handlePermanentDelete(scope.row.id)"><el-icon><Delete /></el-icon></button>
                </el-tooltip>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination background layout="total, sizes, prev, pager, next, jumper" :total="count" :page-size="size" :current-page="current" :page-sizes="[10, 20]" @size-change="sizeChange" @current-change="currentChange" />
      </div>
    </el-card>

    <!-- 批量删除对话框 -->
    <el-dialog v-model="updateIsDeleteDialog" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper warning"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content"><h3>移入回收站</h3><p>确定要删除选中的 {{ articleIds.length }} 篇文章吗？文章将移入回收站。</p></div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="updateIsDeleteDialog = false" class="btn-cancel">取消</el-button>
          <el-button type="warning" @click="updateArticleDelete(null)" class="btn-confirm-warning">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 彻底删除对话框 -->
    <el-dialog v-model="removeDialog" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content"><h3>彻底删除</h3><p>确定要彻底删除选中的 {{ articleIds.length }} 篇文章吗？此操作不可恢复。</p></div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="removeDialog = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteArticles(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 导出对话框 -->
    <el-dialog v-model="exportDialog" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><Download /></el-icon></div>
      <div class="dialog-content"><h3>导出文章</h3><p>确定要导出选中的 {{ articleIds.length }} 篇文章吗？</p></div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="exportDialog = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="exportArticles(null)" class="btn-confirm">确认导出</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import { usePageStateStore } from '@/stores/pageState'
import request from '@/utils/request'
import dayjs from 'dayjs'
import logger from '@/utils/logger'
import { getAuthHeaders } from '@/utils/auth'
import {
  Delete,
  Download,
  Upload,
  Search,
  View,
  Lock,
  Document,
  Clock,
  Warning,
  Edit,
  EditPen,
  RefreshRight,
  Notebook,
  Reading
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const updateIsDeleteDialog = ref(false)
const removeDialog = ref(false)
const exportDialog = ref(false)
const activeStatus = ref('all')
const articles = ref([])
const articleIds = ref([])
const categories = ref([])
const tags = ref([])
const keywords = ref('')
const type = ref('')
const categoryId = ref('')
const tagId = ref('')
const isDelete = ref(0)
const status = ref(null)
const current = ref(1)
const size = ref(10)
const count = ref(0)

const uploadHeaders = getAuthHeaders()

const types = [
  { value: 1, label: '原创' },
  { value: 2, label: '转载' },
  { value: 3, label: '翻译' }
]

const publicCount = computed(() => articles.value.filter(a => a.status === 1 && a.isDelete === 0).length)
const privateCount = computed(() => articles.value.filter(a => a.status === 2 && a.isDelete === 0).length)
const draftCount = computed(() => articles.value.filter(a => a.status === 3 && a.isDelete === 0).length)

const formatDate = (date) => date ? dayjs(date).format('YYYY-MM-DD') : ''

const articleType = (t) => {
  const map = {
    1: { tagType: 'danger', name: '原创' },
    2: { tagType: 'success', name: '转载' },
    3: { tagType: 'primary', name: '翻译' }
  }
  return map[t] || { tagType: '', name: '' }
}

const isActive = (s) => activeStatus.value === s ? 'active-status' : 'status'

const getViewsBarWidth = (views) => {
  const max = Math.max(...articles.value.map(a => a.viewsCount || 0), 1)
  return Math.min(((views || 0) / max) * 100, 100)
}

const getViewsBarColor = (views) => {
  if (views >= 1000) return 'linear-gradient(90deg, #ef4444, #f87171)'
  if (views >= 500) return 'linear-gradient(90deg, #f59e0b, #fbbf24)'
  if (views >= 100) return 'linear-gradient(90deg, #10b981, #34d399)'
  return 'linear-gradient(90deg, #3b82f6, #60a5fa)'
}

const selectionChange = (selection) => {
  articleIds.value = selection.map(item => item.id)
}

const searchArticles = () => {
  current.value = 1
  listArticles()
}

const editArticle = (id) => {
  router.push({ path: `/articles/${id}` })
}

const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该文章吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    updateArticleDelete(id)
  }).catch(() => {})
}

const handleRestore = (id) => {
  ElMessageBox.confirm('确定恢复该文章吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info'
  }).then(() => {
    updateArticleDelete(id)
  }).catch(() => {})
}

const handlePermanentDelete = (id) => {
  ElMessageBox.confirm('确定彻底删除该文章吗？此操作不可恢复！', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'error'
  }).then(() => {
    deleteArticles(id)
  }).catch(() => {})
}

const updateArticleDelete = (id) => {
  const param = {
    ids: id ? [id] : articleIds.value,
    isDelete: isDelete.value === 0 ? 1 : 0
  }
  request.put('/admin/articles', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listArticles()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    updateIsDeleteDialog.value = false
  }).catch(error => {
    ElMessage.error('操作失败')
    logger.error('API Error:', error)
  })
}

const deleteArticles = (id) => {
  const param = { data: id ? [id] : articleIds.value }
  request.delete('/admin/articles/delete', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listArticles()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    removeDialog.value = false
  }).catch(error => {
    ElMessage.error('删除文章失败')
    logger.error('API Error:', error)
  })
}

const exportArticles = (id) => {
  const param = id ? [id] : articleIds.value
  request.post('/admin/articles/export', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      data.data.forEach((item) => {
        downloadFile(item)
      })
      listArticles()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    exportDialog.value = false
  }).catch(error => {
    ElMessage.error('导出文章失败')
    logger.error('API Error:', error)
  })
}

const downloadFile = (url) => {
  const iframe = document.createElement('iframe')
  iframe.style.display = 'none'
  iframe.style.height = 0
  iframe.src = url
  document.body.appendChild(iframe)
  setTimeout(() => {
    iframe.remove()
  }, 30000)
}

const uploadArticle = (data) => {
  if (data.flag) {
    ElNotification.success({ title: '成功', message: '导入成功' })
    listArticles()
  } else {
    ElNotification.error({ title: '失败', message: data.message })
  }
}

const sizeChange = (newSize) => {
  size.value = newSize
  listArticles()
}

const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('articleList', newCurrent)
  listArticles()
}

const changeStatus = (s) => {
  switch (s) {
    case 'all':
      isDelete.value = 0
      status.value = null
      break
    case 'public':
      isDelete.value = 0
      status.value = 1
      break
    case 'private':
      isDelete.value = 0
      status.value = 2
      break
    case 'draft':
      isDelete.value = 0
      status.value = 3
      break
    case 'delete':
      isDelete.value = 1
      status.value = null
      break
  }
  current.value = 1
  activeStatus.value = s
}

const changeTopAndFeatured = (article) => {
  request.put('/admin/articles/topAndFeatured', {
    id: article.id,
    isTop: article.isTop,
    isFeatured: article.isFeatured
  }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '修改成功' })
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  }).catch(error => {
    ElMessage.error('修改失败')
    logger.error('API Error:', error)
  })
}

const listArticles = () => {
  loading.value = true
  request.get('/admin/articles', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value,
      categoryId: categoryId.value,
      status: status.value,
      tagId: tagId.value,
      type: type.value,
      isDelete: isDelete.value
    }
  }).then(({ data }) => {
    articles.value = data.data.records
    count.value = data.data.count
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取文章列表失败')
    logger.error('API Error:', error)
  })
}

const listCategories = () => {
  request.get('/admin/categories/search').then(({ data }) => {
    categories.value = data.data
  })
}

const listTags = () => {
  request.get('/admin/tags/search').then(({ data }) => {
    tags.value = data?.data || []
  })
}

watch([type, categoryId, tagId, status, isDelete], () => {
  current.value = 1
  listArticles()
})

onMounted(() => {
  current.value = pageStateStore.pageState.articleList
  listArticles()
  listCategories()
  listTags()
})
</script>

<style scoped>
.article-page {
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
  flex-wrap: wrap;
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

.capsule.danger-capsule:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.06);
  border-color: rgba(239, 68, 68, 0.15);
}

.capsule.danger-capsule.active {
  color: #fff;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
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
  flex-wrap: wrap;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
  border-radius: 10px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.2s ease;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
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
  padding: 16px 12px;
  border-bottom: 1px solid var(--border-light, #f3f4f6);
}

/* 文章封面 */
.article-cover-wrapper {
  position: relative;
  width: 130px;
  height: 85px;
  overflow: hidden;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.article-cover {
  width: 100%;
  height: 100%;
  border-radius: 10px;
  transition: transform 0.3s ease;
  cursor: pointer;
}

.article-cover:hover {
  transform: scale(1.08);
}

.article-status-badge {
  position: absolute;
  bottom: 6px;
  right: 6px;
  z-index: 10;
}

.article-status-badge :deep(.el-tag) {
  border-radius: 10px;
  font-size: 11px;
  padding: 2px 8px;
}

.article-title-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.5;
}

.category-text {
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
}

.tag-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  justify-content: center;
}

.article-tag {
  border-radius: 8px;
  font-size: 12px;
}

/* 浏览量进度条 */
.views-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.views-bar-container {
  flex: 1;
  height: 6px;
  background: var(--bg-elevated, #e5e7eb);
  border-radius: 3px;
  overflow: hidden;
  min-width: 40px;
}

.views-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.5s ease;
}

.views-value {
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  min-width: 32px;
  text-align: right;
  font-size: 13px;
}

.article-type-tag {
  border-radius: 8px;
  font-weight: 500;
}

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
.action-btn.restore { background: #ecfdf5; color: #10b981; }
.action-btn.restore:hover { background: #10b981; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3); }

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

.dialog-icon-wrapper.primary { background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%); color: #3b82f6; }
.dialog-icon-wrapper.warning { background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%); color: #f59e0b; }
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

.btn-confirm-warning {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  border: none;
  border-radius: 10px;
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
  color: #fff;
}

.btn-confirm-warning:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
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

[data-theme="dark"] .capsule {
  color: var(--text-secondary, #9ca3af);
}

[data-theme="dark"] .capsule:hover {
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.1);
  border-color: rgba(59, 130, 246, 0.2);
}

[data-theme="dark"] .capsule.danger-capsule:hover {
  color: #f87171;
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.2);
}

[data-theme="dark"] .status-capsules {
  border-bottom-color: var(--border-default, #374151);
}

[data-theme="dark"] .modern-table {
  border-color: var(--border-default, #374151);
}

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

[data-theme="dark"] .article-title-text { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .views-bar-container { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .views-value { color: var(--text-primary, #f9fafb); }

[data-theme="dark"] .action-btn.edit { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239, 68, 68, 0.15); }
[data-theme="dark"] .action-btn.restore { background: rgba(16, 185, 129, 0.15); }

[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .dialog-icon-wrapper.primary { background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(59, 130, 246, 0.25) 100%); }
[data-theme="dark"] .dialog-icon-wrapper.warning { background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(245, 158, 11, 0.25) 100%); }
[data-theme="dark"] .dialog-icon-wrapper.danger { background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(239, 68, 68, 0.25) 100%); }

[data-theme="dark"] .pagination-wrapper {
  border-top-color: var(--border-default, #374151);
}

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
}
</style>

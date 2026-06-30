<template>
  <div class="portfolio-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><FolderOpened /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">作品总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><View /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ visibleCount }}</span>
          <span class="stat-label">可见作品</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Star /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ totalStars }}</span>
          <span class="stat-label">总 Star 数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon danger">
          <el-icon><Histogram /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ featuredCount }}</span>
          <span class="stat-label">首页置顶</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button
            type="primary"
            :icon="Refresh"
            :loading="syncing"
            @click="handleSync"
            class="btn-sync">
            <span>同步 GitHub</span>
          </el-button>
          <el-button
            type="danger"
            :icon="Delete"
            :disabled="portfolioIdList.length === 0"
            @click="deleteFlag = true"
            class="btn-batch-delete">
            <span>批量删除 ({{ portfolioIdList.length }})</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索仓库名/描述..."
            class="search-input"
            clearable
            @keyup.enter="searchPortfolios"
            @clear="searchPortfolios" />
          <el-button type="primary" :icon="Search" @click="searchPortfolios" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table
        :data="portfolioList"
        v-loading="loading"
        @selection-change="selectionChange"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }"
        row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="name" label="仓库信息" min-width="280" align="left">
          <template #default="{ row }">
            <div class="repo-info-cell">
              <div class="repo-cover" :style="coverStyle(row)">
                <img v-if="row.cover" :src="row.cover" alt="" />
                <span v-else class="cover-letter">{{ (row.name || '?').charAt(0).toUpperCase() }}</span>
              </div>
              <div class="repo-info">
                <div class="repo-name-row">
                  <span class="repo-name">{{ row.name }}</span>
                  <el-tag v-if="row.isFeatured === 1" type="danger" size="small" effect="dark" round>置顶</el-tag>
                  <el-tag v-if="row.isVisible === 0" type="info" size="small" effect="plain" round>隐藏</el-tag>
                </div>
                <span class="repo-desc">{{ row.description || '暂无描述' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="language" label="语言" width="120" align="center">
          <template #default="{ row }">
            <div class="lang-cell">
              <span class="lang-dot" :style="{ background: langColor(row.language) }"></span>
              <span>{{ row.language || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="stargazersCount" label="Star" width="100" align="center" sortable>
          <template #default="{ row }">
            <div class="star-cell">
              <el-icon class="star-icon"><Star /></el-icon>
              <span>{{ formatCount(row.stargazersCount) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="repoUpdatedAt" label="最近更新" width="160" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDate(row.repoUpdatedAt) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="编辑" placement="top" :show-after="500">
                <button class="action-btn edit" @click="openEditModel(row)"><el-icon><Edit /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="访问仓库" placement="top" :show-after="500">
                <button class="action-btn visit" @click="visitRepo(row.htmlUrl)"><el-icon><TopRight /></el-icon></button>
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
          :page-sizes="[10, 20, 50]"
          @size-change="sizeChange"
          @current-change="currentChange" />
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="deleteFlag" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ portfolioIdList.length }} 个作品吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="deleteFlag = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deletePortfolios(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑对话框 -->
    <el-dialog v-model="editFlag" width="520px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><EditPen /></el-icon></div>
      <div class="dialog-content">
        <h3>编辑作品</h3>
        <p class="dialog-subtitle">{{ editForm.name }}</p>
        <el-form :model="editForm" class="portfolio-form" label-position="top">
          <el-form-item label="自定义封面 URL">
            <el-input
              v-model="editForm.cover"
              placeholder="留空则使用语言色块默认封面"
              class="form-input"
              clearable />
          </el-form-item>
          <el-form-item label="排序权重">
            <el-input-number
              v-model="editForm.sort"
              :min="0"
              :max="9999"
              controls-position="right"
              class="form-input-number" />
            <span class="form-hint">数值越大越靠前</span>
          </el-form-item>
          <el-form-item label="展示状态">
            <el-switch
              v-model="editForm.isVisible"
              :active-value="1"
              :inactive-value="0"
              active-text="显示"
              inactive-text="隐藏"
              inline-prompt />
          </el-form-item>
          <el-form-item label="首页置顶">
            <el-switch
              v-model="editForm.isFeatured"
              :active-value="1"
              :inactive-value="0"
              active-text="置顶"
              inactive-text="普通"
              inline-prompt />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editFlag = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="savePortfolio" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Refresh, Delete, Search, Clock, Edit, EditPen, Warning, TopRight,
  Star, View, FolderOpened, Histogram
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import dayjs from 'dayjs'

const loading = ref(true)
const syncing = ref(false)
const deleteFlag = ref(false)
const editFlag = ref(false)
const portfolioIdList = ref([])
const portfolioList = ref([])
const keywords = ref(null)
const current = ref(1)
const size = ref(10)
const count = ref(0)

const editForm = reactive({
  id: null,
  name: '',
  cover: '',
  sort: 0,
  isFeatured: 0,
  isVisible: 1
})

// 统计
const visibleCount = computed(() => portfolioList.value.filter(p => p.isVisible === 1).length)
const featuredCount = computed(() => portfolioList.value.filter(p => p.isFeatured === 1).length)
const totalStars = computed(() => portfolioList.value.reduce((sum, p) => sum + (p.stargazersCount || 0), 0))

// 语言颜色映射
const LANG_COLORS = {
  JavaScript: '#f1e05a', TypeScript: '#3178c6', Go: '#00ADD8', Java: '#b07219',
  Vue: '#41b883', Python: '#3572A5', HTML: '#e34c26', CSS: '#563d7c',
  Shell: '#89e051', C: '#555555', 'C++': '#f34b7d', Rust: '#dea584',
  PHP: '#4F5D95', Ruby: '#701516', Kotlin: '#A97BFF', Swift: '#F05138',
  Dart: '#00B4AB', Lua: '#000080', Dockerfile: '#384d54', Makefile: '#427819'
}
const langColor = (lang) => LANG_COLORS[lang] || '#888888'

const coverStyle = (row) => ({
  background: `linear-gradient(135deg, ${langColor(row.language)}22, ${langColor(row.language)}44)`,
  borderColor: langColor(row.language)
})

const formatDate = (date) => date ? dayjs(date).format('YYYY-MM-DD') : '-'
const formatCount = (n) => {
  if (n === undefined || n === null) return '0'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return String(n)
}

onMounted(() => { listPortfolios() })

const selectionChange = (selection) => { portfolioIdList.value = selection.map(item => item.id) }
const searchPortfolios = () => { current.value = 1; listPortfolios() }
const sizeChange = (val) => { size.value = val; listPortfolios() }
const currentChange = (val) => { current.value = val; listPortfolios() }

const listPortfolios = () => {
  loading.value = true
  request.get('/admin/portfolios', {
    params: { current: current.value, size: size.value, keywords: keywords.value }
  }).then(({ data }) => {
    if (data && data.data) {
      portfolioList.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => { loading.value = false })
}

// 手动同步 GitHub
const handleSync = () => {
  ElMessageBox.confirm(
    '将从 GitHub 拉取最新仓库数据，已有人工配置（封面/排序/置顶/可见性）不会被覆盖。是否继续？',
    '同步确认',
    { confirmButtonText: '开始同步', cancelButtonText: '取消', type: 'info' }
  ).then(() => {
    syncing.value = true
    request.post('/admin/portfolios/sync').then(({ data }) => {
      syncing.value = false
      if (data.code === 200 || data.flag) {
        ElNotification.success({ title: '同步成功', message: data.message || 'GitHub 仓库同步完成' })
        listPortfolios()
      } else {
        ElNotification.error({ title: '同步失败', message: data.message || '同步过程中发生错误' })
      }
    }).catch(() => { syncing.value = false })
  }).catch(() => {})
}

// 编辑
const openEditModel = (row) => {
  Object.assign(editForm, {
    id: row.id,
    name: row.name,
    cover: row.cover || '',
    sort: row.sort || 0,
    isFeatured: row.isFeatured,
    isVisible: row.isVisible
  })
  editFlag.value = true
}

const savePortfolio = () => {
  request.put('/admin/portfolios', editForm).then(({ data }) => {
    if (data.code === 200 || data.flag) {
      ElNotification.success({ title: '成功', message: data.message || '作品信息已更新' })
      editFlag.value = false
      listPortfolios()
    } else {
      ElNotification.error({ title: '失败', message: data.message || '更新失败' })
    }
  })
}

// 删除
const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该作品吗？', '提示', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
  }).then(() => { deletePortfolios(id) }).catch(() => {})
}

const deletePortfolios = (id) => {
  const param = id ? { data: [id] } : { data: portfolioIdList.value }
  request.delete('/admin/portfolios', param).then(({ data }) => {
    if (data.code === 200 || data.flag) {
      ElNotification.success({ title: '成功', message: data.message || '作品已删除' })
      listPortfolios()
    } else {
      ElNotification.error({ title: '失败', message: data.message || '删除失败' })
    }
    deleteFlag.value = false
  })
}

const visitRepo = (url) => {
  if (url) window.open(url, '_blank')
}
</script>

<style scoped>
.portfolio-page { padding: 0; }

/* 统计卡片 */
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 24px; }
.stat-card {
  background: var(--bg-base, #fff); border-radius: 16px; padding: 20px;
  display: flex; align-items: center; gap: 14px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid var(--border-default, #e5e7eb);
  transition: all 0.3s ease;
}
.stat-card:hover { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(0,0,0,0.08); }
.stat-icon { width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 22px; flex-shrink: 0; }
.stat-icon.primary { background: linear-gradient(135deg, #3b82f6, #60a5fa); color: #fff; }
.stat-icon.success { background: linear-gradient(135deg, #10b981, #34d399); color: #fff; }
.stat-icon.warning { background: linear-gradient(135deg, #f59e0b, #fbbf24); color: #fff; }
.stat-icon.danger { background: linear-gradient(135deg, #ef4444, #f87171); color: #fff; }
.stat-info { display: flex; flex-direction: column; gap: 2px; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1; }
.stat-label { font-size: 13px; color: var(--text-secondary, #6b7280); }

/* 主卡片 */
.main-card { border-radius: 16px; border: 1px solid var(--border-default, #e5e7eb); box-shadow: 0 1px 3px rgba(0,0,0,0.05); background: var(--bg-base, #fff); }
.main-card :deep(.el-card__body) { padding: 24px; }

/* 工具栏 */
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; flex-wrap: wrap; gap: 16px; }
.toolbar-left { display: flex; gap: 12px; }
.toolbar-right { display: flex; align-items: center; gap: 12px; }
.btn-sync { background: linear-gradient(135deg, #8b5cf6, #6366f1); border: none; border-radius: 10px; font-weight: 500; height: 40px; padding: 0 20px; transition: all 0.2s ease; }
.btn-sync:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(139,92,246,0.4); }
.btn-batch-delete { border-radius: 10px; font-weight: 500; height: 40px; padding: 0 20px; transition: all 0.2s ease; }
.btn-batch-delete:not(:disabled):hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239,68,68,0.4); }
.search-input { width: 280px; }
.search-input :deep(.el-input__wrapper) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); transition: all 0.2s ease; }
.search-input :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6; }

/* 表格 */
.modern-table { border-radius: 12px; overflow: hidden; border: 1px solid var(--border-default, #e5e7eb); }
.modern-table :deep(.el-table__header-wrapper th) { background: var(--bg-elevated, #f9fafb); color: var(--text-secondary, #6b7280); font-weight: 600; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; padding: 16px 12px; border-bottom: 1px solid var(--border-default, #e5e7eb); }
.modern-table :deep(.el-table__body tr) { transition: all 0.2s ease; }
.modern-table :deep(.el-table__body tr:hover > td) { background: var(--bg-hover, #f3f4f6) !important; }
.modern-table :deep(.el-table__body td) { padding: 16px 12px; border-bottom: 1px solid var(--border-light, #f3f4f6); }

/* 仓库信息 */
.repo-info-cell { display: flex; align-items: center; gap: 14px; }
.repo-cover {
  width: 52px; height: 52px; border-radius: 12px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  overflow: hidden; border: 2px solid;
}
.repo-cover img { width: 100%; height: 100%; object-fit: cover; }
.cover-letter { font-size: 22px; font-weight: 700; color: #fff; text-shadow: 0 2px 8px rgba(0,0,0,0.3); }
.repo-info { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.repo-name-row { display: flex; align-items: center; gap: 8px; }
.repo-name { font-weight: 600; font-size: 15px; color: var(--text-primary, #1f2937); }
.repo-desc { font-size: 13px; color: var(--text-secondary, #6b7280); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 280px; }

/* 语言 */
.lang-cell { display: flex; align-items: center; justify-content: center; gap: 8px; font-size: 13px; color: var(--text-secondary, #6b7280); }
.lang-dot { display: inline-block; width: 10px; height: 10px; border-radius: 50%; }

/* Star */
.star-cell { display: flex; align-items: center; justify-content: center; gap: 5px; font-size: 14px; font-weight: 500; color: var(--text-primary, #1f2937); }
.star-icon { color: #f59e0b; font-size: 14px; }

/* 时间 */
.time-cell { display: flex; align-items: center; justify-content: center; gap: 8px; color: var(--text-secondary, #6b7280); font-size: 13px; }
.time-icon { color: #3b82f6; }

/* 操作按钮 */
.action-btns { display: flex; justify-content: center; gap: 8px; }
.action-btn { width: 36px; height: 36px; border-radius: 10px; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s ease; font-size: 16px; }
.action-btn.edit { background: #eff6ff; color: #3b82f6; }
.action-btn.edit:hover { background: #3b82f6; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.3); }
.action-btn.visit { background: #f0fdf4; color: #10b981; }
.action-btn.visit:hover { background: #10b981; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(16,185,129,0.3); }
.action-btn.delete { background: #fef2f2; color: #ef4444; }
.action-btn.delete:hover { background: #ef4444; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239,68,68,0.3); }

/* 分页 */
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 24px; padding-top: 16px; border-top: 1px solid var(--border-light, #f3f4f6); }
.pagination-wrapper :deep(.el-pager li) { border-radius: 8px; font-weight: 500; transition: all 0.2s ease; }
.pagination-wrapper :deep(.el-pager li:hover) { background: var(--bg-hover, #f3f4f6); }
.pagination-wrapper :deep(.el-pager li.is-active) { background: linear-gradient(135deg, #3b82f6, #2563eb); }

/* 对话框 */
.modern-dialog :deep(.el-dialog__header) { display: none; }
.modern-dialog :deep(.el-dialog__body) { padding: 32px 32px 24px; }
.modern-dialog :deep(.el-dialog__footer) { padding: 0 32px 32px; }
.dialog-icon-wrapper { width: 64px; height: 64px; border-radius: 16px; display: flex; align-items: center; justify-content: center; font-size: 28px; margin: 0 auto 20px; }
.dialog-icon-wrapper.primary { background: linear-gradient(135deg, #eff6ff, #dbeafe); color: #3b82f6; }
.dialog-icon-wrapper.danger { background: linear-gradient(135deg, #fef2f2, #fee2e2); color: #ef4444; }
.dialog-content { text-align: center; }
.dialog-content h3 { font-size: 20px; font-weight: 600; color: var(--text-primary, #1f2937); margin: 0 0 8px; }
.dialog-content p { font-size: 14px; color: var(--text-secondary, #6b7280); margin: 0; }
.dialog-subtitle { font-family: 'SF Mono', monospace; margin-bottom: 20px !important; }
.dialog-footer { display: flex; gap: 12px; justify-content: center; }
.btn-cancel { border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm { background: linear-gradient(135deg, #3b82f6, #2563eb); border: none; border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }
.btn-confirm-danger { background: linear-gradient(135deg, #ef4444, #dc2626); border: none; border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm-danger:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239,68,68,0.4); }

/* 表单 */
.portfolio-form { margin-top: 20px; text-align: left; }
.portfolio-form :deep(.el-form-item__label) { font-weight: 500; color: var(--text-primary, #1f2937); padding-bottom: 8px; }
.form-input { width: 100%; }
.form-input :deep(.el-input__wrapper) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); height: 44px; }
.form-input :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6; }
.form-input-number { width: 200px; }
.form-hint { margin-left: 12px; font-size: 12px; color: var(--text-secondary, #9ca3af); }

/* 深色模式 */
[data-theme="dark"] .stat-card { background: var(--bg-base, #1f2937); border-color: var(--border-default, #374151); }
[data-theme="dark"] .stat-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .main-card { background: var(--bg-base, #1f2937); border-color: var(--border-default, #374151); }
[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) { background: var(--bg-elevated, #374151); color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) { background: var(--bg-hover, #374151) !important; }
[data-theme="dark"] .repo-name { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .star-cell { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .action-btn.edit { background: rgba(59,130,246,0.15); }
[data-theme="dark"] .action-btn.visit { background: rgba(16,185,129,0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239,68,68,0.15); }
[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .form-input :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .search-input :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .portfolio-form :deep(.el-form-item__label) { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .form-hint { color: var(--text-secondary, #9ca3af); }

/* 响应式 */
@media (max-width: 1024px) { .stats-row { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; }
  .btn-sync, .btn-batch-delete { width: 100%; }
  .search-input { width: 100%; }
  .pagination-wrapper { justify-content: center; }
  .repo-desc { max-width: 160px; }
}
</style>

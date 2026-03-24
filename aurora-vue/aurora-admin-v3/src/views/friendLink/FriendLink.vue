<template>
  <div class="friendlink-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Link /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">友链总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Connection /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ activeLinks }}</span>
          <span class="stat-label">已填链接</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ recentLinks }}</span>
          <span class="stat-label">本月新增</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openModel(null)" class="btn-add">
            <span>新增友链</span>
          </el-button>
          <el-button
            type="danger"
            :icon="Delete"
            :disabled="linkIdList.length === 0"
            @click="deleteFlag = true"
            class="btn-batch-delete">
            <span>批量删除 ({{ linkIdList.length }})</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索友链名..."
            class="search-input"
            clearable
            @keyup.enter="searchLinks"
            @clear="searchLinks" />
          <el-button type="primary" :icon="Search" @click="searchLinks" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table
        :data="linkList"
        v-loading="loading"
        @selection-change="selectionChange"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }"
        row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="linkName" label="友链信息" min-width="260" align="left">
          <template #default="{ row }">
            <div class="link-info-cell">
              <el-avatar :src="row.linkAvatar" :size="44" class="link-avatar">
                <el-icon :size="20"><User /></el-icon>
              </el-avatar>
              <div class="link-info">
                <span class="link-name">{{ row.linkName }}</span>
                <span class="link-intro">{{ row.linkIntro }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="linkAddress" label="链接地址" min-width="200" align="left">
          <template #default="{ row }">
            <div class="link-url-cell">
              <el-icon class="url-icon"><Link /></el-icon>
              <a :href="row.linkAddress" target="_blank" class="link-url" @click.stop>{{ row.linkAddress }}</a>
              <el-icon class="external-icon"><TopRight /></el-icon>
            </div>
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
        <el-table-column label="操作" width="140" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="编辑" placement="top" :show-after="500">
                <button class="action-btn edit" @click="openModel(row)"><el-icon><Edit /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="访问" placement="top" :show-after="500">
                <button class="action-btn visit" @click="visitLink(row.linkAddress)"><el-icon><TopRight /></el-icon></button>
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
          :page-sizes="[10, 20]"
          @size-change="sizeChange"
          @current-change="currentChange" />
      </div>
    </el-card>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="deleteFlag" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ linkIdList.length }} 个友链吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="deleteFlag = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteLink(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="addOrEdit" width="500px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><EditPen /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ linkForm.id ? '编辑友链' : '添加友链' }}</h3>
        <el-form :model="linkForm" class="link-form" label-position="top">
          <el-form-item label="友链名称">
            <el-input v-model="linkForm.linkName" placeholder="请输入友链名称" class="form-input" :prefix-icon="Link" />
          </el-form-item>
          <el-form-item label="友链头像">
            <el-input v-model="linkForm.linkAvatar" placeholder="请输入头像URL" class="form-input">
              <template #prefix>
                <div class="avatar-preview-mini" v-if="linkForm.linkAvatar">
                  <img :src="linkForm.linkAvatar" alt="" />
                </div>
                <el-icon v-else><Picture /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="友链地址">
            <el-input v-model="linkForm.linkAddress" placeholder="https://example.com" class="form-input">
              <template #prefix><el-icon><Link /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item label="友链介绍">
            <el-input v-model="linkForm.linkIntro" placeholder="请输入友链简介" type="textarea" class="form-input" :rows="3" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="addOrEdit = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="addOrEditLink" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search, Clock, Link, Edit, EditPen, Warning, User, TopRight, Picture, Connection } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'

const route = useRoute()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const deleteFlag = ref(false)
const addOrEdit = ref(false)
const linkIdList = ref([])
const linkList = ref([])
const keywords = ref(null)
const current = ref(pageStateStore.friendLink || 1)
const size = ref(10)
const count = ref(0)

const linkForm = reactive({
  id: null,
  linkName: '',
  linkAvatar: '',
  linkIntro: '',
  linkAddress: ''
})

const activeLinks = computed(() => linkList.value.filter(l => l.linkAddress?.trim()).length)
const recentLinks = computed(() => {
  const now = dayjs()
  return linkList.value.filter(l => dayjs(l.createTime).isSame(now, 'month')).length
})

const formatDate = (date) => dayjs(date).format('YYYY-MM-DD')

onMounted(() => { listLinks() })

const selectionChange = (selection) => { linkIdList.value = selection.map(item => item.id) }
const searchLinks = () => { current.value = 1; listLinks() }
const sizeChange = (val) => { size.value = val; listLinks() }
const currentChange = (val) => { current.value = val; pageStateStore.updateFriendLinkPageState(val); listLinks() }

const listLinks = () => {
  request.get('/admin/links', {
    params: { current: current.value, size: size.value, keywords: keywords.value }
  }).then(({ data }) => {
    if (data && data.data) {
      linkList.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => { loading.value = false })
}

const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该友链吗？', '提示', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
  }).then(() => { deleteLink(id) }).catch(() => {})
}

const deleteLink = (id) => {
  const param = id ? { data: [id] } : { data: linkIdList.value }
  request.delete('/admin/links', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listLinks()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    deleteFlag.value = false
  })
}

const openModel = (link) => {
  if (link) {
    Object.assign(linkForm, link)
  } else {
    linkForm.id = null
    linkForm.linkName = ''
    linkForm.linkAvatar = ''
    linkForm.linkIntro = ''
    linkForm.linkAddress = ''
  }
  addOrEdit.value = true
}

const visitLink = (url) => {
  if (url) window.open(url, '_blank')
}

const addOrEditLink = () => {
  if (!linkForm.linkName?.trim()) { ElMessage.error('友链名不能为空'); return }
  if (!linkForm.linkAvatar?.trim()) { ElMessage.error('友链头像不能为空'); return }
  if (!linkForm.linkIntro?.trim()) { ElMessage.error('友链介绍不能为空'); return }
  if (!linkForm.linkAddress?.trim()) { ElMessage.error('友链地址不能为空'); return }
  request.post('/admin/links', linkForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listLinks()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    addOrEdit.value = false
  })
}
</script>

<style scoped>
.friendlink-page { padding: 0; }

/* 统计卡片 */
.stats-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-bottom: 24px; }
.stat-card {
  background: var(--bg-base, #fff); border-radius: 16px; padding: 24px;
  display: flex; align-items: center; gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid var(--border-default, #e5e7eb);
  transition: all 0.3s ease;
}
.stat-card:hover { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(0,0,0,0.08); }
.stat-icon { width: 56px; height: 56px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 24px; flex-shrink: 0; }
.stat-icon.primary { background: linear-gradient(135deg, #3b82f6, #60a5fa); color: #fff; }
.stat-icon.success { background: linear-gradient(135deg, #10b981, #34d399); color: #fff; }
.stat-icon.warning { background: linear-gradient(135deg, #f59e0b, #fbbf24); color: #fff; }
.stat-info { display: flex; flex-direction: column; gap: 4px; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1; }
.stat-label { font-size: 14px; color: var(--text-secondary, #6b7280); }

/* 主卡片 */
.main-card { border-radius: 16px; border: 1px solid var(--border-default, #e5e7eb); box-shadow: 0 1px 3px rgba(0,0,0,0.05); background: var(--bg-base, #fff); }
.main-card :deep(.el-card__body) { padding: 24px; }

/* 工具栏 */
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; flex-wrap: wrap; gap: 16px; }
.toolbar-left { display: flex; gap: 12px; }
.toolbar-right { display: flex; align-items: center; gap: 12px; }
.btn-add { background: linear-gradient(135deg, #3b82f6, #2563eb); border: none; border-radius: 10px; font-weight: 500; height: 40px; padding: 0 20px; transition: all 0.2s ease; }
.btn-add:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }
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

/* 友链信息 */
.link-info-cell { display: flex; align-items: center; gap: 14px; }
.link-avatar { border-radius: 12px; border: 2px solid var(--border-default, #e5e7eb); flex-shrink: 0; }
.link-info { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.link-name { font-weight: 600; font-size: 15px; color: var(--text-primary, #1f2937); }
.link-intro { font-size: 13px; color: var(--text-secondary, #6b7280); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 280px; }

/* 链接地址 */
.link-url-cell { display: flex; align-items: center; gap: 6px; }
.url-icon { color: #10b981; font-size: 14px; flex-shrink: 0; }
.link-url {
  font-family: 'SF Mono', 'Cascadia Code', monospace; font-size: 13px;
  color: var(--color-primary, #3b82f6); text-decoration: none;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 200px;
  transition: color 0.2s ease;
}
.link-url:hover { color: var(--color-primary-light, #60a5fa); text-decoration: underline; }
.external-icon { color: var(--text-secondary, #6b7280); font-size: 14px; flex-shrink: 0; }

/* 时间 */
.time-cell { display: flex; align-items: center; justify-content: center; gap: 8px; color: var(--text-secondary, #6b7280); font-size: 14px; }
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
.dialog-footer { display: flex; gap: 12px; justify-content: center; }
.btn-cancel { border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm { background: linear-gradient(135deg, #3b82f6, #2563eb); border: none; border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }
.btn-confirm-danger { background: linear-gradient(135deg, #ef4444, #dc2626); border: none; border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm-danger:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239,68,68,0.4); }

/* 表单 */
.link-form { margin-top: 20px; text-align: left; }
.link-form :deep(.el-form-item__label) { font-weight: 500; color: var(--text-primary, #1f2937); padding-bottom: 8px; }
.form-input :deep(.el-input__wrapper) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); height: 44px; }
.form-input :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6; }
.form-input :deep(.el-textarea__inner) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); }
.form-input { width: 100%; }

/* 头像预览 */
.avatar-preview-mini { width: 24px; height: 24px; border-radius: 6px; overflow: hidden; flex-shrink: 0; }
.avatar-preview-mini img { width: 100%; height: 100%; object-fit: cover; }

/* 深色模式 */
[data-theme="dark"] .stat-card { background: var(--bg-base, #1f2937); border-color: var(--border-default, #374151); }
[data-theme="dark"] .stat-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .main-card { background: var(--bg-base, #1f2937); border-color: var(--border-default, #374151); }
[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) { background: var(--bg-elevated, #374151); color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) { background: var(--bg-hover, #374151) !important; }
[data-theme="dark"] .link-name { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .link-avatar { border-color: var(--border-default, #374151); }
[data-theme="dark"] .action-btn.edit { background: rgba(59,130,246,0.15); }
[data-theme="dark"] .action-btn.visit { background: rgba(16,185,129,0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239,68,68,0.15); }
[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .form-input :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .form-input :deep(.el-textarea__inner) { background: var(--bg-elevated, #374151); color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .search-input :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .link-form :deep(.el-form-item__label) { color: var(--text-primary, #f9fafb); }

/* 响应式 */
@media (max-width: 1024px) { .stats-row { grid-template-columns: repeat(2, 1fr); } .stat-card:last-child { grid-column: span 2; } }
@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; } .stat-card:last-child { grid-column: span 1; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; }
  .btn-add, .btn-batch-delete { width: 100%; }
  .search-input { width: 100%; }
  .pagination-wrapper { justify-content: center; }
  .link-intro { max-width: 160px; }
  .link-url { max-width: 140px; }
}
</style>

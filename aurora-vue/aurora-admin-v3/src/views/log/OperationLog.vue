<template>
  <div class="log-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><List /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">日志总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><User /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ uniqueOperators }}</span>
          <span class="stat-label">操作人员</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Calendar /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ todayCount }}</span>
          <span class="stat-label">今日操作</span>
        </div>
      </div>
    </div>

    <!-- 搜索工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <label class="select-all" :class="{ active: logIds.length > 0 }">
          <input
            type="checkbox"
            :checked="logIds.length > 0 && logIds.length === logs.length"
            :indeterminate="logIds.length > 0 && logIds.length < logs.length"
            @change="handleCheckAllChange"
          />
          <span>全选</span>
          <span v-if="logIds.length" class="select-badge">{{ logIds.length }}</span>
        </label>
        <div class="toolbar-divider"></div>
        <button class="btn btn-danger-outline" :disabled="!logIds.length" @click="isDelete = true">
          <el-icon><Delete /></el-icon>批量删除
        </button>
      </div>
      <div class="toolbar-right">
        <div class="search-box">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="keywords"
            class="search-input"
            placeholder="搜索模块名或描述..."
            @keyup.enter="searchLogs"
          />
        </div>
      </div>
    </div>

    <!-- 数据表格 -->
    <div class="table-wrapper" v-loading="loading">
      <table class="data-table" v-if="logs.length > 0">
        <thead>
          <tr>
            <th class="col-check">
              <input
                type="checkbox"
                :checked="logIds.length > 0 && logIds.length === logs.length"
                :indeterminate="logIds.length > 0 && logIds.length < logs.length"
                @change="handleCheckAllChange"
              />
            </th>
            <th class="col-module">系统模块</th>
            <th class="col-type">操作类型</th>
            <th class="col-desc">操作描述</th>
            <th class="col-method">请求方式</th>
            <th class="col-user">操作人员</th>
            <th class="col-ip">登录IP</th>
            <th class="col-time">操作日期</th>
            <th class="col-action">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in logs"
            :key="row.id"
            :class="{ 'row-selected': logIds.includes(row.id) }"
          >
            <td class="col-check">
              <input
                type="checkbox"
                :checked="logIds.includes(row.id)"
                @change="toggleSelect(row.id)"
              />
            </td>
            <td class="col-module">
              <span class="cell-tag module">{{ row.optModule }}</span>
            </td>
            <td class="col-type">
              <span class="cell-tag type" :class="optTypeClass(row.optType)">{{ row.optType }}</span>
            </td>
            <td class="col-desc">
              <span class="cell-text" :title="row.optDesc">{{ row.optDesc }}</span>
            </td>
            <td class="col-method">
              <span class="method-badge" :class="methodClass(row.requestMethod)">
                {{ row.requestMethod }}
              </span>
            </td>
            <td class="col-user">
              <div class="user-cell">
                <span class="user-avatar">{{ (row.nickname || '?')[0] }}</span>
                <span class="user-name">{{ row.nickname }}</span>
              </div>
            </td>
            <td class="col-ip">
              <span class="cell-mono">{{ row.ipAddress }}</span>
            </td>
            <td class="col-time">
              <span class="cell-time">{{ formatDateTime(row.createTime) }}</span>
            </td>
            <td class="col-action">
              <div class="action-btns">
                <button class="mini-btn view" @click="check(row)" title="查看详情">
                  <el-icon><View /></el-icon>
                </button>
                <button class="mini-btn delete" @click="handleDelete(row.id)" title="删除">
                  <el-icon><Delete /></el-icon>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="!loading && logs.length === 0" class="empty-state">
        <el-icon class="empty-icon"><DocumentDelete /></el-icon>
        <p>暂无操作日志</p>
        <span class="empty-hint">操作日志将在这里展示</span>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-bar" v-if="count > size">
      <span class="page-info">共 <strong>{{ count }}</strong> 条记录</span>
      <el-pagination
        v-model:current-page="current"
        :page-size="size"
        :total="count"
        background
        layout="prev, pager, next"
        @current-change="currentChange"
      />
    </div>

    <!-- 详情对话框 -->
    <el-dialog v-model="isCheck" width="640px" class="detail-dialog" destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <span class="dialog-icon primary">
            <el-icon><View /></el-icon>
          </span>
          <div>
            <h3 class="dialog-title">操作详情</h3>
            <p class="dialog-subtitle">{{ optLog.optModule }} · {{ optLog.optDesc }}</p>
          </div>
        </div>
      </template>
      <div class="detail-grid">
        <div class="detail-item">
          <span class="detail-label">操作模块</span>
          <span class="detail-value">{{ optLog.optModule }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">操作类型</span>
          <span class="detail-value">
            <span class="cell-tag type" :class="optTypeClass(optLog.optType)">{{ optLog.optType }}</span>
          </span>
        </div>
        <div class="detail-item">
          <span class="detail-label">请求方式</span>
          <span class="detail-value">
            <span class="method-badge" :class="methodClass(optLog.requestMethod)">{{ optLog.requestMethod }}</span>
          </span>
        </div>
        <div class="detail-item">
          <span class="detail-label">操作人员</span>
          <span class="detail-value">{{ optLog.nickname }}</span>
        </div>
        <div class="detail-item full">
          <span class="detail-label">请求地址</span>
          <span class="detail-value mono">{{ optLog.optUri }}</span>
        </div>
        <div class="detail-item full">
          <span class="detail-label">操作方法</span>
          <span class="detail-value mono">{{ optLog.optMethod }}</span>
        </div>
        <div class="detail-item full">
          <span class="detail-label">请求参数</span>
          <div class="detail-code" v-if="optLog.requestParam">{{ optLog.requestParam }}</div>
          <span class="detail-value muted" v-else>-</span>
        </div>
        <div class="detail-item full">
          <span class="detail-label">返回数据</span>
          <div class="detail-code" v-if="optLog.responseData">{{ optLog.responseData }}</div>
          <span class="detail-value muted" v-else>-</span>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="isCheck = false">关闭</button>
        </div>
      </template>
    </el-dialog>

    <!-- 删除确认 -->
    <el-dialog v-model="isDelete" width="420px" class="delete-dialog" destroy-on-close>
      <div class="delete-confirm">
        <div class="delete-icon-wrap danger">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <h3>确认删除</h3>
        <p>确定要删除选中的 <strong>{{ logIds.length }}</strong> 条操作日志吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="isDelete = false">取消</button>
          <button class="btn btn-danger" @click="deleteLog(null)">确认删除</button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification, ElMessageBox } from 'element-plus'
import { Delete, Search, View, WarningFilled, List, User, Calendar, DocumentDelete } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'

const route = useRoute()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const logs = ref([])
const logIds = ref([])
const keywords = ref('')
const current = ref(pageStateStore.operationLog || 1)
const size = ref(10)
const count = ref(0)
const isCheck = ref(false)
const isDelete = ref(false)

const optLog = reactive({
  optModule: '',
  optUri: '',
  requestMethod: '',
  optMethod: '',
  optType: '',
  optDesc: '',
  requestParam: '',
  responseData: '',
  nickname: ''
})

const uniqueOperators = computed(() => {
  const names = new Set(logs.value.map(l => l.nickname).filter(Boolean))
  return names.size
})

const todayCount = computed(() => {
  const today = dayjs().format('YYYY-MM-DD')
  return logs.value.filter(l => dayjs(l.createTime).format('YYYY-MM-DD') === today).length
})

const formatDateTime = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

const methodClass = (method) => {
  const map = { GET: 'get', POST: 'post', PUT: 'put', DELETE: 'del' }
  return map[method] || ''
}

const optTypeClass = (type) => {
  if (!type) return ''
  const t = type.toLowerCase()
  if (t.includes('新增') || t.includes('新增') || t.includes('添加')) return 'create'
  if (t.includes('修改') || t.includes('更新') || t.includes('编辑')) return 'update'
  if (t.includes('删除') || t.includes('移除')) return 'remove'
  if (t.includes('查询') || t.includes('搜索') || t.includes('获取') || t.includes('查看')) return 'query'
  if (t.includes('导出')) return 'export'
  if (t.includes('登录') || t.includes('登出')) return 'auth'
  if (t.includes('上传')) return 'upload'
  return 'other'
}

onMounted(() => {
  listLogs()
})

const selectionChange = (selection) => {
  logIds.value = selection.map(item => item.id)
}

const toggleSelect = (id) => {
  const idx = logIds.value.indexOf(id)
  if (idx > -1) logIds.value.splice(idx, 1)
  else logIds.value.push(id)
}

const handleCheckAllChange = (e) => {
  logIds.value = e.target.checked ? logs.value.map(l => l.id) : []
}

const searchLogs = () => {
  current.value = 1
  pageStateStore.updatePageState('operationLog', current.value)
  listLogs()
}

const sizeChange = (val) => {
  size.value = val
  listLogs()
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updatePageState('operationLog', val)
  listLogs()
}

const listLogs = () => {
  request.get('/admin/operation/logs', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      logs.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该日志吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteLog(id)
  }).catch(() => {})
}

const deleteLog = (id) => {
  const param = id ? { data: [id] } : { data: logIds.value }
  request.delete('/admin/operation/logs', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listLogs()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    isDelete.value = false
    logIds.value = []
  })
}

const check = (log) => {
  Object.assign(optLog, log)
  isCheck.value = true
}
</script>

<style scoped>
.log-page {
  padding: 4px 0;
}

/* ========== 统计卡片 ========== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  padding: 18px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  transition: all 0.25s ease;
}
.stat-card:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.06); }
.stat-icon {
  width: 46px; height: 46px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px; flex-shrink: 0;
}
.stat-icon.primary { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.stat-icon.success { background: linear-gradient(135deg, #e6f4ea, #ceead6); color: #1e8e3e; }
.stat-icon.warning { background: linear-gradient(135deg, #fef7e0, #fde293); color: #e65100; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1.2; }
.stat-label { font-size: 13px; color: var(--text-secondary, #6b7280); margin-top: 2px; }

/* ========== 工具栏 ========== */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  padding: 14px 20px;
  margin-bottom: 16px;
}
.toolbar-left { display: flex; align-items: center; gap: 10px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.toolbar-divider { width: 1px; height: 20px; background: var(--border-color, #e5e7eb); }

/* ========== 按钮 ========== */
.btn {
  height: 34px; padding: 0 14px; border: none; border-radius: 8px;
  font-size: 13px; font-weight: 500; cursor: pointer;
  display: inline-flex; align-items: center; gap: 5px;
  transition: all 0.2s;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-danger { background: #ef4444; color: #fff; box-shadow: 0 2px 8px rgba(239,68,68,0.25); }
.btn-danger:hover:not(:disabled) { background: #dc2626; }
.btn-danger-outline { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }
.btn-danger-outline:hover:not(:disabled) { background: #dc2626; color: #fff; }
.btn-default { background: var(--bg-card, #fff); color: var(--text-primary, #374151); border: 1px solid var(--border-color, #d1d5db); }
.btn-default:hover { background: var(--bg-body, #f9fafb); }

/* ========== 全选 ========== */
.select-all {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px; color: var(--text-secondary, #6b7280);
  cursor: pointer; user-select: none; padding: 4px 8px;
  border-radius: 6px; transition: all 0.2s;
}
.select-all.active { background: #fef2f2; color: #dc2626; }
.select-all input[type="checkbox"] { width: 16px; height: 16px; accent-color: #1a73e8; cursor: pointer; }
.select-badge {
  background: #dc2626; color: #fff; font-size: 11px;
  padding: 0 6px; border-radius: 10px; min-width: 20px;
  text-align: center; line-height: 18px;
}

/* ========== 搜索框 ========== */
.search-box {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute; left: 12px;
  font-size: 14px; color: var(--text-tertiary, #9ca3af);
  pointer-events: none;
}
.search-input {
  width: 240px; height: 34px; padding: 0 12px 0 36px;
  border: 1px solid var(--border-color, #d1d5db); border-radius: 8px;
  font-size: 13px; color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff); outline: none;
  transition: all 0.2s;
}
.search-input:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1); }
.search-input::placeholder { color: var(--text-tertiary, #9ca3af); }

/* ========== 表格 ========== */
.table-wrapper {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  overflow: hidden;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}
.data-table thead th {
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #6b7280);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  background: var(--bg-body, #f9fafb);
  border-bottom: 1px solid var(--border-color, #ebeef5);
  text-align: left;
  white-space: nowrap;
}
.data-table tbody td {
  padding: 12px 16px;
  font-size: 13px;
  color: var(--text-primary, #374151);
  border-bottom: 1px solid var(--border-color, #f3f4f6);
  vertical-align: middle;
}
.data-table tbody tr { transition: background 0.15s; }
.data-table tbody tr:hover { background: var(--bg-body, #f9fafb); }
.data-table tbody tr.row-selected { background: #eff6ff; }
.data-table tbody tr:last-child td { border-bottom: none; }
.col-check { width: 40px; text-align: center !important; }
.col-check input[type="checkbox"] { width: 16px; height: 16px; accent-color: #1a73e8; cursor: pointer; }
.col-module { min-width: 100px; }
.col-type { min-width: 80px; }
.col-desc { min-width: 120px; max-width: 200px; }
.col-method { min-width: 80px; }
.col-user { min-width: 100px; }
.col-ip { min-width: 120px; }
.col-time { min-width: 160px; }
.col-action { width: 80px; text-align: center; }

/* ========== 单元格样式 ========== */
.cell-tag {
  display: inline-block; padding: 2px 8px; border-radius: 6px;
  font-size: 12px; font-weight: 500;
}
.cell-tag.module { background: #eff6ff; color: #2563eb; }
.cell-tag.type { font-size: 11px; padding: 2px 7px; }
.cell-tag.type.create { background: #ecfdf5; color: #059669; }
.cell-tag.type.update { background: #fef3c7; color: #d97706; }
.cell-tag.type.remove { background: #fef2f2; color: #dc2626; }
.cell-tag.type.query { background: #f0f9ff; color: #0284c7; }
.cell-tag.type.export { background: #f5f3ff; color: #7c3aed; }
.cell-tag.type.auth { background: #fdf2f8; color: #db2777; }
.cell-tag.type.upload { background: #ecfdf5; color: #0d9488; }
.cell-tag.type.other { background: #f3f4f6; color: #6b7280; }
.cell-text { display: block; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 200px; }
.cell-mono { font-family: 'JetBrains Mono', 'SF Mono', 'Consolas', monospace; font-size: 12px; color: var(--text-secondary, #6b7280); }
.cell-time { font-size: 12px; color: var(--text-secondary, #6b7280); white-space: nowrap; }

/* ========== 请求方式 ========== */
.method-badge {
  display: inline-block; padding: 2px 8px; border-radius: 5px;
  font-size: 11px; font-weight: 700; letter-spacing: 0.5px;
  font-family: 'JetBrains Mono', 'SF Mono', monospace;
}
.method-badge.get { background: #ecfdf5; color: #059669; }
.method-badge.post { background: #eff6ff; color: #2563eb; }
.method-badge.put { background: #fef3c7; color: #d97706; }
.method-badge.del { background: #fef2f2; color: #dc2626; }

/* ========== 用户单元格 ========== */
.user-cell { display: flex; align-items: center; gap: 8px; }
.user-avatar {
  width: 28px; height: 28px; border-radius: 8px;
  background: linear-gradient(135deg, #e8f0fe, #d2e3fc);
  color: #1a73e8; font-size: 12px; font-weight: 600;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.user-name { font-size: 13px; color: var(--text-primary, #1f2937); }

/* ========== 操作按钮 ========== */
.action-btns { display: flex; align-items: center; gap: 4px; justify-content: center; }
.mini-btn {
  width: 30px; height: 30px; border: none; border-radius: 8px;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  font-size: 14px; transition: all 0.2s;
}
.mini-btn.view { background: #eff6ff; color: #2563eb; }
.mini-btn.view:hover { background: #2563eb; color: #fff; transform: scale(1.1); }
.mini-btn.delete { background: #fef2f2; color: #dc2626; }
.mini-btn.delete:hover { background: #dc2626; color: #fff; transform: scale(1.1); }

/* ========== 空状态 ========== */
.empty-state { padding: 60px 20px; text-align: center; }
.empty-icon { font-size: 56px; color: var(--border-color, #d1d5db); margin-bottom: 16px; }
.empty-state p { font-size: 16px; color: var(--text-secondary, #6b7280); margin: 0 0 6px; font-weight: 500; }
.empty-hint { font-size: 13px; color: var(--text-tertiary, #9ca3af); }

/* ========== 分页 ========== */
.pagination-bar {
  display: flex; align-items: center; justify-content: space-between;
  margin-top: 20px; padding: 0 4px;
}
.page-info { font-size: 13px; color: var(--text-secondary, #6b7280); }
.page-info strong { color: var(--text-primary, #1f2937); }

/* ========== 对话框通用 ========== */
.dialog-header { display: flex; align-items: center; gap: 14px; }
.dialog-icon {
  width: 42px; height: 42px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; font-size: 20px;
}
.dialog-icon.primary { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.dialog-title { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.dialog-subtitle { margin: 3px 0 0; font-size: 13px; color: var(--text-secondary, #6b7280); }
.dialog-footer { display: flex; justify-content: flex-end; gap: 8px; }

.detail-dialog :deep(.el-dialog) { border-radius: 16px; }
.detail-dialog :deep(.el-dialog__header) { padding: 24px 24px 0; margin: 0; }
.detail-dialog :deep(.el-dialog__body) { padding: 16px 24px; }
.detail-dialog :deep(.el-dialog__footer) { padding: 16px 24px; border-top: 1px solid var(--border-color, #f0f0f0); }

.delete-dialog :deep(.el-dialog) { border-radius: 16px; }
.delete-dialog :deep(.el-dialog__header) { padding: 24px 24px 0; margin: 0; }
.delete-dialog :deep(.el-dialog__body) { padding: 16px 24px; }
.delete-dialog :deep(.el-dialog__footer) { padding: 16px 24px; border-top: 1px solid var(--border-color, #f0f0f0); }

/* ========== 详情网格 ========== */
.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.detail-item { display: flex; flex-direction: column; gap: 4px; }
.detail-item.full { grid-column: 1 / -1; }
.detail-label { font-size: 12px; font-weight: 500; color: var(--text-tertiary, #9ca3af); text-transform: uppercase; letter-spacing: 0.5px; }
.detail-value { font-size: 13px; color: var(--text-primary, #1f2937); line-height: 1.5; }
.detail-value.mono { color: var(--text-tertiary, #9ca3af); }
.detail-code {
  background: #f1f5f9;
  border: 1px solid var(--border-color, #e2e8f0);
  border-radius: 8px;
  padding: 12px;
  font-family: 'JetBrains Mono', 'SF Mono', 'Consolas', monospace;
  font-size: 12px;
  color: var(--text-primary, #1f2937);
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 160px;
  overflow-y: auto;
  line-height: 1.6;
}

/* ========== 删除确认 ========== */
.delete-confirm { text-align: center; }
.delete-icon-wrap {
  width: 56px; height: 56px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 28px; margin: 0 auto 16px;
}
.delete-icon-wrap.danger { background: #fef2f2; color: #dc2626; }
.delete-confirm h3 { margin: 0 0 8px; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.delete-confirm p { margin: 0; font-size: 13px; color: var(--text-secondary, #6b7280); line-height: 1.6; }
.delete-confirm strong { color: #dc2626; }

/* ========== 暗色模式 ========== */
[data-theme="dark"] .stat-card { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .toolbar { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .table-wrapper { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .data-table thead th { background: var(--bg-body, #0f172a); border-color: var(--border-color, #334155); color: var(--text-tertiary, #94a3b8); }
[data-theme="dark"] .data-table tbody td { border-color: var(--border-color, #1e293b); color: var(--text-primary, #e2e8f0); }
[data-theme="dark"] .data-table tbody tr:hover { background: rgba(30, 58, 95, 0.5); }
[data-theme="dark"] .data-table tbody tr.row-selected { background: #1e3a5f; }
[data-theme="dark"] .search-input { background: var(--bg-body, #0f172a); border-color: var(--border-color, #334155); color: var(--text-primary, #e2e8f0); }
[data-theme="dark"] .detail-code { background: #0f172a; border-color: var(--border-color, #334155); color: #e2e8f0; }
[data-theme="dark"] .cell-tag.module { background: #1e3a5f; color: #60a5fa; }
[data-theme="dark"] .cell-tag.type.create { background: #064e3b; color: #6ee7b7; }
[data-theme="dark"] .cell-tag.type.update { background: #78350f; color: #fcd34d; }
[data-theme="dark"] .cell-tag.type.remove { background: #7f1d1d; color: #fca5a5; }
[data-theme="dark"] .cell-tag.type.query { background: #0c4a6e; color: #7dd3fc; }
[data-theme="dark"] .method-badge.get { background: #064e3b; color: #6ee7b7; }
[data-theme="dark"] .method-badge.post { background: #1e3a5f; color: #60a5fa; }
[data-theme="dark"] .method-badge.put { background: #78350f; color: #fcd34d; }
[data-theme="dark"] .method-badge.del { background: #7f1d1d; color: #fca5a5; }
[data-theme="dark"] .user-avatar { background: #1e3a5f; color: #60a5fa; }
[data-theme="dark"] .mini-btn.view { background: #1e3a5f; color: #60a5fa; }
[data-theme="dark"] .mini-btn.view:hover { background: #3b82f6; }
[data-theme="dark"] .mini-btn.delete { background: #7f1d1d; color: #fca5a5; }
[data-theme="dark"] .mini-btn.delete:hover { background: #ef4444; }

/* ========== 响应式 ========== */
@media (max-width: 1100px) { .stats-row { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; gap: 10px; }
  .toolbar-left { flex-wrap: wrap; width: 100%; }
  .toolbar-right { width: 100%; }
  .search-box { width: 100%; }
  .search-input { width: 100%; }
  .detail-grid { grid-template-columns: 1fr; }
}
</style>

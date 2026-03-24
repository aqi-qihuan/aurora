<template>
  <div class="quartz-log-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">日志总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><SuccessFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ successCount }}</span>
          <span class="stat-label">执行成功</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon danger">
          <el-icon><CircleCloseFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ failCount }}</span>
          <span class="stat-label">执行失败</span>
        </div>
      </div>
    </div>

    <!-- 搜索筛选栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <div class="search-box">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="searchParams.jobName"
            class="search-input"
            placeholder="搜索任务名称..."
            @keyup.enter="searchLogs"
          />
          <kbd class="search-kbd">Enter</kbd>
        </div>
        <div class="filter-select">
          <el-select
            v-model="searchParams.jobGroup"
            placeholder="任务组名"
            clearable
            style="width: 140px"
            @change="listJobLogs"
          >
            <el-option v-for="g in jobGroups" :key="g" :label="g" :value="g" />
          </el-select>
        </div>
        <div class="filter-select">
          <el-select
            v-model="searchParams.status"
            placeholder="执行状态"
            clearable
            style="width: 120px"
            @change="listJobLogs"
          >
            <el-option value="1" label="成功" />
            <el-option value="0" label="失败" />
          </el-select>
        </div>
        <div class="filter-date">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="—"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            style="width: 240px"
          />
        </div>
      </div>
      <div class="filter-right">
        <button class="filter-btn primary" @click="searchLogs">
          <el-icon><Search /></el-icon>查找
        </button>
        <button class="filter-btn" @click="clearSearch">
          <el-icon><Refresh /></el-icon>重置
        </button>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="action-left">
        <label class="batch-check" :class="{ active: jobLogIds.length }">
          <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
          <span>全选</span>
        </label>
        <button class="action-btn danger" :disabled="!jobLogIds.length" @click="deleteJobLogs">
          <el-icon><Delete /></el-icon>批量删除
          <span v-if="jobLogIds.length" class="badge">{{ jobLogIds.length }}</span>
        </button>
        <button class="action-btn danger-outline" @click="clean">
          <el-icon><Delete /></el-icon>清空日志
        </button>
      </div>
      <div class="action-right">
        <span class="job-id-badge" v-if="jobId > 0">
          <el-icon><Timer /></el-icon>
          任务 #{{ jobId }}
        </span>
      </div>
    </div>

    <!-- 日志表格 -->
    <div class="table-wrapper" v-loading="loading">
      <table class="elegant-table" v-if="jobLogs.length">
        <thead>
          <tr>
            <th class="col-check">
              <input type="checkbox" :checked="isAllSelected" @change="toggleSelectAll" />
            </th>
            <th class="col-index">#</th>
            <th>任务名称</th>
            <th>任务组名</th>
            <th>调用目标</th>
            <th>日志信息</th>
            <th class="col-status">状态</th>
            <th class="col-time">执行时间</th>
            <th class="col-action">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(log, index) in jobLogs"
            :key="log.id"
            class="table-row"
            :class="{ 'row-fail': log.status === 0 }"
          >
            <td class="col-check">
              <input
                type="checkbox"
                :checked="jobLogIds.includes(log.id)"
                @change="toggleSelect(log.id)"
              />
            </td>
            <td class="col-index">{{ (current - 1) * size + index + 1 }}</td>
            <td>
              <div class="cell-name">
                <span class="name-dot" :class="log.status === 1 ? 'success' : 'fail'"></span>
                {{ log.jobName }}
              </div>
            </td>
            <td>
              <span class="group-tag">{{ log.jobGroup }}</span>
            </td>
            <td>
              <code class="cell-code">{{ log.invokeTarget }}</code>
            </td>
            <td>
              <span class="cell-message" :title="log.jobMessage">{{ log.jobMessage || '—' }}</span>
            </td>
            <td class="col-status">
              <span class="status-badge" :class="log.status === 1 ? 'success' : 'fail'">
                <span class="status-dot"></span>
                {{ log.status === 1 ? '成功' : '失败' }}
              </span>
            </td>
            <td class="col-time">
              <span class="cell-time">{{ formatDateTime(log.startTime) }}</span>
            </td>
            <td class="col-action">
              <button class="action-btn detail" @click="changeOpen(log)">
                <el-icon><View /></el-icon>详细
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else-if="!loading" class="empty-state">
        <el-icon class="empty-icon"><Document /></el-icon>
        <p>暂无调度日志</p>
        <span class="empty-hint">任务执行后日志将在此处显示</span>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-bar" v-if="count > 0">
      <span class="page-info">共 {{ count }} 条记录</span>
      <el-pagination
        v-model:current-page="current"
        v-model:page-size="size"
        :total="count"
        :page-sizes="[10, 20]"
        background
        layout="sizes, prev, pager, next, jumper"
        @size-change="listJobLogs"
        @current-change="currentChange"
      />
    </div>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="open"
      title=""
      :width="jobLog.status === 0 ? '820px' : '680px'"
      append-to-body
      destroy-on-close
      class="detail-dialog"
      :show-close="true"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-title-group">
            <span class="dialog-icon" :class="jobLog.status === 1 ? 'success' : 'fail'">
              <el-icon v-if="jobLog.status === 1"><SuccessFilled /></el-icon>
              <el-icon v-else><CircleCloseFilled /></el-icon>
            </span>
            <div>
              <h3 class="dialog-title">调度日志详情</h3>
              <p class="dialog-subtitle">{{ jobLog.jobName }} · {{ jobLog.jobGroup }}</p>
            </div>
          </div>
          <span class="dialog-status" :class="jobLog.status === 1 ? 'success' : 'fail'">
            {{ jobLog.status === 1 ? '执行成功' : '执行失败' }}
          </span>
        </div>
      </template>

      <div class="detail-grid">
        <div class="detail-item">
          <span class="detail-label">日志序号</span>
          <span class="detail-value mono">{{ jobLog.id }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">任务名称</span>
          <span class="detail-value">{{ jobLog.jobName }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">任务分组</span>
          <span class="detail-value">
            <span class="group-tag">{{ jobLog.jobGroup }}</span>
          </span>
        </div>
        <div class="detail-item">
          <span class="detail-label">执行时间</span>
          <span class="detail-value mono">{{ formatDateTime(jobLog.startTime) }}</span>
        </div>
        <div class="detail-item full">
          <span class="detail-label">调用目标</span>
          <span class="detail-value">
            <code class="detail-code">{{ jobLog.invokeTarget }}</code>
          </span>
        </div>
        <div class="detail-item full">
          <span class="detail-label">日志信息</span>
          <span class="detail-value message">{{ jobLog.jobMessage || '无' }}</span>
        </div>
        <div v-if="jobLog.status === 0 && jobLog.exceptionInfo" class="detail-item full">
          <span class="detail-label">异常信息</span>
          <div class="exception-block">
            <pre><code>{{ jobLog.exceptionInfo }}</code></pre>
          </div>
        </div>
      </div>

      <template #footer>
        <button class="dialog-close-btn" @click="open = false">关闭</button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification, ElMessageBox } from 'element-plus'
import { Search, Refresh, Delete, View, Document, Timer, SuccessFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import { usePageStateStore } from '@/stores/pageState'
import request from '@/utils/request'
import logger from '@/utils/logger'

const route = useRoute()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const current = ref(1)
const size = ref(10)
const count = ref(0)
const open = ref(false)
const jobId = ref(0)
const jobLog = ref({})
const searchParams = ref({})
const jobGroups = ref([])
const jobLogIds = ref([])
const jobLogs = ref([])
const dateRange = ref([])

const successCount = computed(() => jobLogs.value.filter(l => l.status === 1).length)
const failCount = computed(() => jobLogs.value.filter(l => l.status === 0).length)
const isAllSelected = computed(() => jobLogs.value.length > 0 && jobLogIds.value.length === jobLogs.value.length)

const formatDateTime = (date) => {
  if (!date) return ''
  const d = new Date(date)
  return d.toLocaleString('zh-CN', { hour12: false })
}

const listJobGroups = async () => {
  try {
    const { data } = await request.get('/admin/jobLogs/jobGroups')
    jobGroups.value = data.data
  } catch (error) {
    logger.error('获取任务组列表失败:', error)
  }
}

const listJobLogs = async () => {
  loading.value = true
  try {
    const params = {
      ...searchParams.value,
      jobId: jobId.value === 0 ? null : jobId.value,
      current: current.value,
      size: size.value,
      startTime: dateRange.value?.[0],
      endTime: dateRange.value?.[1]
    }
    const { data } = await request.get('/admin/jobLogs', { params })
    if (data && data.data) {
      jobLogs.value = data.data.records || []
      count.value = data.data.count || 0
      jobLogIds.value = []
    }
  } catch (error) {
    logger.error('获取日志列表失败:', error)
  } finally {
    loading.value = false
  }
}

const searchLogs = () => {
  current.value = 1
  pageStateStore.updateQuartzLogState(jobId.value, current.value)
  listJobLogs()
}

const clearSearch = () => {
  searchParams.value = {}
  dateRange.value = []
  current.value = 1
  listJobLogs()
}

const toggleSelect = (id) => {
  const idx = jobLogIds.value.indexOf(id)
  if (idx > -1) {
    jobLogIds.value.splice(idx, 1)
  } else {
    jobLogIds.value.push(id)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    jobLogIds.value = []
  } else {
    jobLogIds.value = jobLogs.value.map(l => l.id)
  }
}

const deleteJobLogs = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${jobLogIds.value.length} 条日志吗？`,
      '批量删除',
      { type: 'warning', confirmButtonText: '确认删除', cancelButtonText: '取消' }
    )
    const { data } = await request.delete('/admin/jobLogs', { data: jobLogIds.value })
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '删除成功' })
      listJobLogs()
    } else {
      ElNotification.error({ title: '失败', message: '删除失败' })
    }
  } catch (error) {
    if (error !== 'cancel') {
      logger.error('删除日志失败:', error)
    }
  }
}

const clean = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有调度日志吗？此操作不可恢复！',
      '清空日志',
      { type: 'warning', confirmButtonText: '确认清空', cancelButtonText: '取消' }
    )
    const { data } = await request.delete('/admin/jobLogs/clean')
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '清空成功' })
      listJobLogs()
    } else {
      ElNotification.error({ title: '失败', message: '清空失败' })
    }
  } catch (error) {
    if (error !== 'cancel') {
      logger.error('清空日志失败:', error)
    }
  }
}

const changeOpen = (log) => {
  jobLog.value = { ...log, exceptionInfo: '\n' + (log.exceptionInfo || '') }
  open.value = true
}

const currentChange = (page) => {
  current.value = page
  pageStateStore.updateQuartzLogState(jobId.value, page)
  listJobLogs()
}

onMounted(() => {
  const quartzId = route.params.id
  if (quartzId === 'all') {
    jobId.value = 0
  } else if (quartzId) {
    jobId.value = Number(quartzId)
  }

  const savedState = pageStateStore.pageState.quartzLog
  if (jobId.value === savedState?.jobId) {
    current.value = savedState.current
  } else {
    current.value = 1
    pageStateStore.updateQuartzLogState(jobId.value, current.value)
  }

  listJobLogs()
  listJobGroups()
})
</script>

<style scoped>
.quartz-log-page {
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
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.25s ease;
}
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}
.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}
.stat-icon.primary {
  background: linear-gradient(135deg, #e8f0fe, #d2e3fc);
  color: #1a73e8;
}
.stat-icon.success {
  background: linear-gradient(135deg, #e6f4ea, #ceead6);
  color: #1e8e3e;
}
.stat-icon.danger {
  background: linear-gradient(135deg, #fce8e6, #fad2cf);
  color: #d93025;
}
.stat-info {
  display: flex;
  flex-direction: column;
}
.stat-value {
  font-size: 26px;
  font-weight: 700;
  color: var(--text-primary, #1f2937);
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  margin-top: 2px;
}

/* ========== 搜索筛选栏 ========== */
.filter-bar {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  padding: 16px 20px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}
.filter-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.filter-right {
  display: flex;
  gap: 8px;
}
.search-box {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 10px;
  color: var(--text-tertiary, #9ca3af);
  font-size: 15px;
  pointer-events: none;
}
.search-input {
  width: 200px;
  height: 34px;
  padding: 0 50px 0 32px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  background: var(--bg-body, #f9fafb);
  outline: none;
  transition: all 0.2s;
  box-sizing: border-box;
}
.search-input:focus {
  border-color: var(--color-primary, #1a73e8);
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.1);
  background: var(--bg-card, #fff);
}
.search-kbd {
  position: absolute;
  right: 8px;
  padding: 1px 6px;
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  background: var(--bg-body, #f3f4f6);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 4px;
  font-family: inherit;
  pointer-events: none;
}
.filter-btn {
  height: 34px;
  padding: 0 14px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  background: var(--bg-card, #fff);
  color: var(--text-primary, #374151);
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s;
}
.filter-btn:hover {
  border-color: var(--color-primary, #1a73e8);
  color: var(--color-primary, #1a73e8);
}
.filter-btn.primary {
  background: var(--color-primary, #1a73e8);
  border-color: var(--color-primary, #1a73e8);
  color: #fff;
}
.filter-btn.primary:hover {
  background: #1557b0;
}

/* ========== 操作栏 ========== */
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  padding: 0 4px;
}
.action-left {
  display: flex;
  align-items: center;
  gap: 10px;
}
.batch-check {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  user-select: none;
}
.batch-check input[type="checkbox"] {
  width: 15px;
  height: 15px;
  accent-color: var(--color-primary, #1a73e8);
  cursor: pointer;
}
.action-btn {
  height: 32px;
  padding: 0 12px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: all 0.2s;
  position: relative;
}
.action-btn.danger {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}
.action-btn.danger:hover:not(:disabled) {
  background: #dc2626;
  color: #fff;
}
.action-btn.danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.action-btn.danger-outline {
  background: transparent;
  color: #dc2626;
  border: 1px solid #fecaca;
}
.action-btn.danger-outline:hover {
  background: #fef2f2;
}
.badge {
  background: #dc2626;
  color: #fff;
  font-size: 11px;
  padding: 0 5px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
  line-height: 18px;
}
.job-id-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  background: var(--bg-body, #f3f4f6);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 20px;
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
}

/* ========== 表格 ========== */
.table-wrapper {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  overflow: hidden;
  min-height: 200px;
}
.elegant-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}
.elegant-table thead {
  background: var(--bg-body, #f9fafb);
}
.elegant-table th {
  padding: 12px 16px;
  text-align: left;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #6b7280);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color, #ebeef5);
  white-space: nowrap;
}
.elegant-table td {
  padding: 12px 16px;
  font-size: 13px;
  color: var(--text-primary, #374151);
  border-bottom: 1px solid var(--border-color, #f3f4f6);
  vertical-align: middle;
}
.table-row {
  transition: background 0.15s;
}
.table-row:hover {
  background: var(--bg-hover, #f9fafb);
}
.table-row.row-fail {
  background: rgba(239, 68, 68, 0.02);
}
.table-row.row-fail:hover {
  background: rgba(239, 68, 68, 0.05);
}
.col-check { width: 44px; text-align: center; }
.col-check input[type="checkbox"] { width: 15px; height: 15px; accent-color: var(--color-primary, #1a73e8); cursor: pointer; }
.col-index { width: 50px; text-align: center; color: var(--text-tertiary, #9ca3af); font-size: 12px; }
.col-status { width: 90px; }
.col-time { width: 170px; }
.col-action { width: 80px; text-align: center; }

.cell-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}
.name-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.name-dot.success { background: #22c55e; box-shadow: 0 0 6px rgba(34, 197, 94, 0.4); }
.name-dot.fail { background: #ef4444; box-shadow: 0 0 6px rgba(239, 68, 68, 0.4); }

.group-tag {
  display: inline-block;
  padding: 2px 8px;
  background: var(--bg-body, #f3f4f6);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
}
.cell-code {
  font-family: 'Cascadia Code', 'Fira Code', 'Courier New', monospace;
  font-size: 12px;
  background: var(--bg-body, #f3f4f6);
  padding: 2px 8px;
  border-radius: 4px;
  color: var(--text-secondary, #4b5563);
}
.cell-message {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
}
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}
.status-badge.success {
  background: #ecfdf5;
  color: #059669;
}
.status-badge.fail {
  background: #fef2f2;
  color: #dc2626;
}
.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}
.status-badge.success .status-dot { background: #059669; }
.status-badge.fail .status-dot { background: #dc2626; }
.cell-time {
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
  white-space: nowrap;
}
.action-btn.detail {
  background: #eff6ff;
  color: #2563eb;
  border: none;
  font-size: 12px;
  height: 28px;
  padding: 0 10px;
}
.action-btn.detail:hover {
  background: #2563eb;
  color: #fff;
}

/* ========== 空状态 ========== */
.empty-state {
  padding: 60px 20px;
  text-align: center;
}
.empty-icon {
  font-size: 48px;
  color: var(--border-color, #d1d5db);
  margin-bottom: 12px;
}
.empty-state p {
  font-size: 15px;
  color: var(--text-secondary, #6b7280);
  margin: 0 0 4px;
}
.empty-hint {
  font-size: 13px;
  color: var(--text-tertiary, #9ca3af);
}

/* ========== 分页 ========== */
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
  padding: 0 4px;
}
.page-info {
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
}

/* ========== 详情对话框 ========== */
.detail-dialog :deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
}
.detail-dialog :deep(.el-dialog__header) {
  padding: 0;
  margin: 0;
}
.detail-dialog :deep(.el-dialog__body) {
  padding: 0 24px 24px;
}
.detail-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}
.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 24px 20px;
  background: var(--bg-body, #fafafa);
  border-bottom: 1px solid var(--border-color, #f0f0f0);
}
.dialog-title-group {
  display: flex;
  align-items: center;
  gap: 14px;
}
.dialog-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}
.dialog-icon.success { background: #ecfdf5; color: #059669; }
.dialog-icon.fail { background: #fef2f2; color: #dc2626; }
.dialog-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}
.dialog-subtitle {
  margin: 3px 0 0;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
}
.dialog-status {
  padding: 4px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}
.dialog-status.success { background: #ecfdf5; color: #059669; }
.dialog-status.fail { background: #fef2f2; color: #dc2626; }

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
  padding-top: 20px;
}
.detail-item {
  padding: 12px 0;
  border-bottom: 1px solid var(--border-color, #f3f4f6);
}
.detail-item.full {
  grid-column: 1 / -1;
}
.detail-label {
  display: block;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  margin-bottom: 4px;
  font-weight: 500;
}
.detail-value {
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  word-break: break-all;
}
.detail-value.mono {
  font-family: 'Cascadia Code', 'Fira Code', 'Courier New', monospace;
  font-size: 13px;
}
.detail-value.message {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary, #4b5563);
}
.detail-code {
  font-family: 'Cascadia Code', 'Fira Code', 'Courier New', monospace;
  font-size: 13px;
  background: var(--bg-body, #f3f4f6);
  padding: 6px 12px;
  border-radius: 6px;
  display: block;
  color: var(--text-secondary, #4b5563);
  word-break: break-all;
}

.exception-block {
  margin-top: 4px;
}
.exception-block pre {
  background: #1f2937;
  color: #fca5a5;
  padding: 16px;
  border-radius: 10px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.7;
  margin: 0;
}
.exception-block code {
  font-family: 'Cascadia Code', 'Fira Code', 'Courier New', monospace;
}

.dialog-close-btn {
  height: 36px;
  padding: 0 24px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  background: var(--bg-card, #fff);
  color: var(--text-primary, #374151);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}
.dialog-close-btn:hover {
  background: var(--bg-body, #f9fafb);
}
</style>

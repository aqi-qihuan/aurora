<template>
  <div class="quartz-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Timer /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">任务总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><SuccessFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ activeJobs }}</span>
          <span class="stat-label">运行中</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><VideoPause /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ pausedJobs }}</span>
          <span class="stat-label">已暂停</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="handleAdd" class="btn-add">
            <span>新增任务</span>
          </el-button>
          <el-button :icon="Operation" @click="openLog" class="btn-log">
            <span>调度日志</span>
          </el-button>
          <el-button
            type="danger"
            :icon="Delete"
            :disabled="jobIds.length === 0"
            @click="isDelete = true"
            class="btn-batch-delete">
            <span>批量删除 ({{ jobIds.length }})</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-select v-model="searchParams.jobGroup" placeholder="任务分组" clearable @change="listJobs" class="filter-select">
            <template #prefix><el-icon><Grid /></el-icon></template>
            <el-option v-for="g in jobGroups" :key="g" :label="g" :value="g" />
          </el-select>
          <el-select v-model="searchParams.status" placeholder="任务状态" clearable @change="listJobs" class="filter-select">
            <template #prefix><el-icon><Sort /></el-icon></template>
            <el-option label="正常" :value="1" />
            <el-option label="暂停" :value="0" />
          </el-select>
          <el-input
            v-model="searchParams.jobName"
            :prefix-icon="Search"
            placeholder="搜索任务名..."
            class="search-input"
            clearable
            @keyup.enter="listJobs"
            @clear="listJobs" />
          <el-button type="primary" :icon="Search" @click="listJobs" circle />
        </div>
      </div>

      <!-- 现代化表格 -->
      <el-table
        :data="jobs"
        v-loading="loading"
        @selection-change="selectionChange"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }"
        row-key="id">
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="jobName" label="任务名称" min-width="160" align="left">
          <template #default="{ row }">
            <div class="job-name-cell">
              <div class="job-icon" :class="row.status === 1 ? 'running' : 'paused'">
                <el-icon><Timer /></el-icon>
              </div>
              <span class="job-name-text">{{ row.jobName }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="jobGroup" label="任务分组" width="140" align="center">
          <template #default="{ row }">
            <span class="group-badge">{{ row.jobGroup }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="cronExpression" label="Cron表达式" min-width="160" align="center">
          <template #default="{ row }">
            <span class="cron-expression">{{ row.cronExpression }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="invokeTarget" label="调用目标" min-width="180" align="left" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="invoke-target">{{ row.invokeTarget }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <div class="status-cell">
              <span class="status-dot" :class="row.status === 1 ? 'active' : 'inactive'"></span>
              <span class="status-text" :class="row.status === 1 ? 'active' : 'inactive'">
                {{ row.status === 1 ? '正常' : '暂停' }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDateTime(row.createTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="编辑" placement="top" :show-after="500">
                <button class="action-btn edit" @click="handleChange(row.id)"><el-icon><Edit /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="执行一次" placement="top" :show-after="500">
                <button class="action-btn run" @click="handleRun(row)"><el-icon><VideoPlay /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="查看详情" placement="top" :show-after="500">
                <button class="action-btn view" @click="handleView(row)"><el-icon><View /></el-icon></button>
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
    <el-dialog v-model="isDelete" width="400px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper danger"><el-icon><Warning /></el-icon></div>
      <div class="dialog-content">
        <h3>确认删除</h3>
        <p>确定要删除选中的 {{ jobIds.length }} 个任务吗？此操作不可恢复。</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="isDelete = false" class="btn-cancel">取消</el-button>
          <el-button type="danger" @click="deleteJobs(null)" class="btn-confirm-danger">确认删除</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogFormVisible" width="700px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><EditPen /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ title }}</h3>
        <el-form ref="dataFormRef" :model="job" :rules="rules" class="job-form" label-position="top">
          <div class="form-grid">
            <el-form-item label="任务名称" prop="jobName">
              <el-input v-model="job.jobName" placeholder="请输入任务名称" class="form-input" />
            </el-form-item>
            <el-form-item label="任务分组" prop="jobGroup">
              <el-input v-model="job.jobGroup" placeholder="请输入任务分组" class="form-input" />
            </el-form-item>
          </div>
          <el-form-item prop="invokeTarget">
            <template #label>
              <span>调用方法 <el-tooltip placement="top"><template #content>Bean调用: auroraQuartz.blogParams('blog')<br/>Class调用: com.aurora.quartz.AuroraQuartz.blogParams('blog')</template><el-icon style="margin-left: 4px; color: var(--text-secondary)"><QuestionFilled /></el-icon></el-tooltip></span>
            </template>
            <el-input v-model="job.invokeTarget" placeholder="请输入调用目标字符串" class="form-input">
              <template #prefix><el-icon><Position /></el-icon></template>
            </el-input>
          </el-form-item>
          <el-form-item label="Cron表达式" prop="cronExpression">
            <el-input v-model="job.cronExpression" placeholder="请输入cron执行表达式" class="form-input">
              <template #prefix><el-icon><Clock /></el-icon></template>
              <template #append>
                <button class="cron-gen-btn" type="button" @click="handleShowCron">
                  <el-icon><SetUp /></el-icon>
                  生成表达式
                </button>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="错误策略" prop="misfirePolicy">
            <div class="policy-group">
              <div class="policy-card" :class="{ active: job.misfirePolicy === 0 }" @click="job.misfirePolicy = 0"><span>默认</span></div>
              <div class="policy-card" :class="{ active: job.misfirePolicy === 1 }" @click="job.misfirePolicy = 1"><span>立即执行</span></div>
              <div class="policy-card" :class="{ active: job.misfirePolicy === 2 }" @click="job.misfirePolicy = 2"><span>执行一次</span></div>
              <div class="policy-card" :class="{ active: job.misfirePolicy === 3 }" @click="job.misfirePolicy = 3"><span>放弃执行</span></div>
            </div>
          </el-form-item>
          <div class="form-grid">
            <el-form-item label="是否并发">
              <div class="radio-card-group">
                <div class="radio-card" :class="{ active: job.concurrent === 0 }" @click="job.concurrent = 0"><el-icon><Check /></el-icon><span>允许</span></div>
                <div class="radio-card" :class="{ active: job.concurrent === 1 }" @click="job.concurrent = 1"><el-icon><Close /></el-icon><span>禁止</span></div>
              </div>
            </el-form-item>
            <el-form-item label="状态">
              <div class="radio-card-group">
                <div class="radio-card" :class="{ active: job.status === 1 }" @click="job.status = 1"><el-icon><SuccessFilled /></el-icon><span>正常</span></div>
                <div class="radio-card" :class="{ active: job.status === 0 }" @click="job.status = 0"><el-icon><VideoPause /></el-icon><span>暂停</span></div>
              </div>
            </el-form-item>
          </div>
          <el-form-item label="备注" prop="remark">
            <el-input v-model="job.remark" placeholder="备注信息" type="textarea" class="form-input" :rows="3" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogFormVisible = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="handleEditOrUpdate" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Cron表达式生成器 -->
    <el-dialog v-model="openCron" width="800px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><Clock /></el-icon></div>
      <div class="dialog-content">
        <h3>Cron 表达式生成器</h3>
      </div>
      <Crontab v-model="job.cronExpression" style="margin-top: 16px;" />
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="openCron = false" class="btn-cancel">关闭</el-button>
          <el-button type="primary" @click="openCron = false" class="btn-confirm">确认</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 任务详细对话框 -->
    <el-dialog v-model="openView" width="600px" class="modern-dialog" :show-close="false" @closed="job = {}">
      <div class="dialog-icon-wrapper primary"><el-icon><View /></el-icon></div>
      <div class="dialog-content">
        <h3>任务详细</h3>
      </div>
      <div class="detail-grid" style="margin-top: 20px;">
        <div class="detail-item">
          <span class="detail-label">任务编号</span>
          <span class="detail-value">{{ job.id }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">任务名称</span>
          <span class="detail-value">{{ job.jobName }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">任务分组</span>
          <span class="detail-value">{{ job.jobGroup }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">创建时间</span>
          <span class="detail-value">{{ formatDateTime(job.createTime) }}</span>
        </div>
        <div class="detail-item full-width">
          <span class="detail-label">Cron表达式</span>
          <span class="detail-value cron-value">{{ job.cronExpression }}</span>
        </div>
        <div class="detail-item full-width">
          <span class="detail-label">调用目标</span>
          <span class="detail-value invoke-value">{{ job.invokeTarget }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">下次执行</span>
          <span class="detail-value">{{ formatDateTime(job.nextValidTime) }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">任务状态</span>
          <span class="detail-value" :class="job.status === 1 ? 'status-active' : 'status-paused'">{{ job.status === 1 ? '正常' : '暂停' }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">是否并发</span>
          <span class="detail-value">{{ job.concurrent === 0 ? '允许' : '禁止' }}</span>
        </div>
        <div class="detail-item">
          <span class="detail-label">执行策略</span>
          <span class="detail-value">{{ ['默认策略','立即执行','执行一次','放弃执行'][job.misfirePolicy] || '默认策略' }}</span>
        </div>
        <div class="detail-item full-width" v-if="job.remark">
          <span class="detail-label">备注</span>
          <span class="detail-value">{{ job.remark }}</span>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="openView = false" class="btn-cancel">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Search, Plus, Delete, Operation, Clock, Edit, EditPen,
  Warning, QuestionFilled, Timer, View, VideoPlay,
  VideoPause, SuccessFilled, Grid, Sort, Position, SetUp, Check, Close
} from '@element-plus/icons-vue'
import Crontab from '@/components/Crontab/index.vue'
import { usePageStateStore } from '@/stores/pageState'
import request from '@/utils/request'
import logger from '@/utils/logger'

const router = useRouter()
const pageStateStore = usePageStateStore()

const dataFormRef = ref(null)
const loading = ref(true)
const isDelete = ref(false)
const current = ref(1)
const size = ref(10)
const count = ref(0)
const dialogFormVisible = ref(false)
const openCron = ref(false)
const openView = ref(false)
const title = ref('')
const editOrUpdate = ref(true)

const searchParams = ref({})
const jobGroups = ref([])
const jobs = ref([])
const jobIds = ref([])
const job = ref({})

const rules = {
  jobName: [{ required: true, message: '任务名称不能为空', trigger: 'blur' }],
  invokeTarget: [{ required: true, message: '调用目标字符串不能为空', trigger: 'blur' }],
  cronExpression: [{ required: true, message: 'cron执行表达式不能为空', trigger: 'blur' }]
}

const activeJobs = computed(() => jobs.value.filter(j => j.status === 1).length)
const pausedJobs = computed(() => jobs.value.filter(j => j.status === 0).length)

const formatDateTime = (date) => {
  if (!date) return ''
  const d = new Date(date)
  return d.toLocaleString('zh-CN', { hour12: false })
}

const listJobGroups = async () => {
  try {
    const { data } = await request.get('/admin/jobs/jobGroups')
    jobGroups.value = data.data
  } catch (error) {
    logger.error('获取任务组列表失败:', error)
  }
}

const listJobs = async () => {
  loading.value = true
  try {
    const { data } = await request.get('/admin/jobs', {
      params: { ...searchParams.value, current: current.value, size: size.value }
    })
    if (data && data.data) {
      jobs.value = data.data.records || []
      count.value = data.data.count || 0
    }
  } catch (error) {
    logger.error('获取任务列表失败:', error)
  } finally {
    loading.value = false
  }
}

const reset = () => { searchParams.value = {}; listJobs() }
const selectionChange = (selection) => { jobIds.value = selection.map(item => item.id) }
const sizeChange = (val) => { size.value = val; listJobs() }

const changeStatus = async (job) => {
  try {
    const { data } = await request.put('/admin/jobs/status', { id: job.id, status: job.status })
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '修改成功' })
      listJobs()
    } else {
      ElNotification.error({ title: '失败', message: '修改失败' })
    }
  } catch (error) {
    logger.error('修改状态失败:', error)
  }
}

const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该任务吗？', '提示', {
    confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning'
  }).then(() => { deleteJobs(id) }).catch(() => {})
}

const deleteJobs = async (id) => {
  const ids = id ? [id] : jobIds.value
  try {
    const { data } = await request.delete('/admin/jobs', { data: ids })
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '删除成功' })
      listJobs()
    } else {
      ElNotification.error({ title: '失败', message: '删除失败' })
    }
  } catch (error) {
    logger.error('删除任务失败:', error)
  } finally {
    isDelete.value = false
  }
}

const handleShowCron = () => { openCron.value = true }

const handleAdd = () => {
  editOrUpdate.value = false
  job.value = { status: 1, concurrent: 0, misfirePolicy: 0 }
  title.value = '新增任务'
  dialogFormVisible.value = true
}

const handleChange = async (jobId) => {
  editOrUpdate.value = true
  title.value = '编辑任务'
  try {
    const { data } = await request.get(`/admin/jobs/${jobId}`)
    job.value = data.data
  } catch (error) {
    logger.error('获取任务详情失败:', error)
  }
  dialogFormVisible.value = true
}

const handleEditOrUpdate = async () => {
  if (!dataFormRef.value) return
  await dataFormRef.value.validate(async (valid) => {
    if (!valid) return
    try {
      const method = editOrUpdate.value ? 'put' : 'post'
      const { data } = await request[method]('/admin/jobs', job.value)
      if (data.flag) {
        ElNotification.success({ title: editOrUpdate.value ? '修改成功' : '添加成功', message: data.message })
        listJobs()
      } else {
        ElNotification.error({ title: editOrUpdate.value ? '修改失败' : '添加失败', message: data.message })
      }
      dialogFormVisible.value = false
    } catch (error) {
      logger.error('保存任务失败:', error)
    }
  })
}

const handleRun = async (job) => {
  try {
    const { data } = await request.put('/admin/jobs/run', { id: job.id, jobGroup: job.jobGroup })
    if (data.flag) {
      ElNotification.success({ title: '执行成功', message: data.message })
    } else {
      ElNotification.error({ title: '执行失败', message: data.message })
    }
  } catch (error) {
    logger.error('执行任务失败:', error)
  }
}

const handleView = (jobData) => { openView.value = true; job.value = jobData }
const handleJobLog = (jobId) => { pageStateStore.updateQuartzLogState(jobId, 1); router.push({ path: `/quartz/log/${jobId}` }) }
const openLog = () => { router.push({ path: '/quartz/log/all' }) }
const currentChange = (page) => { current.value = page; pageStateStore.updatePageState('quartz', page); listJobs() }

onMounted(() => {
  current.value = pageStateStore.pageState.quartz || 1
  listJobGroups()
  listJobs()
})
</script>

<style scoped>
.quartz-page { padding: 0; }

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
.btn-log { border-radius: 10px; font-weight: 500; height: 40px; padding: 0 20px; transition: all 0.2s ease; }
.btn-log:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }
.btn-batch-delete { border-radius: 10px; font-weight: 500; height: 40px; padding: 0 20px; transition: all 0.2s ease; }
.btn-batch-delete:not(:disabled):hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239,68,68,0.4); }
.filter-select { width: 140px; }
.filter-select :deep(.el-input__wrapper) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); }
.search-input { width: 220px; }
.search-input :deep(.el-input__wrapper) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); transition: all 0.2s ease; }
.search-input :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6; }

/* 表格 */
.modern-table { border-radius: 12px; overflow: hidden; border: 1px solid var(--border-default, #e5e7eb); }
.modern-table :deep(.el-table__header-wrapper th) { background: var(--bg-elevated, #f9fafb); color: var(--text-secondary, #6b7280); font-weight: 600; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; padding: 16px 12px; border-bottom: 1px solid var(--border-default, #e5e7eb); }
.modern-table :deep(.el-table__body tr) { transition: all 0.2s ease; }
.modern-table :deep(.el-table__body tr:hover > td) { background: var(--bg-hover, #f3f4f6) !important; }
.modern-table :deep(.el-table__body td) { padding: 16px 12px; border-bottom: 1px solid var(--border-light, #f3f4f6); }

/* 任务名称 */
.job-name-cell { display: flex; align-items: center; gap: 12px; }
.job-icon { width: 36px; height: 36px; border-radius: 10px; display: flex; align-items: center; justify-content: center; font-size: 16px; flex-shrink: 0; }
.job-icon.running { background: linear-gradient(135deg, #10b981, #34d399); color: #fff; }
.job-icon.paused { background: var(--bg-elevated, #e5e7eb); color: var(--text-secondary, #6b7280); }
.job-name-text { font-weight: 500; color: var(--text-primary, #1f2937); }

/* 分组徽章 */
.group-badge { display: inline-block; padding: 4px 12px; background: var(--bg-elevated, #f3f4f6); border-radius: 6px; font-size: 13px; font-weight: 500; color: var(--text-secondary, #6b7280); }

/* Cron表达式 */
.cron-expression { font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace; font-size: 13px; padding: 4px 10px; background: var(--bg-elevated, #f3f4f6); border-radius: 6px; color: var(--color-primary, #3b82f6); }

/* 调用目标 */
.invoke-target { font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace; font-size: 13px; color: var(--text-secondary, #6b7280); }

/* 状态 */
.status-cell { display: flex; align-items: center; justify-content: center; gap: 6px; }
.status-dot { width: 8px; height: 8px; border-radius: 50%; }
.status-dot.active { background: #10b981; box-shadow: 0 0 0 3px rgba(16,185,129,0.2); }
.status-dot.inactive { background: #9ca3af; }
.status-text { font-size: 13px; font-weight: 500; }
.status-text.active { color: #10b981; }
.status-text.inactive { color: #9ca3af; }

/* 时间 */
.time-cell { display: flex; align-items: center; justify-content: center; gap: 8px; color: var(--text-secondary, #6b7280); font-size: 14px; }
.time-icon { color: #3b82f6; }

/* 操作按钮 */
.action-btns { display: flex; justify-content: center; gap: 6px; }
.action-btn { width: 34px; height: 34px; border-radius: 8px; border: none; cursor: pointer; display: flex; align-items: center; justify-content: center; transition: all 0.2s ease; font-size: 15px; }
.action-btn.edit { background: #eff6ff; color: #3b82f6; }
.action-btn.edit:hover { background: #3b82f6; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.3); }
.action-btn.run { background: #f0fdf4; color: #10b981; }
.action-btn.run:hover { background: #10b981; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(16,185,129,0.3); }
.action-btn.view { background: #fefce8; color: #f59e0b; }
.action-btn.view:hover { background: #f59e0b; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(245,158,11,0.3); }
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
.job-form { margin-top: 20px; text-align: left; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
.form-input :deep(.el-input__wrapper) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); height: 44px; }
.form-input :deep(.el-input__wrapper.is-focus) { box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6; }
.form-input :deep(.el-textarea__inner) { border-radius: 10px; box-shadow: 0 0 0 1px var(--border-default, #e5e7eb); }
.form-input { width: 100%; }

/* Cron生成按钮 */
.cron-gen-btn {
  display: flex; align-items: center; gap: 4px;
  padding: 0 14px; height: 100%;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff; border: none; border-radius: 0 10px 10px 0;
  font-size: 13px; font-weight: 500;
  cursor: pointer; transition: all 0.2s ease;
}
.cron-gen-btn:hover { opacity: 0.9; }

/* 策略卡片 */
.policy-group { display: flex; gap: 8px; flex-wrap: wrap; }
.policy-card {
  padding: 8px 16px; border: 2px solid var(--border-default, #e5e7eb);
  border-radius: 8px; cursor: pointer;
  font-size: 13px; font-weight: 500;
  color: var(--text-secondary, #6b7280);
  transition: all 0.2s ease;
}
.policy-card:hover { border-color: var(--color-primary-light, #93c5fd); }
.policy-card.active { border-color: var(--color-primary, #3b82f6); color: var(--color-primary, #3b82f6); background: rgba(59,130,246,0.08); }

/* 单选卡片 */
.radio-card-group { display: flex; gap: 12px; }
.radio-card {
  flex: 1; display: flex; align-items: center; justify-content: center; gap: 6px;
  padding: 10px 20px; border: 2px solid var(--border-default, #e5e7eb);
  border-radius: 10px; cursor: pointer;
  font-size: 14px; font-weight: 500; color: var(--text-secondary, #6b7280);
  transition: all 0.2s ease;
}
.radio-card:hover { border-color: var(--color-primary-light, #93c5fd); }
.radio-card.active { border-color: var(--color-primary, #3b82f6); color: var(--color-primary, #3b82f6); background: rgba(59,130,246,0.08); }

/* 详情网格 */
.detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; text-align: left; }
.detail-item { padding: 14px 16px; background: var(--bg-elevated, #f9fafb); border-radius: 10px; }
.detail-item.full-width { grid-column: span 2; }
.detail-label { display: block; font-size: 12px; font-weight: 500; color: var(--text-secondary, #6b7280); margin-bottom: 4px; text-transform: uppercase; letter-spacing: 0.05em; }
.detail-value { display: block; font-size: 14px; font-weight: 500; color: var(--text-primary, #1f2937); word-break: break-all; }
.cron-value { font-family: 'SF Mono', monospace; color: var(--color-primary, #3b82f6); }
.invoke-value { font-family: 'SF Mono', monospace; font-size: 13px; }
.status-active { color: #10b981; }
.status-paused { color: #f59e0b; }

/* 深色模式 */
[data-theme="dark"] .stat-card { background: var(--bg-base, #1f2937); border-color: var(--border-default, #374151); }
[data-theme="dark"] .stat-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .main-card { background: var(--bg-base, #1f2937); border-color: var(--border-default, #374151); }
[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) { background: var(--bg-elevated, #374151); color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) { background: var(--bg-hover, #374151) !important; }
[data-theme="dark"] .job-name-text { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .group-badge { background: var(--bg-elevated, #374151); color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .cron-expression { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .job-icon.paused { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .action-btn.edit { background: rgba(59,130,246,0.15); }
[data-theme="dark"] .action-btn.run { background: rgba(16,185,129,0.15); }
[data-theme="dark"] .action-btn.view { background: rgba(245,158,11,0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239,68,68,0.15); }
[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .dialog-content p { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .detail-item { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .detail-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .policy-card { border-color: var(--border-default, #374151); color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .radio-card { border-color: var(--border-default, #374151); color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .radio-card.active { background: rgba(59,130,246,0.1); }
[data-theme="dark"] .form-input :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .form-input :deep(.el-textarea__inner) { background: var(--bg-elevated, #374151); color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .filter-select :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .search-input :deep(.el-input__wrapper) { background: var(--bg-elevated, #374151); }

/* 响应式 */
@media (max-width: 1024px) { .stats-row { grid-template-columns: repeat(2, 1fr); } .stat-card:last-child { grid-column: span 2; } }
@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; } .stat-card:last-child { grid-column: span 1; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; flex-wrap: wrap; }
  .btn-add, .btn-log, .btn-batch-delete { width: 100%; }
  .search-input, .filter-select { width: 100%; flex: 1; }
  .form-grid { grid-template-columns: 1fr; }
  .detail-grid { grid-template-columns: 1fr; }
  .detail-item.full-width { grid-column: span 1; }
  .policy-group { flex-wrap: wrap; }
  .pagination-wrapper { justify-content: center; }
}
</style>

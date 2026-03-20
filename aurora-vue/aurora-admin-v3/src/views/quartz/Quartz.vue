<template>
  <el-card class="main-card">
    <el-form ref="queryFormRef" :inline="true" label-width="68px">
      <el-form-item label="任务名称" prop="jobName">
        <el-input
          v-model="searchParams.jobName"
          placeholder="请输入任务名称"
          clearable
          @keyup.enter="listJobs"
        />
      </el-form-item>
      <el-form-item label="任务组名" prop="jobGroup">
        <el-select
          v-model="searchParams.jobGroup"
          placeholder="请选择任务组名"
          clearable
          @change="listJobs"
        >
          <el-option v-for="jobGroup in jobGroups" :key="jobGroup" :label="jobGroup" :value="jobGroup" />
        </el-select>
      </el-form-item>
      <el-form-item label="任务状态" prop="status">
        <el-select v-model="searchParams.status" placeholder="请选择任务状态" clearable @change="listJobs">
          <el-option label="正常" :value="1" />
          <el-option label="暂停" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="listJobs">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
        <el-button @click="reset">
          <el-icon><Refresh /></el-icon>
          重置
        </el-button>
      </el-form-item>
    </el-form>

    <el-row :gutter="10" class="mb8">
      <el-col :span="1.5">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增
        </el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="info" @click="openLog">
          <el-icon><Operation /></el-icon>
          日志
        </el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button
          type="danger"
          :disabled="jobIds.length === 0"
          @click="isDelete = true"
        >
          <el-icon><Delete /></el-icon>
          批量删除
        </el-button>
      </el-col>
    </el-row>

    <el-table
      v-loading="loading"
      :data="jobs"
      border
      style="width: 100%"
      @selection-change="selectionChange"
    >
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column label="任务名称" width="160" align="center" prop="jobName" show-overflow-tooltip />
      <el-table-column label="任务组名" align="center" prop="jobGroup">
        <template #default="scope">
          <el-tag>{{ scope.row.jobGroup }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="调用目标字符串" align="center" prop="invokeTarget" show-overflow-tooltip />
      <el-table-column label="cron执行表达式" align="center" prop="cronExpression" show-overflow-tooltip />
      <el-table-column label="状态" align="center">
        <template #default="scope">
          <el-switch
            v-model="scope.row.status"
            active-color="#13ce66"
            inactive-color="#F4F4F5"
            :active-value="1"
            :inactive-value="0"
            @change="changeStatus(scope.row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="创建时间" align="center" width="160">
        <template #default="scope">
          {{ formatDateTime(scope.row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="备注" align="center" width="160" prop="remark" />
      <el-table-column label="操作" align="center" width="200">
        <template #default="scope">
          <el-button link type="primary" @click="handleChange(scope.row.id)">编辑</el-button>
          <el-popconfirm title="确定删除吗?" @confirm="deleteJobs(scope.row.id)">
            <template #reference>
              <el-button link type="danger" style="margin-left: 10px">删除</el-button>
            </template>
          </el-popconfirm>
          <el-dropdown style="margin-left: 10px" @command="(command) => handleCommand(command, scope.row)">
            <el-button link type="primary">
              更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="handleRun">
                  <el-icon><CaretRight /></el-icon>执行一次
                </el-dropdown-item>
                <el-dropdown-item command="handleView">
                  <el-icon><View /></el-icon>任务详细
                </el-dropdown-item>
                <el-dropdown-item command="handleJobLog">
                  <el-icon><Operation /></el-icon>调度日志
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-model:current-page="current"
      v-model:page-size="size"
      :total="count"
      :page-sizes="[10, 20]"
      background
      layout="total, sizes, prev, pager, next, jumper"
      class="pagination-container"
      @size-change="listJobs"
      @current-change="currentChange"
    />

    <!-- 批量删除确认对话框 -->
    <el-dialog v-model="isDelete" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项?</div>
      <template #footer>
        <el-button @click="isDelete = false">取消</el-button>
        <el-button type="primary" @click="deleteJobs(null)">确定</el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogFormVisible" :title="title" width="800px" append-to-body>
      <el-form ref="dataFormRef" :model="job" :rules="rules" label-width="120px">
        <el-row>
          <el-col :span="12">
            <el-form-item label="任务名称" prop="jobName">
              <el-input v-model="job.jobName" placeholder="请输入任务名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="任务分组" prop="jobGroup">
              <el-input v-model="job.jobGroup" placeholder="请输入任务分组" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item prop="invokeTarget">
              <template #label>
                <span>
                  调用方法
                  <el-tooltip placement="top">
                    <template #content>
                      Bean调用示例:auroraQuartz.blogParams('blog')<br />
                      Class类调用示例:com.aurora.quartz.AuroraQuartz.blogParams('blog')<br />
                      参数说明:支持字符串,布尔类型,长整型,浮点型,整型
                    </template>
                    <el-icon style="margin-left: 4px"><QuestionFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-input v-model="job.invokeTarget" placeholder="请输入调用目标字符串" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="cron表达式" prop="cronExpression">
              <el-input v-model="job.cronExpression" placeholder="请输入cron执行表达式">
                <template #append>
                  <el-button type="primary" @click="handleShowCron">
                    生成表达式
                    <el-icon style="margin-left: 4px"><Clock /></el-icon>
                  </el-button>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="错误策略" prop="misfirePolicy">
              <el-radio-group v-model="job.misfirePolicy">
                <el-radio-button :value="0">默认策略</el-radio-button>
                <el-radio-button :value="1">立即执行</el-radio-button>
                <el-radio-button :value="2">执行一次</el-radio-button>
                <el-radio-button :value="3">放弃执行</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="是否并发" prop="concurrent">
              <el-radio-group v-model="job.concurrent">
                <el-radio-button :value="0">允许</el-radio-button>
                <el-radio-button :value="1">禁止</el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-radio-group v-model="job.status">
                <el-radio :value="1">正常</el-radio>
                <el-radio :value="0">暂停</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="备注" prop="remark">
              <el-input v-model="job.remark" placeholder="备注信息" type="textarea" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogFormVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditOrUpdate">确定</el-button>
      </template>
    </el-dialog>

    <!-- Cron表达式生成器 -->
    <el-dialog v-model="openCron" title="Cron表达式生成器" append-to-body destroy-on-close>
      <Crontab v-model="job.cronExpression" />
    </el-dialog>

    <!-- 任务详细对话框 -->
    <el-dialog v-model="openView" title="任务详细" width="700px" append-to-body @closed="job = {}">
      <el-form :model="job" label-width="120px" size="small">
        <el-row>
          <el-col :span="12">
            <el-form-item label="任务编号:">{{ job.id }}</el-form-item>
            <el-form-item label="任务名称:">{{ job.jobName }}</el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="任务分组:">{{ job.jobGroup }}</el-form-item>
            <el-form-item label="创建时间:">{{ formatDateTime(job.createTime) }}</el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="cron表达式:">{{ job.cronExpression }}</el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下次执行时间:">{{ formatDateTime(job.nextValidTime) }}</el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="调用目标方法:">{{ job.invokeTarget }}</el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="任务状态:">
              {{ job.status === 1 ? '正常' : '暂停' }}
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="是否并发:">
              {{ job.concurrent === 0 ? '允许' : '禁止' }}
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="执行策略:">
              {{ ['默认策略', '立即执行', '执行一次', '放弃执行'][job.misfirePolicy] || '默认策略' }}
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="备注:">{{ job.remark }}</el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="openView = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Search,
  Refresh,
  Plus,
  Delete,
  Operation,
  ArrowDown,
  CaretRight,
  View,
  Warning,
  QuestionFilled,
  Clock
} from '@element-plus/icons-vue'
import Crontab from '@/components/Crontab/index.vue'
import { usePageStateStore } from '@/stores/pageState'
import request from '@/utils/request'

const router = useRouter()
const pageStateStore = usePageStateStore()

const queryFormRef = ref(null)
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
    console.error('获取任务组列表失败:', error)
  }
}

const listJobs = async () => {
  loading.value = true
  try {
    const { data } = await request.get('/admin/jobs', {
      params: {
        ...searchParams.value,
        current: current.value,
        size: size.value
      }
    })
    if (data && data.data) {
      jobs.value = data.data.records || []
      count.value = data.data.count || 0
    }
  } catch (error) {
    console.error('获取任务列表失败:', error)
  } finally {
    loading.value = false
  }
}

const reset = () => {
  searchParams.value = {}
  listJobs()
}

const selectionChange = (selection) => {
  jobIds.value = selection.map(item => item.id)
}

const changeStatus = async (job) => {
  try {
    const { data } = await request.put('/admin/jobs/status', {
      id: job.id,
      status: job.status
    })
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '修改成功'
      })
      listJobs()
    } else {
      ElNotification.error({
        title: '失败',
        message: '修改失败'
      })
    }
  } catch (error) {
    console.error('修改状态失败:', error)
  }
}

const deleteJobs = async (id) => {
  const ids = id ? [id] : jobIds.value
  try {
    const { data } = await request.delete('/admin/jobs', { data: ids })
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '删除成功'
      })
      listJobs()
    } else {
      ElNotification.error({
        title: '失败',
        message: '删除失败'
      })
    }
  } catch (error) {
    console.error('删除任务失败:', error)
  } finally {
    isDelete.value = false
  }
}

const handleShowCron = () => {
  openCron.value = true
}

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
    const { data } = await request.get(`/api/admin/jobs/${jobId}`)
    job.value = data.data
  } catch (error) {
    console.error('获取任务详情失败:', error)
  }
  dialogFormVisible.value = true
}

const handleEditOrUpdate = async () => {
  if (!dataFormRef.value) return
  
  await dataFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    try {
      const url = editOrUpdate.value ? '/admin/jobs' : '/admin/jobs'
      const method = editOrUpdate.value ? 'put' : 'post'
      const { data } = await request[method](url, job.value)
      
      if (data.flag) {
        ElNotification.success({
          title: editOrUpdate.value ? '修改成功' : '添加成功',
          message: data.message
        })
        listJobs()
      } else {
        ElNotification.error({
          title: editOrUpdate.value ? '修改失败' : '添加失败',
          message: data.message
        })
      }
      dialogFormVisible.value = false
    } catch (error) {
      console.error('保存任务失败:', error)
    }
  })
}

const handleCommand = (command, row) => {
  switch (command) {
    case 'handleRun':
      handleRun(row)
      break
    case 'handleView':
      handleView(row)
      break
    case 'handleJobLog':
      handleJobLog(row.id)
      break
  }
}

const handleRun = async (job) => {
  try {
    const { data } = await request.put('/admin/jobs/run', {
      id: job.id,
      jobGroup: job.jobGroup
    })
    if (data.flag) {
      ElNotification.success({
        title: '执行成功',
        message: data.message
      })
    } else {
      ElNotification.error({
        title: '执行失败',
        message: data.message
      })
    }
  } catch (error) {
    console.error('执行任务失败:', error)
  }
}

const handleView = (jobData) => {
  openView.value = true
  job.value = jobData
}

const handleJobLog = (jobId) => {
  pageStateStore.updateQuartzLogState(jobId, 1)
  router.push({ path: `/quartz/log/${jobId}` })
}

const openLog = () => {
  router.push({ path: '/quartz/log/all' })
}

const currentChange = (page) => {
  current.value = page
  pageStateStore.updatePageState('quartz', page)
  listJobs()
}

onMounted(() => {
  current.value = pageStateStore.pageState.quartz || 1
  listJobGroups()
  listJobs()
})
</script>

<style scoped>
.mb8 {
  margin-bottom: 8px;
}

.dialog-title-container {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: bold;
}

:deep(.el-textarea__inner) {
  resize: none !important;
}
</style>

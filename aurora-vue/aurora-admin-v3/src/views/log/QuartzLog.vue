<template>
  <el-card class="main-card">
    <el-form ref="formRef" :inline="true" label-width="68px">
      <el-row>
        <el-form-item label="任务名称">
          <el-input
            v-model="searchParams.jobName"
            placeholder="请输入任务名称"
            style="width: 200px"
            @keyup.enter="searchLogs"
          />
        </el-form-item>
        <el-form-item label="任务组名">
          <el-select
            v-model="searchParams.jobGroup"
            placeholder="请选择组名"
            clearable
            @change="listJobLogs"
          >
            <el-option v-for="jobGroup in jobGroups" :key="jobGroup" :label="jobGroup" :value="jobGroup" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行状态">
          <el-select
            v-model="searchParams.status"
            placeholder="请选择任务状态"
            clearable
            @change="listJobLogs"
          >
            <el-option value="1" label="成功" />
            <el-option value="0" label="失败" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行时间">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchLogs">
            <el-icon><Search /></el-icon>
            查找
          </el-button>
          <el-button @click="clearSearch">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-row>
    </el-form>

    <el-row :gutter="10" class="mb8">
      <el-col :span="1.5">
        <el-button
          type="danger"
          :disabled="!jobLogIds.length"
          @click="deleteJobLogs"
        >
          <el-icon><Delete /></el-icon>
          批量删除
        </el-button>
      </el-col>
      <el-col :span="1.5">
        <el-button type="danger" @click="clean">
          <el-icon><Delete /></el-icon>
          清空
        </el-button>
      </el-col>
    </el-row>

    <el-table
      v-loading="loading"
      :data="jobLogs"
      :default-sort="{ prop: 'createTime', order: 'descending' }"
      style="width: 100%"
      @selection-change="selectionChange"
    >
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column label="日志编号" width="80" align="center">
        <template #default="scope">
          {{ scope.$index + 1 }}
        </template>
      </el-table-column>
      <el-table-column label="任务名称" align="center" prop="jobName" show-overflow-tooltip />
      <el-table-column label="任务组名" align="center" prop="jobGroup" show-overflow-tooltip>
        <template #default="scope">
          <el-tag>{{ scope.row.jobGroup }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="调用目标字符串" align="center" prop="invokeTarget" show-overflow-tooltip />
      <el-table-column label="日志信息" align="center" prop="jobMessage" show-overflow-tooltip />
      <el-table-column label="执行状态" align="center" prop="status">
        <template #default="scope">
          <el-tag v-if="scope.row.status === 1" type="success">成功</el-tag>
          <el-tag v-else type="danger">失败</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="执行时间" align="center" width="180">
        <template #default="scope">
          {{ formatDateTime(scope.row.startTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center">
        <template #default="scope">
          <el-button link type="primary" @click="changeOpen(scope.row)">
            <el-icon><View /></el-icon>
            详细
          </el-button>
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
      @size-change="listJobLogs"
      @current-change="currentChange"
    />

    <!-- 详细对话框 -->
    <el-dialog
      v-model="open"
      title="调度日志详细"
      :width="jobLog.status === 1 ? '700px' : '80%'"
      append-to-body
      destroy-on-close
    >
      <el-form :model="jobLog" label-width="100px" size="small">
        <el-row>
          <el-col :span="12">
            <el-form-item label="日志序号:">{{ jobLog.id }}</el-form-item>
            <el-form-item label="任务名称:">{{ jobLog.jobName }}</el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="任务分组:">{{ jobLog.jobGroup }}</el-form-item>
            <el-form-item label="执行时间:">{{ formatDateTime(jobLog.startTime) }}</el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="调用方法:">{{ jobLog.invokeTarget }}</el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="日志信息:">{{ jobLog.jobMessage }}</el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="执行状态:">
              {{ jobLog.status === 1 ? '成功' : '失败' }}
            </el-form-item>
          </el-col>
          <el-col v-if="jobLog.status === 0" :span="24">
            <div>
              <pre><code class="language-java">{{ jobLog.exceptionInfo }}</code></pre>
            </div>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="open = false">关闭</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification } from 'element-plus'
import { Search, Refresh, Delete, View } from '@element-plus/icons-vue'
import { usePageStateStore } from '@/stores/pageState'
import request from '@/utils/request'

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
    console.error('获取任务组列表失败:', error)
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
    }
  } catch (error) {
    console.error('获取日志列表失败:', error)
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
  listJobLogs()
}

const selectionChange = (selection) => {
  jobLogIds.value = selection.map(item => item.id)
}

const deleteJobLogs = async () => {
  try {
    const { data } = await request.delete('/admin/jobLogs', { data: jobLogIds.value })
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '删除成功'
      })
      listJobLogs()
    } else {
      ElNotification.error({
        title: '失败',
        message: '删除失败'
      })
    }
  } catch (error) {
    console.error('删除日志失败:', error)
  }
}

const clean = async () => {
  try {
    const { data } = await request.delete('/admin/jobLogs/clean')
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '清空成功'
      })
      listJobLogs()
    } else {
      ElNotification.error({
        title: '失败',
        message: '清空失败'
      })
    }
  } catch (error) {
    console.error('清空日志失败:', error)
  }
}

const changeOpen = (log) => {
  jobLog.value = { ...log, exceptionInfo: '\n' + log.exceptionInfo }
  open.value = true
}

const currentChange = (page) => {
  current.value = page
  pageStateStore.updateQuartzLogState(jobId.value, page)
  listJobLogs()
}

onMounted(() => {
  const quartzId = route.params.quartzId
  
  if (quartzId === 'all') {
    jobId.value = 0
  } else if (quartzId) {
    jobId.value = quartzId
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
.mb8 {
  margin-bottom: 8px;
}

pre {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
}

code {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
}
</style>

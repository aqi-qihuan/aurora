<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '操作日志' }}</div>
    <div class="operation-container">
      <el-button
        type="danger"
        size="small"
        :disabled="logIds.length === 0"
        @click="isDelete = true">
        <el-icon style="margin-right: 4px"><Delete /></el-icon>
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          size="small"
          placeholder="请输入模块名或描述"
          style="width: 200px"
          @keyup.enter="searchLogs">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" size="small" style="margin-left: 1rem" @click="searchLogs">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <el-table @selection-change="selectionChange" v-loading="loading" :data="logs">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="optModule" label="系统模块" align="center" width="120" />
      <el-table-column width="100" prop="optType" label="操作类型" align="center" />
      <el-table-column prop="optDesc" label="操作描述" align="center" width="150" />
      <el-table-column prop="requestMethod" label="请求方式" align="center" width="100">
        <template #default="scope">
          <el-tag v-if="scope.row.requestMethod" :type="tagType(scope.row.requestMethod)">
            {{ scope.row.requestMethod }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="操作人员" align="center" />
      <el-table-column prop="ipAddress" label="登录IP" align="center" width="130" />
      <el-table-column prop="ipSource" label="登录地址" align="center" width="150" />
      <el-table-column prop="createTime" label="操作日期" align="center" width="190">
        <template #default="scope">
          <el-icon style="margin-right: 5px"><Clock /></el-icon>
          {{ formatDateTime(scope.row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template #default="scope">
          <el-button size="small" type="primary" text @click="check(scope.row)">
            <el-icon style="margin-right: 4px"><View /></el-icon> 查看
          </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 10px" @confirm="deleteLog(scope.row.id)">
            <template #reference>
              <el-button size="small" type="danger" text> 
                <el-icon style="margin-right: 4px"><Delete /></el-icon> 删除 
              </el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      class="pagination-container"
      background
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      :page-sizes="[10, 20]"
      layout="total, sizes, prev, pager, next, jumper" />
    
    <!-- 详情对话框 -->
    <el-dialog v-model="isCheck" width="40%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="margin-right: 8px"><MoreFilled /></el-icon>详细信息
        </div>
      </template>
      <el-form :model="optLog" label-width="100px" size="small">
        <el-form-item label="操作模块：">
          {{ optLog.optModule }}
        </el-form-item>
        <el-form-item label="请求地址：">
          {{ optLog.optUri }}
        </el-form-item>
        <el-form-item label="请求方式：">
          <el-tag :type="tagType(optLog.requestMethod)">
            {{ optLog.requestMethod }}
          </el-tag>
        </el-form-item>
        <el-form-item label="操作方法：">
          {{ optLog.optMethod }}
        </el-form-item>
        <el-form-item label="请求参数：">
          {{ optLog.requestParam }}
        </el-form-item>
        <el-form-item label="返回数据：">
          {{ optLog.responseData }}
        </el-form-item>
        <el-form-item label="操作人员：">
          {{ optLog.nickname }}
        </el-form-item>
      </el-form>
    </el-dialog>
    
    <!-- 删除确认 -->
    <el-dialog v-model="isDelete" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; margin-right: 8px"><WarningFilled /></el-icon>提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <template #footer>
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteLog(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification } from 'element-plus'
import { Delete, Search, Clock, View, MoreFilled, WarningFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'

const route = useRoute()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const logs = ref([])
const logIds = ref([])
const keywords = ref(null)
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
  requestParam: '',
  responseData: '',
  nickname: ''
})

const formatDateTime = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

const tagType = (type) => {
  switch (type) {
    case 'GET': return ''
    case 'POST': return 'success'
    case 'PUT': return 'warning'
    case 'DELETE': return 'danger'
    default: return ''
  }
}

onMounted(() => {
  listLogs()
})

const selectionChange = (selection) => {
  logIds.value = selection.map(item => item.id)
}

const searchLogs = () => {
  current.value = 1
  pageStateStore.updateOperationLogPageState(current.value)
  listLogs()
}

const sizeChange = (val) => {
  size.value = val
  listLogs()
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updateOperationLogPageState(val)
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

const deleteLog = (id) => {
  const param = id ? { data: [id] } : { data: logIds.value }
  request.delete('/admin/operation/logs', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listLogs()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isDelete.value = false
  })
}

const check = (log) => {
  Object.assign(optLog, log)
  isCheck.value = true
}
</script>

<style scoped>
.title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 1rem;
}
.operation-container {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}
</style>

<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '友情链接' }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" @click="openModel(null)">
        <el-icon><Plus /></el-icon>
        新增
      </el-button>
      <el-button
        type="danger"
        size="small"
        :disabled="linkIdList.length === 0"
        @click="deleteFlag = true">
        <el-icon><Delete /></el-icon>
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          size="small"
          placeholder="请输入友链名"
          style="width: 200px"
          @keyup.enter="searchLinks">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" size="small" style="margin-left: 1rem" @click="searchLinks">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <el-table border :data="linkList" @selection-change="selectionChange" v-loading="loading">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="linkAvatar" label="链接头像" align="center" width="180">
        <template #default="scope">
          <el-avatar :src="scope.row.linkAvatar" :size="40" />
        </template>
      </el-table-column>
      <el-table-column prop="linkName" label="链接名" align="center" />
      <el-table-column prop="linkAddress" label="链接地址" align="center" />
      <el-table-column prop="linkIntro" label="链接介绍" align="center" />
      <el-table-column prop="createTime" label="创建时间" width="140" align="center">
        <template #default="scope">
          <el-icon style="margin-right: 5px"><Clock /></el-icon>
          {{ formatDate(scope.row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="160">
        <template #default="scope">
          <el-button type="primary" size="small" @click="openModel(scope.row)"> 编辑 </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 1rem" @confirm="deleteLink(scope.row.id)">
            <template #reference>
              <el-button size="small" type="danger"> 删除 </el-button>
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
    
    <!-- 删除确认 -->
    <el-dialog v-model="deleteFlag" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; margin-right: 8px"><WarningFilled /></el-icon>提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <template #footer>
        <el-button @click="deleteFlag = false">取 消</el-button>
        <el-button type="primary" @click="deleteLink(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
    
    <!-- 添加/编辑对话框 -->
    <el-dialog v-model="addOrEdit" width="30%">
      <template #header>
        <div class="dialog-title-container" ref="linkTitleRef">{{ linkForm.id ? '修改友链' : '添加友链' }}</div>
      </template>
      <el-form label-width="80px" :model="linkForm">
        <el-form-item label="链接名">
          <el-input style="width: 250px" v-model="linkForm.linkName" />
        </el-form-item>
        <el-form-item label="链接头像">
          <el-input style="width: 250px" v-model="linkForm.linkAvatar" />
        </el-form-item>
        <el-form-item label="链接地址">
          <el-input style="width: 250px" v-model="linkForm.linkAddress" />
        </el-form-item>
        <el-form-item label="链接介绍">
          <el-input style="width: 250px" v-model="linkForm.linkIntro" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditLink"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Plus, Delete, Search, Clock, WarningFilled } from '@element-plus/icons-vue'
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
const linkTitleRef = ref(null)

const linkForm = reactive({
  id: null,
  linkName: '',
  linkAvatar: '',
  linkIntro: '',
  linkAddress: ''
})

const formatDate = (date) => {
  return dayjs(date).format('YYYY-MM-DD')
}

onMounted(() => {
  listLinks()
})

const selectionChange = (selection) => {
  linkIdList.value = selection.map(item => item.id)
}

const searchLinks = () => {
  current.value = 1
  listLinks()
}

const sizeChange = (val) => {
  size.value = val
  listLinks()
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updateFriendLinkPageState(val)
  listLinks()
}

const listLinks = () => {
  request.get('/admin/links', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      linkList.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

const deleteLink = (id) => {
  const param = id ? { data: [id] } : { data: linkIdList.value }
  request.delete('/admin/links', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listLinks()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
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

const addOrEditLink = () => {
  if (!linkForm.linkName?.trim()) {
    ElMessage.error('友链名不能为空')
    return false
  }
  if (!linkForm.linkAvatar?.trim()) {
    ElMessage.error('友链头像不能为空')
    return false
  }
  if (!linkForm.linkIntro?.trim()) {
    ElMessage.error('友链介绍不能为空')
    return false
  }
  if (!linkForm.linkAddress?.trim()) {
    ElMessage.error('友链地址不能为空')
    return false
  }
  
  request.post('/admin/links', linkForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listLinks()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    addOrEdit.value = false
  })
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

<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '在线用户' }}</div>
    <div class="operation-container">
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          size="small"
          placeholder="请输入用户昵称"
          style="width: 200px"
          @keyup.enter="listOnlineUsers">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" size="small" style="margin-left: 1rem" @click="listOnlineUsers">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <el-table v-loading="loading" :data="users">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="avatar" label="头像" align="center" width="100">
        <template #default="scope">
          <el-avatar :src="scope.row.avatar" :size="40" />
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="昵称" align="center" />
      <el-table-column prop="ipAddress" label="IP地址" align="center" />
      <el-table-column prop="ipSource" label="登录地址" align="center" width="200" />
      <el-table-column prop="browser" label="浏览器" align="center" width="160" />
      <el-table-column prop="os" label="操作系统" align="center" />
      <el-table-column prop="lastLoginTime" label="登录时间" align="center" width="200">
        <template #default="scope">
          <el-icon style="margin-right: 5px"><Clock /></el-icon>
          {{ formatDateTime(scope.row.lastLoginTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template #default="scope">
          <el-popconfirm title="确定下线吗？" style="margin-left: 10px" @confirm="removeOnlineUser(scope.row)">
            <template #reference>
              <el-button size="small" type="danger" text>
                <el-icon><Delete /></el-icon> 下线
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
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElNotification } from 'element-plus'
import { Search, Clock, Delete } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()
const userStore = useUserStore()

const loading = ref(true)
const users = ref([])
const keywords = ref(null)
const current = ref(pageStateStore.online || 1)
const size = ref(10)
const count = ref(0)

const formatDateTime = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(() => {
  listOnlineUsers()
})

const listOnlineUsers = () => {
  request.get('/admin/users/online', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      users.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

const sizeChange = (val) => {
  size.value = val
  listOnlineUsers()
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updateOnlinePageState(val)
  listOnlineUsers()
}

const removeOnlineUser = (user) => {
  request.delete('/admin/users/' + user.userInfoId + '/online').then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      // 如果是自己被下线，跳转到登录页
      if (user.userInfoId === userStore.userInfo?.id) {
        router.push({ path: '/login' })
        sessionStorage.removeItem('token')
      }
      listOnlineUsers()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
  })
}
</script>

<style scoped>
.operation-container {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-bottom: 1rem;
}
</style>

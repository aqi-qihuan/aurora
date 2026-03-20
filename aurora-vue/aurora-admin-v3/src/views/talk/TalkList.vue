<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '说说管理' }}</div>
    <div class="status-menu">
      <span>状态</span>
      <span @click="changeStatus(null)" :class="isActive(null)">全部</span>
      <span @click="changeStatus(1)" :class="isActive(1)"> 公开 </span>
      <span @click="changeStatus(2)" :class="isActive(2)"> 私密 </span>
    </div>
    <el-empty v-if="talks.length === 0" description="暂无说说" />
    <div class="talk-item" v-for="item of talks" :key="item.id">
      <div class="user-info-wrapper">
        <el-avatar class="user-avatar" :src="item.avatar" :size="36" />
        <div class="user-detail-wrapper">
          <div class="user-nickname">
            <div>{{ item.nickname }}</div>
            <el-dropdown trigger="click" @command="handleCommand">
              <el-icon style="color: #333; cursor: pointer"><MoreFilled /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="'1,' + item.id">
                    <el-icon><Edit /></el-icon>编辑
                  </el-dropdown-item>
                  <el-dropdown-item :command="'2,' + item.id">
                    <el-icon><Delete /></el-icon>删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div class="time">
            {{ formatDateTime(item.createTime) }}
            <span class="top" v-if="item.isTop == 1"> 📌 置顶 </span>
            <span class="secret" v-if="item.status == 2"> 🔒 私密 </span>
          </div>
          <div class="talk-content" v-html="item.content" />
          <el-row :gutter="4" class="talk-images" v-if="item.imgs && item.imgs.length > 0">
            <el-col :md="8" :sm="12" :xs="12" v-for="(img, index) of item.imgs" :key="index">
              <el-image class="images-items" :src="img" :preview-src-list="previews" fit="cover" />
            </el-col>
          </el-row>
        </div>
      </div>
    </div>
    <el-pagination
      v-if="count > 0"
      :hide-on-single-page="false"
      class="pagination-container"
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      layout="prev, pager, next" />
    <el-dialog v-model="isdelete" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; margin-right: 8px"><WarningFilled /></el-icon>提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除该说说？</div>
      <template #footer>
        <el-button @click="isdelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteTalk"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElNotification } from 'element-plus'
import { MoreFilled, Edit, Delete, WarningFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'
import DOMPurify from 'dompurify'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()

// XSS 防护 - HTML 消毒
const sanitizeHtml = (html) => {
  if (!html) return ''
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'br', 'p', 'span', 'img'],
    ALLOWED_ATTR: ['href', 'title', 'target', 'class', 'src', 'alt']
  })
}

const current = ref(pageStateStore.talkList || 1)
const size = ref(5)
const count = ref(0)
const status = ref(null)
const isdelete = ref(false)
const talks = ref([])
const previews = ref([])
const talkId = ref(null)

const formatDateTime = (date) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

onMounted(() => {
  listTalks()
})

const handleCommand = (command) => {
  const arr = command.split(',')
  talkId.value = arr[1]
  switch (arr[0]) {
    case '1':
      router.push({ path: '/talks/' + talkId.value })
      break
    case '2':
      isdelete.value = true
      break
  }
}

const listTalks = () => {
  request.get('/admin/talks', {
    params: {
      current: current.value,
      size: size.value,
      status: status.value
    }
  }).then(({ data }) => {
    if (data && data.data) {
      talks.value = data.data.records || []
      count.value = data.data.count || 0
      // 收集所有图片用于预览
      const allImgs = []
      talks.value.forEach(item => {
        if (item.imgs) {
          allImgs.push(...item.imgs)
        }
      })
      previews.value = allImgs
    }
  })
}

const sizeChange = (val) => {
  previews.value = []
  size.value = val
  listTalks()
}

const currentChange = (val) => {
  previews.value = []
  current.value = val
  pageStateStore.updatePageState('talkList', val)
  listTalks()
}

const changeStatus = (newStatus) => {
  current.value = 1
  previews.value = []
  status.value = newStatus
  listTalks()
}

const deleteTalk = () => {
  request.delete('/admin/talks', { data: [talkId.value] }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listTalks()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isdelete.value = false
  })
}

const isActive = (s) => {
  return status.value === s ? 'active-status' : 'status'
}
</script>

<style scoped>
.status-menu {
  font-size: 14px;
  margin-top: 40px;
  color: #999;
}
.status-menu span {
  margin-right: 24px;
}
.status {
  cursor: pointer;
}
.active-status {
  cursor: pointer;
  color: #333;
  font-weight: bold;
}
.talk-item:not(:first-child) {
  margin-top: 20px;
}
.talk-item {
  padding: 16px 20px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  box-shadow: 0 3px 8px 6px rgb(7 17 27 / 6%);
  transition: all 0.3s ease 0s;
}
.talk-item:hover {
  box-shadow: 0 5px 10px 8px rgb(7 17 27 / 16%);
  transform: translateY(-3px);
}
.user-info-wrapper {
  width: 100%;
  display: flex;
}
.user-avatar {
  border-radius: 50%;
  transition: all 0.5s;
}
.user-avatar:hover {
  transform: rotate(360deg);
}
.user-detail-wrapper {
  margin-left: 10px;
  width: 100%;
}
.user-nickname {
  font-size: 15px;
  font-weight: bold;
  display: flex;
  justify-content: space-between;
}
.time {
  color: #999;
  margin-top: 2px;
  font-size: 12px;
}
.top {
  color: #ff7242;
  margin-left: 10px;
}
.secret {
  color: #999;
  margin-left: 10px;
}
.talk-content {
  margin-top: 8px;
  font-size: 14px;
  line-height: 26px;
  white-space: pre-line;
  word-wrap: break-word;
  word-break: break-all;
}
.talk-images {
  margin-top: 8px;
}
.images-items {
  cursor: pointer;
  object-fit: cover;
  height: 200px;
  width: 100%;
  border-radius: 4px;
}
</style>

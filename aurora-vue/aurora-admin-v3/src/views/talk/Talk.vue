<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '编辑说说' }}</div>
    <div class="talk-container">
      <Editor ref="editorRef" class="editor-wrapper" v-model="talk.content" placeholder="说点什么吧" />
      <div class="operation-wrapper">
        <div class="left-wrapper">
          <el-upload
            action="/api/admin/talks/images"
            multiple
            :headers="headers"
            :before-upload="beforeUpload"
            :on-success="upload"
            :show-file-list="false">
            <el-icon class="operation-btn tupian"><Picture /></el-icon>
          </el-upload>
        </div>
        <div class="right-wrapper">
          <el-switch
            style="margin-right: 16px"
            v-model="talk.isTop"
            inactive-text="置顶"
            :active-value="1"
            :inactive-value="0" />
          <el-dropdown trigger="click" @command="handleCommand" style="margin-right: 16px">
            <span class="talk-status"> {{ dropdownTitle }}<el-icon class="el-icon--right"><ArrowDown /></el-icon> </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="(item, index) of statuses" :key="index" :command="item.status">
                  {{ item.desc }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button type="primary" size="small" @click="saveOrUpdateTalk" :disabled="!talk.content">
            发布
          </el-button>
        </div>
      </div>
      <el-upload
        class="talk-image-upload"
        v-show="uploads.length > 0"
        action="/api/admin/talks/images"
        list-type="picture-card"
        multiple
        :headers="headers"
        :file-list="uploads"
        :before-upload="beforeUpload"
        :on-success="upload"
        :on-remove="handleRemove">
        <el-icon><Plus /></el-icon>
      </el-upload>
    </div>
  </el-card>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Picture, ArrowDown, Plus } from '@element-plus/icons-vue'
import * as imageConversion from 'image-conversion'
import request from '@/utils/request'
import Editor from '@/components/Editor.vue'

const UPLOAD_SIZE = 1024

const route = useRoute()
const router = useRouter()

const editorRef = ref(null)

const talk = reactive({
  id: null,
  content: '',
  isTop: 0,
  status: 1,
  images: ''
})

const statuses = [
  { status: 1, desc: '公开' },
  { status: 2, desc: '私密' }
]

const uploads = ref([])

const headers = computed(() => ({ 
  Authorization: 'Bearer ' + sessionStorage.getItem('token') 
}))

const dropdownTitle = computed(() => {
  const found = statuses.find(item => item.status === talk.status)
  return found ? found.desc : '公开'
})

onMounted(() => {
  if (route.params.talkId) {
    request.get('/admin/talks/' + route.params.talkId).then(({ data }) => {
      if (data && data.data) {
        Object.assign(talk, data.data)
        if (data.data.imgs) {
          data.data.imgs.forEach((item) => {
            uploads.value.push({ url: item })
          })
        }
      }
    })
  }
})

const handleCommand = (command) => {
  talk.status = command
}

const handleRemove = (file) => {
  uploads.value = uploads.value.filter(item => item.url !== file.url)
}

const upload = (response) => {
  if (response && response.data) {
    uploads.value.push({ url: response.data })
  }
}

const beforeUpload = (file) => {
  return new Promise((resolve) => {
    if (file.size / 1024 < UPLOAD_SIZE) {
      resolve(file)
    } else {
      imageConversion.compressAccurately(file, UPLOAD_SIZE).then((res) => {
        resolve(res)
      })
    }
  })
}

const saveOrUpdateTalk = () => {
  if (!talk.content?.trim()) {
    ElMessage.error('说说内容不能为空')
    return false
  }
  
  if (uploads.value.length > 0) {
    const img = uploads.value.map(item => item.url)
    talk.images = JSON.stringify(img)
  } else {
    talk.images = ''
  }
  
  request.post('/admin/talks', talk).then(({ data }) => {
    if (data.flag) {
      if (editorRef.value?.clear) {
        editorRef.value.clear()
      }
      uploads.value = []
      router.push({ path: '/talk-list' })
      ElNotification.success({
        title: '成功',
        message: data.message
      })
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
.tupian {
  margin-left: 3px;
}
.talk-container {
  margin-top: 40px;
}
.editor-wrapper {
  min-height: 150px;
}
.operation-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}
.operation-btn {
  cursor: pointer;
  color: #838383;
  font-size: 20px;
  margin-right: 12px;
}
.talk-status {
  cursor: pointer;
  font-size: 12px;
  color: #999;
}
.left-wrapper {
  display: flex;
  width: 50%;
}
.talk-image-upload {
  margin-top: 8px;
}
</style>

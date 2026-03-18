<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '关于我' }}</div>
    <MdEditor 
      ref="mdEditorRef"
      v-model="aboutContent" 
      @onUploadImg="handleUploadImg" 
      style="height: calc(100vh - 250px); margin-top: 2.25rem" 
    />
    <el-button type="danger" class="edit-btn" @click="updateAbout"> 修改 </el-button>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification } from 'element-plus'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import * as imageConversion from 'image-conversion'
import request from '@/utils/request'

const UPLOAD_SIZE = 1024

const route = useRoute()
const aboutContent = ref('')
const mdEditorRef = ref(null)

onMounted(() => {
  getAbout()
})

const getAbout = () => {
  request.get('/about').then(({ data }) => {
    if (data && data.data) {
      aboutContent.value = data.data.content || ''
    }
  })
}

const handleUploadImg = async (files, callback) => {
  for (const file of files) {
    const formData = new FormData()
    try {
      let uploadFile = file
      if (file.size / 1024 >= UPLOAD_SIZE) {
        uploadFile = await imageConversion.compressAccurately(file, UPLOAD_SIZE)
      }
      formData.append('file', uploadFile)
      const { data } = await request.post('/admin/articles/images', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      if (data && data.data) {
        callback(data.data)
      }
    } catch (error) {
      console.error('上传图片失败:', error)
    }
  }
}

const updateAbout = () => {
  request.put('/admin/about', {
    content: aboutContent.value
  }).then(({ data }) => {
    if (data.flag) {
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
.edit-btn {
  float: right;
  margin: 1rem 0;
}
</style>

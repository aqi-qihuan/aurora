<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || '相册管理' }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" @click="openModel(null)">
        <el-icon><Plus /></el-icon>
        新建相册
      </el-button>
      <div class="search-container">
        <el-button type="primary" link size="small" style="margin-right: 1rem" @click="checkDelete">
          <el-icon><Delete /></el-icon>
          回收站
        </el-button>
        <el-input
          v-model="keywords"
          size="small"
          placeholder="请输入相册名"
          style="width: 200px"
          @keyup.enter="searchAlbums">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" size="small" style="margin-left: 1rem" @click="searchAlbums">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <el-row class="album-container" :gutter="12" v-loading="loading">
      <el-empty v-if="albums == null || albums.length === 0" description="暂无相册" />
      <el-col v-for="item of albums" :key="item.id" :md="6">
        <div class="album-item" @click="checkPhoto(item)">
          <div class="album-opreation">
            <el-dropdown @command="handleCommand">
              <el-icon style="color: #fff; font-size: 18px"><MoreFilled /></el-icon>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="'update' + JSON.stringify(item)">
                    <el-icon><Edit /></el-icon>编辑
                  </el-dropdown-item>
                  <el-dropdown-item :command="'delete' + item.id">
                    <el-icon><Delete /></el-icon>删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div class="album-photo-count">
            <div>{{ item.photoCount }}</div>
            <el-icon v-if="item.status == 2"><Lock /></el-icon>
          </div>
          <el-image fit="cover" class="album-cover" :src="item.albumCover" />
          <div class="album-name">{{ item.albumName }}</div>
        </div>
      </el-col>
    </el-row>
    <el-pagination
      :hide-on-single-page="true"
      class="pagination-container"
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      layout="prev, pager, next" />
    <el-dialog v-model="addOrEdit" width="35%" top="10vh">
      <template #header>
        <div class="dialog-title-container" ref="albumTitleRef" />
      </template>
      <el-form label-width="80px" size="medium" :model="albumForum">
        <el-form-item label="相册名称">
          <el-input style="width: 220px" v-model="albumForum.albumName" />
        </el-form-item>
        <el-form-item label="相册描述">
          <el-input style="width: 220px" v-model="albumForum.albumDesc" />
        </el-form-item>
        <el-form-item label="相册封面">
          <el-upload
            class="upload-cover"
            drag
            :headers="headers"
            :before-upload="beforeUpload"
            action="/api/admin/photos/albums/upload"
            multiple
            :on-success="uploadCover">
            <el-icon class="el-icon--upload" v-if="albumForum.albumCover == ''"><UploadFilled /></el-icon>
            <div class="el-upload__text" v-if="albumForum.albumCover == ''">
              将文件拖到此处,或<em>点击上传</em>
            </div>
            <img v-else :src="albumForum.albumCover" width="360px" height="180px" />
          </el-upload>
        </el-form-item>
        <el-form-item label="发布形式">
          <el-radio-group v-model="albumForum.status">
            <el-radio :value="1">公开</el-radio>
            <el-radio :value="2">私密</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditAlbum">确 定</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="isdelete" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除该相册?</div>
      <template #footer>
        <el-button @click="isdelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteAlbum">确 定</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Plus,
  Delete,
  Search,
  MoreFilled,
  Edit,
  Lock,
  UploadFilled,
  Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()

// 响应式数据
const keywords = ref('')
const loading = ref(true)
const isdelete = ref(false)
const addOrEdit = ref(false)
const albumForum = ref({
  id: null,
  albumName: '',
  albumDesc: '',
  albumCover: '',
  status: 1
})
const albums = ref([])
const current = ref(1)
const size = ref(8)
const count = ref(0)
const headers = ref({ Authorization: 'Bearer ' + sessionStorage.getItem('token') })
const albumTitleRef = ref(null)

// 打开对话框
const openModel = (item) => {
  if (item) {
    albumForum.value = JSON.parse(item)
    nextTick(() => {
      if (albumTitleRef.value) {
        albumTitleRef.value.innerHTML = '修改相册'
      }
    })
  } else {
    albumForum.value = {
      id: null,
      albumName: '',
      albumDesc: '',
      albumCover: '',
      status: 1
    }
    nextTick(() => {
      if (albumTitleRef.value) {
        albumTitleRef.value.innerHTML = '新建相册'
      }
    })
  }
  addOrEdit.value = true
}

// 查看照片
const checkPhoto = (item) => {
  router.push({ path: '/albums/' + item.id })
}

// 查看回收站
const checkDelete = () => {
  router.push({ path: '/photos/delete' })
}

// 获取相册列表
const listAlbums = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin/photos/albums', {
      params: {
        current: current.value,
        size: size.value,
        keywords: keywords.value
      }
    })
    if (data && data.data) {
      albums.value = data.data.records || []
      count.value = data.data.count || 0
    }
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '获取相册列表失败'
    })
  } finally {
    loading.value = false
  }
}

// 添加或修改相册
const addOrEditAlbum = async () => {
  if (albumForum.value.albumName.trim() === '') {
    ElMessage.error('相册名称不能为空')
    return false
  }
  if (albumForum.value.albumDesc.trim() === '') {
    ElMessage.error('相册描述不能为空')
    return false
  }
  if (albumForum.value.albumCover == null || albumForum.value.albumCover === '') {
    ElMessage.error('相册封面不能为空')
    return false
  }
  
  try {
    const { data } = await request.post('/admin/photos/albums', albumForum.value)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listAlbums()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    addOrEdit.value = false
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '操作失败'
    })
  }
}

// 上传封面
const uploadCover = (response) => {
  albumForum.value.albumCover = response.data
}

// 上传前处理
const beforeUpload = (file) => {
  return new Promise((resolve) => {
    const isLt500K = file.size / 1024 < 500
    if (isLt500K) {
      resolve(file)
    } else {
      // 可以添加图片压缩逻辑
      resolve(file)
    }
  })
}

// 下拉菜单命令
const handleCommand = (command) => {
  const type = command.substring(0, 6)
  const data = command.substring(6)
  if (type === 'delete') {
    albumForum.value.id = data
    isdelete.value = true
  } else {
    openModel(data)
  }
}

// 删除相册
const deleteAlbum = async () => {
  try {
    const { data } = await request.delete('/admin/photos/albums/' + albumForum.value.id)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listAlbums()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isdelete.value = false
  } catch (error) {
    ElMessage.error('删除相册失败')
    console.error('API Error:', error)
  }
}

// 搜索相册
const searchAlbums = () => {
  current.value = 1
  listAlbums()
}

// 每页数量变化
const sizeChange = (val) => {
  size.value = val
  listAlbums()
}

// 当前页变化
const currentChange = (val) => {
  current.value = val
  listAlbums()
}

// 初始化
onMounted(() => {
  listAlbums()
})
</script>

<style scoped>
/* ==================== Album Page Modern Styles ==================== */

/* 页面标题 */
.title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text);
  margin-bottom: var(--space-6);
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.title::before {
  content: '';
  width: 4px;
  height: 24px;
  background: linear-gradient(180deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border-radius: var(--radius-full);
}

/* 操作区域 */
.operation-container {
  margin-top: var(--space-6);
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--color-bg-hover);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  margin-bottom: var(--space-6);
}

.operation-container .el-button {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-button:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.operation-container .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

/* 搜索区域 */
.search-container {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-left: auto;
}

.search-container .el-input {
  width: 200px;
}

.search-container .el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.search-container .el-input :deep(.el-input__inner:focus) {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 相册容器 */
.album-container {
  margin-top: var(--space-6);
}

/* 相册项 */
.album-item {
  position: relative;
  cursor: pointer;
  margin-bottom: var(--space-4);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  transition: all var(--duration-base) var(--ease-out);
  background: var(--color-bg-card);
}

.album-item:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover);
}

/* 相册封面 */
.album-cover {
  position: relative;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  width: 100%;
  height: 170px;
  overflow: hidden;
}

.album-cover::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.3) 0%, transparent 50%, rgba(0, 0, 0, 0.5) 100%);
  z-index: 1;
  transition: all var(--duration-base) var(--ease-out);
}

.album-item:hover .album-cover::before {
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.4) 0%, transparent 50%, rgba(0, 0, 0, 0.6) 100%);
}

.album-cover :deep(img) {
  transition: all var(--duration-slow) var(--ease-out);
}

.album-item:hover .album-cover :deep(img) {
  transform: scale(1.05);
}

/* 照片数量 */
.album-photo-count {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--text-xl);
  font-weight: var(--font-bold);
  z-index: 10;
  position: absolute;
  left: var(--space-3);
  right: var(--space-3);
  bottom: var(--space-12);
  color: #fff;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.album-photo-count .el-icon {
  font-size: var(--text-lg);
  opacity: 0.9;
}

/* 相册名称 */
.album-name {
  text-align: center;
  padding: var(--space-3);
  font-weight: var(--font-semibold);
  color: var(--color-text);
  font-size: var(--text-sm);
  background: var(--color-bg-card);
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 操作下拉菜单 */
.album-opreation {
  position: absolute;
  z-index: 10;
  top: var(--space-2);
  right: var(--space-2);
}

.album-opreation :deep(.el-dropdown) {
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(4px);
  border-radius: var(--radius-base);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
  cursor: pointer;
}

.album-opreation :deep(.el-dropdown:hover) {
  background: rgba(255, 255, 255, 0.4);
}

/* 空状态 */
.el-empty {
  padding: var(--space-12) 0;
}

/* 分页 */
.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: var(--space-8);
}

.pagination-container :deep(.el-pagination) {
  font-weight: var(--font-medium);
}

.pagination-container :deep(.el-pagination .el-pager li) {
  border-radius: var(--radius-base);
  transition: all var(--duration-fast) var(--ease-out);
}

.pagination-container :deep(.el-pagination .el-pager li.active) {
  background: var(--color-primary);
}

.pagination-container :deep(.el-pagination .el-pager li:hover) {
  transform: translateY(-1px);
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  font-weight: var(--font-bold);
  font-size: var(--text-lg);
  color: var(--color-text);
}

/* 表单样式 */
:deep(.el-form-item__label) {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

:deep(.el-input .el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

:deep(.el-input .el-input__inner:focus) {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 上传组件 */
.upload-cover :deep(.el-upload) {
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-hover);
  transition: all var(--duration-fast) var(--ease-out);
}

.upload-cover :deep(.el-upload:hover) {
  border-color: var(--color-primary);
  background: var(--color-primary-50);
}

.upload-cover :deep(.el-upload-dragger) {
  width: 360px;
  height: 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
}

.upload-cover :deep(.el-icon--upload) {
  font-size: var(--text-3xl);
  color: var(--color-text-muted);
  margin: 0 0 var(--space-2);
}

.upload-cover :deep(.el-upload__text) {
  color: var(--color-text-secondary);
}

.upload-cover :deep(.el-upload__text em) {
  color: var(--color-primary);
  font-style: normal;
}

/* 单选按钮组 */
.el-radio-group {
  display: flex;
  gap: var(--space-6);
}

.el-radio {
  margin-right: 0;
}

/* 加载动画 */
.album-container :deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .album-name {
  background: var(--color-bg-card);
}

[data-theme="dark"] .album-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.9);
}

[data-theme="dark"] .upload-cover :deep(.el-upload) {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

/* ==================== Responsive ==================== */
@media (max-width: 768px) {
  .title {
    font-size: var(--text-xl);
  }

  .operation-container {
    flex-direction: column;
    align-items: stretch;
  }

  .search-container {
    margin-left: 0;
    width: 100%;
    flex-direction: column;
  }

  .search-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .album-cover {
    height: 140px;
  }

  .upload-cover :deep(.el-upload-dragger) {
    width: 100%;
    height: 150px;
  }
}

@media (max-width: 480px) {
  .album-cover {
    height: 120px;
  }

  .album-photo-count {
    font-size: var(--text-lg);
    bottom: var(--space-10);
  }

  .album-name {
    font-size: var(--text-xs);
    padding: var(--space-2);
  }

  :deep(.el-dialog) {
    width: 90% !important;
  }
}
</style>

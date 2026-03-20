<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '照片管理' }}</div>
    <div class="album-info">
      <el-image fit="cover" class="album-cover" :src="albumInfo.albumCover" />
      <div class="album-detail">
        <div style="margin-bottom: 0.6rem">
          <span class="album-name">{{ albumInfo.albumName }}</span>
          <span class="photo-count">{{ albumInfo.photoCount }}张</span>
        </div>
        <div>
          <span v-if="albumInfo.albumDesc" class="album-desc">
            {{ albumInfo.albumDesc }}
          </span>
          <el-button type="primary" size="small" @click="uploadPhoto = true">
            <el-icon style="margin-right: 4px"><Picture /></el-icon>
            上传照片
          </el-button>
        </div>
      </div>
      <div class="operation">
        <div class="operation-top">
          <div class="all-check">
            <el-checkbox :indeterminate="isIndeterminate" v-model="checkAll" @change="handleCheckAllChange">
              全选
            </el-checkbox>
            <div class="check-count">已选择{{ selectphotoIds.length }}张</div>
          </div>
        </div>
        <div class="operation-buttons">
          <el-button
            type="success"
            @click="movePhoto = true"
            :disabled="selectphotoIds.length === 0"
            size="small">
            <el-icon style="margin-right: 4px"><FolderOpened /></el-icon>
            移动到
          </el-button>
          <el-button
            type="danger"
            @click="batchDeletePhoto = true"
            :disabled="selectphotoIds.length === 0"
            size="small">
            <el-icon style="margin-right: 4px"><Delete /></el-icon>
            批量删除
          </el-button>
        </div>
      </div>
    </div>
    <el-row class="photo-container" :gutter="10" v-loading="loading">
      <el-empty v-if="photos.length === 0" description="暂无照片" />
      <el-checkbox-group v-model="selectphotoIds" @change="handleCheckedPhotoChange">
        <el-col :xs="12" :sm="8" :md="6" :lg="4" v-for="item of photos" :key="item.id">
          <el-checkbox :value="item.id">
            <div class="photo-item">
              <div class="photo-opreation">
                <el-dropdown @command="handleCommand">
                  <el-icon style="color: #fff; cursor: pointer"><MoreFilled /></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="JSON.stringify(item)">
                        <el-icon><Edit /></el-icon>编辑
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
              <el-image fit="cover" class="photo-img" :src="item.photoSrc" :preview-src-list="photoSrcList" />
              <div class="photo-name">{{ item.photoName }}</div>
            </div>
          </el-checkbox>
        </el-col>
      </el-checkbox-group>
    </el-row>
    <el-pagination
      v-if="count > 0"
      :hide-on-single-page="true"
      class="pagination-container"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      layout="prev, pager, next" />
    
    <!-- 上传照片对话框 -->
    <el-dialog v-model="uploadPhoto" width="70%" top="10vh">
      <template #header>
        <div class="dialog-title-container">上传照片</div>
      </template>
      <div class="upload-container">
        <el-upload
          v-show="uploads.length > 0"
          action="/api/admin/photos/upload"
          list-type="picture-card"
          :file-list="uploads"
          multiple
          :headers="headers"
          :before-upload="beforeUpload"
          :on-success="upload"
          :on-remove="handleRemove">
          <el-icon><Plus /></el-icon>
        </el-upload>
        <div class="upload">
          <el-upload
            v-show="uploads.length === 0"
            drag
            action="/api/admin/photos/upload"
            multiple
            :headers="headers"
            :before-upload="beforeUpload"
            :on-success="upload"
            :show-file-list="false">
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">支持上传jpg/png文件</div>
            </template>
          </el-upload>
        </div>
      </div>
      <template #footer>
        <div class="upload-footer">
          <div class="upload-count">共上传{{ uploads.length }}张照片</div>
          <div style="margin-left: auto">
            <el-button @click="uploadPhoto = false">取 消</el-button>
            <el-button @click="savePhotos" type="primary" :disabled="uploads.length === 0"> 开始上传 </el-button>
          </div>
        </div>
      </template>
    </el-dialog>
    
    <!-- 编辑照片对话框 -->
    <el-dialog v-model="editPhoto" width="30%">
      <template #header>
        <div class="dialog-title-container">修改信息</div>
      </template>
      <el-form label-width="80px" :model="photoForm">
        <el-form-item label="照片名称">
          <el-input style="width: 220px" v-model="photoForm.photoName" />
        </el-form-item>
        <el-form-item label="照片描述">
          <el-input style="width: 220px" v-model="photoForm.photoDesc" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editPhoto = false">取 消</el-button>
        <el-button type="primary" @click="updatePhoto"> 确 定 </el-button>
      </template>
    </el-dialog>
    
    <!-- 批量删除对话框 -->
    <el-dialog v-model="batchDeletePhoto" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; margin-right: 8px"><WarningFilled /></el-icon>提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中照片？</div>
      <template #footer>
        <el-button @click="batchDeletePhoto = false">取 消</el-button>
        <el-button type="primary" @click="updatePhotoDelete(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
    
    <!-- 移动照片对话框 -->
    <el-dialog v-model="movePhoto" width="30%">
      <template #header>
        <div class="dialog-title-container">移动照片</div>
      </template>
      <el-empty v-if="albumList.length < 2" description="暂无其他相册" />
      <el-form v-else label-width="80px">
        <el-radio-group v-model="targetAlbumId">
          <div class="album-check-item">
            <template v-for="item of albumList" :key="item.id">
              <el-radio v-if="item.id !== albumInfo.id" :value="item.id" style="margin-bottom: 1rem">
                <div class="album-check">
                  <el-image fit="cover" class="album-check-cover" :src="item.albumCover" />
                  <div style="margin-left: 0.5rem">{{ item.albumName }}</div>
                </div>
              </el-radio>
            </template>
          </div>
        </el-radio-group>
      </el-form>
      <template #footer>
        <el-button @click="movePhoto = false">取 消</el-button>
        <el-button :disabled="targetAlbumId === null" type="primary" @click="updatePhotoAlbum"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Picture, FolderOpened, Delete, MoreFilled, Edit, Plus, UploadFilled, WarningFilled } from '@element-plus/icons-vue'
import * as imageConversion from 'image-conversion'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'

const UPLOAD_SIZE = 1024

const route = useRoute()
const pageStateStore = usePageStateStore()

const loading = ref(true)
const checkAll = ref(false)
const isIndeterminate = ref(false)
const uploadPhoto = ref(false)
const editPhoto = ref(false)
const movePhoto = ref(false)
const batchDeletePhoto = ref(false)
const uploads = ref([])
const photos = ref([])
const photoIds = ref([])
const selectphotoIds = ref([])
const albumList = ref([])
const targetAlbumId = ref(null)

const albumInfo = reactive({
  id: null,
  albumName: '',
  albumDesc: '',
  albumCover: '',
  photoCount: 0
})

const photoForm = reactive({
  id: null,
  photoName: '',
  photoDesc: ''
})

let current = ref(1)
const size = ref(18)
const count = ref(0)

const headers = computed(() => ({ 
  Authorization: 'Bearer ' + sessionStorage.getItem('token') 
}))

const photoSrcList = computed(() => photos.value.map(p => p.photoSrc))

onMounted(() => {
  const albumId = route.params.albumId
  if (albumId === pageStateStore.photo?.albumId) {
    current.value = pageStateStore.photo.current || 1
  } else {
    current.value = 1
    pageStateStore.updatePhotoPageState({
      albumId: route.params.albumId,
      current: current.value
    })
  }
  getAlbumInfo()
  listAlbums()
  listPhotos()
})

watch(photos, (newPhotos) => {
  photoIds.value = newPhotos.map(item => item.id)
})

const getAlbumInfo = () => {
  request.get('/admin/photos/albums/' + route.params.albumId + '/info').then(({ data }) => {
    if (data && data.data) {
      Object.assign(albumInfo, data.data)
    }
  })
}

const listAlbums = () => {
  request.get('/admin/photos/albums/info').then(({ data }) => {
    if (data && data.data) {
      albumList.value = data.data
    }
  })
}

const listPhotos = () => {
  request.get('/admin/photos', {
    params: {
      current: current.value,
      size: size.value,
      albumId: route.params.albumId,
      isDelete: 0
    }
  }).then(({ data }) => {
    if (data && data.data) {
      photos.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updatePhotoPageState({
    albumId: route.params.albumId,
    current: current.value
  })
  listPhotos()
}

const savePhotos = () => {
  const photoUrls = uploads.value.map(item => item.url)
  request.post('/admin/photos', {
    albumId: route.params.albumId,
    photoUrls: photoUrls
  }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      uploads.value = []
      listPhotos()
      getAlbumInfo()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    uploadPhoto.value = false
  })
}

const updatePhoto = () => {
  if (!photoForm.photoName?.trim()) {
    ElMessage.error('照片名称不能为空')
    return false
  }
  request.put('/admin/photos', photoForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listPhotos()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    editPhoto.value = false
  })
}

const updatePhotoAlbum = () => {
  request.put('/admin/photos/album', {
    albumId: targetAlbumId.value,
    photoIds: selectphotoIds.value
  }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      getAlbumInfo()
      listPhotos()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    movePhoto.value = false
  })
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

const handleCheckAllChange = (val) => {
  selectphotoIds.value = val ? photoIds.value : []
  isIndeterminate.value = false
}

const handleCheckedPhotoChange = (value) => {
  const checkedCount = value.length
  checkAll.value = checkedCount === photoIds.value.length
  isIndeterminate.value = checkedCount > 0 && checkedCount < photoIds.value.length
}

const handleCommand = (command) => {
  const item = JSON.parse(command)
  Object.assign(photoForm, item)
  editPhoto.value = true
}

const updatePhotoDelete = (id) => {
  const param = id 
    ? { ids: [id], isDelete: 1 }
    : { ids: selectphotoIds.value, isDelete: 1 }
  request.put('/admin/photos/delete', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listPhotos()
      getAlbumInfo()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    batchDeletePhoto.value = false
  })
}
</script>

<style scoped>
.album-info {
  display: flex;
  margin-top: 2.25rem;
  margin-bottom: 2rem;
}
.album-cover {
  border-radius: 4px;
  width: 5rem;
  height: 5rem;
}
.album-check-cover {
  border-radius: 4px;
  width: 4rem;
  height: 4rem;
}
.album-detail {
  padding-top: 0.4rem;
  margin-left: 0.8rem;
}
.album-desc {
  font-size: 14px;
  margin-right: 0.8rem;
}
.operation {
  padding-top: 1.5rem;
  margin-left: auto;
}
.all-check {
  display: inline-flex;
  align-items: center;
  margin-right: 1rem;
}
.check-count {
  margin-left: 1rem;
  font-size: 12px;
}
.album-name {
  font-size: 1.25rem;
}
.photo-count {
  font-size: 12px;
  margin-left: 0.5rem;
}
.photo-item {
  width: 100%;
  position: relative;
  cursor: pointer;
  margin-bottom: 1rem;
}
.photo-img {
  width: 100%;
  height: 7rem;
  border-radius: 4px;
}
.photo-name {
  font-size: 14px;
  margin-top: 0.3rem;
  text-align: center;
}
.upload-container {
  height: 400px;
}
.upload {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.upload-footer {
  display: flex;
  align-items: center;
}
.photo-opreation {
  position: absolute;
  z-index: 1000;
  top: 0.3rem;
  right: 0.5rem;
}
.album-check {
  display: flex;
  align-items: center;
}
</style>

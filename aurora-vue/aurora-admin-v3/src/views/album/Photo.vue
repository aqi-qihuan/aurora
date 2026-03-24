<template>
  <div class="photo-page">
    <!-- 相册信息横幅 -->
    <div class="album-banner">
      <div class="banner-cover">
        <el-image fit="cover" class="banner-img" :src="albumInfo.albumCover" />
        <div class="banner-overlay"></div>
      </div>
      <div class="banner-content">
        <div class="banner-info">
          <h2 class="banner-title">{{ albumInfo.albumName }}</h2>
          <p class="banner-desc" v-if="albumInfo.albumDesc">{{ albumInfo.albumDesc }}</p>
        </div>
        <div class="banner-actions">
          <button class="btn btn-primary" @click="uploadPhoto = true">
            <el-icon><UploadFilled /></el-icon>上传照片
          </button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Picture /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">照片总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon info">
          <el-icon><Checked /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ selectphotoIds.length }}</span>
          <span class="stat-label">已选中</span>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <label class="select-all" :class="{ active: selectphotoIds.length > 0 }">
          <input
            type="checkbox"
            :checked="checkAll"
            :indeterminate="isIndeterminate"
            @change="handleCheckAllChange"
          />
          <span>全选</span>
          <span v-if="selectphotoIds.length" class="select-badge">{{ selectphotoIds.length }}</span>
        </label>
        <div class="toolbar-divider"></div>
        <button class="btn btn-move" :disabled="!selectphotoIds.length" @click="movePhoto = true">
          <el-icon><FolderOpened /></el-icon>移动到
        </button>
        <button class="btn btn-danger-outline" :disabled="!selectphotoIds.length" @click="batchDeletePhoto = true">
          <el-icon><Delete /></el-icon>批量删除
        </button>
      </div>
      <div class="toolbar-right">
        <button class="btn btn-primary" @click="uploadPhoto = true">
          <el-icon><Plus /></el-icon>上传照片
        </button>
      </div>
    </div>

    <!-- 照片网格 -->
    <div class="photo-grid" v-loading="loading">
      <div
        v-for="item in photos"
        :key="item.id"
        class="photo-card"
        :class="{ selected: selectphotoIds.includes(item.id) }"
      >
        <div class="card-checkbox">
          <input
            type="checkbox"
            :checked="selectphotoIds.includes(item.id)"
            @change="toggleSelect(item.id)"
          />
        </div>
        <div class="card-cover" @click="previewPhoto(item)">
          <el-image
            fit="cover"
            class="cover-img"
            lazy
            :src="item.photoSrc"
            :preview-src-list="photoSrcList"
            :initial-index="photos.findIndex(p => p.id === item.id)"
            :hide-on-click-modal="true"
            close-on-press-escape
          />
          <div class="cover-overlay">
            <el-icon class="preview-icon"><ZoomIn /></el-icon>
          </div>
        </div>
        <div class="card-info">
          <span class="photo-name" :title="item.photoName">{{ item.photoName }}</span>
        </div>
        <div class="card-actions" @click.stop>
          <button class="mini-btn edit" @click="handleCommand(JSON.stringify(item))" title="编辑">
            <el-icon><Edit /></el-icon>
          </button>
        </div>
      </div>

      <div v-if="!loading && photos.length === 0" class="empty-state">
        <el-icon class="empty-icon"><Picture /></el-icon>
        <p>暂无照片</p>
        <span class="empty-hint">点击「上传照片」添加你的第一张照片</span>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-bar" v-if="count > size">
      <el-pagination
        v-model:current-page="current"
        :page-size="size"
        :total="count"
        background
        layout="prev, pager, next"
        @current-change="currentChange"
      />
    </div>

    <!-- 上传照片对话框 -->
    <el-dialog v-model="uploadPhoto" width="720px" top="14vh" class="upload-dialog" destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <span class="dialog-icon primary">
            <el-icon><UploadFilled /></el-icon>
          </span>
          <div>
            <h3 class="dialog-title">上传照片</h3>
            <p class="dialog-subtitle">上传照片到「{{ albumInfo.albumName }}」</p>
          </div>
        </div>
      </template>

      <div class="upload-body">
        <div v-if="uploads.length === 0" class="upload-drag">
          <el-upload
            drag
            action="/api/admin/photos/upload"
            multiple
            :headers="headers"
            :before-upload="beforeUpload"
            :on-success="upload"
            :show-file-list="false"
          >
            <div class="drag-content">
              <el-icon class="drag-icon"><UploadFilled /></el-icon>
              <p class="drag-title">拖拽照片到此处上传</p>
              <p class="drag-sub">或 <em>点击选择文件</em></p>
              <div class="drag-hint">
                <el-icon><InfoFilled /></el-icon>
                <span>支持 JPG、PNG 格式，单张不超过 1MB</span>
              </div>
            </div>
          </el-upload>
        </div>
        <div v-else class="upload-preview">
          <div class="preview-bar">
            <span>已选择 <strong>{{ uploads.length }}</strong> 张照片</span>
            <button class="text-btn" @click="uploads = []">
              <el-icon><Delete /></el-icon>清空
            </button>
          </div>
          <el-upload
            action="/api/admin/photos/upload"
            list-type="picture-card"
            :file-list="uploads"
            multiple
            :headers="headers"
            :before-upload="beforeUpload"
            :on-success="upload"
            :on-remove="handleRemove"
          >
            <div class="add-more">
              <el-icon><Plus /></el-icon>
              <span>继续添加</span>
            </div>
          </el-upload>
        </div>
      </div>

      <template #footer>
        <div class="upload-footer">
          <div class="footer-info">
            <el-icon><Picture /></el-icon>
            <span>共 <strong>{{ uploads.length }}</strong> 张待上传</span>
          </div>
          <div class="footer-btns">
            <button class="btn btn-default" @click="uploadPhoto = false">取消</button>
            <button class="btn btn-primary" @click="savePhotos" :disabled="uploads.length === 0">
              <el-icon><UploadFilled /></el-icon>开始上传
            </button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑照片对话框 -->
    <el-dialog v-model="editPhoto" width="460px" class="edit-dialog" destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <span class="dialog-icon primary">
            <el-icon><Edit /></el-icon>
          </span>
          <div>
            <h3 class="dialog-title">编辑照片</h3>
            <p class="dialog-subtitle">修改照片信息</p>
          </div>
        </div>
      </template>
      <div class="dialog-body">
        <div class="form-group">
          <label class="form-label">照片名称 <span class="required">*</span></label>
          <input v-model="photoForm.photoName" class="form-input" placeholder="请输入照片名称" maxlength="50" />
        </div>
        <div class="form-group">
          <label class="form-label">照片描述</label>
          <textarea v-model="photoForm.photoDesc" class="form-textarea" placeholder="请输入照片描述" rows="3" maxlength="200" />
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="editPhoto = false">取消</button>
          <button class="btn btn-primary" @click="updatePhoto">保存</button>
        </div>
      </template>
    </el-dialog>

    <!-- 批量删除对话框 -->
    <el-dialog v-model="batchDeletePhoto" width="420px" class="delete-dialog" destroy-on-close>
      <div class="delete-confirm">
        <div class="delete-icon-wrap warning">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <h3>删除到回收站</h3>
        <p>确定要将选中的 <strong>{{ selectphotoIds.length }}</strong> 张照片移入回收站吗？</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="batchDeletePhoto = false">取消</button>
          <button class="btn btn-danger" @click="updatePhotoDelete(null)">确认删除</button>
        </div>
      </template>
    </el-dialog>

    <!-- 移动照片对话框 -->
    <el-dialog v-model="movePhoto" width="460px" class="move-dialog" destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <span class="dialog-icon move">
            <el-icon><FolderOpened /></el-icon>
          </span>
          <div>
            <h3 class="dialog-title">移动照片</h3>
            <p class="dialog-subtitle">选择目标相册</p>
          </div>
        </div>
      </template>
      <div class="dialog-body">
        <div v-if="albumList.length < 2" class="move-empty">
          <el-icon><FolderOpened /></el-icon>
          <p>暂无其他相册</p>
        </div>
        <div v-else class="album-select-list">
          <div
            v-for="item in albumList"
            :key="item.id"
            v-show="item.id !== albumInfo.id"
            class="album-select-item"
            :class="{ active: targetAlbumId === item.id }"
            @click="targetAlbumId = item.id"
          >
            <el-image fit="cover" class="item-cover" lazy :src="item.albumCover" />
            <div class="item-info">
              <span class="item-name">{{ item.albumName }}</span>
              <span class="item-count">{{ item.photoCount || 0 }} 张照片</span>
            </div>
            <div class="item-check">
              <el-icon v-if="targetAlbumId === item.id"><SuccessFilled /></el-icon>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="movePhoto = false">取消</button>
          <button class="btn btn-primary" :disabled="!targetAlbumId" @click="updatePhotoAlbum">确认移动</button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Picture, FolderOpened, Delete, Edit, Plus, UploadFilled,
  WarningFilled, InfoFilled, ZoomIn, Checked, SuccessFilled
} from '@element-plus/icons-vue'
import * as imageConversion from 'image-conversion'
import request from '@/utils/request'
import { getAuthHeaders } from '@/utils/auth'
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

const headers = computed(() => getAuthHeaders())
const photoSrcList = computed(() => photos.value.map(p => p.photoSrc))

onMounted(() => {
  const albumId = route.params.albumId
  if (albumId === pageStateStore.photo?.albumId) {
    current.value = pageStateStore.photo.current || 1
  } else {
    current.value = 1
    pageStateStore.updatePhotoPageState({ albumId: route.params.albumId, current: current.value })
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
    if (data && data.data) Object.assign(albumInfo, data.data)
  })
}

const listAlbums = () => {
  request.get('/admin/photos/albums/info').then(({ data }) => {
    if (data && data.data) albumList.value = data.data
  })
}

const listPhotos = () => {
  request.get('/admin/photos', {
    params: { current: current.value, size: size.value, albumId: route.params.albumId, isDelete: 0 }
  }).then(({ data }) => {
    if (data && data.data) {
      photos.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => { loading.value = false })
}

const currentChange = (val) => {
  current.value = val
  pageStateStore.updatePhotoPageState({ albumId: route.params.albumId, current: current.value })
  listPhotos()
}

const savePhotos = () => {
  const photoUrls = uploads.value.map(item => item.url)
  request.post('/admin/photos', { albumId: route.params.albumId, photoUrls }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      uploads.value = []
      listPhotos()
      getAlbumInfo()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    uploadPhoto.value = false
  })
}

const updatePhoto = () => {
  if (!photoForm.photoName?.trim()) {
    ElMessage.error('照片名称不能为空')
    return
  }
  request.put('/admin/photos', photoForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listPhotos()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
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
      ElNotification.success({ title: '成功', message: data.message })
      getAlbumInfo()
      listPhotos()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    movePhoto.value = false
    targetAlbumId.value = null
  })
}

const handleRemove = (file) => {
  uploads.value = uploads.value.filter(item => item.url !== file.url)
}

const upload = (response) => {
  if (response && response.data) uploads.value.push({ url: response.data })
}

const beforeUpload = (file) => {
  return new Promise((resolve) => {
    if (file.size / 1024 < UPLOAD_SIZE) {
      resolve(file)
    } else {
      imageConversion.compressAccurately(file, UPLOAD_SIZE).then((res) => resolve(res))
    }
  })
}

const toggleSelect = (id) => {
  const idx = selectphotoIds.value.indexOf(id)
  if (idx > -1) selectphotoIds.value.splice(idx, 1)
  else selectphotoIds.value.push(id)
  updateCheckAllState()
}

const updateCheckAllState = () => {
  const len = selectphotoIds.value.length
  checkAll.value = len === photoIds.value.length && len > 0
  isIndeterminate.value = len > 0 && len < photoIds.value.length
}

const handleCheckAllChange = (e) => {
  const val = e.target.checked
  selectphotoIds.value = val ? [...photoIds.value] : []
  isIndeterminate.value = false
  checkAll.value = val
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
  const param = id ? { ids: [id], isDelete: 1 } : { ids: selectphotoIds.value, isDelete: 1 }
  request.put('/admin/photos/delete', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listPhotos()
      getAlbumInfo()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    batchDeletePhoto.value = false
  })
}

const previewPhoto = () => {
  // el-image preview-src-list handles click preview
}
</script>

<style scoped>
.photo-page {
  padding: 4px 0;
}

/* ========== 相册横幅 ========== */
.album-banner {
  position: relative;
  border-radius: 16px;
  overflow: hidden;
  margin-bottom: 20px;
  border: 1px solid var(--border-color, #ebeef5);
}
.banner-cover {
  height: 140px;
  overflow: hidden;
}
.banner-img {
  width: 100%;
  height: 100%;
  filter: blur(2px) brightness(0.6);
  transform: scale(1.05);
}
.banner-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(26, 115, 232, 0.3), rgba(99, 102, 241, 0.2));
}
.banner-content {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;
  background: linear-gradient(90deg, rgba(0,0,0,0.4), rgba(0,0,0,0.1));
}
.banner-info { display: flex; flex-direction: column; gap: 4px; }
.banner-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #fff;
  text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.banner-desc {
  margin: 0;
  font-size: 13px;
  color: rgba(255,255,255,0.8);
  max-width: 400px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ========== 统计卡片 ========== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  padding: 18px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  transition: all 0.25s ease;
}
.stat-card:hover { transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.06); }
.stat-icon {
  width: 46px; height: 46px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 20px; flex-shrink: 0;
}
.stat-icon.primary { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.stat-icon.info { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1.2; }
.stat-label { font-size: 13px; color: var(--text-secondary, #6b7280); margin-top: 2px; }

/* ========== 工具栏 ========== */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  padding: 14px 20px;
  margin-bottom: 20px;
}
.toolbar-left { display: flex; align-items: center; gap: 10px; }
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.toolbar-divider { width: 1px; height: 20px; background: var(--border-color, #e5e7eb); }

/* ========== 按钮 ========== */
.btn {
  height: 34px; padding: 0 14px; border: none; border-radius: 8px;
  font-size: 13px; font-weight: 500; cursor: pointer;
  display: inline-flex; align-items: center; gap: 5px;
  transition: all 0.2s;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-primary { background: linear-gradient(135deg, #1a73e8, #4285f4); color: #fff; box-shadow: 0 2px 8px rgba(26,115,232,0.25); }
.btn-primary:hover:not(:disabled) { background: linear-gradient(135deg, #1557b0, #1a73e8); transform: translateY(-1px); }
.btn-move { background: #ecfdf5; color: #059669; border: 1px solid #a7f3d0; }
.btn-move:hover:not(:disabled) { background: #059669; color: #fff; }
.btn-danger-outline { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }
.btn-danger-outline:hover:not(:disabled) { background: #dc2626; color: #fff; }
.btn-danger { background: #ef4444; color: #fff; box-shadow: 0 2px 8px rgba(239,68,68,0.25); }
.btn-danger:hover:not(:disabled) { background: #dc2626; }
.btn-default { background: var(--bg-card, #fff); color: var(--text-primary, #374151); border: 1px solid var(--border-color, #d1d5db); }
.btn-default:hover { background: var(--bg-body, #f9fafb); }

/* ========== 全选 ========== */
.select-all {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px; color: var(--text-secondary, #6b7280);
  cursor: pointer; user-select: none; padding: 4px 8px;
  border-radius: 6px; transition: all 0.2s;
}
.select-all.active { background: #eff6ff; color: #2563eb; }
.select-all input[type="checkbox"] { width: 16px; height: 16px; accent-color: #1a73e8; cursor: pointer; }
.select-badge {
  background: #2563eb; color: #fff; font-size: 11px;
  padding: 0 6px; border-radius: 10px; min-width: 20px;
  text-align: center; line-height: 18px;
}

/* ========== 照片网格 ========== */
.photo-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 14px;
  min-height: 200px;
}

/* ========== 照片卡片 ========== */
.photo-card {
  position: relative;
  background: var(--bg-card, #fff);
  border: 2px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.25s ease;
}
.photo-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.08);
  border-color: transparent;
}
.photo-card.selected {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26,115,232,0.15);
}

.card-checkbox {
  position: absolute; top: 8px; left: 8px; z-index: 5;
}
.card-checkbox input[type="checkbox"] {
  width: 18px; height: 18px; accent-color: #1a73e8; cursor: pointer;
  opacity: 0; transition: opacity 0.2s;
}
.photo-card:hover .card-checkbox input,
.photo-card.selected .card-checkbox input { opacity: 1; }

.card-cover {
  position: relative; width: 100%; aspect-ratio: 1;
  cursor: pointer; overflow: hidden;
}
.cover-img { width: 100%; height: 100%; display: block; transition: transform 0.4s ease; }
.photo-card:hover .cover-img { transform: scale(1.06); }
.cover-overlay {
  position: absolute; inset: 0; background: rgba(0,0,0,0.35);
  display: flex; align-items: center; justify-content: center;
  opacity: 0; transition: opacity 0.25s;
}
.photo-card:hover .cover-overlay { opacity: 1; }
.preview-icon { font-size: 28px; color: #fff; filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3)); }

.card-info { padding: 8px 10px; }
.photo-name {
  display: block; font-size: 12px; color: var(--text-primary, #374151);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  text-align: center; line-height: 1.4;
}

.card-actions {
  position: absolute; bottom: 40px; right: 8px; z-index: 5;
  opacity: 0; transform: translateY(4px); transition: all 0.2s;
}
.photo-card:hover .card-actions { opacity: 1; transform: translateY(0); }
.mini-btn {
  width: 30px; height: 30px; border: none; border-radius: 8px;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  font-size: 14px; transition: all 0.2s; backdrop-filter: blur(6px);
}
.mini-btn.edit { background: rgba(26,115,232,0.85); color: #fff; }
.mini-btn.edit:hover { background: #1a73e8; transform: scale(1.1); }

/* ========== 空状态 ========== */
.empty-state { grid-column: 1 / -1; padding: 60px 20px; text-align: center; }
.empty-icon { font-size: 56px; color: var(--border-color, #d1d5db); margin-bottom: 16px; }
.empty-state p { font-size: 16px; color: var(--text-secondary, #6b7280); margin: 0 0 6px; font-weight: 500; }
.empty-hint { font-size: 13px; color: var(--text-tertiary, #9ca3af); }

/* ========== 分页 ========== */
.pagination-bar { display: flex; justify-content: center; margin-top: 24px; }

/* ========== 对话框通用 ========== */
.dialog-header { display: flex; align-items: center; gap: 14px; }
.dialog-icon {
  width: 42px; height: 42px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center; font-size: 20px;
}
.dialog-icon.primary { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.dialog-icon.move { background: linear-gradient(135deg, #e6f4ea, #ceead6); color: #1e8e3e; }
.dialog-title { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.dialog-subtitle { margin: 3px 0 0; font-size: 13px; color: var(--text-secondary, #6b7280); }
.dialog-body { padding: 4px 0; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 8px; }

.upload-dialog :deep(.el-dialog) { border-radius: 16px; }
.upload-dialog :deep(.el-dialog__header) { padding: 24px 24px 0; margin: 0; }
.upload-dialog :deep(.el-dialog__body) { padding: 16px 24px; }
.upload-dialog :deep(.el-dialog__footer) { padding: 16px 24px; border-top: 1px solid var(--border-color, #f0f0f0); }

.edit-dialog :deep(.el-dialog),
.delete-dialog :deep(.el-dialog),
.move-dialog :deep(.el-dialog) { border-radius: 16px; }
.edit-dialog :deep(.el-dialog__header),
.delete-dialog :deep(.el-dialog__header),
.move-dialog :deep(.el-dialog__header) { padding: 24px 24px 0; margin: 0; }
.edit-dialog :deep(.el-dialog__body),
.delete-dialog :deep(.el-dialog__body),
.move-dialog :deep(.el-dialog__body) { padding: 16px 24px; }
.edit-dialog :deep(.el-dialog__footer),
.delete-dialog :deep(.el-dialog__footer),
.move-dialog :deep(.el-dialog__footer) { padding: 16px 24px; border-top: 1px solid var(--border-color, #f0f0f0); }

/* ========== 表单 ========== */
.form-group { margin-bottom: 18px; }
.form-label { display: block; font-size: 13px; font-weight: 500; color: var(--text-primary, #374151); margin-bottom: 6px; }
.required { color: #ef4444; }
.form-input {
  width: 100%; height: 38px; padding: 0 12px;
  border: 1px solid var(--border-color, #d1d5db); border-radius: 8px;
  font-size: 13px; color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff); outline: none;
  transition: all 0.2s; box-sizing: border-box;
}
.form-input:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1); }
.form-textarea {
  width: 100%; padding: 10px 12px;
  border: 1px solid var(--border-color, #d1d5db); border-radius: 8px;
  font-size: 13px; color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff); outline: none;
  transition: all 0.2s; resize: vertical; font-family: inherit;
  box-sizing: border-box; line-height: 1.5;
}
.form-textarea:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1); }

/* ========== 上传拖拽 ========== */
.upload-drag :deep(.el-upload) { width: 100%; }
.upload-drag :deep(.el-upload-dragger) {
  width: 100%; height: 280px;
  border: 2px dashed var(--border-color, #d1d5db); border-radius: 14px;
  background: var(--bg-body, #f9fafb); transition: all 0.3s;
  display: flex; align-items: center; justify-content: center;
}
.upload-drag :deep(.el-upload-dragger:hover) {
  border-color: #1a73e8; background: #eff6ff;
}
.drag-content { display: flex; flex-direction: column; align-items: center; gap: 10px; }
.drag-icon { font-size: 52px; color: #1a73e8; opacity: 0.7; }
.drag-title { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.drag-sub { margin: 0; font-size: 13px; color: var(--text-secondary, #6b7280); }
.drag-sub em { color: #1a73e8; font-style: normal; font-weight: 500; }
.drag-hint {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 14px; background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 20px; font-size: 12px; color: var(--text-tertiary, #9ca3af);
}

.preview-bar {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px; padding-bottom: 10px;
  border-bottom: 1px solid var(--border-color, #ebeef5);
  font-size: 14px; color: var(--text-secondary, #6b7280);
}
.preview-bar strong { color: #1a73e8; }
.text-btn {
  background: none; border: none; color: #6b7280; font-size: 12px;
  cursor: pointer; display: inline-flex; align-items: center; gap: 3px;
}
.text-btn:hover { color: #ef4444; }

.upload-preview :deep(.el-upload-list--picture-card .el-upload-list__item) { border-radius: 10px; }
.upload-preview :deep(.el-upload--picture-card) {
  border-radius: 10px; border: 2px dashed var(--border-color, #d1d5db);
  background: var(--bg-body, #f9fafb);
}
.upload-preview :deep(.el-upload--picture-card:hover) { border-color: #1a73e8; }
.add-more { display: flex; flex-direction: column; align-items: center; gap: 4px; color: var(--text-tertiary, #9ca3af); }
.add-more .el-icon { font-size: 22px; }
.add-more span { font-size: 11px; }

.upload-footer {
  display: flex; align-items: center; justify-content: space-between;
}
.footer-info {
  display: flex; align-items: center; gap: 6px;
  font-size: 13px; color: var(--text-secondary, #6b7280);
}
.footer-info .el-icon { color: #1a73e8; }
.footer-info strong { color: #1a73e8; }
.footer-btns { display: flex; gap: 8px; }

/* ========== 删除确认 ========== */
.delete-confirm { text-align: center; }
.delete-icon-wrap {
  width: 56px; height: 56px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 28px; margin: 0 auto 16px;
}
.delete-icon-wrap.warning { background: #fef3e0; color: #e65100; }
.delete-confirm h3 { margin: 0 0 8px; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.delete-confirm p { margin: 0; font-size: 13px; color: var(--text-secondary, #6b7280); line-height: 1.6; }
.delete-confirm strong { color: #ef4444; }

/* ========== 移动相册列表 ========== */
.move-empty {
  text-align: center; padding: 40px 20px;
  color: var(--text-tertiary, #9ca3af);
}
.move-empty .el-icon { font-size: 40px; margin-bottom: 8px; }
.move-empty p { margin: 0; font-size: 14px; }

.album-select-list {
  display: flex; flex-direction: column; gap: 8px;
  max-height: 320px; overflow-y: auto;
}
.album-select-item {
  display: flex; align-items: center; gap: 12px;
  padding: 10px 12px; border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 10px; cursor: pointer; transition: all 0.2s;
}
.album-select-item:hover { border-color: #93c5fd; background: #eff6ff; }
.album-select-item.active { border-color: #1a73e8; background: linear-gradient(135deg, #eff6ff, #dbeafe); }
.item-cover {
  width: 44px; height: 44px; border-radius: 8px;
  object-fit: cover; flex-shrink: 0;
}
.item-info { flex: 1; min-width: 0; }
.item-name {
  display: block; font-size: 14px; font-weight: 500;
  color: var(--text-primary, #1f2937);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.item-count { font-size: 12px; color: var(--text-tertiary, #9ca3af); }
.item-check { color: #1a73e8; font-size: 18px; flex-shrink: 0; }

/* ========== 暗色模式 ========== */
[data-theme="dark"] .stat-card {
  background: var(--bg-card, #1e293b);
  border-color: var(--border-color, #334155);
}
[data-theme="dark"] .stat-value {
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  color: var(--neon-blue, #00D4FF);
}
[data-theme="dark"] .toolbar { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .photo-card {
  background: var(--bg-card, #1e293b);
  border-color: var(--border-color, #334155);
  transition: all 0.25s ease;
}
[data-theme="dark"] .photo-card:hover {
  border-color: rgba(0, 212, 255, 0.3);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.1);
}
[data-theme="dark"] .photo-card.selected {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 3px rgba(0, 212, 255, 0.2), 0 0 15px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .select-all.active {
  background: rgba(0, 212, 255, 0.1);
  color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .select-badge {
  background: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.4);
}
[data-theme="dark"] .form-input, [data-theme="dark"] .form-textarea {
  background: var(--bg-card, #1e293b);
  border-color: var(--border-color, #334155);
  color: var(--text-primary, #e2e8f0);
}
[data-theme="dark"] .form-input:focus, [data-theme="dark"] .form-textarea:focus {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15), 0 0 12px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .upload-drag :deep(.el-upload-dragger) {
  background: var(--bg-body, #0f172a);
  border-color: rgba(0, 212, 255, 0.2);
  transition: all 0.25s ease;
}
[data-theme="dark"] .upload-drag :deep(.el-upload-dragger:hover) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.15);
}
[data-theme="dark"] .album-select-item {
  background: var(--bg-card, #1e293b);
  border-color: var(--border-color, #334155);
  transition: all 0.25s ease;
}
[data-theme="dark"] .album-select-item:hover {
  border-color: rgba(0, 212, 255, 0.3);
  background: rgba(0, 212, 255, 0.05);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.1);
}
[data-theme="dark"] .album-select-item.active {
  border-color: var(--neon-blue, #00D4FF);
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.08), rgba(0, 212, 255, 0.03));
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.15);
}

/* ========== 响应式 ========== */
@media (max-width: 1400px) { .photo-grid { grid-template-columns: repeat(5, 1fr); } }
@media (max-width: 1100px) { .photo-grid { grid-template-columns: repeat(4, 1fr); } }
@media (max-width: 768px) {
  .photo-grid { grid-template-columns: repeat(3, 1fr); gap: 10px; }
  .toolbar { flex-direction: column; gap: 10px; }
  .toolbar-left { flex-wrap: wrap; }
  .banner-content { padding: 0 16px; }
  .banner-title { font-size: 18px; }
  .banner-cover { height: 110px; }
}
@media (max-width: 480px) {
  .photo-grid { grid-template-columns: repeat(2, 1fr); gap: 8px; }
  .banner-cover { height: 90px; }
  .banner-title { font-size: 16px; }
  .banner-desc { display: none; }
}
</style>

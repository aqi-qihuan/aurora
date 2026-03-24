<template>
  <div class="album-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><PictureFilled /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">相册总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><View /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ publicCount }}</span>
          <span class="stat-label">公开相册</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Lock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ privateCount }}</span>
          <span class="stat-label">私密相册</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon purple">
          <el-icon><PictureRounded /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ totalPhotos }}</span>
          <span class="stat-label">照片总数</span>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <button class="btn btn-primary" @click="openModel(null)">
          <el-icon><Plus /></el-icon>新建相册
        </button>
        <button class="btn btn-ghost" @click="checkDelete">
          <el-icon><Delete /></el-icon>回收站
        </button>
      </div>
      <div class="toolbar-right">
        <div class="search-box">
          <el-icon class="search-icon"><Search /></el-icon>
          <input
            v-model="keywords"
            class="search-input"
            placeholder="搜索相册名称..."
            @keyup.enter="searchAlbums"
          />
          <kbd class="search-kbd">Enter</kbd>
        </div>
      </div>
    </div>

    <!-- 相册卡片网格 -->
    <div class="album-grid" v-loading="loading">
      <div
        v-for="item in albums"
        :key="item.id"
        class="album-card"
        @click="checkPhoto(item)"
      >
        <!-- 封面区域 -->
        <div class="album-cover-wrap">
          <el-image fit="cover" class="album-cover" lazy :src="item.albumCover" />
          <div class="cover-overlay">
            <div class="cover-stats">
              <span class="photo-count">
                <el-icon><Picture /></el-icon>
                {{ item.photoCount || 0 }}
              </span>
              <span v-if="item.status === 2" class="private-badge">
                <el-icon><Lock /></el-icon>私密
              </span>
            </div>
          </div>
          <!-- 操作菜单 -->
          <div class="card-actions" @click.stop>
            <el-dropdown @command="handleCommand" trigger="click">
              <button class="action-trigger" @click.stop>
                <el-icon><MoreFilled /></el-icon>
              </button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="'update' + JSON.stringify(item)">
                    <el-icon><Edit /></el-icon>编辑相册
                  </el-dropdown-item>
                  <el-dropdown-item :command="'delete' + item.id" divided>
                    <el-icon style="color:#ef4444"><Delete /></el-icon>
                    <span style="color:#ef4444">删除相册</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
        <!-- 信息区域 -->
        <div class="album-info">
          <h3 class="album-name" :title="item.albumName">{{ item.albumName }}</h3>
          <p class="album-desc" :title="item.albumDesc">{{ item.albumDesc || '暂无描述' }}</p>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && (!albums || albums.length === 0)" class="empty-state">
        <el-icon class="empty-icon"><PictureFilled /></el-icon>
        <p>暂无相册</p>
        <span class="empty-hint">点击「新建相册」创建你的第一个相册</span>
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

    <!-- 新建/编辑对话框 -->
    <el-dialog v-model="addOrEdit" width="520px" top="8vh" class="album-dialog" destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <span class="dialog-icon primary">
            <el-icon><PictureFilled /></el-icon>
          </span>
          <div>
            <h3 class="dialog-title">{{ albumForum.id ? '编辑相册' : '新建相册' }}</h3>
            <p class="dialog-subtitle">{{ albumForum.id ? '修改相册信息' : '创建一个新的相册' }}</p>
          </div>
        </div>
      </template>

      <div class="dialog-body">
        <div class="form-group">
          <label class="form-label">相册名称 <span class="required">*</span></label>
          <input
            v-model="albumForum.albumName"
            class="form-input"
            placeholder="请输入相册名称"
            maxlength="30"
          />
        </div>
        <div class="form-group">
          <label class="form-label">相册描述 <span class="required">*</span></label>
          <textarea
            v-model="albumForum.albumDesc"
            class="form-textarea"
            placeholder="请输入相册描述"
            rows="3"
            maxlength="120"
          />
        </div>
        <div class="form-group">
          <label class="form-label">相册封面 <span class="required">*</span></label>
          <div class="cover-upload">
            <el-upload
              class="uploader"
              drag
              :headers="headers"
              :before-upload="beforeUpload"
              action="/api/admin/photos/albums/upload"
              :on-success="uploadCover"
              :show-file-list="false"
              accept="image/*"
            >
              <template v-if="albumForum.albumCover">
                <img :src="albumForum.albumCover" class="cover-preview" />
                <div class="cover-change">
                  <el-icon><RefreshRight /></el-icon>
                  <span>更换封面</span>
                </div>
              </template>
              <template v-else>
                <el-icon class="upload-icon"><UploadFilled /></el-icon>
                <p class="upload-text">将图片拖到此处，或<em>点击上传</em></p>
                <p class="upload-hint">支持 JPG/PNG，建议 720x360</p>
              </template>
            </el-upload>
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">发布形式</label>
          <div class="status-cards">
            <div
              class="status-card"
              :class="{ active: albumForum.status === 1 }"
              @click="albumForum.status = 1"
            >
              <el-icon><View /></el-icon>
              <span class="status-name">公开</span>
              <span class="status-desc">所有人可见</span>
            </div>
            <div
              class="status-card"
              :class="{ active: albumForum.status === 2 }"
              @click="albumForum.status = 2"
            >
              <el-icon><Lock /></el-icon>
              <span class="status-name">私密</span>
              <span class="status-desc">仅自己可见</span>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="addOrEdit = false">取消</button>
          <button class="btn btn-primary" @click="addOrEditAlbum">确定</button>
        </div>
      </template>
    </el-dialog>

    <!-- 删除确认对话框 -->
    <el-dialog v-model="isdelete" width="420px" class="delete-dialog" destroy-on-close>
      <div class="delete-confirm">
        <div class="delete-icon-wrap">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <h3>确认删除</h3>
        <p>删除后相册将进入回收站，确定要删除该相册吗？</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="isdelete = false">取消</button>
          <button class="btn btn-danger" @click="deleteAlbum">确认删除</button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Plus, Delete, Search, MoreFilled, Edit, Lock,
  UploadFilled, WarningFilled, PictureFilled, PictureRounded,
  Picture, View, RefreshRight
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import logger from '@/utils/logger'
import { getAuthHeaders } from '@/utils/auth'
import { createBeforeUploadHandler } from '@/utils/imageUtils'

const route = useRoute()
const router = useRouter()

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
const headers = computed(() => getAuthHeaders())

const publicCount = computed(() => albums.value.filter(a => a.status === 1).length)
const privateCount = computed(() => albums.value.filter(a => a.status === 2).length)
const totalPhotos = computed(() => albums.value.reduce((sum, a) => sum + (a.photoCount || 0), 0))

const openModel = (item) => {
  if (item) {
    albumForum.value = JSON.parse(item)
  } else {
    albumForum.value = { id: null, albumName: '', albumDesc: '', albumCover: '', status: 1 }
  }
  addOrEdit.value = true
}

const checkPhoto = (item) => {
  router.push({ path: '/albums/' + item.id })
}

const checkDelete = () => {
  router.push({ path: '/photos/delete' })
}

const listAlbums = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin/photos/albums', {
      params: { current: current.value, size: size.value, keywords: keywords.value }
    })
    if (data && data.data) {
      albums.value = data.data.records || []
      count.value = data.data.count || 0
    }
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '获取相册列表失败' })
  } finally {
    loading.value = false
  }
}

const addOrEditAlbum = async () => {
  if (!albumForum.value.albumName.trim()) {
    ElMessage.error('相册名称不能为空')
    return
  }
  if (!albumForum.value.albumDesc.trim()) {
    ElMessage.error('相册描述不能为空')
    return
  }
  if (!albumForum.value.albumCover) {
    ElMessage.error('相册封面不能为空')
    return
  }
  try {
    const { data } = await request.post('/admin/photos/albums', albumForum.value)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listAlbums()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    addOrEdit.value = false
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '操作失败' })
  }
}

const uploadCover = (response) => {
  albumForum.value.albumCover = response.data
}

const beforeUpload = createBeforeUploadHandler(500)

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

const deleteAlbum = async () => {
  try {
    const { data } = await request.delete('/admin/photos/albums/' + albumForum.value.id)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listAlbums()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    isdelete.value = false
  } catch (error) {
    ElMessage.error('删除相册失败')
    logger.error('API Error:', error)
  }
}

const searchAlbums = () => {
  current.value = 1
  listAlbums()
}

const sizeChange = (val) => {
  size.value = val
  listAlbums()
}

const currentChange = (val) => {
  current.value = val
  listAlbums()
}

onMounted(() => {
  listAlbums()
})
</script>

<style scoped>
.album-page {
  padding: 4px 0;
}

/* ========== 统计卡片 ========== */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
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
.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.06);
}
.stat-icon {
  width: 46px;
  height: 46px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}
.stat-icon.primary { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.stat-icon.success { background: linear-gradient(135deg, #e6f4ea, #ceead6); color: #1e8e3e; }
.stat-icon.warning { background: linear-gradient(135deg, #fef3e0, #fde8c8); color: #e65100; }
.stat-icon.purple { background: linear-gradient(135deg, #f3e8ff, #e9d5ff); color: #7c3aed; }
.stat-info { display: flex; flex-direction: column; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1.2; }
.stat-label { font-size: 13px; color: var(--text-secondary, #6b7280); margin-top: 2px; }

/* ========== 工具栏 ========== */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 12px;
  padding: 14px 20px;
}
.toolbar-left { display: flex; gap: 8px; }
.toolbar-right { display: flex; align-items: center; }

/* ========== 按钮 ========== */
.btn {
  height: 36px;
  padding: 0 16px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}
.btn-primary {
  background: linear-gradient(135deg, #1a73e8, #4285f4);
  color: #fff;
  box-shadow: 0 2px 8px rgba(26, 115, 232, 0.25);
}
.btn-primary:hover {
  background: linear-gradient(135deg, #1557b0, #1a73e8);
  box-shadow: 0 4px 12px rgba(26, 115, 232, 0.35);
  transform: translateY(-1px);
}
.btn-ghost {
  background: transparent;
  color: var(--text-secondary, #6b7280);
}
.btn-ghost:hover { background: var(--bg-body, #f3f4f6); color: var(--text-primary, #374151); }
.btn-default {
  background: var(--bg-card, #fff);
  color: var(--text-primary, #374151);
  border: 1px solid var(--border-color, #d1d5db);
}
.btn-default:hover { background: var(--bg-body, #f9fafb); }
.btn-danger {
  background: #ef4444;
  color: #fff;
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.25);
}
.btn-danger:hover { background: #dc2626; }

/* ========== 搜索框 ========== */
.search-box {
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 10px;
  color: var(--text-tertiary, #9ca3af);
  font-size: 15px;
  pointer-events: none;
}
.search-input {
  width: 220px;
  height: 36px;
  padding: 0 50px 0 32px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  background: var(--bg-body, #f9fafb);
  outline: none;
  transition: all 0.2s;
  box-sizing: border-box;
}
.search-input:focus {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.1);
  background: var(--bg-card, #fff);
}
.search-kbd {
  position: absolute;
  right: 8px;
  padding: 1px 6px;
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
  background: var(--bg-body, #f3f4f6);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 4px;
  font-family: inherit;
  pointer-events: none;
}

/* ========== 相册网格 ========== */
.album-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
  min-height: 200px;
}

/* ========== 相册卡片 ========== */
.album-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 14px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}
.album-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.1);
  border-color: transparent;
}

/* 封面 */
.album-cover-wrap {
  position: relative;
  width: 100%;
  height: 170px;
  overflow: hidden;
}
.album-cover {
  width: 100%;
  height: 100%;
  display: block;
  transition: transform 0.5s cubic-bezier(0.25, 0.46, 0.45, 0.94);
}
.album-card:hover .album-cover {
  transform: scale(1.06);
}
.cover-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(0,0,0,0.15) 0%, transparent 40%, rgba(0,0,0,0.5) 100%);
  display: flex;
  align-items: flex-end;
  padding: 12px;
  opacity: 0;
  transition: opacity 0.3s;
}
.album-card:hover .cover-overlay {
  opacity: 1;
}
.cover-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}
.photo-count {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 10px;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(6px);
  border-radius: 20px;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
}
.private-badge {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 3px 10px;
  background: rgba(234, 179, 8, 0.85);
  backdrop-filter: blur(6px);
  border-radius: 20px;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
}

/* 操作按钮 */
.card-actions {
  position: absolute;
  top: 10px;
  right: 10px;
  opacity: 0;
  transition: opacity 0.2s;
}
.album-card:hover .card-actions { opacity: 1; }
.action-trigger {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(6px);
  color: var(--text-primary, #374151);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
.action-trigger:hover {
  background: #fff;
  transform: scale(1.05);
}

/* 信息区域 */
.album-info {
  padding: 14px 16px 16px;
}
.album-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}
.album-desc {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

/* ========== 空状态 ========== */
.empty-state {
  grid-column: 1 / -1;
  padding: 60px 20px;
  text-align: center;
}
.empty-icon { font-size: 56px; color: var(--border-color, #d1d5db); margin-bottom: 16px; }
.empty-state p { font-size: 16px; color: var(--text-secondary, #6b7280); margin: 0 0 6px; font-weight: 500; }
.empty-hint { font-size: 13px; color: var(--text-tertiary, #9ca3af); }

/* ========== 分页 ========== */
.pagination-bar {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

/* ========== 对话框 ========== */
.album-dialog :deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
}
.album-dialog :deep(.el-dialog__header) {
  padding: 24px 24px 0;
  margin: 0;
}
.album-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}
.album-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 14px;
}
.dialog-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}
.dialog-icon.primary { background: linear-gradient(135deg, #e8f0fe, #d2e3fc); color: #1a73e8; }
.dialog-title { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.dialog-subtitle { margin: 3px 0 0; font-size: 13px; color: var(--text-secondary, #6b7280); }

/* ========== 表单 ========== */
.form-group {
  margin-bottom: 20px;
}
.form-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary, #374151);
  margin-bottom: 6px;
}
.required { color: #ef4444; }
.form-input {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff);
  outline: none;
  transition: all 0.2s;
  box-sizing: border-box;
}
.form-input:focus {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.1);
}
.form-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff);
  outline: none;
  transition: all 0.2s;
  resize: vertical;
  font-family: inherit;
  box-sizing: border-box;
  line-height: 1.5;
}
.form-textarea:focus {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.1);
}

/* ========== 封面上传 ========== */
.cover-upload {
  width: 100%;
}
.uploader :deep(.el-upload) {
  width: 100%;
  border: 2px dashed var(--border-color, #d1d5db);
  border-radius: 12px;
  transition: all 0.2s;
  overflow: hidden;
}
.uploader :deep(.el-upload:hover) {
  border-color: #1a73e8;
}
.uploader :deep(.el-upload-dragger) {
  width: 100%;
  height: 170px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: var(--bg-body, #f9fafb);
  border: none;
  border-radius: 10px;
  padding: 20px;
}
.cover-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 8px;
}
.cover-change {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 13px;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}
.uploader :deep(.el-upload:hover) .cover-change { opacity: 1; }
.upload-icon { font-size: 32px; color: var(--text-tertiary, #9ca3af); margin-bottom: 8px; }
.upload-text { font-size: 13px; color: var(--text-secondary, #6b7280); margin: 0; }
.upload-text em { color: #1a73e8; font-style: normal; }
.upload-hint { font-size: 12px; color: var(--text-tertiary, #9ca3af); margin: 6px 0 0; }

/* ========== 状态选择卡片 ========== */
.status-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.status-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 12px;
  border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  background: var(--bg-card, #fff);
}
.status-card:hover {
  border-color: #93c5fd;
  background: #eff6ff;
}
.status-card.active {
  border-color: #1a73e8;
  background: linear-gradient(135deg, #eff6ff, #dbeafe);
}
.status-card .el-icon {
  font-size: 24px;
  color: var(--text-secondary, #6b7280);
  transition: color 0.2s;
}
.status-card.active .el-icon { color: #1a73e8; }
.status-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}
.status-desc {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

/* ========== 删除确认 ========== */
.delete-dialog :deep(.el-dialog) {
  border-radius: 16px;
}
.delete-dialog :deep(.el-dialog__body) {
  padding: 32px 24px;
}
.delete-confirm {
  text-align: center;
}
.delete-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #fef2f2;
  color: #ef4444;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  margin: 0 auto 16px;
}
.delete-confirm h3 {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}
.delete-confirm p {
  margin: 0;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  line-height: 1.6;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* ========== 暗色模式适配 ========== */
[data-theme="dark"] .album-card {
  background: var(--bg-card, #1e293b);
  border-color: var(--border-color, #334155);
  transition: all 0.25s ease;
}
[data-theme="dark"] .album-card:hover {
  border-color: rgba(0, 212, 255, 0.3);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4), 0 0 15px rgba(0, 212, 255, 0.1);
}
[data-theme="dark"] .stat-card {
  background: var(--bg-card, #1e293b);
  border-color: var(--border-color, #334155);
}
[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 15px var(--primary-glow);
}
[data-theme="dark"] .stat-value {
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  color: var(--neon-blue, #00D4FF);
}
[data-theme="dark"] .toolbar { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .uploader :deep(.el-upload-dragger) {
  background: var(--bg-body, #0f172a);
  border-color: rgba(0, 212, 255, 0.2);
  transition: all 0.25s ease;
}
[data-theme="dark"] .uploader :deep(.el-upload-dragger:hover) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.15);
}
[data-theme="dark"] .status-card { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .status-card:hover {
  border-color: var(--neon-blue, #00D4FF);
  background: rgba(0, 212, 255, 0.05);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.15);
}
[data-theme="dark"] .status-card.active {
  border-color: var(--neon-blue, #00D4FF);
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.08), rgba(0, 212, 255, 0.03));
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
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

/* ========== 响应式 ========== */
@media (max-width: 1200px) {
  .album-grid { grid-template-columns: repeat(3, 1fr); }
  .stats-row { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .album-grid { grid-template-columns: repeat(2, 1fr); gap: 12px; }
  .toolbar { flex-direction: column; gap: 10px; }
  .toolbar-right { width: 100%; }
  .search-input { width: 100%; }
  .stats-row { grid-template-columns: repeat(2, 1fr); gap: 10px; }
  .album-cover-wrap { height: 140px; }
}
@media (max-width: 480px) {
  .album-grid { grid-template-columns: 1fr; }
  .album-cover-wrap { height: 180px; }
  .status-cards { grid-template-columns: 1fr; }
  .search-kbd { display: none; }
}
</style>

<template>
  <div class="recycle-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Delete /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ count }}</span>
          <span class="stat-label">回收照片</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon info">
          <el-icon><Checked /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ selectPhotoIds.length }}</span>
          <span class="stat-label">已选中</span>
        </div>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <label class="select-all" :class="{ active: selectPhotoIds.length > 0 }">
          <input
            type="checkbox"
            :checked="checkAll"
            :indeterminate="isIndeterminate"
            @change="handleCheckAllChange"
          />
          <span>全选</span>
          <span v-if="selectPhotoIds.length" class="select-badge">{{ selectPhotoIds.length }}</span>
        </label>
        <div class="toolbar-divider"></div>
        <button class="btn btn-success" :disabled="!selectPhotoIds.length" @click="updatePhotoDelete(null)">
          <el-icon><RefreshRight /></el-icon>批量恢复
        </button>
        <button class="btn btn-danger" :disabled="!selectPhotoIds.length" @click="batchDeletePhoto = true">
          <el-icon><Delete /></el-icon>彻底删除
        </button>
      </div>
      <div class="toolbar-right">
        <span class="toolbar-hint">选中照片可恢复或彻底删除</span>
      </div>
    </div>

    <!-- 照片网格 -->
    <div class="photo-grid" v-loading="loading">
      <div
        v-for="item in photos"
        :key="item.id"
        class="photo-card"
        :class="{ selected: selectPhotoIds.includes(item.id) }"
      >
        <!-- 选择框 -->
        <div class="card-checkbox">
          <input
            type="checkbox"
            :checked="selectPhotoIds.includes(item.id)"
            @change="toggleSelect(item.id)"
          />
        </div>
        <!-- 封面 -->
        <div class="card-cover" @click="previewPhoto(item)">
          <el-image fit="cover" class="cover-img" lazy :src="item.photoSrc" :preview-src-list="photoSrcList" :initial-index="photos.indexOf(item)" />
          <div class="cover-overlay">
            <el-icon class="preview-icon"><ZoomIn /></el-icon>
          </div>
        </div>
        <!-- 信息 -->
        <div class="card-info">
          <span class="photo-name" :title="item.photoName">{{ item.photoName }}</span>
        </div>
        <!-- 底部操作 -->
        <div class="card-actions">
          <button class="mini-btn restore" @click="updatePhotoDelete(item.id)" title="恢复照片">
            <el-icon><RefreshRight /></el-icon>
          </button>
          <button class="mini-btn remove" @click="confirmSingleDelete(item)" title="彻底删除">
            <el-icon><Delete /></el-icon>
          </button>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!loading && photos.length === 0" class="empty-state">
        <el-icon class="empty-icon"><Delete /></el-icon>
        <p>回收站是空的</p>
        <span class="empty-hint">删除的照片会在这里保留一段时间</span>
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

    <!-- 批量删除确认 -->
    <el-dialog v-model="batchDeletePhoto" width="420px" class="delete-dialog" destroy-on-close>
      <div class="delete-confirm">
        <div class="delete-icon-wrap danger">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <h3>彻底删除确认</h3>
        <p>确定要彻底删除选中的 <strong>{{ selectPhotoIds.length }}</strong> 张照片吗？此操作不可恢复！</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="batchDeletePhoto = false">取消</button>
          <button class="btn btn-danger" @click="deletePhotos">确认删除</button>
        </div>
      </template>
    </el-dialog>

    <!-- 单张删除确认 -->
    <el-dialog v-model="singleDeleteDialog" width="420px" class="delete-dialog" destroy-on-close>
      <div class="delete-confirm">
        <div class="delete-icon-wrap danger">
          <el-icon><WarningFilled /></el-icon>
        </div>
        <h3>彻底删除确认</h3>
        <p>确定要彻底删除这张照片吗？此操作不可恢复！</p>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <button class="btn btn-default" @click="singleDeleteDialog = false">取消</button>
          <button class="btn btn-danger" @click="deleteSinglePhoto">确认删除</button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification } from 'element-plus'
import { RefreshRight, Delete, WarningFilled, ZoomIn, Checked } from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()

const loading = ref(true)
const batchDeletePhoto = ref(false)
const singleDeleteDialog = ref(false)
const isIndeterminate = ref(false)
const checkAll = ref(false)
const photos = ref([])
const photoIds = ref([])
const selectPhotoIds = ref([])
const pendingDeleteId = ref(null)
const current = ref(1)
const size = ref(18)
const count = ref(0)

const photoSrcList = computed(() => photos.value.map(p => p.photoSrc))

onMounted(() => {
  listPhotos()
})

watch(photos, (newPhotos) => {
  photoIds.value = newPhotos.map(item => item.id)
  selectPhotoIds.value = []
  checkAll.value = false
  isIndeterminate.value = false
})

const listPhotos = () => {
  request.get('/admin/photos', {
    params: { current: current.value, size: size.value, isDelete: 1 }
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

const sizeChange = (val) => {
  size.value = val
  listPhotos()
}

const currentChange = (val) => {
  current.value = val
  listPhotos()
}

const toggleSelect = (id) => {
  const idx = selectPhotoIds.value.indexOf(id)
  if (idx > -1) {
    selectPhotoIds.value.splice(idx, 1)
  } else {
    selectPhotoIds.value.push(id)
  }
  updateCheckAllState()
}

const updateCheckAllState = () => {
  const len = selectPhotoIds.value.length
  checkAll.value = len === photoIds.value.length && len > 0
  isIndeterminate.value = len > 0 && len < photoIds.value.length
}

const handleCheckAllChange = (e) => {
  const val = e.target.checked
  selectPhotoIds.value = val ? [...photoIds.value] : []
  isIndeterminate.value = false
  checkAll.value = val
}

const handleCheckedPhotoChange = (value) => {
  const checkedCount = value.length
  checkAll.value = checkedCount === photoIds.value.length
  isIndeterminate.value = checkedCount > 0 && checkedCount < photoIds.value.length
}

const updatePhotoDelete = (id) => {
  const param = id
    ? { ids: [id], isDelete: 0 }
    : { ids: selectPhotoIds.value, isDelete: 0 }
  request.put('/admin/photos/delete', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listPhotos()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    batchDeletePhoto.value = false
  })
}

const confirmSingleDelete = (item) => {
  pendingDeleteId.value = item.id
  singleDeleteDialog.value = true
}

const deleteSinglePhoto = () => {
  if (!pendingDeleteId.value) return
  request.delete('/admin/photos', { data: [pendingDeleteId.value] }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listPhotos()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    singleDeleteDialog.value = false
    pendingDeleteId.value = null
  })
}

const deletePhotos = () => {
  request.delete('/admin/photos', { data: selectPhotoIds.value }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listPhotos()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    batchDeletePhoto.value = false
  })
}

const previewPhoto = (item) => {
  // el-image preview-src-list handles click preview
}
</script>

<style scoped>
.recycle-page {
  padding: 4px 0;
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
.stat-icon.warning { background: linear-gradient(135deg, #fef3e0, #fde8c8); color: #e65100; }
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
.toolbar-right { display: flex; align-items: center; }
.toolbar-divider { width: 1px; height: 20px; background: var(--border-color, #e5e7eb); }
.toolbar-hint { font-size: 12px; color: var(--text-tertiary, #9ca3af); }

/* ========== 按钮 ========== */
.btn {
  height: 34px;
  padding: 0 14px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transition: all 0.2s;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-success { background: #ecfdf5; color: #059669; border: 1px solid #a7f3d0; }
.btn-success:hover:not(:disabled) { background: #059669; color: #fff; }
.btn-danger { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }
.btn-danger:hover:not(:disabled) { background: #dc2626; color: #fff; }
.btn-default { background: var(--bg-card, #fff); color: var(--text-primary, #374151); border: 1px solid var(--border-color, #d1d5db); }
.btn-default:hover { background: var(--bg-body, #f9fafb); }

/* ========== 全选 ========== */
.select-all {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  user-select: none;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}
.select-all.active { background: #eff6ff; color: #2563eb; }
.select-all input[type="checkbox"] { width: 16px; height: 16px; accent-color: #1a73e8; cursor: pointer; }
.select-badge {
  background: #2563eb;
  color: #fff;
  font-size: 11px;
  padding: 0 6px;
  border-radius: 10px;
  min-width: 20px;
  text-align: center;
  line-height: 18px;
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
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
  border-color: transparent;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}
.photo-card.selected {
  border-color: #1a73e8;
  box-shadow: 0 0 0 3px rgba(26, 115, 232, 0.15);
}

/* 选择框 */
.card-checkbox {
  position: absolute;
  top: 8px;
  left: 8px;
  z-index: 5;
}
.card-checkbox input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: #1a73e8;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
.photo-card:hover .card-checkbox input,
.photo-card.selected .card-checkbox input {
  opacity: 1;
}

/* 封面 */
.card-cover {
  position: relative;
  width: 100%;
  aspect-ratio: 1;
  cursor: pointer;
  overflow: hidden;
}
.cover-img {
  width: 100%;
  height: 100%;
  display: block;
  transition: transform 0.4s ease;
}
.photo-card:hover .cover-img {
  transform: scale(1.06);
}
.cover-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.25s;
}
.photo-card:hover .cover-overlay { opacity: 1; }
.preview-icon {
  font-size: 28px;
  color: #fff;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.3));
}

/* 信息 */
.card-info {
  padding: 8px 10px;
}
.photo-name {
  display: block;
  font-size: 12px;
  color: var(--text-primary, #374151);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: center;
  line-height: 1.4;
}

/* 底部操作 */
.card-actions {
  position: absolute;
  bottom: 44px;
  right: 8px;
  z-index: 5;
  display: flex;
  gap: 4px;
  opacity: 0;
  transform: translateY(4px);
  transition: all 0.2s;
}
.photo-card:hover .card-actions {
  opacity: 1;
  transform: translateY(0);
}
.mini-btn {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s;
  backdrop-filter: blur(6px);
}
.mini-btn.restore {
  background: rgba(5, 150, 105, 0.85);
  color: #fff;
}
.mini-btn.restore:hover { background: #059669; transform: scale(1.1); }
.mini-btn.remove {
  background: rgba(220, 38, 38, 0.85);
  color: #fff;
}
.mini-btn.remove:hover { background: #dc2626; transform: scale(1.1); }

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

/* ========== 删除确认对话框 ========== */
.delete-dialog :deep(.el-dialog) {
  border-radius: 16px;
}
.delete-dialog :deep(.el-dialog__body) {
  padding: 32px 24px;
}
.delete-confirm { text-align: center; }
.delete-icon-wrap {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  margin: 0 auto 16px;
}
.delete-icon-wrap.danger { background: #fef2f2; color: #ef4444; }
.delete-confirm h3 { margin: 0 0 8px; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.delete-confirm p { margin: 0; font-size: 13px; color: var(--text-secondary, #6b7280); line-height: 1.6; }
.delete-confirm strong { color: #ef4444; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 8px; }

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

/* ========== 响应式 ========== */
@media (max-width: 1400px) {
  .photo-grid { grid-template-columns: repeat(5, 1fr); }
}
@media (max-width: 1100px) {
  .photo-grid { grid-template-columns: repeat(4, 1fr); }
}
@media (max-width: 768px) {
  .photo-grid { grid-template-columns: repeat(3, 1fr); gap: 10px; }
  .toolbar { flex-direction: column; gap: 10px; align-items: flex-start; }
  .toolbar-right { display: none; }
  .stats-row { gap: 10px; }
}
@media (max-width: 480px) {
  .photo-grid { grid-template-columns: repeat(2, 1fr); gap: 8px; }
}
</style>

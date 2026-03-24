<template>
  <div class="about-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><EditPen /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ charCount }}</span>
          <span class="stat-label">字符数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ lineCount }}</span>
          <span class="stat-label">行数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Clock /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ readTime }}</span>
          <span class="stat-label">预计阅读</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <div class="editor-card">
      <!-- 工具栏 -->
      <div class="editor-toolbar">
        <div class="toolbar-left">
          <div class="toolbar-title">
            <el-icon><Reading /></el-icon>
            <span>关于我</span>
          </div>
          <span class="toolbar-divider"></span>
          <span class="edit-hint" v-if="!aboutContent">点击编辑器开始编写你的"关于我"页面</span>
          <span class="edit-hint saved" v-else-if="saveSuccess">
            <el-icon><CircleCheck /></el-icon>
            内容已保存
          </span>
          <span class="edit-hint" v-else>
            <el-icon><Edit /></el-icon>
            编辑中...
          </span>
        </div>
        <div class="toolbar-right">
          <el-button @click="previewContent" :icon="View" class="btn-preview">
            <span>预览</span>
          </el-button>
          <button class="btn-save" :disabled="saving" @click="updateAbout">
            <el-icon v-if="saving" class="is-loading"><Loading /></el-icon>
            <el-icon v-else><Check /></el-icon>
            {{ saving ? '保存中...' : '保存' }}
          </button>
        </div>
      </div>

      <!-- 编辑器容器 -->
      <div class="editor-wrapper">
        <MdEditor
          ref="mdEditorRef"
          v-model="aboutContent"
          @onUploadImg="handleUploadImg"
          @onChange="onContentChange"
          style="height: calc(100vh - 300px); border: none; border-radius: 0 0 16px 16px;" />
      </div>
    </div>

    <!-- 预览对话框 -->
    <el-dialog v-model="showPreview" width="700px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><Reading /></el-icon></div>
      <div class="dialog-content">
        <h3>关于我 - 预览</h3>
      </div>
      <div class="preview-content" v-html="previewHtml"></div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showPreview = false" class="btn-cancel">关闭</el-button>
          <button class="btn-confirm" @click="showPreview = false; updateAbout()">
            <el-icon><Check /></el-icon>
            保存内容
          </button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification } from 'element-plus'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import * as imageConversion from 'image-conversion'
import { EditPen, Document, Clock, Reading, Edit, Check, View, CircleCheck, Loading } from '@element-plus/icons-vue'
import request from '@/utils/request'
import logger from '@/utils/logger'

const UPLOAD_SIZE = 1024

const route = useRoute()
const aboutContent = ref('')
const mdEditorRef = ref(null)
const saving = ref(false)
const saveSuccess = ref(false)
const showPreview = ref(false)

const charCount = computed(() => aboutContent.value.length)
const lineCount = computed(() => aboutContent.value ? aboutContent.value.split('\n').length : 0)
const readTime = computed(() => {
  const words = aboutContent.value.length
  const minutes = Math.max(1, Math.ceil(words / 500))
  return `${minutes}分`
})

const previewHtml = computed(() => {
  if (!aboutContent.value) return '<p style="color: var(--text-secondary); text-align: center; padding: 40px;">暂无内容</p>'
  return aboutContent.value
})

const onContentChange = () => {
  if (saveSuccess.value) saveSuccess.value = false
}

const previewContent = () => {
  if (!aboutContent.value) {
    ElNotification.info({ title: '提示', message: '请先编写内容' })
    return
  }
  showPreview.value = true
}

onMounted(() => { getAbout() })

const getAbout = () => {
  request.get('/about').then(({ data }) => {
    if (data && data.data) {
      aboutContent.value = data.data.content || ''
    }
  }).catch(() => {})
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
      if (data && data.data) { callback(data.data) }
    } catch (error) {
      logger.error('上传图片失败:', error)
    }
  }
}

const updateAbout = () => {
  if (saving.value) return
  saving.value = true
  request.put('/admin/about', { content: aboutContent.value }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      saveSuccess.value = true
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  }).catch((error) => {
    ElNotification.error({ title: '失败', message: error.message || '保存失败' })
  }).finally(() => {
    saving.value = false
  })
}
</script>

<style scoped>
.about-page { padding: 0; }

/* 统计卡片 */
.stats-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-bottom: 24px; }
.stat-card {
  background: var(--bg-base, #fff); border-radius: 16px; padding: 24px;
  display: flex; align-items: center; gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid var(--border-default, #e5e7eb);
  transition: all 0.3s ease;
}
.stat-card:hover { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(0,0,0,0.08); }
.stat-icon { width: 56px; height: 56px; border-radius: 14px; display: flex; align-items: center; justify-content: center; font-size: 24px; flex-shrink: 0; }
.stat-icon.primary { background: linear-gradient(135deg, #3b82f6, #60a5fa); color: #fff; }
.stat-icon.success { background: linear-gradient(135deg, #10b981, #34d399); color: #fff; }
.stat-icon.warning { background: linear-gradient(135deg, #f59e0b, #fbbf24); color: #fff; }
.stat-info { display: flex; flex-direction: column; gap: 4px; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1; }
.stat-label { font-size: 14px; color: var(--text-secondary, #6b7280); }

/* 编辑器卡片 */
.editor-card {
  background: var(--bg-base, #fff);
  border-radius: 16px;
  border: 1px solid var(--border-default, #e5e7eb);
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  overflow: hidden;
}

/* 工具栏 */
.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: var(--bg-elevated, #f9fafb);
  border-bottom: 1px solid var(--border-default, #e5e7eb);
}
.toolbar-left { display: flex; align-items: center; gap: 12px; }
.toolbar-right { display: flex; align-items: center; gap: 12px; }
.toolbar-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 16px; font-weight: 600;
  color: var(--text-primary, #1f2937);
}
.toolbar-title .el-icon { color: var(--color-primary, #3b82f6); }
.toolbar-divider { width: 1px; height: 20px; background: var(--border-default, #e5e7eb); }
.edit-hint {
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  display: flex; align-items: center; gap: 6px;
}
.edit-hint.saved { color: #10b981; }
.edit-hint .el-icon { font-size: 14px; }

/* 预览按钮 */
.btn-preview {
  border-radius: 10px; font-weight: 500; height: 40px; padding: 0 20px;
  transition: all 0.2s ease;
}
.btn-preview:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.1); }

/* 保存按钮 */
.btn-save {
  display: flex; align-items: center; gap: 6px;
  padding: 0 24px; height: 40px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-save:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }
.btn-save:disabled { opacity: 0.7; cursor: not-allowed; transform: none; box-shadow: none; }

/* 编辑器 */
.editor-wrapper {
  :deep(.md-editor) {
    border: none;
  }
  :deep(.md-editor .md-editor-toolbar) {
    border-bottom: 1px solid var(--border-default, #e5e7eb);
  }
  :deep(.md-editor-dark) {
    --md-bk-color: var(--bg-base, #fff);
  }
}

/* 预览对话框 */
.modern-dialog :deep(.el-dialog__header) { display: none; }
.modern-dialog :deep(.el-dialog__body) { padding: 32px 32px 24px; max-height: 70vh; overflow-y: auto; }
.modern-dialog :deep(.el-dialog__footer) { padding: 0 32px 32px; }
.dialog-icon-wrapper { width: 64px; height: 64px; border-radius: 16px; display: flex; align-items: center; justify-content: center; font-size: 28px; margin: 0 auto 20px; }
.dialog-icon-wrapper.primary { background: linear-gradient(135deg, #eff6ff, #dbeafe); color: #3b82f6; }
.dialog-content { text-align: center; }
.dialog-content h3 { font-size: 20px; font-weight: 600; color: var(--text-primary, #1f2937); margin: 0 0 8px; }
.preview-content {
  margin-top: 20px;
  padding: 24px;
  background: var(--bg-elevated, #f9fafb);
  border-radius: 12px;
  border: 1px solid var(--border-default, #e5e7eb);
  text-align: left;
  font-size: 15px;
  line-height: 1.8;
  color: var(--text-primary, #1f2937);
  max-height: 50vh;
  overflow-y: auto;
}
.preview-content :deep(img) { max-width: 100%; border-radius: 8px; }
.preview-content :deep(h1), .preview-content :deep(h2), .preview-content :deep(h3) { margin: 16px 0 8px; }
.preview-content :deep(p) { margin: 8px 0; }
.dialog-footer { display: flex; gap: 12px; justify-content: center; }
.btn-cancel { border-radius: 10px; height: 44px; padding: 0 24px; font-weight: 500; }
.btn-confirm {
  display: flex; align-items: center; gap: 6px;
  padding: 0 24px; height: 44px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-confirm:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }

/* 深色模式 - 极客风 */
[data-theme="dark"] .stat-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}
[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 15px var(--primary-glow);
}
[data-theme="dark"] .stat-value {
  color: var(--text-primary, #f9fafb);
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
}
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .editor-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}
[data-theme="dark"] .editor-card:hover {
  border-color: rgba(0, 212, 255, 0.2);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.08);
}
[data-theme="dark"] .editor-toolbar {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(51, 65, 85, 0.7) 100%);
  border-color: var(--border-default, #374151);
}
[data-theme="dark"] .toolbar-title {
  color: var(--neon-blue, #00D4FF);
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
}
[data-theme="dark"] .edit-hint { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .preview-content {
  background: var(--bg-elevated, #374151);
  border-color: rgba(0, 212, 255, 0.2);
  color: var(--text-primary, #f9fafb);
  box-shadow: inset 0 0 30px rgba(0, 212, 255, 0.03);
}
[data-theme="dark"] .btn-preview:hover {
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.3);
}
[data-theme="dark"] .btn-save {
  background: linear-gradient(135deg, var(--neon-blue, #00D4FF) 0%, var(--neon-purple, #BF5AF2) 100%);
  box-shadow: 0 4px 14px rgba(0, 212, 255, 0.4);
}

/* 响应式 */
@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; }
  .editor-toolbar { flex-direction: column; gap: 12px; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; justify-content: space-between; }
  .btn-preview { flex: 1; }
  .btn-save { flex: 1; justify-content: center; }
  .stat-value { font-size: 24px; }
}
</style>

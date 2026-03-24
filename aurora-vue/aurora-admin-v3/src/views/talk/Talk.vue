<template>
  <div class="talk-publish-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-info">
        <div class="header-icon">
          <el-icon><EditPen /></el-icon>
        </div>
        <div class="header-text">
          <h2 class="header-title">{{ isEdit ? '编辑说说' : '发布说说' }}</h2>
          <p class="header-desc">{{ isEdit ? '修改说说内容和状态' : '分享你的想法和动态' }}</p>
        </div>
      </div>
    </div>

    <!-- 编辑区域 -->
    <el-card class="main-card">
      <!-- 编辑器 -->
      <div class="editor-section">
        <Editor ref="editorRef" class="editor-wrapper" :value="talk.content" @input="talk.content = $event" placeholder="说点什么吧..." />
        <div class="editor-footer">
          <span class="char-count" :class="{ 'has-content': talk.content?.length > 0 }">
            {{ talk.content?.length || 0 }} 字
          </span>
        </div>
      </div>

      <!-- 操作栏 -->
      <div class="operation-bar">
        <div class="operation-left">
          <!-- 图片上传按钮 -->
          <el-tooltip content="上传图片" placement="top">
            <el-upload
              action="/api/admin/talks/images"
              multiple
              :headers="headers"
              :before-upload="beforeUpload"
              :on-success="upload"
              :show-file-list="false">
              <button class="op-btn upload-btn">
                <el-icon><Picture /></el-icon>
                <span>图片</span>
              </button>
            </el-upload>
          </el-tooltip>
        </div>
        <div class="operation-right">
          <!-- 置顶开关 -->
          <div class="op-toggle">
            <el-icon class="toggle-icon"><Top /></el-icon>
            <span class="toggle-label">置顶</span>
            <el-switch
              v-model="talk.isTop"
              :active-value="1"
              :inactive-value="0"
              inline-prompt
              active-text="是"
              inactive-text="否" />
          </div>

          <!-- 状态选择器 -->
          <el-dropdown trigger="click" @command="handleCommand">
            <button class="op-btn status-btn" :class="{ 'status-private': talk.status === 2 }">
              <el-icon><View v-if="talk.status === 1" /><Lock v-else /></el-icon>
              <span>{{ dropdownTitle }}</span>
              <el-icon class="arrow-icon"><ArrowDown /></el-icon>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="(item, index) of statuses" :key="index" :command="item.status">
                  <el-icon><View v-if="item.status === 1" /><Lock v-else /></el-icon>
                  {{ item.desc }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <!-- 发布按钮 -->
          <button
            class="publish-btn"
            :disabled="!talk.content?.trim()"
            @click="saveOrUpdateTalk">
            <el-icon><Promotion /></el-icon>
            <span>{{ isEdit ? '保存修改' : '发布说说' }}</span>
          </button>
        </div>
      </div>

      <!-- 图片上传预览区 -->
      <transition name="slide-fade">
        <div v-if="uploads.length > 0" class="image-upload-section">
          <div class="section-header">
            <div class="section-title">
              <el-icon><PictureFilled /></el-icon>
              <span>已上传图片 ({{ uploads.length }}/9)</span>
            </div>
            <button class="clear-all-btn" @click="clearAllImages">
              <el-icon><Delete /></el-icon>
              <span>清空</span>
            </button>
          </div>
          <el-upload
            class="talk-image-upload"
            action="/api/admin/talks/images"
            list-type="picture-card"
            multiple
            :headers="headers"
            :file-list="uploads"
            :before-upload="beforeUpload"
            :on-success="upload"
            :on-remove="handleRemove">
            <div v-if="uploads.length < 9" class="upload-trigger">
              <el-icon><Plus /></el-icon>
              <span>添加</span>
            </div>
          </el-upload>
        </div>
      </transition>
    </el-card>

    <!-- 提示卡片 -->
    <el-card class="tips-card">
      <div class="tips-header">
        <el-icon><InfoFilled /></el-icon>
        <span>发布小贴士</span>
      </div>
      <div class="tips-list">
        <div class="tip-item">
          <el-icon><PictureFilled /></el-icon>
          <span>支持上传多张图片，最多 9 张</span>
        </div>
        <div class="tip-item">
          <el-icon><Top /></el-icon>
          <span>置顶的说说将展示在最前面</span>
        </div>
        <div class="tip-item">
          <el-icon><Lock /></el-icon>
          <span>私密说说仅自己可见</span>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Picture, ArrowDown, Plus, EditPen, Promotion,
  Top, View, Lock, Delete, InfoFilled, PictureFilled
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import Editor from '@/components/Editor.vue'
import { getAuthHeaders } from '@/utils/auth'
import { createBeforeUploadHandler } from '@/utils/imageUtils'

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

const headers = computed(() => getAuthHeaders())

const isEdit = computed(() => !!talk.id)

const isWriteMode = computed(() => route.params.talkId === 'write')

const dropdownTitle = computed(() => {
  const found = statuses.find(item => item.status === talk.status)
  return found ? found.desc : '公开'
})

onMounted(() => {
  // "write" 标识发布模式，不需要加载已有数据
  if (route.params.talkId && route.params.talkId !== 'write') {
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

const clearAllImages = () => {
  uploads.value = []
}

const upload = (response) => {
  if (response && response.data) {
    uploads.value.push({ url: response.data })
  }
}

const beforeUpload = createBeforeUploadHandler(1024)

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
/* ==================== 页面容器 ==================== */
.talk-publish-page {
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* ==================== 页面头部 ==================== */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 14px;
}

.header-icon {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: #fff;
  flex-shrink: 0;
}

.header-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary, #1a1a2e);
  margin: 0;
  line-height: 1.3;
}

.header-desc {
  font-size: 13px;
  color: var(--text-muted, #8e8ea0);
  margin: 2px 0 0;
}

/* ==================== 主内容卡片 ==================== */
.main-card {
  border-radius: 16px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.06));
  box-shadow: var(--shadow-sm, 0 1px 3px rgba(0, 0, 0, 0.04));
}

.main-card :deep(.el-card__body) {
  padding: 24px;
}

/* ==================== 编辑器 ==================== */
.editor-section {
  margin-bottom: 20px;
}

.editor-wrapper {
  min-height: 160px;
  border-radius: 12px;
  border: 2px solid var(--border-light, rgba(0, 0, 0, 0.08));
  transition: border-color 0.3s;
}

.editor-wrapper:focus-within {
  border-color: #667eea;
  box-shadow: 0 0 0 4px rgba(102, 126, 234, 0.1);
}

.editor-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 6px;
}

.char-count {
  font-size: 12px;
  color: var(--text-muted, #aaa);
  font-variant-numeric: tabular-nums;
}

.char-count.has-content {
  color: #667eea;
}

/* ==================== 操作栏 ==================== */
.operation-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-top: 1px solid var(--border-light, rgba(0, 0, 0, 0.06));
  flex-wrap: wrap;
  gap: 12px;
}

.operation-left {
  display: flex;
  gap: 8px;
}

.operation-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 操作按钮 */
.op-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 10px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.08));
  background: var(--bg-card, #fff);
  color: var(--text-secondary, #666);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.op-btn:hover {
  border-color: #667eea;
  color: #667eea;
  background: rgba(102, 126, 234, 0.04);
}

.op-btn .arrow-icon {
  font-size: 12px;
  transition: transform 0.3s;
}

.op-btn:hover .arrow-icon {
  transform: rotate(180deg);
}

.upload-btn .el-icon {
  font-size: 17px;
}

.status-btn.status-private {
  color: #f5576c;
  border-color: rgba(245, 87, 108, 0.25);
  background: rgba(245, 87, 108, 0.04);
}

.status-btn.status-private:hover {
  border-color: #f5576c;
  background: rgba(245, 87, 108, 0.08);
}

/* 置顶开关 */
.op-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 10px;
  background: var(--bg-secondary, #f5f5f7);
}

.toggle-icon {
  font-size: 15px;
  color: var(--text-muted, #999);
}

.toggle-label {
  font-size: 13px;
  color: var(--text-secondary, #666);
  font-weight: 500;
}

/* 发布按钮 */
.publish-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 24px;
  border-radius: 12px;
  border: none;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 3px 12px rgba(102, 126, 234, 0.25);
}

.publish-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.4);
}

.publish-btn:active:not(:disabled) {
  transform: translateY(0);
}

.publish-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  box-shadow: none;
}

/* ==================== 图片上传区 ==================== */
.image-upload-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid var(--border-light, rgba(0, 0, 0, 0.06));
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
}

.section-title .el-icon {
  color: #667eea;
}

.clear-all-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border-radius: 8px;
  border: none;
  background: rgba(245, 87, 108, 0.08);
  color: #f5576c;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.clear-all-btn:hover {
  background: rgba(245, 87, 108, 0.15);
}

/* 上传组件美化 */
.talk-image-upload :deep(.el-upload--picture-card) {
  width: 100px;
  height: 100px;
  border-radius: 12px;
  border: 2px dashed var(--border-light, rgba(0, 0, 0, 0.1));
  transition: all 0.3s;
}

.talk-image-upload :deep(.el-upload--picture-card):hover {
  border-color: #667eea;
  background: rgba(102, 126, 234, 0.04);
}

.talk-image-upload :deep(.el-upload-list--picture-card .el-upload-list__item) {
  width: 100px;
  height: 100px;
  border-radius: 12px;
  border: none;
  box-shadow: var(--shadow-sm, 0 2px 8px rgba(0, 0, 0, 0.08));
  overflow: hidden;
}

.talk-image-upload :deep(.el-upload-list--picture-card .el-upload-list__item:hover) {
  box-shadow: var(--shadow-md, 0 4px 16px rgba(0, 0, 0, 0.15));
}

.talk-image-upload :deep(.el-upload-list__item-status-label) {
  right: 6px;
  top: 6px;
}

.talk-image-upload :deep(.el-upload-list__item-actions) {
  border-radius: 12px;
}

.upload-trigger {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  height: 100%;
  color: var(--text-muted, #bbb);
  font-size: 12px;
}

.upload-trigger .el-icon {
  font-size: 24px;
}

/* 过渡动画 */
.slide-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.slide-fade-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* ==================== 提示卡片 ==================== */
.tips-card {
  border-radius: 14px;
  border: 1px solid var(--border-light, rgba(0, 0, 0, 0.04));
  background: var(--bg-elevated, rgba(255, 255, 255, 0.5));
}

.tips-card :deep(.el-card__body) {
  padding: 16px 20px;
}

.tips-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1a1a2e);
  margin-bottom: 12px;
}

.tips-header .el-icon {
  color: #667eea;
  font-size: 17px;
}

.tips-list {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.tip-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-muted, #8e8ea0);
}

.tip-item .el-icon {
  font-size: 15px;
  color: #a78bfa;
}

/* ==================== 深色模式 - 极客风 ==================== */
[data-theme="dark"] .main-card {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.06));
}

[data-theme="dark"] .tips-card {
  background: var(--bg-elevated, rgba(30, 30, 46, 0.6));
  border-color: rgba(0, 212, 255, 0.15);
  box-shadow: inset 0 0 20px rgba(0, 212, 255, 0.03);
}

[data-theme="dark"] .op-btn {
  background: var(--bg-card, #1e1e2e);
  border-color: var(--border-light, rgba(255, 255, 255, 0.08));
  color: var(--text-secondary, #aaa);
  transition: all 0.25s ease;
}

[data-theme="dark"] .op-btn:hover {
  border-color: var(--neon-blue, #00D4FF);
  color: var(--neon-blue, #00D4FF);
  background: rgba(0, 212, 255, 0.08);
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.2);
}

[data-theme="dark"] .op-toggle {
  background: var(--bg-secondary, #2a2a3e);
}

[data-theme="dark"] .status-btn.status-private {
  background: rgba(255, 45, 146, 0.1);
  color: var(--neon-pink, #FF2D92);
}

[data-theme="dark"] .status-btn.status-private:hover {
  background: rgba(255, 45, 146, 0.18);
  box-shadow: 0 0 8px rgba(255, 45, 146, 0.2);
}

[data-theme="dark"] .clear-all-btn {
  background: rgba(255, 45, 146, 0.1);
  color: var(--neon-pink, #FF2D92);
}

[data-theme="dark"] .clear-all-btn:hover {
  background: rgba(255, 45, 146, 0.2);
  box-shadow: 0 0 10px rgba(255, 45, 146, 0.2);
}

[data-theme="dark"] .editor-wrapper {
  border-color: rgba(0, 212, 255, 0.2);
  box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.05);
}

/* ==================== 响应式 ==================== */
@media (max-width: 768px) {
  .operation-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .operation-right {
    flex-wrap: wrap;
    gap: 10px;
  }

  .op-toggle {
    flex: 1;
    justify-content: space-between;
  }

  .publish-btn {
    width: 100%;
    justify-content: center;
    padding: 12px 24px;
  }

  .tips-list {
    flex-direction: column;
    gap: 8px;
  }

  .talk-image-upload :deep(.el-upload--picture-card),
  .talk-image-upload :deep(.el-upload-list--picture-card .el-upload-list__item) {
    width: 80px;
    height: 80px;
  }
}

@media (max-width: 480px) {
  .main-card :deep(.el-card__body) {
    padding: 16px;
  }

  .header-title {
    font-size: 17px;
  }

  .header-icon {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    font-size: 18px;
  }

  .tips-card :deep(.el-card__body) {
    padding: 12px 16px;
  }
}
</style>

<template>
  <div class="article-edit-page">
    <!-- 页面头部 -->
    <div class="edit-header">
      <div class="header-left">
        <h1 class="page-title">{{ route.meta.title || '编辑文章' }}</h1>
        <span class="edit-status" :class="{ saved: hasContent }">
          <span class="status-dot"></span>
          {{ hasContent ? '已输入内容' : '空白草稿' }}
        </span>
      </div>
      <div class="header-actions">
        <el-button
          v-if="!article.id || article.status === 3"
          class="draft-btn"
          @click="saveArticleDraft">
          <el-icon><Document /></el-icon>
          <span>保存草稿</span>
        </el-button>
        <el-button type="primary" class="publish-btn" @click="openModel">
          <el-icon><Promotion /></el-icon>
          <span>发布文章</span>
        </el-button>
      </div>
    </div>

    <!-- 标题输入区 -->
    <div class="title-input-area">
      <input
        v-model="article.articleTitle"
        class="title-input"
        placeholder="输入文章标题..."
        maxlength="100"
      />
      <div class="title-meta">
        <span class="char-count" :class="{ warning: article.articleTitle.length > 80 }">
          {{ article.articleTitle.length }} / 100
        </span>
      </div>
    </div>

    <!-- 编辑器区域 -->
    <div class="editor-area">
      <Editor v-model="article.articleContent" placeholder="开始撰写你的文章内容..." />
    </div>

    <!-- 发布对话框 -->
    <el-dialog
      v-model="addOrEdit"
      width="560px"
      top="12vh"
      class="publish-dialog"
      :close-on-click-modal="false"
      destroy-on-close>
      <template #header>
        <div class="dialog-header">
          <div class="dialog-icon">
            <el-icon><Promotion /></el-icon>
          </div>
          <div class="dialog-title-group">
            <h3 class="dialog-title">发布文章</h3>
            <p class="dialog-subtitle">设置文章的分类、标签等信息</p>
          </div>
        </div>
      </template>

      <div class="publish-form">
        <!-- 文章分类 -->
        <div class="form-section">
          <label class="form-label">
            <span class="label-required">*</span>
            文章分类
          </label>
          <div class="tag-select-area">
            <span
              v-if="article.categoryName"
              class="selected-tag category-tag">
              <el-icon><Folder /></el-icon>
              {{ article.categoryName }}
              <button class="tag-remove" @click="removeCategory">
                <el-icon><Close /></el-icon>
              </button>
            </span>
            <el-select
              v-if="!article.categoryName"
              v-model="article.categoryId"
              placeholder="选择文章分类"
              class="form-select"
              @change="handleCategoryChange">
              <el-option
                v-for="item in categories"
                :key="item.id"
                :label="item.categoryName"
                :value="item.id"
              />
            </el-select>
          </div>
        </div>

        <!-- 文章标签 -->
        <div class="form-section">
          <label class="form-label">文章标签</label>
          <div class="tag-select-area">
            <div class="tag-list" v-if="article.tagNames.length">
              <span
                v-for="(item, index) in article.tagNames"
                :key="index"
                class="selected-tag label-tag">
                <el-icon><PriceTag /></el-icon>
                {{ item }}
                <button class="tag-remove" @click="removeTag(item)">
                  <el-icon><Close /></el-icon>
                </button>
              </span>
            </div>
            <el-select
              v-if="article.tagNames.length < 3"
              v-model="selectedTag"
              placeholder="添加标签（最多3个）"
              class="form-select"
              @change="addTag">
              <el-option
                v-for="item in availableTags"
                :key="item.id"
                :label="item.tagName"
                :value="item.tagName"
              />
            </el-select>
            <span class="tag-hint" v-if="article.tagNames.length >= 3">
              已达标签上限
            </span>
          </div>
        </div>

        <!-- 文章类型 -->
        <div class="form-section">
          <label class="form-label">
            <span class="label-required">*</span>
            文章类型
          </label>
          <div class="type-selector">
            <label
              v-for="item in typeList"
              :key="item.type"
              :class="['type-option', { active: article.type === item.type }]"
              @click="article.type = item.type">
              <span class="type-radio">
                <span class="type-radio-inner"></span>
              </span>
              <span class="type-label">{{ item.desc }}</span>
            </label>
          </div>
        </div>

        <!-- 原文地址 -->
        <div class="form-section" v-if="article.type !== 1">
          <label class="form-label">
            <span class="label-required">*</span>
            原文地址
          </label>
          <input
            v-model="article.originalUrl"
            type="url"
            class="form-input"
            placeholder="https://example.com/article"
          />
        </div>

        <!-- 文章摘要 -->
        <div class="form-section">
          <label class="form-label">文章摘要</label>
          <div class="textarea-wrapper">
            <textarea
              v-model="article.articleAbstract"
              class="form-textarea"
              placeholder="简要描述文章内容（选填，不填将自动截取）"
              maxlength="200"
              rows="3"
            ></textarea>
            <span class="textarea-count">{{ article.articleAbstract.length }} / 200</span>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <button class="btn-cancel" @click="addOrEdit = false">取消</button>
          <button class="btn-publish" @click="saveOrUpdateArticle">
            <el-icon><Promotion /></el-icon>
            发布文章
          </button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import request from '@/utils/request'
import Editor from '@/components/Editor.vue'

const route = useRoute()
const router = useRouter()

const publishFormRef = ref(null)

// 响应式数据
const addOrEdit = ref(false)
const categories = ref([])
const tags = ref([])
const selectedTag = ref('')
const isSaving = ref(false)
const hasUnsaved = ref(false)

const article = reactive({
  id: null,
  articleTitle: '',
  articleContent: '',
  articleAbstract: '',
  categoryId: null,
  categoryName: '',
  tagNames: [],
  type: 1,
  originalUrl: '',
  status: 3
})

const typeList = [
  { type: 1, desc: '原创', icon: 'edit' },
  { type: 2, desc: '转载', icon: 'link' },
  { type: 3, desc: '翻译', icon: 'translate' }
]

// 计算属性
const hasContent = computed(() => {
  return article.articleTitle.trim() || article.articleContent.trim()
})

const availableTags = computed(() => {
  return tags.value.filter(
    t => !article.tagNames.includes(t.tagName)
  )
})

// 打开发布对话框
const openModel = () => {
  if (!article.articleTitle.trim()) {
    ElMessage.warning('请输入文章标题')
    return
  }
  if (!article.articleContent.trim()) {
    ElMessage.warning('请输入文章内容')
    return
  }
  addOrEdit.value = true
}

// 保存草稿
const saveArticleDraft = () => {
  if (!article.articleTitle.trim()) {
    ElMessage.warning('请输入文章标题')
    return
  }
  article.status = 3
  doSave()
}

// 保存或更新文章
const saveOrUpdateArticle = () => {
  if (!article.categoryId) {
    ElMessage.warning('请选择文章分类')
    return
  }
  if (article.type !== 1 && !article.originalUrl.trim()) {
    ElMessage.warning('请填写原文链接')
    return
  }
  article.status = 1
  doSave()
}

const doSave = () => {
  if (isSaving.value) return
  isSaving.value = true

  const url = article.id ? '/admin/articles/update' : '/admin/articles/save'
  request.post(url, article).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: article.id ? '修改成功' : '发布成功'
      })
      addOrEdit.value = false
      hasUnsaved.value = false
      router.push({ path: '/articles' })
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  }).catch(() => {
    ElNotification.error({ title: '失败', message: '网络错误，请重试' })
  }).finally(() => {
    isSaving.value = false
  })
}

// 离开页面提醒
const beforeUnloadHandler = (e) => {
  if (hasUnsaved.value) {
    e.preventDefault()
    e.returnValue = ''
  }
}

// 分类选择变化
const handleCategoryChange = (id) => {
  const category = categories.value.find(item => item.id === id)
  if (category) {
    article.categoryName = category.categoryName
  }
  hasUnsaved.value = true
}

// 移除分类
const removeCategory = () => {
  article.categoryId = null
  article.categoryName = ''
  hasUnsaved.value = true
}

// 添加标签
const addTag = (tagName) => {
  if (tagName && !article.tagNames.includes(tagName)) {
    article.tagNames.push(tagName)
  }
  selectedTag.value = ''
  hasUnsaved.value = true
}

// 移除标签
const removeTag = (tag) => {
  const index = article.tagNames.indexOf(tag)
  if (index > -1) {
    article.tagNames.splice(index, 1)
  }
  hasUnsaved.value = true
}

// 获取分类列表
const listCategories = () => {
  request.get('/admin/categories/search').then(({ data }) => {
    categories.value = data.data
  })
}

// 获取标签列表
const listTags = () => {
  request.get('/admin/tags/search').then(({ data }) => {
    tags.value = data.data
  })
}

// 获取文章详情
const getArticle = () => {
  const id = route.params.id
  if (id && id !== 'undefined') {
    request.get(`/admin/articles/${id}`).then(({ data }) => {
      Object.assign(article, data.data)
    })
  }
}

// 初始化
onMounted(() => {
  listCategories()
  listTags()
  getArticle()
  window.addEventListener('beforeunload', beforeUnloadHandler)
})

onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', beforeUnloadHandler)
})
</script>

<style scoped>
/* ===== 页面容器 ===== */
.article-edit-page {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 24px 80px;
  animation: fadeIn 0.4s ease;
}

/* ===== 头部区域 ===== */
.edit-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 28px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-light);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.edit-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
  background: var(--bg-surface);
  padding: 4px 12px;
  border-radius: var(--radius-full);
}

.edit-status.saved {
  color: var(--success);
  background: var(--success-light);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-muted);
}

.edit-status.saved .status-dot {
  background: var(--success);
  box-shadow: 0 0 6px var(--success-glow);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* ===== 按钮 ===== */
.draft-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  border-radius: var(--radius-lg);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  cursor: pointer;
  transition: all var(--duration-fast) var(--ease-out);
}

.draft-btn:hover {
  color: var(--primary);
  border-color: var(--primary);
  background: var(--primary-light);
}

.publish-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 22px;
  border-radius: var(--radius-lg);
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: var(--gradient-primary);
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.3);
  transition: all var(--duration-fast) var(--ease-out);
}

.publish-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
}

.publish-btn:active {
  transform: translateY(0);
}

/* ===== 标题输入区 ===== */
.title-input-area {
  margin-bottom: 24px;
}

.title-input {
  width: 100%;
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  background: transparent;
  border: none;
  outline: none;
  padding: 0 0 8px 0;
  border-bottom: 2px solid var(--border-light);
  transition: border-color var(--duration-fast) var(--ease-out);
  font-family: var(--font-sans);
  line-height: 1.4;
}

.title-input::placeholder {
  color: var(--text-muted);
  font-weight: 400;
}

.title-input:focus {
  border-bottom-color: var(--primary);
}

.title-meta {
  display: flex;
  justify-content: flex-end;
  margin-top: 6px;
}

.char-count {
  font-size: 12px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.char-count.warning {
  color: var(--warning);
}

/* ===== 编辑器区域 ===== */
.editor-area {
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out);
  background: var(--bg-elevated);
}

.editor-area:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.editor-area :deep(.edit-container) {
  min-height: 520px;
  padding: 20px 24px;
  font-size: 16px;
  line-height: 1.8;
  background: var(--bg-elevated);
  border: none;
  border-radius: 0;
}

.editor-area :deep(.edit-container:empty::before) {
  font-size: 16px;
}

/* ===== 发布对话框 ===== */
:deep(.publish-dialog) {
  border-radius: var(--radius-xl);
  overflow: hidden;
}

:deep(.publish-dialog .el-dialog__header) {
  padding: 24px 28px 16px;
  margin: 0;
  border-bottom: 1px solid var(--border-light);
}

:deep(.publish-dialog .el-dialog__body) {
  padding: 24px 28px;
  max-height: 60vh;
  overflow-y: auto;
}

:deep(.publish-dialog .el-dialog__footer) {
  padding: 16px 28px 24px;
  border-top: 1px solid var(--border-light);
}

.dialog-header {
  display: flex;
  align-items: center;
  gap: 14px;
}

.dialog-icon {
  width: 42px;
  height: 42px;
  border-radius: var(--radius-lg);
  background: var(--primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary);
  font-size: 20px;
}

.dialog-title-group {
  flex: 1;
}

.dialog-title {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 2px;
}

.dialog-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

/* ===== 表单 ===== */
.publish-form {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 4px;
}

.label-required {
  color: var(--danger);
  font-size: 14px;
}

.form-select {
  width: 100%;
}

.tag-select-area {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.selected-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: var(--radius-full);
  font-size: 13px;
  font-weight: 500;
  transition: all var(--duration-fast) var(--ease-out);
}

.category-tag {
  background: var(--success-light);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.label-tag {
  background: var(--primary-light);
  color: var(--primary);
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.tag-remove {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: transparent;
  border: none;
  cursor: pointer;
  color: inherit;
  opacity: 0.6;
  transition: all var(--duration-fast) var(--ease-out);
  padding: 0;
  font-size: 12px;
}

.tag-remove:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.1);
}

.tag-hint {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

/* ===== 类型选择器 ===== */
.type-selector {
  display: flex;
  gap: 10px;
}

.type-option {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-default);
  background: var(--bg-surface);
  cursor: pointer;
  transition: all var(--duration-fast) var(--ease-out);
}

.type-option:hover {
  border-color: var(--primary);
  background: var(--primary-light);
}

.type-option.active {
  border-color: var(--primary);
  background: var(--primary-light);
  box-shadow: 0 0 0 2px var(--primary-glow);
}

.type-radio {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid var(--border-default);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--duration-fast) var(--ease-out);
  flex-shrink: 0;
}

.type-option.active .type-radio {
  border-color: var(--primary);
}

.type-radio-inner {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--primary);
  transform: scale(0);
  transition: transform var(--duration-fast) var(--ease-spring);
}

.type-option.active .type-radio-inner {
  transform: scale(1);
}

.type-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

/* ===== 输入框 ===== */
.form-input {
  width: 100%;
  height: 40px;
  padding: 0 14px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-default);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 14px;
  font-family: var(--font-mono);
  outline: none;
  transition: all var(--duration-fast) var(--ease-out);
}

.form-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.form-input::placeholder {
  color: var(--text-muted);
  font-family: var(--font-sans);
}

/* ===== 文本域 ===== */
.textarea-wrapper {
  position: relative;
}

.form-textarea {
  width: 100%;
  padding: 12px 14px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-default);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 14px;
  font-family: var(--font-sans);
  line-height: 1.6;
  outline: none;
  resize: vertical;
  min-height: 80px;
  transition: border-color var(--duration-fast) var(--ease-out), box-shadow var(--duration-fast) var(--ease-out);
}

.form-textarea:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.form-textarea::placeholder {
  color: var(--text-muted);
}

.textarea-count {
  position: absolute;
  bottom: 10px;
  right: 14px;
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
  pointer-events: none;
}

/* ===== 对话框底部按钮 ===== */
.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.btn-cancel {
  padding: 9px 22px;
  border-radius: var(--radius-lg);
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  cursor: pointer;
  transition: all var(--duration-fast) var(--ease-out);
}

.btn-cancel:hover {
  color: var(--text-primary);
  border-color: var(--border-default);
  background: var(--bg-hover);
}

.btn-publish {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 9px 26px;
  border-radius: var(--radius-lg);
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: var(--gradient-primary);
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
  transition: all var(--duration-fast) var(--ease-out);
}

.btn-publish:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 18px rgba(59, 130, 246, 0.4);
}

.btn-publish:active {
  transform: translateY(0);
}

/* ===== 暗色模式适配 ===== */
[data-theme="dark"] .title-input {
  color: var(--text-primary);
}

[data-theme="dark"] .tag-remove:hover {
  background: rgba(255, 255, 255, 0.15);
}

[data-theme="dark"] .form-input,
[data-theme="dark"] .form-textarea {
  background: var(--bg-deep);
  border-color: var(--border-default);
}

[data-theme="dark"] .type-option {
  background: var(--bg-deep);
  border-color: var(--border-default);
}

[data-theme="dark"] .type-option:hover {
  background: var(--bg-surface);
}

[data-theme="dark"] .type-option.active {
  background: var(--bg-surface);
}

/* ===== 响应式 ===== */
@media (max-width: 767px) {
  .article-edit-page {
    padding: 16px 12px 60px;
  }

  .edit-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
    margin-bottom: 20px;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }

  .page-title {
    font-size: 18px;
  }

  .title-input {
    font-size: 22px;
  }

  .type-selector {
    flex-direction: column;
  }

  .editor-area :deep(.edit-container) {
    min-height: 400px;
    padding: 16px;
    font-size: 15px;
  }

  :deep(.publish-dialog) {
    width: 92vw !important;
    margin: 0 auto;
  }

  :deep(.publish-dialog .el-dialog__header),
  :deep(.publish-dialog .el-dialog__body),
  :deep(.publish-dialog .el-dialog__footer) {
    padding-left: 20px;
    padding-right: 20px;
  }
}

@media (max-width: 1023px) and (min-width: 768px) {
  .article-edit-page {
    padding: 24px 20px 60px;
  }
}
</style>

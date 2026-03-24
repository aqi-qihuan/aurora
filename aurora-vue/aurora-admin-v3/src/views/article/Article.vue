<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="article-title-container">
      <el-input v-model="article.articleTitle" placeholder="输入文章标题" />
      <el-button
        type="danger"
        class="save-btn"
        @click="saveArticleDraft"
        v-if="article.id == null || article.status == 3">
        保存草稿
      </el-button>
      <el-button type="danger" @click="openModel" style="margin-left: 10px"> 发布文章 </el-button>
    </div>
    <MdEditor
      ref="mdRef"
      v-model="article.articleContent"
      @onUploadImg="uploadImg"
      style="height: calc(100vh - 260px)"
    />
    <el-dialog v-model="addOrEdit" width="40%" top="3vh" custom-class="elegant-dialog">
      <div class="dialog-title-container">发布文章</div>
      <el-form label-width="80px" :model="article">
        <el-form-item label="文章分类">
          <el-tag
            type="success"
            v-show="article.categoryName"
            style="margin: 0 1rem 0 0"
            :closable="true"
            @close="removeCategory">
            {{ article.categoryName }}
          </el-tag>
          <el-popover placement="bottom-start" :width="460" trigger="click" v-if="!article.categoryName">
            <div class="popover-title">分类</div>
            <el-autocomplete
              style="width: 100%"
              v-model="categoryName"
              :fetch-suggestions="searchCategories"
              placeholder="请输入分类名搜索，enter可添加自定义分类"
              :trigger-on-focus="false"
              @keyup.enter="saveCategory"
              @select="handleSelectCategories">
              <template #default="{ item }">
                <div>{{ item.categoryName }}</div>
              </template>
            </el-autocomplete>
            <div class="popover-container">
              <div v-for="item of categorys" :key="item.id" class="category-item" @click="addCategory(item)">
                {{ item.categoryName }}
              </div>
            </div>
            <template #reference>
              <el-button type="success" plain size="small"> 添加分类 </el-button>
            </template>
          </el-popover>
        </el-form-item>
        <el-form-item label="文章标签">
          <el-tag
            v-for="(item, index) of article.tagNames"
            :key="index"
            style="margin: 0 1rem 0 0"
            :closable="true"
            @close="removeTag(item)">
            {{ item }}
          </el-tag>
          <el-popover placement="bottom-start" :width="460" trigger="click" v-if="article.tagNames.length < 3">
            <div class="popover-title">标签</div>
            <el-autocomplete
              style="width: 100%"
              v-model="tagName"
              :fetch-suggestions="searchTags"
              placeholder="请输入标签名搜索，enter可添加自定义标签"
              :trigger-on-focus="false"
              @keyup.enter="saveTag"
              @select="handleSelectTag">
              <template #default="{ item }">
                <div>{{ item.tagName }}</div>
              </template>
            </el-autocomplete>
            <div class="popover-container">
              <div style="margin-bottom: 1rem">添加标签</div>
              <el-tag v-for="(item, index) in tagList" :key="index" :class="tagClass(item)" @click="addTag(item)">
                {{ item.tagName }}
              </el-tag>
            </div>
            <template #reference>
              <el-button type="primary" plain size="small"> 添加标签 </el-button>
            </template>
          </el-popover>
        </el-form-item>
        <el-form-item label="文章类型">
          <el-select v-model="article.type" placeholder="请选择类型">
            <el-option v-for="item in typeList" :key="item.type" :label="item.desc" :value="item.type" />
          </el-select>
        </el-form-item>
        <el-form-item label="原文地址" v-if="article.type != 1">
          <el-input v-model="article.originalUrl" placeholder="请填写原文链接" />
        </el-form-item>
        <el-form-item label="上传封面">
          <el-upload
            class="upload-cover"
            drag
            action="/api/admin/articles/images"
            multiple
            :headers="headers"
            :before-upload="beforeUpload"
            :on-success="uploadCover">
            <el-icon class="el-icon--upload" v-if="article.articleCover == ''"><Upload /></el-icon>
            <div class="el-upload__text" v-if="article.articleCover == ''">将文件拖到此处，或<em>点击上传</em></div>
            <img v-else :src="article.articleCover" width="360px" height="180px" />
          </el-upload>
        </el-form-item>
        <el-form-item label="置顶">
          <el-switch
            v-model="article.isTop"
            active-color="#13ce66"
            inactive-color="#F4F4F5"
            :active-value="1"
            :inactive-value="0" />
        </el-form-item>
        <el-form-item label="推荐">
          <el-switch
            v-model="article.isFeatured"
            active-color="#13ce66"
            inactive-color="#F4F4F5"
            :active-value="1"
            :inactive-value="0" />
        </el-form-item>
        <el-form-item label="发布形式">
          <el-radio-group v-model="article.status">
            <el-radio :value="1">公开</el-radio>
            <el-radio :value="2">密码</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="访问密码" v-if="article.status == 2">
          <el-input v-model="article.password" placeholder="请填写文章访问密码" />
        </el-form-item>
        <el-form-item label="文章摘要">
          <el-input type="textarea" :autosize="true" v-model="article.articleAbstract" placeholder="默认取文章前500个字符" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="danger" @click="saveOrUpdateArticle"> 发 表 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { ElMessage, ElNotification } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import * as imageConversion from 'image-conversion'
import dayjs from 'dayjs'
import request from '@/utils/request'
import logger from '@/utils/logger'
import { createBeforeUploadHandler } from '@/utils/imageUtils'
import { getAuthHeaders } from '@/utils/auth'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

// 常量配置
const UPLOAD_SIZE = 500 // KB

// 响应式数据
const addOrEdit = ref(false)
const autoSave = ref(true)
const categoryName = ref('')
const tagName = ref('')
const categorys = ref([])
const tagList = ref([])
const mdRef = ref(null)

const typeList = ref([
  { type: 1, desc: '原创' },
  { type: 2, desc: '转载' },
  { type: 3, desc: '翻译' }
])

const article = reactive({
  id: null,
  articleTitle: dayjs(new Date()).format('YYYY-MM-DD'),
  articleContent: '',
  articleAbstract: '',
  articleCover: '',
  categoryName: null,
  tagNames: [],
  isTop: 0,
  isFeatured: 0,
  type: 1,
  status: 1,
  originalUrl: '',
  password: ''
})

const headers = ref(getAuthHeaders())

// 分类相关方法
const listCategories = async () => {
  try {
    const { data } = await request.get('/admin/categories/search')
    categorys.value = data.data || []
  } catch (error) {
    logger.error('获取分类列表失败:', error)
  }
}

const searchCategories = async (keywords, cb) => {
  try {
    const { data } = await request.get('/admin/categories/search', {
      params: { keywords }
    })
    cb(data.data || [])
  } catch (error) {
    cb([])
  }
}

const handleSelectCategories = (item) => {
  addCategory({ categoryName: item.categoryName })
}

const saveCategory = () => {
  if (categoryName.value.trim() !== '') {
    addCategory({ categoryName: categoryName.value })
    categoryName.value = ''
  }
}

const addCategory = (item) => {
  article.categoryName = item.categoryName
}

const removeCategory = () => {
  article.categoryName = null
}

// 标签相关方法
const listTags = async () => {
  try {
    const { data } = await request.get('/admin/tags/search')
    tagList.value = data.data || []
  } catch (error) {
    logger.error('获取标签列表失败:', error)
  }
}

const searchTags = async (keywords, cb) => {
  try {
    const { data } = await request.get('/admin/tags/search', {
      params: { keywords }
    })
    cb(data.data || [])
  } catch (error) {
    cb([])
  }
}

const handleSelectTag = (item) => {
  addTag({ tagName: item.tagName })
}

const saveTag = () => {
  if (tagName.value.trim() !== '') {
    addTag({ tagName: tagName.value })
    tagName.value = ''
  }
}

const addTag = (item) => {
  if (article.tagNames.indexOf(item.tagName) === -1) {
    article.tagNames.push(item.tagName)
  }
}

const removeTag = (item) => {
  const index = article.tagNames.indexOf(item)
  article.tagNames.splice(index, 1)
}

const tagClass = (item) => {
  const index = article.tagNames.indexOf(item.tagName)
  return index !== -1 ? 'tag-item-select' : 'tag-item'
}

// 上传相关方法
const uploadCover = (response) => {
  if (response.data) {
    article.articleCover = response.data
  }
}

const beforeUpload = createBeforeUploadHandler(UPLOAD_SIZE)

const uploadImg = async (files, callback) => {
  const file = files[0]
  const formdata = new FormData()
  
  try {
    if (file.size / 1024 < UPLOAD_SIZE) {
      formdata.append('file', file)
      const { data } = await request.post('/admin/articles/images', formdata)
      callback([data.data])
    } else {
      const res = await imageConversion.compressAccurately(file, UPLOAD_SIZE)
      formdata.append('file', new File([res], file.name, { type: file.type }))
      const { data } = await request.post('/admin/articles/images', formdata)
      callback([data.data])
    }
  } catch (error) {
    ElMessage.error('图片上传失败')
    callback([])
  }
}

// 文章保存相关方法
const openModel = () => {
  if (!article.articleTitle?.trim()) {
    ElMessage.error('文章标题不能为空')
    return false
  }
  if (!article.articleContent?.trim()) {
    ElMessage.error('文章内容不能为空')
    return false
  }
  listCategories()
  listTags()
  addOrEdit.value = true
}

const saveArticleDraft = async () => {
  if (!article.articleTitle?.trim()) {
    ElMessage.error('文章标题不能为空')
    return false
  }
  if (!article.articleContent?.trim()) {
    ElMessage.error('文章内容不能为空')
    return false
  }
  
  article.status = 3
  try {
    const { data } = await request.post('/admin/articles', article)
    if (data.flag) {
      if (article.id === null) {
        appStore.removeTab({ path: '/article/publish', name: '发布文章' })
      } else {
        appStore.removeTab({ path: `/article/${article.id}`, name: '修改文章' })
      }
      sessionStorage.removeItem('article')
      router.push({ path: '/article-list' })
      ElNotification.success({
        title: '成功',
        message: '保存草稿成功'
      })
    } else {
      ElNotification.error({
        title: '失败',
        message: '保存草稿失败'
      })
    }
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: '保存草稿失败'
    })
  }
  autoSave.value = false
}

const saveOrUpdateArticle = async () => {
  if (!article.articleTitle?.trim()) {
    ElMessage.error('文章标题不能为空')
    return false
  }
  if (!article.articleContent?.trim()) {
    ElMessage.error('文章内容不能为空')
    return false
  }
  if (!article.categoryName) {
    ElMessage.error('文章分类不能为空')
    return false
  }
  if (article.tagNames.length === 0) {
    ElMessage.error('文章标签不能为空')
    return false
  }
  if (!article.articleCover?.trim()) {
    ElMessage.error('文章封面不能为空')
    return false
  }

  try {
    const { data } = await request.post('/admin/articles', article)
    if (data.flag) {
      if (article.id === null) {
        appStore.removeTab({ path: '/article/publish', name: '发布文章' })
      } else {
        appStore.removeTab({ path: `/article/${article.id}`, name: '修改文章' })
      }
      sessionStorage.removeItem('article')
      router.push({ path: '/article-list' })
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
    addOrEdit.value = false
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: '发表文章失败'
    })
  }
  autoSave.value = false
}

const autoSaveArticle = async () => {
  if (
    autoSave.value &&
    article.articleTitle?.trim() &&
    article.articleContent?.trim() &&
    article.id != null
  ) {
    try {
      const { data } = await request.post('/admin/articles', article)
      if (data.flag) {
        ElNotification.success({
          title: '成功',
          message: '自动保存成功'
        })
      } else {
        ElNotification.error({
          title: '失败',
          message: '自动保存失败'
        })
      }
    } catch (error) {
      logger.error('自动保存失败:', error)
    }
  }
  if (autoSave.value && article.id == null) {
    sessionStorage.setItem('article', JSON.stringify(article))
  }
}

// 生命周期钩子
onMounted(() => {
  const path = route.path
  const arr = path.split('/')
  const articleId = arr[2]
  
  if (articleId) {
    request.get('/admin/articles/' + articleId).then(({ data }) => {
      if (data.data) {
        Object.assign(article, data.data)
      }
    })
  } else {
    const savedArticle = sessionStorage.getItem('article')
    if (savedArticle) {
      try {
        const parsedArticle = JSON.parse(savedArticle)
        Object.assign(article, parsedArticle)
      } catch (error) {
        logger.error('解析保存的文章失败:', error)
      }
    }
  }
})

onUnmounted(() => {
  autoSaveArticle()
})
</script>

<style scoped>
.article-title-container {
  display: flex;
  align-items: center;
  margin-bottom: 1.25rem;
  margin-top: 2.25rem;
}
.save-btn {
  margin-left: 0.75rem;
  background: var(--bg-elevated, #fff);
  color: var(--danger, #f56c6c);
}
.tag-item {
  margin-right: 1rem;
  margin-bottom: 1rem;
  cursor: pointer;
}
.tag-item-select {
  margin-right: 1rem;
  margin-bottom: 1rem;
  cursor: not-allowed;
  color: #ccccd8 !important;
}
.category-item {
  cursor: pointer;
  padding: 0.6rem 0.5rem;
}
.category-item:hover {
  background-color: #f0f9eb;
  color: #67c23a;
}
.popover-title {
  margin-bottom: 1rem;
  text-align: center;
}
.popover-container {
  margin-top: 1rem;
  height: 260px;
  overflow-y: auto;
}
.dialog-title-container {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 20px;
}

/* 深色模式 - 极客风 */
[data-theme="dark"] .save-btn {
  background: var(--bg-elevated, #374151);
  color: #F87171;
  border-color: rgba(239, 68, 68, 0.3);
}
[data-theme="dark"] .save-btn:hover {
  background: rgba(239, 68, 68, 0.15);
  box-shadow: 0 0 10px rgba(239, 68, 68, 0.3);
}
[data-theme="dark"] .category-item:hover {
  background-color: rgba(0, 255, 136, 0.08);
  color: var(--neon-green, #00FF88);
}
[data-theme="dark"] .popover-title {
  color: var(--neon-blue, #00D4FF);
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
}
</style>

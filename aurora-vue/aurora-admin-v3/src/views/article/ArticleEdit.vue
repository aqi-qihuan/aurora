<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="article-title-container">
      <el-input v-model="article.articleTitle" size="large" placeholder="输入文章标题" />
      <el-button
        type="info"
        size="large"
        class="save-btn"
        @click="saveArticleDraft"
        v-if="!article.id || article.status === 3">
        保存草稿
      </el-button>
      <el-button type="primary" size="large" @click="openModel" style="margin-left: 10px"> 发布文章 </el-button>
    </div>
    
    <!-- 这里暂时使用 textarea 替代 mavon-editor -->
    <el-input
      v-model="article.articleContent"
      type="textarea"
      :rows="20"
      placeholder="请输入文章内容 (Markdown 格式)"
      style="margin-top: 20px; font-family: monospace;"
    />
    
    <el-dialog v-model="addOrEdit" width="40%" top="3vh">
      <template #header>
        <div class="dialog-title-container">发布文章</div>
      </template>
      <el-form label-width="80px" size="medium" :model="article">
        <el-form-item label="文章分类">
          <el-tag
            type="success"
            v-show="article.categoryName"
            style="margin: 0 1rem 0 0"
            closable
            @close="removeCategory">
            {{ article.categoryName }}
          </el-tag>
          <el-select
            v-if="!article.categoryName"
            v-model="article.categoryId"
            placeholder="请选择分类"
            style="width: 200px"
            @change="handleCategoryChange">
            <el-option
              v-for="item in categories"
              :key="item.id"
              :label="item.categoryName"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文章标签">
          <el-tag
            v-for="(item, index) in article.tagNames"
            :key="index"
            style="margin: 0 1rem 0 0"
            closable
            @close="removeTag(item)">
            {{ item }}
          </el-tag>
          <el-select
            v-if="article.tagNames.length < 3"
            v-model="selectedTag"
            placeholder="请选择标签"
            style="width: 200px"
            @change="addTag">
            <el-option
              v-for="item in tags"
              :key="item.id"
              :label="item.tagName"
              :value="item.tagName"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文章类型">
          <el-select v-model="article.type" placeholder="请选择类型">
            <el-option
              v-for="item in typeList"
              :key="item.type"
              :label="item.desc"
              :value="item.type"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="原文地址" v-if="article.type !== 1">
          <el-input v-model="article.originalUrl" placeholder="请填写原文链接" />
        </el-form-item>
        <el-form-item label="文章摘要">
          <el-input
            v-model="article.articleAbstract"
            type="textarea"
            :rows="3"
            placeholder="请输入文章摘要"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateArticle"> 发 布 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()

// 响应式数据
const addOrEdit = ref(false)
const categories = ref([])
const tags = ref([])
const selectedTag = ref('')

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
  { type: 1, desc: '原创' },
  { type: 2, desc: '转载' },
  { type: 3, desc: '翻译' }
]

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
  article.status = 3
  saveOrUpdateArticle()
}

// 保存或更新文章
const saveOrUpdateArticle = () => {
  if (!article.categoryId) {
    ElMessage.warning('请选择文章分类')
    return
  }
  
  const url = article.id ? '/admin/articles/update' : '/admin/articles/save'
  request.post(url, article).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: article.id ? '修改成功' : '发布成功'
      })
      addOrEdit.value = false
      router.push({ path: '/articles' })
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  })
}

// 分类选择变化
const handleCategoryChange = (id) => {
  const category = categories.value.find(item => item.id === id)
  if (category) {
    article.categoryName = category.categoryName
  }
}

// 移除分类
const removeCategory = () => {
  article.categoryId = null
  article.categoryName = ''
}

// 添加标签
const addTag = (tagName) => {
  if (tagName && !article.tagNames.includes(tagName)) {
    article.tagNames.push(tagName)
  }
  selectedTag.value = ''
}

// 移除标签
const removeTag = (tag) => {
  const index = article.tagNames.indexOf(tag)
  if (index > -1) {
    article.tagNames.splice(index, 1)
  }
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
})
</script>

<style scoped>
.title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e4e7ed;
}

.article-title-container {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.article-title-container .el-input {
  flex: 1;
}

.save-btn {
  margin-left: 20px;
}

.dialog-title-container {
  font-size: 16px;
  font-weight: 600;
}
</style>

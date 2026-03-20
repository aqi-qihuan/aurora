<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="article-status-menu">
      <span>状态</span>
      <span @click="changeStatus('all')" :class="isActive('all')">全部</span>
      <span @click="changeStatus('public')" :class="isActive('public')"> 公开 </span>
      <span @click="changeStatus('private')" :class="isActive('private')"> 私密 </span>
      <span @click="changeStatus('draft')" :class="isActive('draft')"> 草稿箱 </span>
      <span @click="changeStatus('delete')" :class="isActive('delete')"> 回收站 </span>
    </div>
    <div class="operation-container">
      <el-button
        v-if="isDelete === 0"
        type="danger"
        size="small"
        :icon="Delete"
        :disabled="articleIds.length === 0"
        @click="updateIsDeleteDialog = true">
        批量删除
      </el-button>
      <el-button
        v-else
        type="danger"
        size="small"
        :icon="Delete"
        :disabled="articleIds.length === 0"
        @click="removeDialog = true">
        批量删除
      </el-button>
      <el-button
        type="success"
        size="small"
        :icon="Download"
        :disabled="articleIds.length === 0"
        style="margin-right: 1rem"
        @click="exportDialog = true">
        批量导出
      </el-button>
      <el-upload
        action="/api/admin/articles/import"
        multiple
        :limit="9"
        :show-file-list="false"
        :headers="uploadHeaders"
        :on-success="uploadArticle">
        <el-button type="primary" size="small" :icon="Upload"> 批量导入 </el-button>
      </el-upload>
      <div style="margin-left: auto">
        <el-select
          clearable
          v-model="type"
          placeholder="请选择文章类型"
          size="small"
          style="margin-right: 1rem; width: 180px">
          <el-option label="全部" value="" />
          <el-option v-for="item in types" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select
          clearable
          size="small"
          v-model="categoryId"
          filterable
          placeholder="请选择分类"
          style="margin-right: 1rem; width: 180px">
          <el-option label="全部" value="" />
          <el-option v-for="item in categories" :key="item.id" :label="item.categoryName" :value="item.id" />
        </el-select>
        <el-select
          clearable
          size="small"
          v-model="tagId"
          filterable
          placeholder="请选择标签"
          style="margin-right: 1rem; width: 180px">
          <el-option label="全部" value="" />
          <el-option v-for="item in tags" :key="item.id" :label="item.tagName" :value="item.id" />
        </el-select>
        <el-input
          clearable
          v-model="keywords"
          :prefix-icon="Search"
          size="small"
          placeholder="请输入文章名"
          style="width: 200px"
          @keyup.enter="searchArticles" />
        <el-button type="primary" size="small" :icon="Search" style="margin-left: 1rem" @click="searchArticles">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="articles"
      @selection-change="selectionChange"
      v-loading="loading"
      class="article-table"
      :header-cell-style="{ background: 'var(--bg-elevated)', color: 'var(--text-primary)', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="articleCover" label="文章封面" width="200" align="center">
        <template #default="scope">
          <div class="article-cover-wrapper">
            <el-image
              class="article-cover"
              :src="scope.row.articleCover || 'https://static.talkxj.com/articles/c5cc2b2561bd0e3060a500198a4ad37d.png'"
              :preview-src-list="[scope.row.articleCover]"
              fit="cover" />
            <div class="article-status-badge">
              <el-tag
                v-if="scope.row.status === 1"
                size="small"
                type="success"
                effect="dark">
                <el-icon><View /></el-icon> 公开
              </el-tag>
              <el-tag
                v-if="scope.row.status === 2"
                size="small"
                type="info"
                effect="dark">
                <el-icon><Lock /></el-icon> 私密
              </el-tag>
              <el-tag
                v-if="scope.row.status === 3"
                size="small"
                type="warning"
                effect="dark">
                <el-icon><Document /></el-icon> 草稿
              </el-tag>
            </div>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="articleTitle" label="标题" align="center" />
      <el-table-column prop="categoryName" label="分类" width="110" align="center" />
      <el-table-column prop="tagDTOs" label="标签" width="170" align="center">
        <template #default="scope">
          <el-tag v-for="item in scope.row.tagDTOs" :key="item.tagId" style="margin-right: 0.2rem; margin-top: 0.2rem">
            {{ item.tagName }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="viewsCount" label="浏览量" width="70" align="center">
        <template #default="scope">
          <span>{{ scope.row.viewsCount || 0 }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="类型" width="80" align="center">
        <template #default="scope">
          <el-tag :type="articleType(scope.row.type).tagType">
            {{ articleType(scope.row.type).name }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="发表时间" width="130" align="center">
        <template #default="scope">
          <el-icon><Clock /></el-icon>
          {{ formatDate(scope.row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column prop="isTop" label="置顶" width="80" align="center">
        <template #default="scope">
          <el-switch
            v-model="scope.row.isTop"
            :active-value="1"
            :inactive-value="0"
            :disabled="scope.row.isDelete === 1"
            @change="changeTopAndFeatured(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column prop="isFeatured" label="推荐" width="80" align="center">
        <template #default="scope">
          <el-switch
            v-model="scope.row.isFeatured"
            :active-value="1"
            :inactive-value="0"
            :disabled="scope.row.isDelete === 1"
            @change="changeTopAndFeatured(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template #default="scope">
          <el-button type="primary" size="small" @click="editArticle(scope.row.id)" v-if="scope.row.isDelete === 0">
            编辑
          </el-button>
          <el-popconfirm
            title="确定删除吗？"
            style="margin-left: 10px"
            @confirm="updateArticleDelete(scope.row.id)"
            v-if="scope.row.isDelete === 0">
            <template #reference>
              <el-button size="small" type="danger"> 删除 </el-button>
            </template>
          </el-popconfirm>
          <el-popconfirm
            title="确定恢复吗？"
            v-if="scope.row.isDelete === 1"
            @confirm="updateArticleDelete(scope.row.id)">
            <template #reference>
              <el-button size="small" type="success"> 恢复 </el-button>
            </template>
          </el-popconfirm>
          <el-popconfirm
            style="margin-left: 10px"
            v-if="scope.row.isDelete === 1"
            title="确定彻底删除吗？"
            @confirm="deleteArticles(scope.row.id)">
            <template #reference>
              <el-button size="small" type="danger"> 删除 </el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      class="pagination-container"
      background
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      :page-sizes="[10, 20]"
      layout="total, sizes, prev, pager, next, jumper" />
    
    <!-- 批量删除对话框 -->
    <el-dialog v-model="updateIsDeleteDialog" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <template #footer>
        <el-button @click="updateIsDeleteDialog = false">取 消</el-button>
        <el-button type="primary" @click="updateArticleDelete(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
    
    <!-- 彻底删除对话框 -->
    <el-dialog v-model="removeDialog" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否彻底删除选中项？</div>
      <template #footer>
        <el-button @click="removeDialog = false">取 消</el-button>
        <el-button type="primary" @click="deleteArticles(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
    
    <!-- 导出对话框 -->
    <el-dialog v-model="exportDialog" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否导出选中文章？</div>
      <template #footer>
        <el-button @click="exportDialog = false">取 消</el-button>
        <el-button type="primary" @click="exportArticles(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { usePageStateStore } from '@/stores/pageState'
import request from '@/utils/request'
import dayjs from 'dayjs'
import {
  Delete,
  Download,
  Upload,
  Search,
  View,
  Lock,
  Document,
  Clock,
  Warning
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const pageStateStore = usePageStateStore()

// 响应式数据
const loading = ref(true)
const updateIsDeleteDialog = ref(false)
const removeDialog = ref(false)
const exportDialog = ref(false)
const activeStatus = ref('all')
const articles = ref([])
const articleIds = ref([])
const categories = ref([])
const tags = ref([])
const keywords = ref('')
const type = ref('')
const categoryId = ref('')
const tagId = ref('')
const isDelete = ref(0)
const status = ref(null)
const current = ref(1)
const size = ref(10)
const count = ref(0)

const uploadHeaders = {
  Authorization: 'Bearer ' + sessionStorage.getItem('token')
}

const types = [
  { value: 1, label: '原创' },
  { value: 2, label: '转载' },
  { value: 3, label: '翻译' }
]

// 格式化日期
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD') : ''
}

// 文章类型
const articleType = (type) => {
  const map = {
    1: { tagType: 'danger', name: '原创' },
    2: { tagType: 'success', name: '转载' },
    3: { tagType: 'primary', name: '翻译' }
  }
  return map[type] || { tagType: '', name: '' }
}

// 状态激活样式
const isActive = (s) => {
  return activeStatus.value === s ? 'active-status' : 'status'
}

// 选择变化
const selectionChange = (selection) => {
  articleIds.value = selection.map(item => item.id)
}

// 搜索文章
const searchArticles = () => {
  current.value = 1
  listArticles()
}

// 编辑文章
const editArticle = (id) => {
  router.push({ path: `/articles/${id}` })
}

// 更新删除状态
const updateArticleDelete = (id) => {
  const param = {
    ids: id ? [id] : articleIds.value,
    isDelete: isDelete.value === 0 ? 1 : 0
  }
  request.put('/admin/articles', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listArticles()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    updateIsDeleteDialog.value = false
  })
}

// 彻底删除
const deleteArticles = (id) => {
  const param = { data: id ? [id] : articleIds.value }
  request.delete('/admin/articles/delete', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      listArticles()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    removeDialog.value = false
  }).catch(error => {
    ElMessage.error('删除文章失败')
    console.error('API Error:', error)
  })
}

// 导出文章
const exportArticles = (id) => {
  const param = id ? [id] : articleIds.value
  request.post('/admin/articles/export', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
      data.data.forEach((item) => {
        downloadFile(item)
      })
      listArticles()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
    exportDialog.value = false
  }).catch(error => {
    ElMessage.error('导出文章失败')
    console.error('API Error:', error)
  })
}

// 下载文件
const downloadFile = (url) => {
  const iframe = document.createElement('iframe')
  iframe.style.display = 'none'
  iframe.style.height = 0
  iframe.src = url
  document.body.appendChild(iframe)
  setTimeout(() => {
    iframe.remove()
  }, 30000)
}

// 上传文章
const uploadArticle = (data) => {
  if (data.flag) {
    ElNotification.success({ title: '成功', message: '导入成功' })
    listArticles()
  } else {
    ElNotification.error({ title: '失败', message: data.message })
  }
}

// 分页大小变化
const sizeChange = (newSize) => {
  size.value = newSize
  listArticles()
}

// 页码变化
const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('articleList', newCurrent)
  listArticles()
}

// 改变状态
const changeStatus = (s) => {
  switch (s) {
    case 'all':
      isDelete.value = 0
      status.value = null
      break
    case 'public':
      isDelete.value = 0
      status.value = 1
      break
    case 'private':
      isDelete.value = 0
      status.value = 2
      break
    case 'draft':
      isDelete.value = 0
      status.value = 3
      break
    case 'delete':
      isDelete.value = 1
      status.value = null
      break
  }
  current.value = 1
  activeStatus.value = s
}

// 修改置顶和推荐
const changeTopAndFeatured = (article) => {
  request.put('/admin/articles/topAndFeatured', {
    id: article.id,
    isTop: article.isTop,
    isFeatured: article.isFeatured
  }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '修改成功' })
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  })
}

// 获取文章列表
const listArticles = () => {
  loading.value = true
  request.get('/admin/articles', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value,
      categoryId: categoryId.value,
      status: status.value,
      tagId: tagId.value,
      type: type.value,
      isDelete: isDelete.value
    }
  }).then(({ data }) => {
    articles.value = data.data.records
    count.value = data.data.count
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取文章列表失败')
    console.error('API Error:', error)
  })
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
    tags.value = data?.data || []
  })
}

// 监听筛选条件变化
watch([type, categoryId, tagId, status, isDelete], () => {
  current.value = 1
  listArticles()
})

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.articleList
  listArticles()
  listCategories()
  listTags()
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

.operation-container {
  margin-top: 1.5rem;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.article-status-menu {
  font-size: 14px;
  margin-top: 20px;
  color: #999;
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 2px solid #f0f0f0;
}

.article-status-menu span {
  margin-right: 24px;
  padding: 8px 16px;
  border-radius: 20px;
  transition: all 0.3s ease;
  position: relative;
}

.status {
  cursor: pointer;
  color: #999;
}

.status:hover {
  color: #409eff;
  background: rgba(64, 158, 255, 0.05);
}

.active-status {
  cursor: pointer;
  color: #409eff;
  font-weight: bold;
  background: rgba(64, 158, 255, 0.1);
}

.active-status::after {
  content: '';
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 30px;
  height: 3px;
  background: #409eff;
  border-radius: 2px;
}

.article-table {
  margin-top: 20px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.article-table :deep(.el-table__body tr:hover > td) {
  background-color: #f5f7fa !important;
}

.article-cover-wrapper {
  position: relative;
  width: 100%;
  height: 120px;
  overflow: hidden;
  border-radius: 6px;
}

.article-cover {
  width: 100%;
  height: 100%;
  border-radius: 6px;
  transition: transform 0.3s ease;
  cursor: pointer;
}

.article-cover:hover {
  transform: scale(1.05);
}

.article-status-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  z-index: 10;
}

.article-status-badge :deep(.el-tag) {
  border-radius: 12px;
  font-size: 12px;
  padding: 4px 10px;
}

.pagination-container {
  float: right;
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
}

.dialog-title-container {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}
</style>

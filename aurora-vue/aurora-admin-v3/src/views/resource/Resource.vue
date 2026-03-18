<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || '资源管理' }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" @click="openModel(null)">
        <el-icon><Plus /></el-icon>
        新增模块
      </el-button>
      <div class="search-container">
        <el-input
          v-model="keywords"
          size="small"
          placeholder="请输入资源名"
          style="width: 200px"
          @keyup.enter="listResources">
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button type="primary" size="small" style="margin-left: 1rem" @click="listResources">
          <el-icon><Search /></el-icon>
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      v-loading="loading"
      :data="resources"
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      class="resource-table">
      <el-table-column prop="resourceName" label="资源名" width="220" />
      <el-table-column prop="url" label="资源路径" width="300" />
      <el-table-column prop="requestMethod" label="请求方式">
        <template #default="scope">
          <el-tag v-if="scope.row.requestMethod" :type="tagType(scope.row.requestMethod)">
            {{ scope.row.requestMethod }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="isAnonymous" label="匿名访问" align="center">
        <template #default="scope">
          <el-switch
            v-if="scope.row.url"
            v-model="scope.row.isAnonymous"
            active-color="#13ce66"
            inactive-color="#F4F4F5"
            :active-value="1"
            :inactive-value="0"
            @change="changeResource(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center">
        <template #default="scope">
          <el-icon style="margin-right: 5px"><Clock /></el-icon>
          {{ formatDate(scope.row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="200">
        <template #default="scope">
          <el-button
            type="primary"
            link
            size="small"
            @click="openAddResourceModel(scope.row)"
            v-if="scope.row.children">
            <el-icon><Plus /></el-icon> 新增
          </el-button>
          <el-button type="primary" link size="small" @click="openEditResourceModel(scope.row)">
            <el-icon><Edit /></el-icon> 修改
          </el-button>
          <el-popconfirm
            title="确定删除吗？"
            style="margin-left: 10px"
            @confirm="deleteResource(scope.row.id)">
            <template #reference>
              <el-button type="danger" link size="small">
                <el-icon><Delete /></el-icon> 删除
              </el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="addModule" width="30%">
      <template #header>
        <div class="dialog-title-container" ref="moduleTitleRef" />
      </template>
      <el-form label-width="80px" size="medium" :model="resourceForm">
        <el-form-item label="模块名">
          <el-input v-model="resourceForm.resourceName" style="width: 220px" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="addModule = false">取 消</el-button>
          <el-button type="primary" @click="addOrEditResource">确 定</el-button>
        </span>
      </template>
    </el-dialog>
    <el-dialog v-model="addResource" width="30%">
      <template #header>
        <div class="dialog-title-container" ref="resourceTitleRef" />
      </template>
      <el-form label-width="80px" size="medium" :model="resourceForm">
        <el-form-item label="资源名">
          <el-input v-model="resourceForm.resourceName" style="width: 220px" />
        </el-form-item>
        <el-form-item label="资源路径">
          <el-input v-model="resourceForm.url" style="width: 220px" />
        </el-form-item>
        <el-form-item label="请求方式">
          <el-radio-group v-model="resourceForm.requestMethod">
            <el-radio label="GET">GET</el-radio>
            <el-radio label="POST">POST</el-radio>
            <el-radio label="PUT">PUT</el-radio>
            <el-radio label="DELETE">DELETE</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="addResource = false">取 消</el-button>
          <el-button type="primary" @click="addOrEditResource">确 定</el-button>
        </span>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import {
  Plus,
  Search,
  Clock,
  Edit,
  Delete
} from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()

// 响应式数据
const loading = ref(true)
const keywords = ref('')
const resources = ref([])
const addModule = ref(false)
const addResource = ref(false)
const resourceForm = ref({})
const moduleTitleRef = ref(null)
const resourceTitleRef = ref(null)

// 日期格式化
const formatDate = (date) => {
  if (!date) return ''
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hour = String(d.getHours()).padStart(2, '0')
  const minute = String(d.getMinutes()).padStart(2, '0')
  const second = String(d.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`
}

// 标签类型
const tagType = (type) => {
  switch (type) {
    case 'GET':
      return ''
    case 'POST':
      return 'success'
    case 'PUT':
      return 'warning'
    case 'DELETE':
      return 'danger'
    default:
      return ''
  }
}

// 获取资源列表
const listResources = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin/resources', {
      params: {
        keywords: keywords.value
      }
    })
    resources.value = data.data
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '获取资源列表失败'
    })
  } finally {
    loading.value = false
  }
}

// 修改资源
const changeResource = async (resource) => {
  try {
    const { data } = await request.post('/admin/resources', resource)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listResources()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '修改资源失败'
    })
  }
}

// 打开模块对话框
const openModel = (resource) => {
  if (resource != null) {
    resourceForm.value = JSON.parse(JSON.stringify(resource))
    if (moduleTitleRef.value) {
      moduleTitleRef.value.innerHTML = '修改模块'
    }
  } else {
    resourceForm.value = {}
    if (moduleTitleRef.value) {
      moduleTitleRef.value.innerHTML = '添加模块'
    }
  }
  addModule.value = true
}

// 打开编辑资源对话框
const openEditResourceModel = (resource) => {
  if (resource.url == null) {
    openModel(resource)
    return false
  }
  resourceForm.value = JSON.parse(JSON.stringify(resource))
  nextTick(() => {
    if (resourceTitleRef.value) {
      resourceTitleRef.value.innerHTML = '修改资源'
    }
  })
  addResource.value = true
}

// 打开添加资源对话框
const openAddResourceModel = (resource) => {
  resourceForm.value = {}
  resourceForm.value.parentId = resource.id
  nextTick(() => {
    if (resourceTitleRef.value) {
      resourceTitleRef.value.innerHTML = '添加资源'
    }
  })
  addResource.value = true
}

// 删除资源
const deleteResource = async (id) => {
  try {
    const { data } = await request.delete(`/api/admin/resources/${id}`)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listResources()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
  } catch (error) {
    ElNotification.error({
      title: '失败',
      message: error.message || '删除资源失败'
    })
  }
}

// 添加或修改资源
const addOrEditResource = async () => {
  if (!resourceForm.value.resourceName || resourceForm.value.resourceName.trim() === '') {
    ElMessage.error('资源名不能为空')
    return false
  }
  
  try {
    const { data } = await request.post('/admin/resources', resourceForm.value)
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listResources()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    addModule.value = false
    addResource.value = false
  } catch (error) {
    ElMessage.error('保存资源失败')
    console.error('API Error:', error)
  }
}

// 初始化
onMounted(() => {
  listResources()
})
</script>

<style scoped>
/* ==================== Resource Page Modern Styles ==================== */

/* 页面标题 */
.title {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text);
  margin-bottom: var(--space-6);
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.title::before {
  content: '';
  width: 4px;
  height: 24px;
  background: linear-gradient(180deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border-radius: var(--radius-full);
}

/* 操作区域 */
.operation-container {
  margin-top: var(--space-6);
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--color-bg-hover);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  margin-bottom: var(--space-6);
}

.operation-container .el-button {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-button:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.operation-container .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

/* 搜索区域 */
.search-container {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-left: auto;
}

.search-container .el-input {
  width: 200px;
}

.search-container .el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.search-container .el-input :deep(.el-input__inner:focus) {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 资源表格 */
.resource-table {
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.resource-table :deep(.el-table__header-wrapper) {
  background: var(--color-bg-hover);
}

.resource-table :deep(.el-table__header th) {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.resource-table :deep(.el-table__body td) {
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.resource-table :deep(.el-table__body tr) {
  transition: all var(--duration-fast) var(--ease-out);
}

.resource-table :deep(.el-table__body tr:hover > td) {
  background-color: var(--color-primary-50) !important;
}

/* 请求方式标签 */
.resource-table :deep(.el-tag) {
  border-radius: var(--radius-base);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-3);
  transition: all var(--duration-fast) var(--ease-out);
}

.resource-table :deep(.el-tag:hover) {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* Switch 开关样式 */
:deep(.el-switch .el-switch__core) {
  border-radius: var(--radius-full);
}

/* 操作按钮 */
.el-button--text {
  font-weight: var(--font-medium);
  transition: all var(--duration-fast) var(--ease-out);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-base);
  color: var(--color-primary);
}

.el-button--text:hover {
  background: var(--color-primary-50);
  transform: translateY(-1px);
}

.el-button--text .el-icon {
  margin-right: var(--space-1);
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  font-weight: var(--font-bold);
  font-size: var(--text-lg);
  color: var(--color-text);
}

/* 表单样式 */
:deep(.el-form-item__label) {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

:deep(.el-input .el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

:deep(.el-input .el-input__inner:focus) {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

:deep(.el-radio-group) {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
}

:deep(.el-radio) {
  margin-right: 0;
}

/* 加载动画 */
.resource-table :deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .resource-table :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.9);
}

/* ==================== Responsive ==================== */
@media (max-width: 768px) {
  .title {
    font-size: var(--text-xl);
  }

  .operation-container {
    flex-direction: column;
    align-items: stretch;
  }

  .search-container {
    margin-left: 0;
    width: 100%;
  }

  .search-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .resource-table {
    font-size: var(--text-xs);
  }

  .el-button--text {
    padding: var(--space-1) var(--space-2);
  }
}

@media (max-width: 480px) {
  :deep(.el-dialog) {
    width: 90% !important;
  }

  :deep(.el-form-item__label) {
    float: none;
    display: block;
    text-align: left;
    margin-bottom: var(--space-2);
  }

  :deep(.el-form-item__content) {
    margin-left: 0 !important;
  }

  :deep(.el-input) {
    width: 100% !important;
  }

  :deep(.el-radio-group) {
    flex-direction: column;
    gap: var(--space-2);
  }
}
</style>

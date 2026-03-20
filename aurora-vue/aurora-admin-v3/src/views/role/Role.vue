<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" :icon="Plus" @click="openMenuModel(null)"> 新增 </el-button>
      <el-button
        type="danger"
        size="small"
        :icon="Delete"
        :disabled="roleIds.length == 0"
        @click="isDelete = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          :prefix-icon="Search"
          size="small"
          placeholder="请输入角色名"
          style="width: 200px"
          @keyup.enter="searchRoles" />
        <el-button type="primary" size="small" :icon="Search" style="margin-left: 1rem" @click="searchRoles">
          搜索
        </el-button>
      </div>
    </div>
    <el-table border :data="roles" @selection-change="selectionChange" v-loading="loading">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="roleName" label="角色名" align="center" />
      <el-table-column prop="roleLabel" label="权限标签" align="center">
        <template #default="{ row }">
          <el-tag>
            {{ row.roleName }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="150" align="center">
        <template #default="{ row }">
          <el-icon style="margin-right: 5px"><Clock /></el-icon>
          {{ formatDate(row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="220">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="openMenuModel(row)">
            <el-icon><Edit /></el-icon> 菜单权限
          </el-button>
          <el-button type="primary" link size="small" @click="openResourceModel(row)">
            <el-icon><FolderChecked /></el-icon> 资源权限
          </el-button>
          <el-popconfirm title="确定删除吗？" @confirm="deleteRoles(row.id)">
            <template #reference>
              <el-button size="small" type="danger" link> <el-icon><Delete /></el-icon> 删除 </el-button>
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
    <el-dialog v-model="roleMenu" width="30%">
      <template #header>
        <div class="dialog-title-container">{{ roleTitle }}</div>
      </template>
      <el-form label-width="80px" size="medium" :model="roleForm">
        <el-form-item label="角色名">
          <el-input v-model="roleForm.roleName" style="width: 250px" />
        </el-form-item>
        <el-form-item label="菜单权限">
          <el-tree :data="menus" :default-checked-keys="roleForm.menuIds" show-checkbox node-key="id" ref="menuTreeRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleMenu = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateRoleMenu"> 确 定 </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="roleResource" width="30%" top="9vh">
      <template #header>
        <div class="dialog-title-container">修改资源权限</div>
      </template>
      <el-form label-width="80px" size="medium" :model="roleForm">
        <el-form-item label="角色名">
          <el-input v-model="roleForm.roleName" style="width: 250px" />
        </el-form-item>
        <el-form-item label="资源权限">
          <el-tree
            :data="resources"
            :default-checked-keys="roleForm.resourceIds"
            show-checkbox
            node-key="id"
            ref="resourceTreeRef" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="roleResource = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateRoleResource"> 确 定 </el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="isDelete" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; font-size: 1.5rem; margin-right: 8px"><Warning /></el-icon>
          提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <template #footer>
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteRoles(null)"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, nextTick, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { 
  Plus, 
  Delete, 
  Search, 
  Edit, 
  FolderChecked, 
  Clock,
  Warning
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { usePageStateStore } from '@/stores/pageState'
import dayjs from 'dayjs'

const route = useRoute()
const pageStateStore = usePageStateStore()

// Refs
const menuTreeRef = ref(null)
const resourceTreeRef = ref(null)

// 响应式数据
const loading = ref(true)
const isDelete = ref(false)
const roles = ref([])
const roleIds = ref([])
const keywords = ref('')
const current = ref(1)
const size = ref(10)
const count = ref(0)
const roleMenu = ref(false)
const roleResource = ref(false)
const roleTitle = ref('')
const resources = ref([])
const menus = ref([])
const roleForm = reactive({
  roleName: '',
  roleLabel: '',
  resourceIds: [],
  menuIds: []
})

// 日期格式化
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

// 搜索角色
const searchRoles = () => {
  current.value = 1
  listRoles()
}

// 分页大小变化
const sizeChange = (newSize) => {
  size.value = newSize
  listRoles()
}

// 页码变化
const currentChange = (newCurrent) => {
  current.value = newCurrent
  pageStateStore.updatePageState('role', newCurrent)
  listRoles()
}

// 选择变化
const selectionChange = (selectedRoles) => {
  roleIds.value = selectedRoles.map(item => item.id)
}

// 获取角色列表
const listRoles = () => {
  loading.value = true
  request.get('/admin/roles', {
    params: {
      current: current.value,
      size: size.value,
      keywords: keywords.value
    }
  }).then(({ data }) => {
    roles.value = data.data.records
    count.value = data.data.count
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取角色列表失败')
    console.error('API Error:', error)
  })
  
  request.get('/admin/role/resources').then(({ data }) => {
    resources.value = data?.data || []
  }).catch(error => {
    console.error('API Error:', error)
  })
  
  request.get('/admin/role/menus').then(({ data }) => {
    menus.value = data.data
  }).catch(error => {
    console.error('API Error:', error)
  })
}

// 删除角色
const deleteRoles = (id) => {
  const param = id == null ? { data: roleIds.value } : { data: [id] }
  request.delete('/admin/roles', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listRoles()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    isDelete.value = false
  }).catch(error => {
    ElMessage.error('删除失败')
    console.error('API Error:', error)
  })
}

// 打开菜单权限对话框
const openMenuModel = (role) => {
  nextTick(() => {
    if (menuTreeRef.value) {
      menuTreeRef.value.setCheckedKeys([])
    }
  })
  
  roleTitle.value = role ? '修改角色' : '新增角色'
  
  if (role != null) {
    Object.assign(roleForm, JSON.parse(JSON.stringify(role)))
  } else {
    Object.assign(roleForm, {
      roleName: '',
      roleLabel: '',
      resourceIds: [],
      menuIds: []
    })
  }
  roleMenu.value = true
}

// 打开资源权限对话框
const openResourceModel = (role) => {
  nextTick(() => {
    if (resourceTreeRef.value) {
      resourceTreeRef.value.setCheckedKeys([])
    }
  })
  Object.assign(roleForm, JSON.parse(JSON.stringify(role)))
  roleResource.value = true
}

// 保存资源权限
const saveOrUpdateRoleResource = () => {
  roleForm.menuIds = null
  roleForm.resourceIds = resourceTreeRef.value ? resourceTreeRef.value.getCheckedKeys() : []
  request.post('/admin/role', roleForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listRoles()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    roleResource.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    console.error('API Error:', error)
  })
}

// 保存菜单权限
const saveOrUpdateRoleMenu = () => {
  if (roleForm.roleName.trim() == '') {
    ElMessage.error('角色名不能为空')
    return false
  }
  roleForm.resourceIds = null
  const checkedKeys = menuTreeRef.value ? menuTreeRef.value.getCheckedKeys() : []
  const halfCheckedKeys = menuTreeRef.value ? menuTreeRef.value.getHalfCheckedKeys() : []
  roleForm.menuIds = checkedKeys.concat(halfCheckedKeys)
  request.post('/admin/role', roleForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listRoles()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    roleMenu.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    console.error('API Error:', error)
  })
}

// 初始化
onMounted(() => {
  current.value = pageStateStore.pageState.role
  listRoles()
})
</script>

<style scoped>
/* ==================== Role Page Modern Styles ====================
 * 基于 UI/UX Pro Max 设计系统
 * 配色: Primary #2563EB, CTA #F97316
 */

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

/* 操作区域 - 现代化工具栏 */
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

.operation-container .el-button:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.operation-container .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

.operation-container .el-button--danger {
  background: linear-gradient(135deg, var(--color-error) 0%, #f87171 100%);
  border: none;
}

/* 搜索区域 */
.operation-container > div:last-child {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-left: auto;
}

.operation-container .el-input {
  width: 200px;
}

.operation-container .el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-input :deep(.el-input__inner):focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 角色表格 - 现代化数据表格 */
.el-table {
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.el-table :deep(.el-table__header-wrapper) {
  background: var(--color-bg-hover);
}

.el-table :deep(.el-table__header th) {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.el-table :deep(.el-table__body td) {
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.el-table :deep(.el-table__body tr) {
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table :deep(.el-table__body tr:hover > td) {
  background-color: var(--color-primary-50) !important;
}

/* 角色标签 */
.el-table :deep(.el-tag) {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-3);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table :deep(.el-tag):hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
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

/* 分页 - 现代化样式 */
.pagination-container {
  float: right;
  margin-top: var(--space-6);
  margin-bottom: var(--space-4);
}

.pagination-container :deep(.el-pagination) {
  font-weight: var(--font-medium);
}

.pagination-container :deep(.el-pagination .el-pager li) {
  border-radius: var(--radius-base);
  transition: all var(--duration-fast) var(--ease-out);
}

.pagination-container :deep(.el-pagination .el-pager li.is-active) {
  background: var(--color-primary);
}

.pagination-container :deep(.el-pagination .el-pager li):hover {
  transform: translateY(-1px);
}

.pagination-container :deep(.el-pagination button) {
  border-radius: var(--radius-base);
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  font-weight: var(--font-bold);
  font-size: var(--text-lg);
  color: var(--color-text);
}

.dialog-title-container .el-icon {
  font-size: var(--text-2xl);
  margin-right: var(--space-2);
  color: var(--color-warning);
}

/* 表单样式 */
:deep(.el-form-item__label) {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

.el-input :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-input :deep(.el-input__inner):focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 树形控件 */
.el-tree {
  background: var(--color-bg-hover);
  border-radius: var(--radius-base);
  padding: var(--space-3);
  border: 1px solid var(--color-border);
}

.el-tree :deep(.el-tree-node__content) {
  border-radius: var(--radius-sm);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-tree :deep(.el-tree-node__content):hover {
  background: var(--color-primary-50);
}

/* 加载动画 */
.el-table :deep(.el-loading-mask) {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .el-table :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.9);
}

[data-theme="dark"] .el-tree {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
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

  .operation-container > div:last-child {
    margin-left: 0;
    width: 100%;
  }

  .operation-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
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

  .el-input {
    width: 100% !important;
  }
}
</style>

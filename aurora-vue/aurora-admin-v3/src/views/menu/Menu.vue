<template>
  <el-card class="main-card">
    <div class="title">{{ route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" :icon="Plus" @click="openModel(null)"> 新增菜单 </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          :prefix-icon="Search"
          size="small"
          placeholder="请输入菜单名"
          style="width: 200px"
          @keyup.enter="listMenus" />
        <el-button type="primary" size="small" :icon="Search" style="margin-left: 1rem" @click="listMenus">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      v-loading="loading"
      :data="menus"
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="name" label="菜单名称" width="140" />
      <el-table-column prop="icon" align="center" label="图标" width="100">
        <template #default="{ row }">
          <i :class="'iconfont ' + row.icon" />
        </template>
      </el-table-column>
      <el-table-column prop="orderNum" align="center" label="排序" width="100" />
      <el-table-column prop="path" label="访问路径" />
      <el-table-column prop="component" label="组件路径" />
      <el-table-column prop="isHidden" label="隐藏" align="center" width="80">
        <template #default="{ row }">
          <el-switch
            v-model="row.isHidden"
            :active-color="'#13ce66'"
            :inactive-color="'#F4F4F5'"
            :active-value="1"
            :inactive-value="0"
            @change="changeDisable(row)" />
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center" width="150">
        <template #default="{ row }">
          <el-icon style="margin-right: 5px"><Clock /></el-icon>
          {{ formatDate(row.createTime) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="200">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="openModel(row, 1)" v-if="row.children">
            <el-icon><Plus /></el-icon> 新增
          </el-button>
          <el-button type="primary" link size="small" @click="openModel(row, 2)">
            <el-icon><Edit /></el-icon> 修改
          </el-button>
          <el-popconfirm title="确定删除吗？" @confirm="deleteMenu(row.id)">
            <template #reference>
              <el-button size="small" type="danger" link> <el-icon><Delete /></el-icon> 删除 </el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="addMenu" width="30%" top="12vh">
      <template #header>
        <div class="dialog-title-container">{{ menuTitle }}</div>
      </template>
      <el-form label-width="80px" size="medium" :model="menuForm">
        <el-form-item label="菜单类型" v-if="show">
          <el-radio-group v-model="isCatalog">
            <el-radio :label="true">目录</el-radio>
            <el-radio :label="false">一级菜单</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称">
          <el-input v-model="menuForm.name" style="width: 220px" />
        </el-form-item>
        <el-form-item label="菜单图标">
          <el-popover placement="bottom-start" width="300" trigger="click">
            <template #default>
              <el-row>
                <el-col v-for="(item, index) in icons" :key="index" :md="12" :gutter="10">
                  <div class="icon-item" @click="checkIcon(item)"><i :class="'iconfont ' + item" /> {{ item }}</div>
                </el-col>
              </el-row>
            </template>
            <template #reference>
              <el-input
                :prefix-icon="'iconfont ' + menuForm.icon"
                v-model="menuForm.icon"
                style="width: 220px" />
            </template>
          </el-popover>
        </el-form-item>
        <el-form-item label="组件路径" v-show="!isCatalog">
          <el-input v-model="menuForm.component" style="width: 220px" />
        </el-form-item>
        <el-form-item label="访问路径">
          <el-input v-model="menuForm.path" style="width: 220px" />
        </el-form-item>
        <el-form-item label="显示排序">
          <el-input-number v-model="menuForm.orderNum" controls-position="right" :min="1" :max="10" />
        </el-form-item>
        <el-form-item label="显示状态">
          <el-radio-group v-model="menuForm.isHidden">
            <el-radio :label="0">显示</el-radio>
            <el-radio :label="1">隐藏</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addMenu = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateMenu"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { 
  Plus, 
  Search, 
  Edit, 
  Delete, 
  Clock
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import dayjs from 'dayjs'

const route = useRoute()

// 响应式数据
const keywords = ref('')
const loading = ref(true)
const addMenu = ref(false)
const isCatalog = ref(true)
const show = ref(true)
const menuTitle = ref('')
const menus = ref([])
const menuForm = reactive({
  id: null,
  name: '',
  icon: '',
  component: '',
  path: '',
  orderNum: 1,
  parentId: null,
  isHidden: 0
})

const icons = [
  'el-icon-myshouye',
  'el-icon-myfabiaowenzhang',
  'el-icon-myyonghuliebiao',
  'el-icon-myxiaoxi',
  'el-icon-myliuyan',
  'el-icon-myshouye',
  'el-icon-myfabiaowenzhang',
  'el-icon-myyonghuliebiao',
  'el-icon-myxiaoxi',
  'el-icon-myliuyan'
]

// 日期格式化
const formatDate = (date) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'
}

// 获取菜单列表
const listMenus = () => {
  loading.value = true
  request.get('/admin/menus', {
    params: {
      keywords: keywords.value
    }
  }).then(({ data }) => {
    menus.value = data.data
    loading.value = false
  }).catch(error => {
    loading.value = false
    ElMessage.error('获取菜单列表失败')
    console.error('API Error:', error)
  })
}

// 打开对话框
const openModel = (menu, type) => {
  if (menu) {
    show.value = false
    isCatalog.value = false
    switch (type) {
      case 1:
        Object.assign(menuForm, {
          id: null,
          name: '',
          icon: '',
          component: '',
          path: '',
          orderNum: 1,
          parentId: null,
          isHidden: 0
        })
        menuTitle.value = '新增菜单'
        menuForm.parentId = JSON.parse(JSON.stringify(menu.id))
        break
      case 2:
        menuTitle.value = '修改菜单'
        Object.assign(menuForm, JSON.parse(JSON.stringify(menu)))
        break
    }
  } else {
    menuTitle.value = '新增菜单'
    show.value = true
    Object.assign(menuForm, {
      id: null,
      name: '',
      icon: '',
      component: 'Layout',
      path: '',
      orderNum: 1,
      parentId: null,
      isHidden: 0
    })
  }
  addMenu.value = true
}

// 选择图标
const checkIcon = (icon) => {
  menuForm.icon = icon
}

// 切换隐藏状态
const changeDisable = (menu) => {
  const params = {
    id: menu.id,
    isHidden: menu.isHidden
  }
  request.put('/admin/menus/isHidden', params).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '修改成功'
      })
    } else {
      ElNotification.error({
        title: '失败',
        message: '修改失败'
      })
    }
  }).catch(error => {
    ElMessage.error('操作失败')
    console.error('API Error:', error)
  })
}

// 保存或更新菜单
const saveOrUpdateMenu = () => {
  if (menuForm.name.trim() == '') {
    ElMessage.error('菜单名不能为空')
    return false
  }
  if (menuForm.icon.trim() == '') {
    ElMessage.error('菜单icon不能为空')
    return false
  }
  if (menuForm.component.trim() == '') {
    ElMessage.error('菜单组件路径不能为空')
    return false
  }
  if (menuForm.path.trim() == '') {
    ElMessage.error('菜单访问路径不能为空')
    return false
  }
  request.post('/admin/menus', menuForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '操作成功'
      })
      listMenus()
    } else {
      ElNotification.error({
        title: '失败',
        message: '操作失败'
      })
    }
    addMenu.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    console.error('API Error:', error)
  })
}

// 删除菜单
const deleteMenu = (id) => {
  request.delete('/admin/menus/' + id).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '删除成功'
      })
      listMenus()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
  }).catch(error => {
    ElMessage.error('删除失败')
    console.error('API Error:', error)
  })
}

// 初始化
onMounted(() => {
  listMenus()
})
</script>

<style scoped>
/* ==================== Menu Page Modern Styles ====================
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

.operation-container .el-button:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.operation-container .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
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

/* 菜单表格 - 现代化树形表格 */
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

/* 菜单名称 */
.el-table :deep(.cell) {
  font-size: var(--text-sm);
  color: var(--color-text);
}

/* 图标样式 */
.iconfont {
  font-size: var(--text-lg);
  color: var(--color-primary);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table :deep(tr:hover) .iconfont {
  transform: scale(1.1);
}

/* Switch 开关样式 */
:deep(.el-switch__core) {
  border-radius: var(--radius-full);
}

/* 操作按钮 */
.el-button--text {
  font-weight: var(--font-medium);
  transition: all var(--duration-fast) var(--ease-out);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-base);
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

/* 图标选择器 */
.icon-item {
  cursor: pointer;
  padding: var(--space-2) var(--space-3);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-radius: var(--radius-base);
  transition: all var(--duration-fast) var(--ease-out);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.icon-item:hover {
  background: var(--color-primary-50);
  color: var(--color-primary);
  transform: translateX(4px);
}

.icon-item i {
  color: var(--color-primary);
  font-size: var(--text-base);
}

/* 表单样式 */
:deep(.el-form-item__label) {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

.el-input :deep(.el-input__inner),
.el-input-number :deep(.el-input__inner) {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-input :deep(.el-input__inner):focus,
.el-input-number :deep(.el-input__inner):focus {
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

/* 数字输入器 */
.el-input-number {
  width: 120px;
}

.el-input-number :deep(.el-input-number__decrease),
.el-input-number :deep(.el-input-number__increase) {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
  color: var(--color-text-secondary);
}

.el-input-number :deep(.el-input-number__decrease):hover,
.el-input-number :deep(.el-input-number__increase):hover {
  color: var(--color-primary);
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

[data-theme="dark"] .icon-item:hover {
  background: var(--color-bg-active);
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

  .el-table {
    font-size: var(--text-xs);
  }

  .el-table :deep(.el-table__header th) {
    padding: var(--space-2) var(--space-3) !important;
  }

  .el-table :deep(.el-table__body td) {
    padding: var(--space-2) var(--space-3) !important;
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

  .el-input,
  .el-input-number {
    width: 100% !important;
  }
}
</style>

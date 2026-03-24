<template>
  <div class="menu-page">
    <!-- 页面头部 - 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Menu /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ menuCount }}</span>
          <span class="stat-label">菜单总数</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><View /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ visibleCount }}</span>
          <span class="stat-label">可见菜单</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Hide /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ hiddenCount }}</span>
          <span class="stat-label">隐藏菜单</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="openModel(null)" class="btn-add">
            <span>新增菜单</span>
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="keywords"
            :prefix-icon="Search"
            placeholder="搜索菜单名..."
            class="search-input"
            clearable
            @keyup.enter="listMenus"
            @clear="listMenus" />
          <el-button type="primary" :icon="Search" @click="listMenus" circle />
        </div>
      </div>

      <!-- 现代化树形表格 -->
      <el-table
        v-loading="loading"
        :data="menus"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        class="modern-table"
        :header-cell-style="{ background: 'transparent' }">
        <el-table-column prop="name" label="菜单名称" min-width="180">
          <template #default="{ row }">
            <div class="menu-name-cell">
              <div class="menu-icon-wrapper">
                <i :class="'iconfont ' + row.icon" />
              </div>
              <span class="menu-name-text">{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="orderNum" label="排序" width="90" align="center">
          <template #default="{ row }">
            <span class="order-badge">{{ row.orderNum }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="访问路径" min-width="180">
          <template #default="{ row }">
            <span v-if="row.path" class="menu-path">{{ row.path }}</span>
            <span v-else class="menu-path-empty">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="component" label="组件路径" min-width="180">
          <template #default="{ row }">
            <span v-if="row.component" class="menu-component">{{ row.component }}</span>
            <span v-else class="menu-path-empty">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="isHidden" label="状态" width="110" align="center">
          <template #default="{ row }">
            <div class="switch-cell">
              <span class="switch-label" :class="{ active: row.isHidden === 0 }">
                {{ row.isHidden === 0 ? '显示' : '隐藏' }}
              </span>
              <el-switch
                v-model="row.isHidden"
                :active-value="1"
                :inactive-value="0"
                @change="changeDisable(row)" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" align="center">
          <template #default="{ row }">
            <div class="time-cell">
              <el-icon class="time-icon"><Clock /></el-icon>
              <span>{{ formatDate(row.createTime) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-btns">
              <el-tooltip content="新增子菜单" placement="top" :show-after="500" v-if="row.children">
                <button class="action-btn add" @click="openModel(row, 1)"><el-icon><Plus /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="修改" placement="top" :show-after="500">
                <button class="action-btn edit" @click="openModel(row, 2)"><el-icon><Edit /></el-icon></button>
              </el-tooltip>
              <el-tooltip content="删除" placement="top" :show-after="500">
                <button class="action-btn delete" @click="handleDelete(row.id)"><el-icon><Delete /></el-icon></button>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑菜单对话框 -->
    <el-dialog v-model="addMenu" width="500px" class="modern-dialog" :show-close="false">
      <div class="dialog-icon-wrapper primary"><el-icon><EditPen /></el-icon></div>
      <div class="dialog-content">
        <h3>{{ menuTitle }}</h3>
        <el-form :model="menuForm" class="menu-form" label-position="top">
          <el-form-item label="菜单类型" v-if="show">
            <div class="type-select">
              <button
                class="type-option"
                :class="{ active: isCatalog }"
                @click="isCatalog = true">
                <el-icon><FolderOpened /></el-icon>
                <span>目录</span>
              </button>
              <button
                class="type-option"
                :class="{ active: !isCatalog }"
                @click="isCatalog = false">
                <el-icon><Document /></el-icon>
                <span>一级菜单</span>
              </button>
            </div>
          </el-form-item>
          <el-form-item label="菜单名称">
            <el-input v-model="menuForm.name" placeholder="请输入菜单名称" class="form-input" />
          </el-form-item>
          <el-form-item label="菜单图标">
            <el-popover placement="bottom-start" width="300" trigger="click">
              <template #default>
                <div class="icon-grid">
                  <div
                    v-for="(item, index) in icons"
                    :key="index"
                    class="icon-item"
                    :class="{ active: menuForm.icon === item }"
                    @click="checkIcon(item)">
                    <i :class="'iconfont ' + item" />
                  </div>
                </div>
              </template>
              <template #reference>
                <div class="icon-input-wrapper" @click.stop>
                  <div class="icon-preview">
                    <i v-if="menuForm.icon" :class="'iconfont ' + menuForm.icon" />
                    <el-icon v-else><QuestionFilled /></el-icon>
                  </div>
                  <el-input v-model="menuForm.icon" placeholder="选择或输入图标" class="form-input" />
                </div>
              </template>
            </el-popover>
          </el-form-item>
          <el-form-item label="组件路径" v-show="!isCatalog">
            <el-input v-model="menuForm.component" placeholder="如: layout/Index" class="form-input" />
          </el-form-item>
          <el-form-item label="访问路径">
            <el-input v-model="menuForm.path" placeholder="/path" class="form-input" />
          </el-form-item>
          <el-form-item label="显示排序">
            <el-input-number v-model="menuForm.orderNum" controls-position="right" :min="1" :max="10" class="order-input" />
          </el-form-item>
          <el-form-item label="显示状态">
            <div class="type-select">
              <button
                class="type-option"
                :class="{ active: menuForm.isHidden === 0 }"
                @click="menuForm.isHidden = 0">
                <el-icon><View /></el-icon>
                <span>显示</span>
              </button>
              <button
                class="type-option"
                :class="{ active: menuForm.isHidden === 1 }"
                @click="menuForm.isHidden = 1">
                <el-icon><Hide /></el-icon>
                <span>隐藏</span>
              </button>
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="addMenu = false" class="btn-cancel">取消</el-button>
          <el-button type="primary" @click="saveOrUpdateMenu" class="btn-confirm">确认保存</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElNotification, ElMessageBox } from 'element-plus'
import {
  Plus,
  Search,
  Edit,
  Delete,
  Clock,
  Menu,
  View,
  Hide,
  FolderOpened,
  Document,
  EditPen,
  QuestionFilled
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import dayjs from 'dayjs'
import logger from '@/utils/logger'

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
  'el-icon-myliuyan'
]

// 统计数据
const flatMenus = (items) => {
  let result = []
  items.forEach(item => {
    result.push(item)
    if (item.children && item.children.length) {
      result = result.concat(flatMenus(item.children))
    }
  })
  return result
}

const allFlat = computed(() => flatMenus(menus.value))
const menuCount = computed(() => allFlat.value.length)
const visibleCount = computed(() => allFlat.value.filter(m => m.isHidden === 0).length)
const hiddenCount = computed(() => allFlat.value.filter(m => m.isHidden === 1).length)

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
    logger.error('API Error:', error)
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
        menuTitle.value = '新增子菜单'
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
      ElNotification.success({ title: '成功', message: '修改成功' })
    } else {
      ElNotification.error({ title: '失败', message: '修改失败' })
    }
  }).catch(error => {
    ElMessage.error('操作失败')
    logger.error('API Error:', error)
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
      ElNotification.success({ title: '成功', message: '操作成功' })
      listMenus()
    } else {
      ElNotification.error({ title: '失败', message: '操作失败' })
    }
    addMenu.value = false
  }).catch(error => {
    ElMessage.error('保存失败')
    logger.error('API Error:', error)
  })
}

// 删除菜单
const handleDelete = (id) => {
  ElMessageBox.confirm('确定删除该菜单吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    deleteMenu(id)
  }).catch(() => {})
}

const deleteMenu = (id) => {
  request.delete('/admin/menus/' + id).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '删除成功' })
      listMenus()
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  }).catch(error => {
    ElMessage.error('删除失败')
    logger.error('API Error:', error)
  })
}

// 初始化
onMounted(() => {
  listMenus()
})
</script>

<style scoped>
.menu-page {
  padding: 0;
}

/* 统计卡片 */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: var(--bg-base, #fff);
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid var(--border-default, #e5e7eb);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.08);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  flex-shrink: 0;
}

.stat-icon.primary { background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%); color: #fff; }
.stat-icon.success { background: linear-gradient(135deg, #10b981 0%, #34d399 100%); color: #fff; }
.stat-icon.warning { background: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%); color: #fff; }

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary, #1f2937);
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary, #6b7280);
}

/* 主卡片 */
.main-card {
  border-radius: 16px;
  border: 1px solid var(--border-default, #e5e7eb);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  background: var(--bg-base, #fff);
}

.main-card :deep(.el-card__body) {
  padding: 24px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  flex-wrap: wrap;
  gap: 16px;
}

.toolbar-left {
  display: flex;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-add {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
  border-radius: 10px;
  font-weight: 500;
  height: 40px;
  padding: 0 20px;
  transition: all 0.2s ease;
}

.btn-add:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.search-input {
  width: 280px;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  transition: all 0.2s ease;
}

.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2), 0 0 0 1px #3b82f6;
}

/* 现代化表格 */
.modern-table {
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--border-default, #e5e7eb);
}

.modern-table :deep(.el-table__header-wrapper th) {
  background: var(--bg-elevated, #f9fafb);
  color: var(--text-secondary, #6b7280);
  font-weight: 600;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 16px 12px;
  border-bottom: 1px solid var(--border-default, #e5e7eb);
}

.modern-table :deep(.el-table__body tr) {
  transition: all 0.2s ease;
}

.modern-table :deep(.el-table__body tr:hover > td) {
  background: var(--bg-hover, #f3f4f6) !important;
}

.modern-table :deep(.el-table__body td) {
  padding: 16px 12px;
  border-bottom: 1px solid var(--border-light, #f3f4f6);
}

/* 菜单名称单元格 */
.menu-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.menu-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}

.menu-name-text {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

/* 排序徽章 */
.order-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: var(--bg-elevated, #f3f4f6);
  font-weight: 600;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
}

/* 路径样式 */
.menu-path, .menu-component {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  background: var(--bg-elevated, #f3f4f6);
  padding: 4px 10px;
  border-radius: 6px;
  display: inline-block;
}

.menu-path-empty {
  color: var(--text-secondary, #9ca3af);
}

/* 开关单元格 */
.switch-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
}

.switch-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
  min-width: 28px;
}

.switch-label.active {
  color: #16a34a;
}

/* 时间单元格 */
.time-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-secondary, #6b7280);
  font-size: 14px;
}

.time-icon {
  color: #3b82f6;
}

/* 操作按钮 */
.action-btns {
  display: flex;
  justify-content: center;
  gap: 8px;
}

.action-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 16px;
}

.action-btn.add { background: #f0fdf4; color: #16a34a; }
.action-btn.add:hover { background: #16a34a; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(22, 163, 74, 0.3); }
.action-btn.edit { background: #eff6ff; color: #3b82f6; }
.action-btn.edit:hover { background: #3b82f6; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3); }
.action-btn.delete { background: #fef2f2; color: #ef4444; }
.action-btn.delete:hover { background: #ef4444; color: #fff; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3); }

/* 优雅对话框 */
.modern-dialog :deep(.el-dialog__header) {
  display: none;
}

.modern-dialog :deep(.el-dialog__body) {
  padding: 32px 32px 24px;
  max-height: 70vh;
  overflow-y: auto;
}

.modern-dialog :deep(.el-dialog__footer) {
  padding: 0 32px 32px;
}

.dialog-icon-wrapper {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  margin: 0 auto 20px;
}

.dialog-icon-wrapper.primary { background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%); color: #3b82f6; }

.dialog-content {
  text-align: center;
}

.dialog-content h3 {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin: 0 0 8px;
}

.dialog-content p {
  font-size: 14px;
  color: var(--text-secondary, #6b7280);
  margin: 0;
}

.dialog-footer {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.btn-cancel {
  border-radius: 10px;
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
}

.btn-confirm {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
  border-radius: 10px;
  height: 44px;
  padding: 0 24px;
  font-weight: 500;
}

.btn-confirm:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

/* 表单样式 */
.menu-form {
  margin-top: 24px;
  text-align: left;
}

.menu-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  padding-bottom: 8px;
}

.form-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  height: 44px;
}

.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2), 0 0 0 1px #3b82f6;
}

/* 类型选择器 */
.type-select {
  display: flex;
  gap: 10px;
  width: 100%;
}

.type-option {
  flex: 1;
  padding: 12px 16px;
  border-radius: 10px;
  border: 2px solid var(--border-default, #e5e7eb);
  background: var(--bg-base, #fff);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s ease;
  color: var(--text-secondary, #6b7280);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.type-option.active {
  border-color: #3b82f6;
  background: #eff6ff;
  color: #3b82f6;
}

.type-option:hover {
  transform: translateY(-1px);
}

/* 图标输入 */
.icon-input-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-preview {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
  color: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

/* 图标网格 */
.icon-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.icon-item {
  padding: 10px;
  border-radius: 10px;
  border: 2px solid var(--border-default, #e5e7eb);
  cursor: pointer;
  text-align: center;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--text-primary, #1f2937);
}

.icon-item:hover {
  border-color: #3b82f6;
  background: #eff6ff;
  transform: translateY(-1px);
}

.icon-item.active {
  border-color: #3b82f6;
  background: #eff6ff;
  color: #3b82f6;
}

/* 数字输入器 */
.order-input {
  width: 120px;
}

.order-input :deep(.el-input__wrapper) {
  border-radius: 10px;
}

/* 深色模式 */
[data-theme="dark"] .stat-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .stat-value { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }

[data-theme="dark"] .main-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}

[data-theme="dark"] .modern-table :deep(.el-table__header-wrapper th) {
  background: var(--bg-elevated, #374151);
  color: var(--text-secondary, #9ca3af);
}

[data-theme="dark"] .modern-table :deep(.el-table__body tr:hover > td) {
  background: var(--bg-hover, #374151) !important;
}

[data-theme="dark"] .menu-icon-wrapper { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .menu-name-text { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .menu-path, .menu-component { background: var(--bg-elevated, #374151); }
[data-theme="dark"] .order-badge { background: var(--bg-elevated, #374151); }

[data-theme="dark"] .action-btn.add { background: rgba(22, 163, 74, 0.15); }
[data-theme="dark"] .action-btn.edit { background: rgba(59, 130, 246, 0.15); }
[data-theme="dark"] .action-btn.delete { background: rgba(239, 68, 68, 0.15); }

[data-theme="dark"] .dialog-content h3 { color: var(--text-primary, #f9fafb); }

[data-theme="dark"] .type-option {
  border-color: var(--border-default, #374151);
  background: var(--bg-base, #1f2937);
  color: var(--text-secondary, #9ca3af);
}

[data-theme="dark"] .type-option.active {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

[data-theme="dark"] .icon-preview { background: rgba(59, 130, 246, 0.15); }

[data-theme="dark"] .icon-item {
  border-color: var(--border-default, #374151);
  color: var(--text-primary, #f9fafb);
}

[data-theme="dark"] .icon-item:hover,
[data-theme="dark"] .icon-item.active {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.15);
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-row { grid-template-columns: repeat(2, 1fr); }
  .stat-card:last-child { grid-column: span 2; }
}

@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; }
  .stat-card:last-child { grid-column: span 1; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left, .toolbar-right { width: 100%; }
  .btn-add { width: 100%; }
  .search-input { width: 100%; }
}

@media (max-width: 480px) {
  .main-card :deep(.el-card__body) { padding: 16px; }
  .stat-card { padding: 16px; }
  .stat-icon { width: 48px; height: 48px; font-size: 20px; }
  .stat-value { font-size: 24px; }
  .modern-dialog :deep(.el-dialog) { width: 92% !important; }
  .type-select { flex-direction: column; }
}
</style>

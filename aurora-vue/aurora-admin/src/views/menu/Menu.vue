<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openModel(null)"> 新增菜单 </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入菜单名"
          style="width: 200px"
          @keyup.enter.native="listMenus" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="listMenus">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      v-loading="loading"
      :data="menus"
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }">
      <el-table-column prop="name" label="菜单名称" width="140" />
      <el-table-column prop="icon" align="center" label="图标" width="100">
        <template slot-scope="scope">
          <i :class="'iconfont ' + scope.row.icon" />
        </template>
      </el-table-column>
      <el-table-column prop="orderNum" align="center" label="排序" width="100" />
      <el-table-column prop="path" label="访问路径" />
      <el-table-column prop="component" label="组件路径" />
      <el-table-column prop="isHidden" label="隐藏" align="center" width="80">
        <template slot-scope="scope">
          <el-switch
            v-model="scope.row.isHidden"
            active-color="#13ce66"
            inactive-color="#F4F4F5"
            :active-value="1"
            :inactive-value="0"
            @change="changeDisable(scope.row)" />
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center" width="150">
        <template slot-scope="scope">
          <i class="el-icon-time" style="margin-right: 5px" />
          {{ scope.row.createTime | date }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="200">
        <template slot-scope="scope">
          <el-button type="text" size="mini" @click="openModel(scope.row, 1)" v-if="scope.row.children">
            <i class="el-icon-plus" /> 新增
          </el-button>
          <el-button type="text" size="mini" @click="openModel(scope.row, 2)">
            <i class="el-icon-edit" /> 修改
          </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 10px" @confirm="deleteMenu(scope.row.id)">
            <el-button size="mini" type="text" slot="reference"> <i class="el-icon-delete" /> 删除 </el-button>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog :visible.sync="addMenu" width="30%" top="12vh">
      <div class="dialog-title-container" slot="title" ref="menuTitle" />
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
            <el-row>
              <el-col v-for="(item, index) of icons" :key="index" :md="12" :gutter="10">
                <div class="icon-item" @click="checkIcon(item)"><i :class="'iconfont ' + item" /> {{ item }}</div>
              </el-col>
            </el-row>
            <el-input
              :prefix-icon="'iconfont ' + menuForm.icon"
              slot="reference"
              v-model="menuForm.icon"
              style="width: 220px" />
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
      <div slot="footer">
        <el-button @click="addMenu = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateMenu"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.listMenus()
  },
  data() {
    return {
      keywords: '',
      loading: true,
      addMenu: false,
      isCatalog: true,
      show: true,
      menus: [],
      menuForm: {
        id: null,
        name: '',
        icon: '',
        component: '',
        path: '',
        orderNum: 1,
        parentId: null,
        isHidden: 0
      },
      icons: [
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
    }
  },
  methods: {
    listMenus() {
      this.axios
        .get('/api/admin/menus', {
          params: {
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.menus = data.data
          this.loading = false
        })
    },
    openModel(menu, type) {
      if (menu) {
        this.show = false
        this.isCatalog = false
        switch (type) {
          case 1:
            this.menuForm = {
              id: null,
              name: '',
              icon: '',
              component: '',
              path: '',
              orderNum: 1,
              parentId: null,
              isHidden: 0
            }
            this.$nextTick(() => {
              if (this.$refs.menuTitle) {
                this.$refs.menuTitle.innerHTML = '新增菜单'
              }
            })
            this.menuForm.parentId = JSON.parse(JSON.stringify(menu.id))
            break
          case 2:
            this.$nextTick(() => {
              if (this.$refs.menuTitle) {
                this.$refs.menuTitle.innerHTML = '修改菜单'
              }
            })
            this.menuForm = JSON.parse(JSON.stringify(menu))
            break
        }
      } else {
        this.$nextTick(() => {
          if (this.$refs.menuTitle) {
            this.$refs.menuTitle.innerHTML = '新增菜单'
          }
        })
        this.show = true
        this.menuForm = {
          id: null,
          name: '',
          icon: '',
          component: 'Layout',
          path: '',
          orderNum: 1,
          parentId: null,
          isHidden: 0
        }
      }
      this.addMenu = true
    },
    checkIcon(icon) {
      this.menuForm.icon = icon
    },
    changeDisable(menu) {
      let params = {
        id: menu.id,
        isHidden: menu.isHidden
      }
      this.axios.put('/api/admin/menus/isHidden', params).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: '修改成功'
          })
        } else {
          this.$notify.error({
            title: '失败',
            message: '修改失败'
          })
        }
      })
    },
    saveOrUpdateMenu() {
      if (this.menuForm.name.trim() == '') {
        this.$message.error('菜单名不能为空')
        return false
      }
      if (this.menuForm.icon.trim() == '') {
        this.$message.error('菜单icon不能为空')
        return false
      }
      if (this.menuForm.component.trim() == '') {
        this.$message.error('菜单组件路径不能为空')
        return false
      }
      if (this.menuForm.path.trim() == '') {
        this.$message.error('菜单访问路径不能为空')
        return false
      }
      this.axios.post('/api/admin/menus', this.menuForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: '操作成功'
          })
          this.listMenus()
        } else {
          this.$notify.error({
            title: '失败',
            message: '操作失败'
          })
        }
        this.addMenu = false
      })
    },
    deleteMenu(id) {
      this.axios.delete('/api/admin/menus/' + id).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: '删除成功'
          })
          this.listMenus()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
      })
    }
  }
}
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

.operation-container .el-input ::v-deep .el-input__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-input ::v-deep .el-input__inner:focus {
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

.el-table ::v-deep .el-table__header-wrapper {
  background: var(--color-bg-hover);
}

.el-table ::v-deep .el-table__header th {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.el-table ::v-deep .el-table__body td {
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.el-table ::v-deep .el-table__body tr {
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table ::v-deep .el-table__body tr:hover > td {
  background-color: var(--color-primary-50) !important;
}

/* 菜单名称 */
.el-table ::v-deep .cell {
  font-size: var(--text-sm);
  color: var(--color-text);
}

/* 图标样式 */
.iconfont {
  font-size: var(--text-lg);
  color: var(--color-primary);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table ::v-deep tr:hover .iconfont {
  transform: scale(1.1);
}

/* Switch 开关样式 */
.el-switch ::v-deep .el-switch__core {
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

.el-button--text i {
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
.el-form-item__label {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

.el-input ::v-deep .el-input__inner,
.el-input-number ::v-deep .el-input__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-input ::v-deep .el-input__inner:focus,
.el-input-number ::v-deep .el-input__inner:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

.el-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
}

.el-radio {
  margin-right: 0;
}

/* 数字输入器 */
.el-input-number {
  width: 120px;
}

.el-input-number ::v-deep .el-input-number__decrease,
.el-input-number ::v-deep .el-input-number__increase {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
  color: var(--color-text-secondary);
}

.el-input-number ::v-deep .el-input-number__decrease:hover,
.el-input-number ::v-deep .el-input-number__increase:hover {
  color: var(--color-primary);
}

/* 加载动画 */
.el-table ::v-deep .el-loading-mask {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .el-table ::v-deep .el-loading-mask {
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

  .el-table ::v-deep .el-table__header th {
    padding: var(--space-2) var(--space-3) !important;
  }

  .el-table ::v-deep .el-table__body td {
    padding: var(--space-2) var(--space-3) !important;
  }

  .el-button--text {
    padding: var(--space-1) var(--space-2);
  }
}

@media (max-width: 480px) {
  .el-dialog {
    width: 90% !important;
  }

  .el-form-item__label {
    float: none;
    display: block;
    text-align: left;
    margin-bottom: var(--space-2);
  }

  .el-form-item__content {
    margin-left: 0 !important;
  }

  .el-input,
  .el-input-number {
    width: 100% !important;
  }
}
</style>

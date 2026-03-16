<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openMenuModel(null)"> 新增 </el-button>
      <el-button
        type="danger"
        size="small"
        icon="el-icon-delete"
        :disabled="this.roleIds.length == 0"
        @click="isDelete = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入角色名"
          style="width: 200px"
          @keyup.enter.native="searchRoles" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="searchRoles">
          搜索
        </el-button>
      </div>
    </div>
    <el-table border :data="roles" @selection-change="selectionChange" v-loading="loading">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="roleName" label="角色名" align="center" />
      <el-table-column prop="roleLabel" label="权限标签" align="center">
        <template slot-scope="scope">
          <el-tag>
            {{ scope.row.roleName }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="150" align="center">
        <template slot-scope="scope">
          <i class="el-icon-time" style="margin-right: 5px" />
          {{ scope.row.createTime | date }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="220">
        <template slot-scope="scope">
          <el-button type="text" size="mini" @click="openMenuModel(scope.row)">
            <i class="el-icon-edit" /> 菜单权限
          </el-button>
          <el-button type="text" size="mini" @click="openResourceModel(scope.row)">
            <i class="el-icon-folder-checked" /> 资源权限
          </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 10px" @confirm="deleteRoles(scope.row.id)">
            <el-button size="mini" type="text" slot="reference"> <i class="el-icon-delete" /> 删除 </el-button>
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
    <el-dialog :visible.sync="roleMenu" width="30%">
      <div class="dialog-title-container" slot="title" ref="roleTitle" />
      <el-form label-width="80px" size="medium" :model="roleForm">
        <el-form-item label="角色名">
          <el-input v-model="roleForm.roleName" style="width: 250px" />
        </el-form-item>
        <el-form-item label="菜单权限">
          <el-tree :data="menus" :default-checked-keys="roleForm.menuIds" show-checkbox node-key="id" ref="menuTree" />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="roleMenu = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateRoleMenu"> 确 定 </el-button>
      </div>
    </el-dialog>
    <el-dialog :visible.sync="roleResource" width="30%" top="9vh">
      <div class="dialog-title-container" slot="title">修改资源权限</div>
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
            ref="resourceTree" />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="roleResource = false">取 消</el-button>
        <el-button type="primary" @click="saveOrUpdateRoleResource"> 确 定 </el-button>
      </div>
    </el-dialog>
    <el-dialog :visible.sync="isDelete" width="30%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-warning" style="color: #ff9900" />提示</div>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <div slot="footer">
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteRoles(null)"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.current = this.$store.state.pageState.role
    this.listRoles()
  },
  data: function () {
    return {
      loading: true,
      isDelete: false,
      roles: [],
      roleIds: [],
      keywords: null,
      current: 1,
      size: 10,
      count: 0,
      roleMenu: false,
      roleResource: false,
      resources: [],
      menus: [],
      roleForm: {
        roleName: '',
        roleLabel: '',
        resourceIds: [],
        menuIds: []
      }
    }
  },
  methods: {
    searchRoles() {
      this.current = 1
      this.listRoles()
    },
    sizeChange(size) {
      this.size = size
      this.listRoles()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateRolePageState', current)
      this.listRoles()
    },
    selectionChange(roles) {
      this.roleIds = []
      roles.forEach((item) => {
        this.roleIds.push(item.id)
      })
    },
    listRoles() {
      this.axios
        .get('/api/admin/roles', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.roles = data.data.records
          this.count = data.data.count
          this.loading = false
        })
      this.axios.get('/api/admin/role/resources').then(({ data }) => {
        this.resources = data.data
      })
      this.axios.get('/api/admin/role/menus').then(({ data }) => {
        this.menus = data.data
      })
    },
    deleteRoles(id) {
      var param = {}
      if (id == null) {
        param = { data: this.roleIds }
      } else {
        param = { data: [id] }
      }
      this.axios.delete('/api/admin/roles', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listRoles()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.isDelete = false
      })
    },
    openMenuModel(role) {
      this.$nextTick(function () {
        if (this.$refs.menuTree) {
          this.$refs.menuTree.setCheckedKeys([])
        }
      })
      this.$nextTick(() => {
        if (this.$refs.roleTitle) {
          this.$refs.roleTitle.innerHTML = role ? '修改角色' : '新增角色'
        }
      })
      if (role != null) {
        this.roleForm = JSON.parse(JSON.stringify(role))
      } else {
        this.roleForm = {
          roleName: '',
          roleLabel: '',
          resourceIds: [],
          menuIds: []
        }
      }
      this.roleMenu = true
    },
    openResourceModel(role) {
      this.$nextTick(function () {
        this.$refs.resourceTree.setCheckedKeys([])
      })
      this.roleForm = JSON.parse(JSON.stringify(role))
      this.roleResource = true
    },
    saveOrUpdateRoleResource() {
      this.roleForm.menuIds = null
      this.roleForm.resourceIds = this.$refs.resourceTree.getCheckedKeys()
      this.axios.post('/api/admin/role', this.roleForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listRoles()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.roleResource = false
      })
    },
    saveOrUpdateRoleMenu() {
      if (this.roleForm.roleName.trim() == '') {
        this.$message.error('角色名不能为空')
        return false
      }
      this.roleForm.resourceIds = null
      this.roleForm.menuIds = this.$refs.menuTree.getCheckedKeys().concat(this.$refs.menuTree.getHalfCheckedKeys())
      this.axios.post('/api/admin/role', this.roleForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listRoles()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.roleMenu = false
      })
    }
  }
}
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

/* 角色表格 - 现代化数据表格 */
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

/* 角色标签 */
.el-table ::v-deep .el-tag {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-3);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table ::v-deep .el-tag:hover {
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

.el-button--text i {
  margin-right: var(--space-1);
}

/* 分页 - 现代化样式 */
.pagination-container {
  float: right;
  margin-top: var(--space-6);
  margin-bottom: var(--space-4);
}

.pagination-container ::v-deep .el-pagination {
  font-weight: var(--font-medium);
}

.pagination-container ::v-deep .el-pagination .el-pager li {
  border-radius: var(--radius-base);
  transition: all var(--duration-fast) var(--ease-out);
}

.pagination-container ::v-deep .el-pagination .el-pager li.active {
  background: var(--color-primary);
}

.pagination-container ::v-deep .el-pagination .el-pager li:hover {
  transform: translateY(-1px);
}

.pagination-container ::v-deep .el-pagination button {
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

.dialog-title-container i {
  font-size: var(--text-2xl);
  margin-right: var(--space-2);
  color: var(--color-warning);
}

/* 表单样式 */
.el-form-item__label {
  font-weight: var(--font-medium);
  color: var(--color-text);
}

.el-input ::v-deep .el-input__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-input ::v-deep .el-input__inner:focus {
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

.el-tree ::v-deep .el-tree-node__content {
  border-radius: var(--radius-sm);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-tree ::v-deep .el-tree-node__content:hover {
  background: var(--color-primary-50);
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

  .el-input {
    width: 100% !important;
  }
}
</style>

<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openModel(null)"> 新增模块 </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入资源名"
          style="width: 200px"
          @keyup.enter.native="listResources" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="listResources">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      v-loading="loading"
      :data="resources"
      row-key="id"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }">
      <el-table-column prop="resourceName" label="资源名" width="220" />
      <el-table-column prop="url" label="资源路径" width="300" />
      <el-table-column prop="requestMethod" label="请求方式">
        <template slot-scope="scope" v-if="scope.row.requestMethod">
          <el-tag :type="tagType(scope.row.requestMethod)">
            {{ scope.row.requestMethod }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="isAnonymous" label="匿名访问" align="center">
        <template slot-scope="scope">
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
        <template slot-scope="scope">
          <i class="el-icon-time" style="margin-right: 5px" />
          {{ scope.row.createTime | date }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="200">
        <template slot-scope="scope">
          <el-button type="text" size="mini" @click="openAddResourceModel(scope.row)" v-if="scope.row.children">
            <i class="el-icon-plus" /> 新增
          </el-button>
          <el-button type="text" size="mini" @click="openEditResourceModel(scope.row)">
            <i class="el-icon-edit" /> 修改
          </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 10px" @confirm="deleteResource(scope.row.id)">
            <el-button size="mini" type="text" slot="reference"> <i class="el-icon-delete" /> 删除 </el-button>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog :visible.sync="addModule" width="30%">
      <div class="dialog-title-container" slot="title" ref="moduleTitle" />
      <el-form label-width="80px" size="medium" :model="resourceForm">
        <el-form-item label="模块名">
          <el-input v-model="resourceForm.resourceName" style="width: 220px" />
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="addModule = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditResource"> 确 定 </el-button>
      </span>
    </el-dialog>
    <el-dialog :visible.sync="addResource" width="30%">
      <div class="dialog-title-container" slot="title" ref="resourceTitle" />
      <el-form label-width="80px" size="medium" :model="resourceForm">
        <el-form-item label="资源名">
          <el-input v-model="resourceForm.resourceName" style="width: 220px" />
        </el-form-item>
        <el-form-item label="资源路径">
          <el-input v-model="resourceForm.url" style="width: 220px" />
        </el-form-item>
        <el-form-item label="请求方式">
          <el-radio-group v-model="resourceForm.requestMethod">
            <el-radio :label="'GET'">GET</el-radio>
            <el-radio :label="'POST'">POST</el-radio>
            <el-radio :label="'PUT'">PUT</el-radio>
            <el-radio :label="'DELETE'">DELETE</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button @click="addResource = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditResource"> 确 定 </el-button>
      </span>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.listResources()
  },
  data() {
    return {
      loading: true,
      keywords: '',
      resources: [],
      addModule: false,
      addResource: false,
      resourceForm: {}
    }
  },
  methods: {
    listResources() {
      this.axios
        .get('/api/admin/resources', {
          params: {
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.resources = data.data
          this.loading = false
        })
    },
    changeResource(resource) {
      this.axios.post('/api/admin/resources', resource).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listResources()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
      })
    },
    openModel(resource) {
      if (resource != null) {
        this.resourceForm = JSON.parse(JSON.stringify(resource))
        this.$refs.moduleTitle.innerHTML = '修改模块'
      } else {
        this.resourceForm = {}
        this.$refs.moduleTitle.innerHTML = '添加模块'
      }
      this.addModule = true
    },
    openEditResourceModel(resource) {
      if (resource.url == null) {
        this.openModel(resource)
        return false
      }
      this.resourceForm = JSON.parse(JSON.stringify(resource))
      this.$nextTick(() => {
        if (this.$refs.resourceTitle) {
          this.$refs.resourceTitle.innerHTML = '修改资源'
        }
      })
      this.addResource = true
    },
    openAddResourceModel(resource) {
      this.resourceForm = {}
      this.resourceForm.parentId = resource.id
      this.$nextTick(() => {
        if (this.$refs.resourceTitle) {
          this.$refs.resourceTitle.innerHTML = '添加资源'
        }
      })
      this.addResource = true
    },
    deleteResource(id) {
      this.axios.delete('/api/admin/resources/' + id).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listResources()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
      })
    },
    addOrEditResource() {
      if (this.resourceForm.resourceName.trim() == '') {
        this.$message.error('资源名不能为空')
        return false
      }
      this.axios.post('/api/admin/resources', this.resourceForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listResources()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.addModule = false
        this.addResource = false
      })
      .catch(error => {
        this.$message.error('保存资源失败')
        console.error('API Error:', error)
      })
    }
  },
  computed: {
    tagType() {
      return function (type) {
        switch (type) {
          case 'GET':
            return ''
          case 'POST':
            return 'success'
          case 'PUT':
            return 'warning'
          case 'DELETE':
            return 'danger'
        }
      }
    }
  }
}
</script>

<style scoped>
/* ==================== Resource Page Modern Styles ====================
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

/* 资源表格 - 现代化树形表格 */
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

/* 请求方式标签 */
.el-table ::v-deep .el-tag {
  border-radius: var(--radius-base);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-3);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table ::v-deep .el-tag:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
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
  color: var(--color-primary);
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

.el-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
}

.el-radio {
  margin-right: 0;
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

  .el-radio-group {
    flex-direction: column;
    gap: var(--space-2);
  }
}
</style>

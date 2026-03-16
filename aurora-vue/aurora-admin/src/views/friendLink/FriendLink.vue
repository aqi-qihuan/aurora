<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openModel(null)"> 新增 </el-button>
      <el-button
        type="danger"
        size="small"
        icon="el-icon-delete"
        :disabled="linkIdList.length == 0"
        @click="deleteFlag = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入友链名"
          style="width: 200px"
          @keyup.enter.native="searchLinks" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="searchLinks">
          搜索
        </el-button>
      </div>
    </div>
    <el-table border :data="linkList" @selection-change="selectionChange" v-loading="loading">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="linkAvatar" label="链接头像" align="center" width="180">
        <template slot-scope="scope">
          <img :src="scope.row.linkAvatar" width="40" height="40" />
        </template>
      </el-table-column>
      <el-table-column prop="linkName" label="链接名" align="center" />
      <el-table-column prop="linkAddress" label="链接地址" align="center" />
      <el-table-column prop="linkIntro" label="链接介绍" align="center" />
      <el-table-column prop="createTime" label="创建时间" width="140" align="center">
        <template slot-scope="scope">
          <i class="el-icon-time" style="margin-right: 5px" />
          {{ scope.row.createTime | date }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="160">
        <template slot-scope="scope">
          <el-button type="primary" size="mini" @click="openModel(scope.row)"> 编辑 </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 1rem" @confirm="deleteLink(scope.row.id)">
            <el-button size="mini" type="danger" slot="reference"> 删除 </el-button>
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
    <el-dialog :visible.sync="deleteFlag" width="30%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-warning" style="color: #ff9900" />提示</div>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <div slot="footer">
        <el-button @click="deleteFlag = false">取 消</el-button>
        <el-button type="primary" @click="deleteLink(null)"> 确 定 </el-button>
      </div>
    </el-dialog>
    <el-dialog :visible.sync="addOrEdit" width="30%">
      <div class="dialog-title-container" slot="title" ref="linkTitle" />
      <el-form label-width="80px" size="medium" :model="linkForm">
        <el-form-item label="链接名">
          <el-input style="width: 250px" v-model="linkForm.linkName" />
        </el-form-item>
        <el-form-item label="链接头像">
          <el-input style="width: 250px" v-model="linkForm.linkAvatar" />
        </el-form-item>
        <el-form-item label="链接地址">
          <el-input style="width: 250px" v-model="linkForm.linkAddress" />
        </el-form-item>
        <el-form-item label="链接介绍">
          <el-input style="width: 250px" v-model="linkForm.linkIntro" />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditCategory"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.current = this.$store.state.pageState.friendLink
    this.listLinks()
  },
  data: function () {
    return {
      loading: true,
      deleteFlag: false,
      addOrEdit: false,
      linkIdList: [],
      linkList: [],
      linkForm: {
        id: null,
        linkName: '',
        linkAvatar: '',
        linkIntro: '',
        linkAddress: ''
      },
      keywords: null,
      current: 1,
      size: 10,
      count: 0
    }
  },
  methods: {
    selectionChange(linkList) {
      this.linkIdList = []
      linkList.forEach((item) => {
        this.linkIdList.push(item.id)
      })
    },
    searchLinks() {
      this.current = 1
      this.listLinks()
    },
    sizeChange(size) {
      this.size = size
      this.listLinks()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateFriendLinkPageState', current)
      this.listLinks()
    },
    deleteLink(id) {
      var param = {}
      if (id == null) {
        param = { data: this.linkIdList }
      } else {
        param = { data: [id] }
      }
      this.axios.delete('/api/admin/links', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listLinks()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.deleteFlag = false
      })
    },
    openModel(link) {
      if (link != null) {
        this.linkForm = JSON.parse(JSON.stringify(link))
        this.$nextTick(() => {
          if (this.$refs.linkTitle) {
            this.$refs.linkTitle.innerHTML = '修改友链'
          }
        })
      } else {
        this.linkForm.id = null
        this.linkForm.linkName = ''
        this.linkForm.linkAvatar = ''
        this.linkForm.linkIntro = ''
        this.linkForm.linkAddress = ''
        this.$nextTick(() => {
          if (this.$refs.linkTitle) {
            this.$refs.linkTitle.innerHTML = '添加友链'
          }
        })
      }
      this.addOrEdit = true
    },
    addOrEditCategory() {
      if (this.linkForm.linkName.trim() == '') {
        this.$message.error('友链名不能为空')
        return false
      }
      if (this.linkForm.linkAvatar.trim() == '') {
        this.$message.error('友链头像不能为空')
        return false
      }
      if (this.linkForm.linkIntro.trim() == '') {
        this.$message.error('友链介绍不能为空')
        return false
      }
      if (this.linkForm.linkAddress.trim() == '') {
        this.$message.error('友链地址不能为空')
        return false
      }
      this.axios.post('/api/admin/links', this.linkForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listLinks()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.addOrEdit = false
      })
    },
    listLinks() {
      this.axios
        .get('/api/admin/links', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.linkList = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    }
  }
}
</script>

<style scoped>
/* ==================== Friend Link Page Modern Styles ====================
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

/* 友链表格 - 现代化数据表格 */
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

/* 链接头像 */
.el-table img {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-full);
  object-fit: cover;
  border: 2px solid var(--color-border);
  transition: all var(--duration-base) var(--ease-out);
}

.el-table img:hover {
  transform: scale(1.1);
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}

/* 操作按钮 */
.el-table .el-button {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  transition: all var(--duration-fast) var(--ease-out);
}

.el-table .el-button:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.el-table .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

.el-table .el-button--danger {
  background: linear-gradient(135deg, var(--color-error) 0%, #f87171 100%);
  border: none;
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

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
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

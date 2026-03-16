<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openModel(null)"> 新增 </el-button>
      <el-button
        type="danger"
        size="small"
        icon="el-icon-delete"
        :disabled="tagIds.length == 0"
        @click="isDelete = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入标签名"
          style="width: 200px"
          @keyup.enter.native="searchTags" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="searchTags">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="tags"
      v-loading="loading"
      @selection-change="selectionChange"
      class="tag-table"
      :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="tagName" label="标签名" align="center" min-width="150">
        <template slot-scope="scope">
          <el-tag
            size="medium"
            :type="getTagType(scope.row.tagName)"
            effect="plain"
            class="tag-item">
            <i class="el-icon-price-tag" style="margin-right: 5px" />
            {{ scope.row.tagName }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="articleCount" label="文章量" align="center" width="120" sortable>
        <template slot-scope="scope">
          <el-tag size="small" :type="getArticleCountType(scope.row.articleCount)" effect="plain">
            {{ scope.row.articleCount }} 篇
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" align="center" width="160" sortable>
        <template slot-scope="scope">
          <div class="create-time">
            <i class="el-icon-time" />
            {{ scope.row.createTime | date }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template slot-scope="scope">
          <div class="action-buttons">
            <el-button
              type="primary"
              size="mini"
              icon="el-icon-edit"
              @click="openModel(scope.row)"
              circle />
            <el-popconfirm
              title="确定删除吗？"
              @confirm="deleteTag(scope.row.id)">
              <el-button
                size="mini"
                type="danger"
                icon="el-icon-delete"
                slot="reference"
                circle />
            </el-popconfirm>
          </div>
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
    <el-dialog :visible.sync="isDelete" width="30%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-warning" style="color: #ff9900" />提示</div>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <div slot="footer">
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteTag(null)"> 确 定 </el-button>
      </div>
    </el-dialog>
    <el-dialog :visible.sync="addOrEdit" width="30%">
      <div class="dialog-title-container" slot="title" ref="tagTitle" />
      <el-form label-width="80px" size="medium" :model="tagForm">
        <el-form-item label="标签名">
          <el-input style="width: 220px" v-model="tagForm.tagName" />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditTag"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.current = this.$store.state.pageState.tag
    this.listTags()
  },
  data: function () {
    return {
      isDelete: false,
      loading: true,
      addOrEdit: false,
      keywords: null,
      tags: [],
      tagIds: [],
      tagForm: {
        id: null,
        tagName: ''
      },
      current: 1,
      size: 10,
      count: 0
    }
  },
  methods: {
    selectionChange(tags) {
      this.tagIds = []
      tags.forEach((item) => {
        this.tagIds.push(item.id)
      })
    },
    searchTags() {
      this.current = 1
      this.listTags()
    },
    sizeChange(size) {
      this.size = size
      this.listTags()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateTagPageState', current)
      this.listTags()
    },
    deleteTag(id) {
      var param = {}
      if (id == null) {
        param = { data: this.tagIds }
      } else {
        param = { data: [id] }
      }
      this.axios.delete('/api/admin/tags', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listTags()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
      })
      this.isDelete = false
    },
    listTags() {
      this.axios
        .get('/api/admin/tags', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.tags = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    },
    openModel(tag) {
      if (tag != null) {
        this.tagForm = JSON.parse(JSON.stringify(tag))
        this.$nextTick(() => {
          if (this.$refs.tagTitle) {
            this.$refs.tagTitle.innerHTML = '修改标签'
          }
        })
      } else {
        this.tagForm.id = null
        this.tagForm.tagName = ''
        this.$nextTick(() => {
          if (this.$refs.tagTitle) {
            this.$refs.tagTitle.innerHTML = '添加标签'
          }
        })
      }
      this.addOrEdit = true
    },
    addOrEditTag() {
      if (this.tagForm.tagName.trim() == '') {
        this.$message.error('标签名不能为空')
        return false
      }
      this.axios.post('/api/admin/tags', this.tagForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listTags()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.addOrEdit = false
      })
    }
  },
  computed: {
    getTagType() {
      return function (name) {
        const colors = ['primary', 'success', 'info', 'warning', 'danger']
        const index = name.length % colors.length
        return colors[index]
      }
    },
    getArticleCountType() {
      return function (count) {
        if (count >= 50) return 'danger'
        if (count >= 30) return 'warning'
        if (count >= 10) return 'success'
        return 'info'
      }
    }
  }
}
</script>

<style scoped>
/* ==================== Tag Page Modern Styles ====================
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

/* 标签表格 - 现代化数据表格 */
.tag-table {
  margin-top: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.tag-table ::v-deep .el-table__header-wrapper {
  background: var(--color-bg-hover);
}

.tag-table ::v-deep .el-table__header th {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.tag-table ::v-deep .el-table__body td {
  padding: var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.tag-table ::v-deep .el-table__body tr {
  transition: all var(--duration-fast) var(--ease-out);
}

.tag-table ::v-deep .el-table__body tr:hover > td {
  background-color: var(--color-primary-50) !important;
}

.tag-table ::v-deep .el-table__row {
  animation: slideIn var(--duration-base) var(--ease-out);
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 标签样式 */
.tag-item {
  font-size: var(--text-sm);
  font-weight: var(--font-medium);
  transition: all var(--duration-base) var(--ease-out);
  cursor: default;
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-base);
}

.tag-item:hover {
  transform: scale(1.05) translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 文章数量标签 */
.tag-table ::v-deep .el-tag {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
}

.tag-table ::v-deep .el-tag:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 创建时间 */
.create-time {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
}

.create-time i {
  color: var(--color-secondary);
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-2);
}

.action-buttons .el-button {
  transition: all var(--duration-fast) var(--ease-out);
  border-radius: var(--radius-base);
}

.action-buttons .el-button:hover {
  transform: translateY(-2px) scale(1.05);
  box-shadow: var(--shadow-md);
}

.action-buttons .el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
}

.action-buttons .el-button--danger {
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
.tag-table ::v-deep .el-loading-mask {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .tag-table ::v-deep .el-loading-mask {
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
  .tag-table ::v-deep .el-table__header {
    display: none;
  }

  .tag-table ::v-deep .el-table__row {
    display: flex;
    flex-direction: column;
    padding: var(--space-4);
    margin-bottom: var(--space-3);
    background: var(--color-bg-card);
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border);
    box-shadow: var(--shadow-sm);
  }

  .tag-table ::v-deep .el-table__row td {
    border: none;
    padding: var(--space-2) 0 !important;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .tag-table ::v-deep .el-table__row td::before {
    content: attr(data-label);
    font-weight: var(--font-semibold);
    color: var(--color-text-secondary);
    font-size: var(--text-xs);
  }

  .tag-item {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>

<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openModel(null)"> 新增 </el-button>
      <el-button
        type="danger"
        size="small"
        icon="el-icon-delete"
        :disabled="this.categoryIds.length == 0"
        @click="isDelete = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入分类名"
          style="width: 200px"
          @keyup.enter.native="searchCategories" />
        <el-button
          type="primary"
          size="small"
          icon="el-icon-search"
          style="margin-left: 1rem"
          @click="searchCategories">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="categories"
      @selection-change="selectionChange"
      v-loading="loading"
      class="category-table"
      :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="categoryName" label="分类名" align="center" min-width="150">
        <template slot-scope="scope">
          <div class="category-name">
            <i class="el-icon-folder-opened" style="margin-right: 8px; color: #409eff" />
            {{ scope.row.categoryName }}
          </div>
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
              @confirm="deleteCategory(scope.row.id)">
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
        <el-button type="primary" @click="deleteCategory(null)"> 确 定 </el-button>
      </div>
    </el-dialog>
    <el-dialog :visible.sync="addOrEdit" width="30%">
      <div class="dialog-title-container" slot="title" ref="categoryTitle" />
      <el-form label-width="80px" size="medium" :model="categoryForm">
        <el-form-item label="分类名">
          <el-input v-model="categoryForm.categoryName" style="width: 220px" />
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
    this.current = this.$store.state.pageState.category
    this.listCategories()
  },
  data: function () {
    return {
      isDelete: false,
      loading: true,
      addOrEdit: false,
      keywords: null,
      categoryIds: [],
      categories: [],
      categoryForm: {
        id: null,
        categoryName: ''
      },
      current: 1,
      size: 10,
      count: 0
    }
  },
  methods: {
    selectionChange(categories) {
      this.categoryIds = []
      categories.forEach((item) => {
        this.categoryIds.push(item.id)
      })
    },
    searchCategories() {
      this.current = 1
      this.listCategories()
    },
    sizeChange(size) {
      this.size = size
      this.listCategories()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateCategoryPageState', current)
      this.listCategories()
    },
    deleteCategory(id) {
      let param = {}
      if (id == null) {
        param = { data: this.categoryIds }
      } else {
        param = { data: [id] }
      }
      this.axios.delete('/api/admin/categories', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listCategories()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.isDelete = false
      })
    },
    listCategories() {
      this.axios
        .get('/api/admin/categories', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.categories = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    },
    openModel(category) {
      if (category != null) {
        this.categoryForm = JSON.parse(JSON.stringify(category))
        this.$nextTick(() => {
          if (this.$refs.categoryTitle) {
            this.$refs.categoryTitle.innerHTML = '修改分类'
          }
        })
      } else {
        this.categoryForm.id = null
        this.categoryForm.categoryName = ''
        this.$nextTick(() => {
          if (this.$refs.categoryTitle) {
            this.$refs.categoryTitle.innerHTML = '添加分类'
          }
        })
      }
      this.addOrEdit = true
    },
    addOrEditCategory() {
      if (this.categoryForm.categoryName.trim() == '') {
        this.$message.error('分类名不能为空')
        return false
      }
      this.axios.post('/api/admin/categories', this.categoryForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listCategories()
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
/* 操作区域 */
.operation-container {
  margin-top: 1.5rem;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

/* 分类表格 */
.category-table {
  margin-top: 20px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.category-table ::v-deep .el-table__body tr:hover > td {
  background-color: #f5f7fa !important;
}

/* 分类名称 */
.category-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.category-name:hover {
  color: #409eff;
}

/* 创建时间 */
.create-time {
  font-size: 13px;
  color: #909399;
}

.create-time i {
  margin-right: 4px;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
}

.action-buttons .el-button {
  transition: all 0.3s ease;
}

.action-buttons .el-button:hover {
  transform: translateY(-2px);
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  font-weight: bold;
  font-size: 16px;
}

.dialog-title-container i {
  font-size: 1.5rem;
  margin-right: 0.5rem;
}

/* 表单优化 */
.el-input ::v-deep .el-input__inner {
  border-radius: 20px;
}

.el-button {
  border-radius: 20px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

/* 分页 */
.pagination-container {
  float: right;
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
}

/* 加载动画 */
.category-table ::v-deep .el-loading-mask {
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.9);
}

/* 表格动画 */
.category-table ::v-deep .el-table__body tr {
  transition: all 0.3s ease;
}

.category-table ::v-deep .el-table__row {
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
</style>

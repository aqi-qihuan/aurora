<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button
        type="danger"
        size="small"
        icon="el-icon-delete"
        :disabled="this.logIds.length == 0"
        @click="isDelete = true">
        批量删除
      </el-button>
      <div style="margin-left: auto">
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入模块名或描述"
          style="width: 200px"
          @keyup.enter.native="searchLogs" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="searchLogs">
          搜索
        </el-button>
      </div>
    </div>
    <el-table @selection-change="selectionChange" v-loading="loading" :data="logs">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="optModule" label="系统模块" align="center" width="120" />
      <el-table-column width="100" prop="optType" label="操作类型" align="center" />
      <el-table-column prop="optDesc" label="操作描述" align="center" width="150" />
      <el-table-column prop="requestMethod" label="请求方式" align="center" width="100">
        <template slot-scope="scope" v-if="scope.row.requestMethod">
          <el-tag :type="tagType(scope.row.requestMethod)">
            {{ scope.row.requestMethod }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="操作人员" align="center" />
      <el-table-column prop="ipAddress" label="登录ip" align="center" width="130" />
      <el-table-column prop="ipSource" label="登录地址" align="center" width="150" />
      <el-table-column prop="createTime" label="操作日期" align="center" width="190">
        <template slot-scope="scope">
          <i class="el-icon-time" style="margin-right: 5px" />
          {{ scope.row.createTime | dateTime }}
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template slot-scope="scope">
          <el-button size="mini" type="text" slot="reference" @click="check(scope.row)">
            <i class="el-icon-view" /> 查看
          </el-button>
          <el-popconfirm title="确定删除吗？" style="margin-left: 10px" @confirm="deleteLog(scope.row.id)">
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
    <el-dialog :visible.sync="isCheck" width="40%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-more" />详细信息</div>
      <el-form ref="form" :model="optLog" label-width="100px" size="mini">
        <el-form-item label="操作模块：">
          {{ optLog.optModule }}
        </el-form-item>
        <el-form-item label="请求地址：">
          {{ optLog.optUri }}
        </el-form-item>
        <el-form-item label="请求方式：">
          <el-tag :type="tagType(optLog.requestMethod)">
            {{ optLog.requestMethod }}
          </el-tag>
        </el-form-item>
        <el-form-item label="操作方法：">
          {{ optLog.optMethod }}
        </el-form-item>
        <el-form-item label="请求参数：">
          {{ optLog.requestParam }}
        </el-form-item>
        <el-form-item label="返回数据：">
          {{ optLog.responseData }}
        </el-form-item>
        <el-form-item label="操作人员：">
          {{ optLog.nickname }}
        </el-form-item>
      </el-form>
    </el-dialog>
    <el-dialog :visible.sync="isDelete" width="30%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-warning" style="color: #ff9900" />提示</div>
      <div style="font-size: 1rem">是否删除选中项？</div>
      <div slot="footer">
        <el-button @click="isDelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteLog(null)"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.current = this.$store.state.pageState.operationLog
    this.listLogs()
  },
  data() {
    return {
      loading: true,
      logs: [],
      logIds: [],
      keywords: null,
      current: 1,
      size: 10,
      count: 0,
      isCheck: false,
      isDelete: false,
      optLog: {}
    }
  },
  methods: {
    selectionChange(logs) {
      this.logIds = []
      logs.forEach((item) => {
        this.logIds.push(item.id)
      })
    },
    searchLogs() {
      this.current = 1
      this.$store.commit('updateOperationLogPageState', this.current)
      this.listLogs()
    },
    sizeChange(size) {
      this.size = size
      this.listLogs()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateOperationLogPageState', current)
      this.listLogs()
    },
    listLogs() {
      this.axios
        .get('/api/admin/operation/logs', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.logs = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    },
    deleteLog(id) {
      var param = {}
      if (id != null) {
        param = { data: [id] }
      } else {
        param = { data: this.logIds }
      }
      this.axios.delete('/api/admin/operation/logs', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listLogs()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.isDelete = false
      })
    },
    check(optLog) {
      this.optLog = JSON.parse(JSON.stringify(optLog))
      this.isCheck = true
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
/* ==================== Operation Log Page Modern Styles ====================
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

/* 日志表格 - 现代化数据表格 */
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
  color: var(--color-primary);
}

/* 详情表单 */
.el-form-item {
  margin-bottom: var(--space-4);
}

.el-form-item__label {
  font-weight: var(--font-semibold);
  color: var(--color-text);
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
}
</style>

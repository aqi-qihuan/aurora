<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="review-menu">
      <span>状态</span>
      <span @click="changeReview(null)" :class="isReview == null ? 'active-review' : 'review'"> 全部 </span>
      <span @click="changeReview(1)" :class="isReview == 1 ? 'active-review' : 'review'"> 正常 </span>
      <span @click="changeReview(0)" :class="isReview == 0 ? 'active-review' : 'review'"> 审核中 </span>
    </div>
    <div class="operation-container">
      <el-button
        type="danger"
        size="small"
        icon="el-icon-delete"
        :disabled="commentIds.length == 0"
        @click="remove = true">
        批量删除
      </el-button>
      <el-button
        type="success"
        size="small"
        icon="el-icon-success"
        :disabled="commentIds.length == 0"
        @click="updateCommentReview(null)">
        批量通过
      </el-button>
      <div style="margin-left: auto">
        <el-select clearable v-model="type" placeholder="请选择来源" size="small" style="margin-right: 1rem">
          <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入用户昵称"
          style="width: 200px"
          @keyup.enter.native="searchComments" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="searchComments">
          搜索
        </el-button>
      </div>
    </div>
    <el-table
      border
      :data="comments"
      @selection-change="selectionChange"
      v-loading="loading"
      class="comment-table"
      :header-cell-style="{ background: '#f5f7fa', color: '#606266', fontWeight: '600' }">
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column prop="avatar" label="头像" align="center" width="100">
        <template slot-scope="scope">
          <el-avatar :size="40" :src="scope.row.avatar" />
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="评论人" align="center" width="120">
        <template slot-scope="scope">
          <div class="nickname">
            <i class="el-icon-user" style="margin-right: 5px; color: #409eff" />
            {{ scope.row.nickname }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="replyNickname" label="回复人" align="center" width="120">
        <template slot-scope="scope">
          <div class="reply-nickname">
            <i v-if="scope.row.replyNickname" class="el-icon-chat-line-round" style="margin-right: 5px" />
            {{ scope.row.replyNickname || '无' }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="articleTitle" label="文章标题" align="center" min-width="180">
        <template slot-scope="scope">
          <el-tooltip :content="scope.row.articleTitle" placement="top" :disabled="!scope.row.articleTitle || scope.row.articleTitle.length <= 20">
            <div class="article-title">
              <i class="el-icon-document" style="margin-right: 5px" />
              {{ scope.row.articleTitle || '无' }}
            </div>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column prop="commentContent" label="评论内容" align="center" min-width="200">
        <template slot-scope="scope">
          <div class="comment-content" v-html="scope.row.commentContent" />
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="评论时间" width="160" align="center" sortable>
        <template slot-scope="scope">
          <div class="create-time">
            <i class="el-icon-time" />
            {{ scope.row.createTime | date }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="isReview" label="状态" width="100" align="center">
        <template slot-scope="scope">
          <el-tag
            v-if="scope.row.isReview == 0"
            type="warning"
            size="small"
            effect="plain">
            <i class="el-icon-loading" /> 审核中
          </el-tag>
          <el-tag
            v-if="scope.row.isReview == 1"
            type="success"
            size="small"
            effect="plain">
            <i class="el-icon-check" /> 正常
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="来源" align="center" width="100">
        <template slot-scope="scope">
          <el-tag
            :type="getSourceType(scope.row.type).tagType"
            size="small"
            effect="plain">
            <i :class="getSourceType(scope.row.type).icon" />
            {{ getSourceType(scope.row.type).name }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="180" fixed="right">
        <template slot-scope="scope">
          <div class="action-buttons">
            <el-button
              v-if="scope.row.isReview == 0"
              type="success"
              size="mini"
              icon="el-icon-check"
              circle
              @click="updateCommentReview(scope.row.id)" />
            <el-popconfirm
              title="确定删除吗？"
              @confirm="deleteComments(scope.row.id)">
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
    <el-dialog :visible.sync="remove" width="30%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-warning" style="color: #ff9900" />提示</div>
      <div style="font-size: 1rem">是否彻底删除选中项？</div>
      <div slot="footer">
        <el-button @click="remove = false">取 消</el-button>
        <el-button type="primary" @click="deleteComments(null)"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.current = this.$store.state.pageState.comment
    this.listComments()
  },
  data: function () {
    return {
      loading: true,
      remove: false,
      options: [
        {
          value: 1,
          label: '文章'
        },
        {
          value: 2,
          label: '留言'
        },
        {
          value: 3,
          label: '关于我'
        },
        {
          value: 4,
          label: '友链'
        },
        {
          value: 5,
          label: '说说'
        }
      ],
      comments: [],
      commentIds: [],
      type: null,
      keywords: null,
      isReview: null,
      current: 1,
      size: 10,
      count: 0
    }
  },
  methods: {
    selectionChange(comments) {
      this.commentIds = []
      comments.forEach((item) => {
        this.commentIds.push(item.id)
      })
    },
    searchComments() {
      this.current = 1
      this.listComments()
    },
    sizeChange(size) {
      this.size = size
      this.listComments()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateCommentPageState', current)
      this.listComments()
    },
    changeReview(review) {
      this.current = 1
      this.isReview = review
    },
    updateCommentReview(id) {
      let param = {}
      if (id != null) {
        param.ids = [id]
      } else {
        param.ids = this.commentIds
      }
      param.isReview = 1
      this.axios.put('/api/admin/comments/review', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listComments()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
      })
    },
    deleteComments(id) {
      var param = {}
      if (id == null) {
        param = { data: this.commentIds }
      } else {
        param = { data: [id] }
      }
      this.axios.delete('/api/admin/comments', param).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listComments()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.remove = false
      })
      .catch(error => {
        this.$message.error('删除评论失败')
        console.error('API Error:', error)
      })
    },
    listComments() {
      this.axios
        .get('/api/admin/comments', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords,
            type: this.type,
            isReview: this.isReview
          }
        })
        .then(({ data }) => {
          this.comments = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    }
  },
  computed: {
    getSourceType() {
      return function (type) {
        const types = {
          1: { name: '文章', tagType: 'primary', icon: 'el-icon-document' },
          2: { name: '留言', tagType: 'danger', icon: 'el-icon-chat-line-round' },
          3: { name: '关于我', tagType: 'success', icon: 'el-icon-user' },
          4: { name: '友链', tagType: 'warning', icon: 'el-icon-link' },
          5: { name: '说说', tagType: 'info', icon: 'el-icon-edit-outline' }
        }
        return types[type] || { name: '未知', tagType: 'info', icon: 'el-icon-question' }
      }
    }
  },
  watch: {
    isReview() {
      this.current = 1
      this.listComments()
    },
    type() {
      this.current = 1
      this.listComments()
    }
  }
}
</script>

<style scoped>
/* ==================== Comment Page Modern Styles ====================
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

/* 审核菜单 - 现代化标签页样式 */
.review-menu {
  font-size: var(--text-sm);
  margin-top: var(--space-4);
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  padding: var(--space-3) 0;
  border-bottom: 2px solid var(--color-border);
  gap: var(--space-2);
}

.review-menu > span:first-child {
  font-weight: var(--font-semibold);
  color: var(--color-text);
  margin-right: var(--space-4);
}

.review-menu span {
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-full);
  transition: all var(--duration-base) var(--ease-out);
  position: relative;
  cursor: pointer;
}

.review {
  color: var(--color-text-secondary);
  background: transparent;
}

.review:hover {
  color: var(--color-primary);
  background: var(--color-primary-50);
}

.active-review {
  color: var(--color-primary);
  font-weight: var(--font-semibold);
  background: var(--color-primary-50);
}

.active-review::after {
  content: '';
  position: absolute;
  bottom: -11px;
  left: 50%;
  transform: translateX(-50%);
  width: 24px;
  height: 3px;
  background: var(--color-primary);
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

.operation-container .el-button--danger {
  background: linear-gradient(135deg, var(--color-error) 0%, #f87171 100%);
  border: none;
}

.operation-container .el-button--success {
  background: linear-gradient(135deg, var(--color-success) 0%, #34d399 100%);
  border: none;
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

.operation-container .el-select,
.operation-container .el-input {
  width: 180px;
}

.operation-container .el-input ::v-deep .el-input__inner,
.operation-container .el-select ::v-deep .el-input__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-input ::v-deep .el-input__inner:focus,
.operation-container .el-select ::v-deep .el-input__inner:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 评论表格 - 现代化数据表格 */
.comment-table {
  margin-top: var(--space-6);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  background: var(--color-bg-card);
}

.comment-table ::v-deep .el-table__header-wrapper {
  background: var(--color-bg-hover);
}

.comment-table ::v-deep .el-table__header th {
  background: var(--color-bg-hover) !important;
  color: var(--color-text);
  font-weight: var(--font-semibold);
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: var(--space-3) var(--space-4) !important;
  border-bottom: 1px solid var(--color-border);
}

.comment-table ::v-deep .el-table__body td {
  padding: var(--space-4) !important;
  border-bottom: 1px solid var(--color-border-light);
}

.comment-table ::v-deep .el-table__body tr {
  transition: all var(--duration-fast) var(--ease-out);
}

.comment-table ::v-deep .el-table__body tr:hover > td {
  background-color: var(--color-primary-50) !important;
}

.comment-table ::v-deep .el-table__row {
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

/* 头像 */
.comment-table ::v-deep .el-avatar {
  border: 2px solid var(--color-border);
  box-shadow: var(--shadow-sm);
  transition: all var(--duration-base) var(--ease-out);
}

.comment-table ::v-deep .el-avatar:hover {
  transform: scale(1.15);
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}

/* 昵称 */
.nickname {
  font-size: var(--text-sm);
  color: var(--color-text);
  font-weight: var(--font-semibold);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
}

.nickname i {
  color: var(--color-primary);
}

.reply-nickname {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
}

.reply-nickname i {
  color: var(--color-secondary);
}

/* 文章标题 */
.article-title {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  transition: color var(--duration-fast) var(--ease-out);
}

.article-title i {
  color: var(--color-primary);
  flex-shrink: 0;
}

.article-title:hover {
  color: var(--color-primary);
}

/* 评论内容 */
.comment-content {
  display: inline-block;
  max-width: 280px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--text-sm);
  color: var(--color-text);
  line-height: var(--leading-relaxed);
  padding: var(--space-2);
  background: var(--color-bg-hover);
  border-radius: var(--radius-md);
  transition: all var(--duration-base) var(--ease-out);
}

.comment-content:hover {
  white-space: normal;
  overflow: visible;
  max-width: 400px;
  background: var(--color-bg-card);
  box-shadow: var(--shadow-lg);
  z-index: 10;
  position: relative;
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

/* 状态标签 */
.comment-table ::v-deep .el-tag {
  border-radius: var(--radius-base);
  font-weight: var(--font-medium);
  font-size: var(--text-xs);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
}

.comment-table ::v-deep .el-tag:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.comment-table ::v-deep .el-tag--warning {
  background: var(--color-warning-light);
  border-color: var(--color-warning);
  color: var(--color-warning);
}

.comment-table ::v-deep .el-tag--success {
  background: var(--color-success-light);
  border-color: var(--color-success);
  color: var(--color-success);
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

.action-buttons .el-button--success {
  background: linear-gradient(135deg, var(--color-success) 0%, #34d399 100%);
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

/* 加载动画 */
.comment-table ::v-deep .el-loading-mask {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .comment-content {
  background: var(--color-bg-active);
}

[data-theme="dark"] .comment-content:hover {
  background: var(--color-bg-card);
}

[data-theme="dark"] .comment-table ::v-deep .el-loading-mask {
  background: rgba(15, 23, 42, 0.9);
}

/* ==================== Responsive ==================== */
@media (max-width: 768px) {
  .title {
    font-size: var(--text-xl);
  }

  .review-menu {
    flex-wrap: wrap;
    gap: var(--space-2);
  }

  .review-menu span {
    padding: var(--space-1) var(--space-3);
    font-size: var(--text-xs);
  }

  .operation-container {
    flex-direction: column;
    align-items: stretch;
  }

  .operation-container > div:last-child {
    margin-left: 0;
    flex-direction: column;
    width: 100%;
  }

  .operation-container .el-select,
  .operation-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .comment-content {
    max-width: 150px;
  }

  .article-title {
    max-width: 120px;
  }

  .action-buttons {
    flex-direction: row;
  }

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .review-menu > span:first-child {
    width: 100%;
    margin-bottom: var(--space-2);
  }

  .comment-table ::v-deep .el-table__header {
    display: none;
  }

  .comment-table ::v-deep .el-table__row {
    display: flex;
    flex-direction: column;
    padding: var(--space-4);
    margin-bottom: var(--space-3);
    background: var(--color-bg-card);
    border-radius: var(--radius-lg);
    border: 1px solid var(--color-border);
    box-shadow: var(--shadow-sm);
  }

  .comment-table ::v-deep .el-table__row td {
    border: none;
    padding: var(--space-2) 0 !important;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .comment-table ::v-deep .el-table__row td::before {
    content: attr(data-label);
    font-weight: var(--font-semibold);
    color: var(--color-text-secondary);
    font-size: var(--text-xs);
  }

  .comment-content {
    max-width: none;
    white-space: normal;
  }

  .article-title {
    max-width: none;
  }
}
</style>

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
/* 评论内容 */
.comment-content {
  display: inline-block;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.comment-content:hover {
  white-space: normal;
  overflow: visible;
}

/* 操作区域 */
.operation-container {
  margin-top: 1.5rem;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

/* 审核菜单 */
.review-menu {
  font-size: 14px;
  margin-top: 40px;
  color: #999;
  display: flex;
  align-items: center;
  padding: 10px 0;
  border-bottom: 2px solid #f0f0f0;
}

.review-menu span {
  margin-right: 24px;
  padding: 8px 16px;
  border-radius: 20px;
  transition: all 0.3s ease;
  position: relative;
}

.review {
  cursor: pointer;
  color: #999;
}

.review:hover {
  color: #409eff;
  background: rgba(64, 158, 255, 0.05);
}

.active-review {
  cursor: pointer;
  color: #409eff;
  font-weight: bold;
  background: rgba(64, 158, 255, 0.1);
}

.active-review::after {
  content: '';
  position: absolute;
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
  width: 30px;
  height: 3px;
  background: #409eff;
  border-radius: 2px;
}

/* 评论表格 */
.comment-table {
  margin-top: 20px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.comment-table ::v-deep .el-table__body tr:hover > td {
  background-color: #f5f7fa !important;
}

/* 昵称 */
.nickname {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.reply-nickname {
  font-size: 13px;
  color: #909399;
}

/* 文章标题 */
.article-title {
  font-size: 13px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.article-title:hover {
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

/* 选择器和输入框 */
.el-select ::v-deep .el-input__inner,
.el-input ::v-deep .el-input__inner {
  border-radius: 20px;
}

/* 按钮优化 */
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
.comment-table ::v-deep .el-loading-mask {
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.9);
}

/* 表格动画 */
.comment-table ::v-deep .el-table__body tr {
  transition: all 0.3s ease;
}

.comment-table ::v-deep .el-table__row {
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

/* 头像动画 */
.el-avatar {
  transition: transform 0.3s ease;
}

.el-avatar:hover {
  transform: scale(1.1);
}

/* 标签动画 */
.el-tag {
  transition: all 0.3s ease;
}

.el-tag:hover {
  transform: translateY(-2px);
}
</style>

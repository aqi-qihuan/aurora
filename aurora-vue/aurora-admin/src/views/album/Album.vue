<template>
  <el-card class="main-card">
    <div class="title">{{ this.$route.name }}</div>
    <div class="operation-container">
      <el-button type="primary" size="small" icon="el-icon-plus" @click="openModel(null)"> 新建相册 </el-button>
      <div style="margin-left: auto">
        <el-button type="text" size="small" icon="el-icon-delete" style="margin-right: 1rem" @click="checkDelete">
          回收站
        </el-button>
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入相册名"
          style="width: 200px"
          @keyup.enter.native="searchAlbums" />
        <el-button type="primary" size="small" icon="el-icon-search" style="margin-left: 1rem" @click="searchAlbums">
          搜索
        </el-button>
      </div>
    </div>
    <el-row class="album-container" :gutter="12" v-loading="loading">
      <el-empty v-if="albums == null" description="暂无相册" />
      <el-col v-for="item of albums" :key="item.id" :md="6">
        <div class="album-item" @click="checkPhoto(item)">
          <div class="album-opreation">
            <el-dropdown @command="handleCommand">
              <i class="el-icon-more" style="color: #fff" />
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item :command="'update' + JSON.stringify(item)">
                  <i class="el-icon-edit" />编辑
                </el-dropdown-item>
                <el-dropdown-item :command="'delete' + item.id"> <i class="el-icon-delete" />删除 </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
          <div class="album-photo-count">
            <div>{{ item.photoCount }}</div>
            <i v-if="item.status == 2" class="iconfont el-icon-mymima" />
          </div>
          <el-image fit="cover" class="album-cover" :src="item.albumCover" />
          <div class="album-name">{{ item.albumName }}</div>
        </div>
      </el-col>
    </el-row>
    <el-pagination
      :hide-on-single-page="true"
      class="pagination-container"
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      layout="prev, pager, next" />
    <el-dialog :visible.sync="addOrEdit" width="35%" top="10vh">
      <div class="dialog-title-container" slot="title" ref="albumTitle" />
      <el-form label-width="80px" size="medium" :model="albumForum">
        <el-form-item label="相册名称">
          <el-input style="width: 220px" v-model="albumForum.albumName" />
        </el-form-item>
        <el-form-item label="相册描述">
          <el-input style="width: 220px" v-model="albumForum.albumDesc" />
        </el-form-item>
        <el-form-item label="相册封面">
          <el-upload
            class="upload-cover"
            drag
            :headers="headers"
            :before-upload="beforeUpload"
            action="/api/admin/photos/albums/upload"
            multiple
            :on-success="uploadCover">
            <i class="el-icon-upload" v-if="albumForum.albumCover == ''" />
            <div class="el-upload__text" v-if="albumForum.albumCover == ''">将文件拖到此处，或<em>点击上传</em></div>
            <img v-else :src="albumForum.albumCover" width="360px" height="180px" />
          </el-upload>
        </el-form-item>
        <el-form-item label="发布形式">
          <el-radio-group v-model="albumForum.status">
            <el-radio :label="1">公开</el-radio>
            <el-radio :label="2">私密</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="addOrEdit = false">取 消</el-button>
        <el-button type="primary" @click="addOrEditAlbum"> 确 定 </el-button>
      </div>
    </el-dialog>
    <el-dialog :visible.sync="isdelete" width="30%">
      <div class="dialog-title-container" slot="title"><i class="el-icon-warning" style="color: #ff9900" />提示</div>
      <div style="font-size: 1rem">是否删除该相册？</div>
      <div slot="footer">
        <el-button @click="isdelete = false">取 消</el-button>
        <el-button type="primary" @click="deleteAlbum"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
import * as imageConversion from 'image-conversion'
export default {
  created() {
    this.listAlbums()
  },
  data: function () {
    return {
      keywords: '',
      loading: true,
      isdelete: false,
      addOrEdit: false,
      albumForum: {
        id: null,
        albumName: '',
        albumDesc: '',
        albumCover: '',
        status: 1
      },
      albums: [],
      current: 1,
      size: 8,
      count: 0,
      headers: { Authorization: 'Bearer ' + sessionStorage.getItem('token') }
    }
  },
  methods: {
    openModel(item) {
      if (item) {
        console.log(item)
        this.albumForum = JSON.parse(item)
        this.$refs.albumTitle.innerHTML = '修改相册'
      } else {
        this.albumForum = {
          id: null,
          albumName: '',
          albumDesc: '',
          albumCover: '',
          status: 1
        }
        this.$refs.albumTitle.innerHTML = '新建相册'
      }
      this.$nextTick(() => {
        if (this.$refs.albumTitle) {
          this.$refs.albumTitle.innerHTML = item ? '修改相册' : '新建相册'
        }
      })
      this.addOrEdit = true
    },
    checkPhoto(item) {
      this.$router.push({ path: '/albums/' + item.id })
    },
    checkDelete() {
      this.$router.push({ path: '/photos/delete' })
    },
    listAlbums() {
      this.axios
        .get('/api/admin/photos/albums', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords
          }
        })
        .then(({ data }) => {
          this.albums = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    },
    addOrEditAlbum() {
      if (this.albumForum.albumName.trim() == '') {
        this.$message.error('相册名称不能为空')
        return false
      }
      if (this.albumForum.albumDesc.trim() == '') {
        this.$message.error('相册描述不能为空')
        return false
      }
      if (this.albumForum.albumCover == null) {
        this.$message.error('相册封面不能为空')
        return false
      }
      this.axios.post('/api/admin/photos/albums', this.albumForum).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listAlbums()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
      })
      this.addOrEdit = false
    },
    uploadCover(response) {
      this.albumForum.albumCover = response.data
    },
    beforeUpload(file) {
      return new Promise((resolve) => {
        if (file.size / 1024 < this.config.UPLOAD_SIZE) {
          resolve(file)
        }
        imageConversion.compressAccurately(file, this.config.UPLOAD_SIZE).then((res) => {
          resolve(res)
        })
      })
    },
    handleCommand(command) {
      const type = command.substring(0, 6)
      const data = command.substring(6)
      if (type == 'delete') {
        this.albumForum.id = data
        this.isdelete = true
      } else {
        this.openModel(data)
      }
    },
    deleteAlbum() {
      this.axios.delete('/api/admin/photos/albums/' + this.albumForum.id).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listAlbums()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.isdelete = false
      })
      .catch(error => {
        this.$message.error('删除相册失败')
        console.error('API Error:', error)
      })
    },
    searchAlbums() {
      this.current = 1
      this.listAlbums()
    },
    sizeChange(size) {
      this.size = size
      this.listAlbums()
    },
    currentChange(current) {
      this.current = current
      this.listAlbums()
    }
  }
}
</script>

<style scoped>
/* ==================== Album Page Modern Styles ====================
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

.operation-container .el-button--text {
  color: var(--color-text-secondary);
  transition: all var(--duration-fast) var(--ease-out);
}

.operation-container .el-button--text:hover {
  color: var(--color-error);
  background: var(--color-error-light);
}

/* 相册容器 */
.album-container {
  margin-top: var(--space-6);
}

/* 相册项 - 现代化卡片 */
.album-item {
  position: relative;
  cursor: pointer;
  margin-bottom: var(--space-4);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
  transition: all var(--duration-base) var(--ease-out);
  background: var(--color-bg-card);
}

.album-item:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover);
}

/* 相册封面 */
.album-cover {
  position: relative;
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  width: 100%;
  height: 170px;
  overflow: hidden;
}

.album-cover::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.3) 0%, transparent 50%, rgba(0, 0, 0, 0.5) 100%);
  z-index: 1;
  transition: all var(--duration-base) var(--ease-out);
}

.album-item:hover .album-cover::before {
  background: linear-gradient(to bottom, rgba(0, 0, 0, 0.4) 0%, transparent 50%, rgba(0, 0, 0, 0.6) 100%);
}

.album-cover ::v-deep img {
  transition: all var(--duration-slow) var(--ease-out);
}

.album-item:hover .album-cover ::v-deep img {
  transform: scale(1.05);
}

/* 照片数量 */
.album-photo-count {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: var(--text-xl);
  font-weight: var(--font-bold);
  z-index: 10;
  position: absolute;
  left: var(--space-3);
  right: var(--space-3);
  bottom: var(--space-12);
  color: #fff;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.album-photo-count i {
  font-size: var(--text-lg);
  opacity: 0.9;
}

/* 相册名称 */
.album-name {
  text-align: center;
  padding: var(--space-3);
  font-weight: var(--font-semibold);
  color: var(--color-text);
  font-size: var(--text-sm);
  background: var(--color-bg-card);
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 操作下拉菜单 */
.album-opreation {
  position: absolute;
  z-index: 10;
  top: var(--space-2);
  right: var(--space-2);
}

.album-opreation .el-dropdown {
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(4px);
  border-radius: var(--radius-base);
  padding: var(--space-1) var(--space-2);
  transition: all var(--duration-fast) var(--ease-out);
}

.album-opreation .el-dropdown:hover {
  background: rgba(255, 255, 255, 0.4);
}

.album-opreation i {
  font-size: var(--text-lg);
  color: #fff;
  cursor: pointer;
}

/* 空状态 */
.el-empty {
  padding: var(--space-12) 0;
}

/* 分页 */
.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: var(--space-8);
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

/* 上传组件 */
.upload-cover ::v-deep .el-upload {
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-hover);
  transition: all var(--duration-fast) var(--ease-out);
}

.upload-cover ::v-deep .el-upload:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-50);
}

.upload-cover ::v-deep .el-upload-dragger {
  width: 360px;
  height: 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
}

.upload-cover ::v-deep .el-icon-upload {
  font-size: var(--text-3xl);
  color: var(--color-text-muted);
  margin: 0 0 var(--space-2);
}

.upload-cover ::v-deep .el-upload__text {
  color: var(--color-text-secondary);
}

.upload-cover ::v-deep .el-upload__text em {
  color: var(--color-primary);
  font-style: normal;
}

/* 单选按钮组 */
.el-radio-group {
  display: flex;
  gap: var(--space-6);
}

.el-radio {
  margin-right: 0;
}

/* 加载动画 */
.album-container ::v-deep .el-loading-mask {
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .operation-container {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .album-name {
  background: var(--color-bg-card);
}

[data-theme="dark"] .album-container ::v-deep .el-loading-mask {
  background: rgba(15, 23, 42, 0.9);
}

[data-theme="dark"] .upload-cover ::v-deep .el-upload {
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
    flex-direction: column;
  }

  .operation-container .el-input {
    width: 100%;
  }

  .operation-container .el-button {
    width: 100%;
  }

  .album-cover {
    height: 140px;
  }

  .upload-cover ::v-deep .el-upload-dragger {
    width: 100%;
    height: 150px;
  }
}

@media (max-width: 480px) {
  .album-cover {
    height: 120px;
  }

  .album-photo-count {
    font-size: var(--text-lg);
    bottom: var(--space-10);
  }

  .album-name {
    font-size: var(--text-xs);
    padding: var(--space-2);
  }

  .el-dialog {
    width: 90% !important;
  }
}
</style>

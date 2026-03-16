<template>
  <el-card class="main-card">
    <el-tabs v-model="activeName">
      <el-tab-pane label="网站信息" name="info">
        <el-form label-width="100px" :model="websiteConfigForm" label-position="left">
          <el-form-item label="作者头像">
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleAuthorAvatarSuccess">
              <img v-if="websiteConfigForm.authorAvatar" :src="websiteConfigForm.authorAvatar" class="avatar" />
              <i v-else class="el-icon-plus avatar-uploader-icon" />
            </el-upload>
          </el-form-item>
          <el-form-item label="网站logo">
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleLogoSuccess">
              <img v-if="websiteConfigForm.logo" :src="websiteConfigForm.logo" class="avatar" />
              <i v-else class="el-icon-plus avatar-uploader-icon" />
            </el-upload>
          </el-form-item>
          <el-form-item label="favicon">
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleFaviconSuccess">
              <img v-if="websiteConfigForm.favicon" :src="websiteConfigForm.favicon" class="avatar" />
              <i v-else class="el-icon-plus avatar-uploader-icon" />
            </el-upload>
          </el-form-item>
          <el-form-item label="网站名称">
            <el-input v-model="websiteConfigForm.name" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="网站英文名称">
            <el-input v-model="websiteConfigForm.englishName" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="网站作者">
            <el-input v-model="websiteConfigForm.author" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="网页标题">
            <el-input v-model="websiteConfigForm.websiteTitle" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="作者介绍">
            <el-input v-model="websiteConfigForm.authorIntro" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="多语言">
            <el-radio-group v-model="websiteConfigForm.multiLanguage">
              <el-radio :label="0">关闭</el-radio>
              <el-radio :label="1">开启</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="网站创建日期">
            <el-date-picker
              style="width: 400px"
              value-format="yyyy-MM-dd"
              v-model="websiteConfigForm.websiteCreateTime"
              type="date"
              placeholder="选择日期" />
          </el-form-item>
          <el-form-item label="网站公告">
            <el-input
              v-model="websiteConfigForm.notice"
              placeholder="请输入公告内容"
              style="width: 400px"
              type="textarea"
              :rows="5" />
          </el-form-item>
          <el-form-item label="工信部备案号">
            <el-input v-model="websiteConfigForm.beianNumber" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="公安部备案号">
            <el-input v-model="websiteConfigForm.gonganBeianNumber" size="small" style="width: 400px" />
          </el-form-item>
          <el-form-item label="qq登录">
            <el-radio-group v-model="websiteConfigForm.qqLogin">
              <el-radio :label="0">关闭</el-radio>
              <el-radio :label="1">开启</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-button type="primary" size="medium" style="margin-left: 6.3rem" @click="updateWebsiteConfig">
            修改
          </el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="社交信息" name="notice">
        tip:空白默认不显示
        <el-form label-width="70px" :model="websiteConfigForm">
          <el-form-item label="Github">
            <el-input v-model="websiteConfigForm.github" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="Gitee">
            <el-input v-model="websiteConfigForm.gitee" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="QQ">
            <el-input v-model="websiteConfigForm.qq" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="WeChat">
            <el-input v-model="websiteConfigForm.weChat" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="微博">
            <el-input v-model="websiteConfigForm.weibo" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="CSDN">
            <el-input v-model="websiteConfigForm.csdn" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="知乎">
            <el-input v-model="websiteConfigForm.zhihu" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="掘金">
            <el-input v-model="websiteConfigForm.juejin" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="twitter">
            <el-input v-model="websiteConfigForm.twitter" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-form-item label="stackoverflow">
            <el-input v-model="websiteConfigForm.stackoverflow" size="small" style="width: 400px; margin-right: 1rem" />
          </el-form-item>
          <el-button type="primary" size="medium" style="margin-left: 4.375rem" @click="updateWebsiteConfig">
            修改
          </el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="其他设置" name="password">
        <el-form label-width="120px" :model="websiteConfigForm" label-position="left">
          <el-row style="width: 600px">
            <el-col :md="12">
              <el-form-item label="用户头像">
                <el-upload
                  class="avatar-uploader"
                  action="/api/admin/config/images"
                  :headers="headers"
                  :show-file-list="false"
                  :before-upload="beforeUpload"
                  :on-success="handleUserAvatarSuccess">
                  <img v-if="websiteConfigForm.userAvatar" :src="websiteConfigForm.userAvatar" class="avatar" />
                  <i v-else class="el-icon-plus avatar-uploader-icon" />
                </el-upload>
              </el-form-item>
            </el-col>
            <el-col :md="12">
              <el-form-item label="游客头像">
                <el-upload
                  class="avatar-uploader"
                  action="/api/admin/config/images"
                  :headers="headers"
                  :show-file-list="false"
                  :before-upload="beforeUpload"
                  :on-success="handleTouristAvatarSuccess">
                  <img v-if="websiteConfigForm.touristAvatar" :src="websiteConfigForm.touristAvatar" class="avatar" />
                  <i v-else class="el-icon-plus avatar-uploader-icon" />
                </el-upload>
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="邮箱通知">
            <el-radio-group v-model="websiteConfigForm.isEmailNotice">
              <el-radio :label="1">开启</el-radio>
              <el-radio :label="0">关闭</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="评论审核">
            <el-radio-group v-model="websiteConfigForm.isCommentReview">
              <el-radio :label="1">开启</el-radio>
              <el-radio :label="0">关闭</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="打赏状态">
            <el-radio-group v-model="websiteConfigForm.isReward">
              <el-radio :label="1">开启</el-radio>
              <el-radio :label="0">关闭</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-row style="width: 600px" v-show="websiteConfigForm.isReward == 1">
            <el-col :md="12">
              <el-form-item label="微信收款码">
                <el-upload
                  class="avatar-uploader"
                  action="/api/admin/config/images"
                  :headers="headers"
                  :show-file-list="false"
                  :before-upload="beforeUpload"
                  :on-success="handleWeiXinSuccess">
                  <img v-if="websiteConfigForm.weiXinQRCode" :src="websiteConfigForm.weiXinQRCode" class="avatar" />
                  <i v-else class="el-icon-plus avatar-uploader-icon" />
                </el-upload>
              </el-form-item>
            </el-col>
            <el-col :md="12">
              <el-form-item label="支付宝收款码">
                <el-upload
                  class="avatar-uploader"
                  action="/api/admin/config/images"
                  :headers="headers"
                  :show-file-list="false"
                  :before-upload="beforeUpload"
                  :on-success="handleAlipaySuccess">
                  <img v-if="websiteConfigForm.alipayQRCode" :src="websiteConfigForm.alipayQRCode" class="avatar" />
                  <i v-else class="el-icon-plus avatar-uploader-icon" />
                </el-upload>
              </el-form-item>
            </el-col>
          </el-row>
          <el-button type="primary" size="medium" style="margin-left: 6.3rem" @click="updateWebsiteConfig">
            修改
          </el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script>
import * as imageConversion from 'image-conversion'

export default {
  created() {
    this.getWebsiteConfig()
  },
  data: function () {
    return {
      websiteConfigForm: {},
      activeName: 'info',
      headers: { Authorization: 'Bearer ' + sessionStorage.getItem('token') }
    }
  },
  methods: {
    getWebsiteConfig() {
      this.axios.get('/api/admin/website/config').then(({ data }) => {
        this.websiteConfigForm = data.data
      })
    },
    handleAuthorAvatarSuccess(response) {
      this.websiteConfigForm.authorAvatar = response.data
    },
    handleFaviconSuccess(response) {
      this.websiteConfigForm.favicon = response.data
    },
    handleLogoSuccess(response) {
      this.websiteConfigForm.logo = response.data
    },
    handleUserAvatarSuccess(response) {
      this.websiteConfigForm.userAvatar = response.data
    },
    handleTouristAvatarSuccess(response) {
      this.websiteConfigForm.touristAvatar = response.data
    },
    handleWeiXinSuccess(response) {
      this.websiteConfigForm.weiXinQRCode = response.data
    },
    handleAlipaySuccess(response) {
      this.websiteConfigForm.alipayQRCode = response.data
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
    updateWebsiteConfig() {
      this.axios.put('/api/admin/website/config', this.websiteConfigForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
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
/* ==================== Website Config Page Modern Styles ====================
 * 基于 UI/UX Pro Max 设计系统
 * 配色: Primary #2563EB, CTA #F97316
 */

/* 标签页样式 */
.el-tabs ::v-deep .el-tabs__header {
  margin-bottom: var(--space-6);
}

.el-tabs ::v-deep .el-tabs__nav-wrap::after {
  background: var(--color-border);
}

.el-tabs ::v-deep .el-tabs__item {
  font-weight: var(--font-medium);
  color: var(--color-text-secondary);
  transition: all var(--duration-fast) var(--ease-out);
  padding: 0 var(--space-6);
  height: 40px;
  line-height: 40px;
}

.el-tabs ::v-deep .el-tabs__item:hover {
  color: var(--color-primary);
}

.el-tabs ::v-deep .el-tabs__item.is-active {
  color: var(--color-primary);
  font-weight: var(--font-semibold);
}

.el-tabs ::v-deep .el-tabs__active-bar {
  background: linear-gradient(90deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  height: 3px;
  border-radius: var(--radius-full);
}

/* 表单样式 */
.el-form {
  max-width: 600px;
}

.el-form-item {
  margin-bottom: var(--space-6);
}

.el-form-item__label {
  font-weight: var(--font-semibold);
  color: var(--color-text);
  padding-right: var(--space-4);
}

/* 输入框样式 */
.el-input ::v-deep .el-input__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
  padding: 0 var(--space-4);
}

.el-input ::v-deep .el-input__inner:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

.el-input ::v-deep .el-input__inner:hover {
  border-color: var(--color-primary-light);
}

/* 文本域样式 */
.el-textarea ::v-deep .el-textarea__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
  background: var(--color-bg-card);
  transition: all var(--duration-fast) var(--ease-out);
  padding: var(--space-3) var(--space-4);
}

.el-textarea ::v-deep .el-textarea__inner:focus {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 日期选择器 */
.el-date-editor {
  width: 100%;
}

.el-date-editor ::v-deep .el-input__inner {
  border-radius: var(--radius-base);
  border-color: var(--color-border);
}

/* 单选按钮组 */
.el-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-6);
}

.el-radio {
  margin-right: 0;
  font-weight: var(--font-medium);
}

.el-radio ::v-deep .el-radio__input.is-checked .el-radio__inner {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

.el-radio ::v-deep .el-radio__input.is-checked + .el-radio__label {
  color: var(--color-primary);
}

/* 上传组件 */
.avatar-uploader ::v-deep .el-upload {
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  cursor: pointer;
  position: relative;
  overflow: hidden;
  background: var(--color-bg-hover);
  transition: all var(--duration-fast) var(--ease-out);
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-uploader ::v-deep .el-upload:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-50);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.avatar-uploader-icon {
  font-size: var(--text-2xl);
  color: var(--color-text-muted);
  transition: all var(--duration-fast) var(--ease-out);
}

.avatar-uploader ::v-deep .el-upload:hover .avatar-uploader-icon {
  color: var(--color-primary);
  transform: scale(1.1);
}

.avatar {
  width: 120px;
  height: 120px;
  display: block;
  object-fit: cover;
  border-radius: var(--radius-lg);
}

/* 提交按钮 */
.el-button--primary {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-primary-light) 100%);
  border: none;
  border-radius: var(--radius-base);
  font-weight: var(--font-semibold);
  padding: var(--space-3) var(--space-8);
  transition: all var(--duration-fast) var(--ease-out);
  box-shadow: var(--shadow-sm);
}

.el-button--primary:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  background: linear-gradient(135deg, var(--color-primary-light) 0%, var(--color-primary) 100%);
}

.el-button--primary:active {
  transform: translateY(0);
}

/* 提示文字 */
.el-tab-pane > div:first-child {
  color: var(--color-text-muted);
  font-size: var(--text-sm);
  margin-bottom: var(--space-4);
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-hover);
  border-radius: var(--radius-base);
  border-left: 3px solid var(--color-primary);
}

/* 两列布局 */
.el-row {
  margin-bottom: var(--space-4);
}

/* ==================== Dark Mode ==================== */
[data-theme="dark"] .el-tabs ::v-deep .el-tabs__nav-wrap::after {
  background: var(--color-border);
}

[data-theme="dark"] .el-input ::v-deep .el-input__inner,
[data-theme="dark"] .el-textarea ::v-deep .el-textarea__inner {
  background: var(--color-bg-card);
  border-color: var(--color-border);
  color: var(--color-text);
}

[data-theme="dark"] .avatar-uploader ::v-deep .el-upload {
  background: var(--color-bg-hover);
  border-color: var(--color-border);
}

[data-theme="dark"] .el-tab-pane > div:first-child {
  background: var(--color-bg-hover);
}

/* ==================== Responsive ==================== */
@media (max-width: 768px) {
  .el-form {
    max-width: 100%;
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
  .el-textarea,
  .el-date-editor {
    width: 100% !important;
  }

  .el-radio-group {
    flex-direction: column;
    gap: var(--space-2);
  }

  .el-row {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .el-col {
    width: 100% !important;
  }

  .el-button--primary {
    width: 100%;
    margin-left: 0 !important;
  }
}

@media (max-width: 480px) {
  .el-tabs ::v-deep .el-tabs__item {
    padding: 0 var(--space-3);
    font-size: var(--text-sm);
  }

  .avatar-uploader ::v-deep .el-upload {
    width: 100px;
    height: 100px;
  }

  .avatar {
    width: 100px;
    height: 100px;
  }

  .avatar-uploader-icon {
    font-size: var(--text-xl);
  }
}
</style>

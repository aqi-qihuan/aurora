<template>
  <div class="website-page">
    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon primary">
          <el-icon><Monitor /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">3</span>
          <span class="stat-label">配置分类</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon success">
          <el-icon><Link /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ socialCount }}</span>
          <span class="stat-label">社交链接</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon warning">
          <el-icon><Setting /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ activeFeatures }}</span>
          <span class="stat-label">已启用功能</span>
        </div>
      </div>
    </div>

    <!-- 主内容卡片 -->
    <el-card class="main-card">
      <!-- Tab 头部 -->
      <div class="tab-header">
        <div
          v-for="tab in tabs"
          :key="tab.key"
          class="tab-item"
          :class="{ active: activeName === tab.key }"
          @click="activeName = tab.key">
          <el-icon><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </div>
      </div>

      <!-- 网站信息 -->
      <div v-show="activeName === 'info'" class="tab-content">
        <div class="section-title">
          <el-icon><InfoFilled /></el-icon>
          基本信息
        </div>
        <div class="form-grid">
          <div class="form-group upload-group">
            <label class="form-label">作者头像</label>
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleAuthorAvatarSuccess">
              <img v-if="websiteConfigForm.authorAvatar" :src="websiteConfigForm.authorAvatar" class="avatar" />
              <div v-else class="avatar-placeholder">
                <el-icon><Plus /></el-icon>
                <span>上传头像</span>
              </div>
            </el-upload>
          </div>
          <div class="form-group upload-group">
            <label class="form-label">网站Logo</label>
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleLogoSuccess">
              <img v-if="websiteConfigForm.logo" :src="websiteConfigForm.logo" class="avatar" />
              <div v-else class="avatar-placeholder">
                <el-icon><Plus /></el-icon>
                <span>上传Logo</span>
              </div>
            </el-upload>
          </div>
          <div class="form-group upload-group">
            <label class="form-label">Favicon</label>
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleFaviconSuccess">
              <img v-if="websiteConfigForm.favicon" :src="websiteConfigForm.favicon" class="avatar" />
              <div v-else class="avatar-placeholder">
                <el-icon><Plus /></el-icon>
                <span>上传图标</span>
              </div>
            </el-upload>
          </div>
          <div class="form-group">
            <label class="form-label">网站名称</label>
            <el-input v-model="websiteConfigForm.name" placeholder="请输入网站名称" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">网站英文名称</label>
            <el-input v-model="websiteConfigForm.englishName" placeholder="请输入英文名称" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">网站作者</label>
            <el-input v-model="websiteConfigForm.author" placeholder="请输入作者名" class="form-input" />
          </div>
          <div class="form-group full-width">
            <label class="form-label">网页标题</label>
            <el-input v-model="websiteConfigForm.websiteTitle" placeholder="请输入网页标题" class="form-input" />
          </div>
          <div class="form-group full-width">
            <label class="form-label">作者介绍</label>
            <el-input v-model="websiteConfigForm.authorIntro" placeholder="请输入作者介绍" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">多语言</label>
            <div class="radio-card-group">
              <div class="radio-card" :class="{ active: websiteConfigForm.multiLanguage === 0 }" @click="websiteConfigForm.multiLanguage = 0">
                <el-icon><Close /></el-icon>
                <span>关闭</span>
              </div>
              <div class="radio-card" :class="{ active: websiteConfigForm.multiLanguage === 1 }" @click="websiteConfigForm.multiLanguage = 1">
                <el-icon><Check /></el-icon>
                <span>开启</span>
              </div>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">网站创建日期</label>
            <el-date-picker
              class="form-input"
              value-format="YYYY-MM-DD"
              v-model="websiteConfigForm.websiteCreateTime"
              type="date"
              placeholder="选择日期" />
          </div>
          <div class="form-group">
            <label class="form-label">QQ登录</label>
            <div class="radio-card-group">
              <div class="radio-card" :class="{ active: websiteConfigForm.qqLogin === 0 }" @click="websiteConfigForm.qqLogin = 0">
                <el-icon><Close /></el-icon>
                <span>关闭</span>
              </div>
              <div class="radio-card" :class="{ active: websiteConfigForm.qqLogin === 1 }" @click="websiteConfigForm.qqLogin = 1">
                <el-icon><Check /></el-icon>
                <span>开启</span>
              </div>
            </div>
          </div>
          <div class="form-group full-width">
            <label class="form-label">网站公告</label>
            <el-input
              v-model="websiteConfigForm.notice"
              placeholder="请输入公告内容"
              class="form-input"
              type="textarea"
              :rows="4" />
          </div>
          <div class="form-group">
            <label class="form-label">工信部备案号</label>
            <el-input v-model="websiteConfigForm.beianNumber" placeholder="请输入备案号" class="form-input" />
          </div>
          <div class="form-group">
            <label class="form-label">公安部备案号</label>
            <el-input v-model="websiteConfigForm.gonganBeianNumber" placeholder="请输入备案号" class="form-input" />
          </div>
        </div>
        <div class="save-bar">
          <button class="btn-save" @click="updateWebsiteConfig">
            <el-icon><Check /></el-icon>
            保存网站信息
          </button>
        </div>
      </div>

      <!-- 社交信息 -->
      <div v-show="activeName === 'notice'" class="tab-content">
        <div class="tip-box">
          <el-icon><InfoFilled /></el-icon>
          <span>空白默认不显示，填写链接即可在博客前台展示对应社交图标</span>
        </div>
        <div class="section-title">
          <el-icon><Share /></el-icon>
          社交平台链接
        </div>
        <div class="form-grid">
          <div class="form-group" v-for="item in socialFields" :key="item.key">
            <label class="form-label">
              <span class="social-icon-label" :style="{ color: item.color }">
                <el-icon><component :is="item.icon" /></el-icon>
                {{ item.label }}
              </span>
            </label>
            <el-input v-model="websiteConfigForm[item.key]" :placeholder="item.placeholder" class="form-input">
              <template #prefix>
                <el-icon :style="{ color: item.color }"><component :is="item.icon" /></el-icon>
              </template>
            </el-input>
          </div>
        </div>
        <div class="save-bar">
          <button class="btn-save" @click="updateWebsiteConfig">
            <el-icon><Check /></el-icon>
            保存社交信息
          </button>
        </div>
      </div>

      <!-- 其他设置 -->
      <div v-show="activeName === 'settings'" class="tab-content">
        <div class="section-title">
          <el-icon><Avatar /></el-icon>
          用户头像设置
        </div>
        <div class="form-grid two-col-upload">
          <div class="form-group upload-group">
            <label class="form-label">用户默认头像</label>
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleUserAvatarSuccess">
              <img v-if="websiteConfigForm.userAvatar" :src="websiteConfigForm.userAvatar" class="avatar" />
              <div v-else class="avatar-placeholder">
                <el-icon><Plus /></el-icon>
                <span>用户头像</span>
              </div>
            </el-upload>
          </div>
          <div class="form-group upload-group">
            <label class="form-label">游客默认头像</label>
            <el-upload
              class="avatar-uploader"
              action="/api/admin/config/images"
              :headers="headers"
              :show-file-list="false"
              :before-upload="beforeUpload"
              :on-success="handleTouristAvatarSuccess">
              <img v-if="websiteConfigForm.touristAvatar" :src="websiteConfigForm.touristAvatar" class="avatar" />
              <div v-else class="avatar-placeholder">
                <el-icon><Plus /></el-icon>
                <span>游客头像</span>
              </div>
            </el-upload>
          </div>
        </div>

        <div class="section-title" style="margin-top: 28px">
          <el-icon><SetUp /></el-icon>
          功能开关
        </div>
        <div class="feature-grid">
          <div class="feature-card" :class="{ active: websiteConfigForm.isEmailNotice === 1 }">
            <div class="feature-icon">
              <el-icon><Message /></el-icon>
            </div>
            <div class="feature-info">
              <span class="feature-name">邮箱通知</span>
              <span class="feature-desc">新评论时发送邮件通知</span>
            </div>
            <div class="feature-switch" @click="websiteConfigForm.isEmailNotice = websiteConfigForm.isEmailNotice === 1 ? 0 : 1">
              <div class="switch-track" :class="{ on: websiteConfigForm.isEmailNotice === 1 }">
                <div class="switch-thumb"></div>
              </div>
            </div>
          </div>
          <div class="feature-card" :class="{ active: websiteConfigForm.isCommentReview === 1 }">
            <div class="feature-icon">
              <el-icon><ChatLineSquare /></el-icon>
            </div>
            <div class="feature-info">
              <span class="feature-name">评论审核</span>
              <span class="feature-desc">评论需审核后才显示</span>
            </div>
            <div class="feature-switch" @click="websiteConfigForm.isCommentReview = websiteConfigForm.isCommentReview === 1 ? 0 : 1">
              <div class="switch-track" :class="{ on: websiteConfigForm.isCommentReview === 1 }">
                <div class="switch-thumb"></div>
              </div>
            </div>
          </div>
          <div class="feature-card" :class="{ active: websiteConfigForm.isReward === 1 }">
            <div class="feature-icon">
              <el-icon><Present /></el-icon>
            </div>
            <div class="feature-info">
              <span class="feature-name">打赏功能</span>
              <span class="feature-desc">在文章底部显示打赏</span>
            </div>
            <div class="feature-switch" @click="websiteConfigForm.isReward = websiteConfigForm.isReward === 1 ? 0 : 1">
              <div class="switch-track" :class="{ on: websiteConfigForm.isReward === 1 }">
                <div class="switch-thumb"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- 收款码上传 -->
        <div v-show="websiteConfigForm.isReward === 1" class="qrcode-section">
          <div class="section-title">
            <el-icon><Wallet /></el-icon>
            收款码设置
          </div>
          <div class="form-grid two-col-upload">
            <div class="form-group upload-group">
              <label class="form-label">微信收款码</label>
              <el-upload
                class="avatar-uploader qrcode-uploader"
                action="/api/admin/config/images"
                :headers="headers"
                :show-file-list="false"
                :before-upload="beforeUpload"
                :on-success="handleWeiXinSuccess">
                <img v-if="websiteConfigForm.weiXinQRCode" :src="websiteConfigForm.weiXinQRCode" class="avatar qrcode-img" />
                <div v-else class="avatar-placeholder qrcode-placeholder">
                  <el-icon><Plus /></el-icon>
                  <span>微信收款码</span>
                </div>
              </el-upload>
            </div>
            <div class="form-group upload-group">
              <label class="form-label">支付宝收款码</label>
              <el-upload
                class="avatar-uploader qrcode-uploader"
                action="/api/admin/config/images"
                :headers="headers"
                :show-file-list="false"
                :before-upload="beforeUpload"
                :on-success="handleAlipaySuccess">
                <img v-if="websiteConfigForm.alipayQRCode" :src="websiteConfigForm.alipayQRCode" class="avatar qrcode-img" />
                <div v-else class="avatar-placeholder qrcode-placeholder">
                  <el-icon><Plus /></el-icon>
                  <span>支付宝收款码</span>
                </div>
              </el-upload>
            </div>
          </div>
        </div>
        <div class="save-bar">
          <button class="btn-save" @click="updateWebsiteConfig">
            <el-icon><Check /></el-icon>
            保存其他设置
          </button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElNotification } from 'element-plus'
import {
  Plus, Check, Close, Monitor, Link, Setting,
  InfoFilled, Share, Avatar, SetUp, Message,
  ChatLineSquare, Present, Wallet, EditPen,
  Platform, ChatRound, Notebook, Reading, Promotion,
  Position, Trophy, Connection
} from '@element-plus/icons-vue'
import { createBeforeUploadHandler } from '@/utils/imageUtils'
import { getAuthHeaders } from '@/utils/auth'
import request from '@/utils/request'

const tabs = [
  { key: 'info', label: '网站信息', icon: EditPen },
  { key: 'notice', label: '社交信息', icon: Share },
  { key: 'settings', label: '其他设置', icon: Setting }
]

const socialFields = [
  { key: 'github', label: 'Github', icon: Link, color: '#24292e', placeholder: 'https://github.com/username' },
  { key: 'gitee', label: 'Gitee', icon: Link, color: '#c71d23', placeholder: 'https://gitee.com/username' },
  { key: 'qq', label: 'QQ', icon: ChatRound, color: '#12b7f5', placeholder: 'QQ号码' },
  { key: 'weChat', label: '微信', icon: ChatLineSquare, color: '#07c160', placeholder: '微信号' },
  { key: 'weibo', label: '微博', icon: Platform, color: '#e6162d', placeholder: '微博链接' },
  { key: 'csdn', label: 'CSDN', icon: Notebook, color: '#cf2228', placeholder: 'CSDN链接' },
  { key: 'zhihu', label: '知乎', icon: Reading, color: '#0084ff', placeholder: '知乎链接' },
  { key: 'juejin', label: '掘金', icon: Trophy, color: '#1e80ff', placeholder: '掘金链接' },
  { key: 'twitter', label: 'Twitter', icon: Promotion, color: '#1da1f2', placeholder: 'Twitter链接' },
  { key: 'stackoverflow', label: 'StackOverflow', icon: Position, color: '#f48024', placeholder: 'StackOverflow链接' }
]

const websiteConfigForm = ref({
  authorAvatar: '',
  logo: '',
  favicon: '',
  name: '',
  englishName: '',
  author: '',
  websiteTitle: '',
  authorIntro: '',
  multiLanguage: 0,
  websiteCreateTime: '',
  notice: '',
  beianNumber: '',
  gonganBeianNumber: '',
  qqLogin: 0,
  github: '',
  gitee: '',
  qq: '',
  weChat: '',
  weibo: '',
  csdn: '',
  zhihu: '',
  juejin: '',
  twitter: '',
  stackoverflow: '',
  userAvatar: '',
  touristAvatar: '',
  isEmailNotice: 0,
  isCommentReview: 0,
  isReward: 0,
  weiXinQRCode: '',
  alipayQRCode: ''
})
const activeName = ref('info')
const headers = ref(getAuthHeaders())

const socialCount = computed(() => {
  const keys = ['github', 'gitee', 'qq', 'weChat', 'weibo', 'csdn', 'zhihu', 'juejin', 'twitter', 'stackoverflow']
  return keys.filter(k => websiteConfigForm.value[k]?.trim()).length
})

const activeFeatures = computed(() => {
  let count = 0
  if (websiteConfigForm.value.isEmailNotice === 1) count++
  if (websiteConfigForm.value.isCommentReview === 1) count++
  if (websiteConfigForm.value.isReward === 1) count++
  if (websiteConfigForm.value.qqLogin === 1) count++
  return count
})

const beforeUpload = createBeforeUploadHandler(500)

const handleAuthorAvatarSuccess = (response) => { websiteConfigForm.value.authorAvatar = response.data }
const handleFaviconSuccess = (response) => { websiteConfigForm.value.favicon = response.data }
const handleLogoSuccess = (response) => { websiteConfigForm.value.logo = response.data }
const handleUserAvatarSuccess = (response) => { websiteConfigForm.value.userAvatar = response.data }
const handleTouristAvatarSuccess = (response) => { websiteConfigForm.value.touristAvatar = response.data }
const handleWeiXinSuccess = (response) => { websiteConfigForm.value.weiXinQRCode = response.data }
const handleAlipaySuccess = (response) => { websiteConfigForm.value.alipayQRCode = response.data }

const getWebsiteConfig = async () => {
  try {
    const { data } = await request.get('/admin/website/config')
    if (data && data.data) {
      websiteConfigForm.value = { ...websiteConfigForm.value, ...data.data }
    }
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '获取网站配置失败' })
  }
}

const updateWebsiteConfig = async () => {
  try {
    const { data } = await request.put('/admin/website/config', websiteConfigForm.value)
    if (data.flag) {
      ElNotification.success({ title: '成功', message: data.message })
    } else {
      ElNotification.error({ title: '失败', message: data.message })
    }
  } catch (error) {
    ElNotification.error({ title: '失败', message: error.message || '更新网站配置失败' })
  }
}

onMounted(() => { getWebsiteConfig() })
</script>

<style scoped>
.website-page { padding: 0; }

/* 统计卡片 */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}
.stat-card {
  background: var(--bg-base, #fff);
  border-radius: 16px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  border: 1px solid var(--border-default, #e5e7eb);
  transition: all 0.3s ease;
}
.stat-card:hover { transform: translateY(-4px); box-shadow: 0 12px 24px rgba(0,0,0,0.08); }
.stat-icon {
  width: 56px; height: 56px; border-radius: 14px;
  display: flex; align-items: center; justify-content: center;
  font-size: 24px; flex-shrink: 0;
}
.stat-icon.primary { background: linear-gradient(135deg, #3b82f6, #60a5fa); color: #fff; }
.stat-icon.success { background: linear-gradient(135deg, #10b981, #34d399); color: #fff; }
.stat-icon.warning { background: linear-gradient(135deg, #f59e0b, #fbbf24); color: #fff; }
.stat-info { display: flex; flex-direction: column; gap: 4px; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary, #1f2937); line-height: 1; }
.stat-label { font-size: 14px; color: var(--text-secondary, #6b7280); }

/* 主卡片 */
.main-card {
  border-radius: 16px;
  border: 1px solid var(--border-default, #e5e7eb);
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
  background: var(--bg-base, #fff);
}
.main-card :deep(.el-card__body) { padding: 0; }

/* Tab 头部 */
.tab-header {
  display: flex;
  padding: 0 24px;
  border-bottom: 1px solid var(--border-default, #e5e7eb);
  background: var(--bg-elevated, #f9fafb);
  border-radius: 16px 16px 0 0;
}
.tab-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 24px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  border-bottom: 3px solid transparent;
  transition: all 0.2s ease;
  margin-bottom: -1px;
}
.tab-item:hover { color: var(--color-primary, #3b82f6); }
.tab-item.active {
  color: var(--color-primary, #3b82f6);
  font-weight: 600;
  border-bottom-color: var(--color-primary, #3b82f6);
}

/* Tab 内容 */
.tab-content { padding: 32px; }

/* 段落标题 */
.section-title {
  display: flex; align-items: center; gap: 8px;
  font-size: 16px; font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-light, #f3f4f6);
}

/* 表单网格 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}
.form-grid .full-width { grid-column: span 2; }
.form-grid .upload-group { grid-column: span 1; }
.form-grid.two-col-upload { grid-template-columns: repeat(2, 1fr); }

/* 表单组 */
.form-group { display: flex; flex-direction: column; gap: 8px; }
.form-label {
  font-size: 14px; font-weight: 500;
  color: var(--text-primary, #1f2937);
}

/* 输入框 */
.form-input :deep(.el-input__wrapper) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  height: 44px;
  transition: all 0.2s ease;
}
.form-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6;
}
.form-input :deep(.el-textarea__inner) {
  border-radius: 10px;
  box-shadow: 0 0 0 1px var(--border-default, #e5e7eb);
  transition: all 0.2s ease;
}
.form-input :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px rgba(59,130,246,0.2), 0 0 0 1px #3b82f6;
}
.form-input { width: 100%; }
.form-input :deep(.el-date-editor) { width: 100%; }

/* 上传组件 */
.avatar-uploader :deep(.el-upload) {
  border: 2px dashed var(--border-default, #e5e7eb);
  border-radius: 14px;
  cursor: pointer;
  overflow: hidden;
  background: var(--bg-elevated, #f9fafb);
  transition: all 0.3s ease;
  width: 120px; height: 120px;
  display: flex; align-items: center; justify-content: center;
}
.avatar-uploader :deep(.el-upload:hover) {
  border-color: var(--color-primary, #3b82f6);
  background: rgba(59,130,246,0.05);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59,130,246,0.15);
}
.avatar-placeholder {
  display: flex; flex-direction: column;
  align-items: center; gap: 6px;
  color: var(--text-secondary, #6b7280);
  font-size: 12px;
}
.avatar-placeholder .el-icon { font-size: 24px; }
.avatar {
  width: 120px; height: 120px;
  display: block; object-fit: cover;
  border-radius: 14px;
}

/* 单选卡片 */
.radio-card-group { display: flex; gap: 12px; }
.radio-card {
  flex: 1;
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 10px 20px;
  border: 2px solid var(--border-default, #e5e7eb);
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px; font-weight: 500;
  color: var(--text-secondary, #6b7280);
  transition: all 0.2s ease;
}
.radio-card:hover { border-color: var(--color-primary-light, #93c5fd); }
.radio-card.active {
  border-color: var(--color-primary, #3b82f6);
  color: var(--color-primary, #3b82f6);
  background: rgba(59,130,246,0.08);
}

/* 提示框 */
.tip-box {
  display: flex; align-items: center; gap: 10px;
  padding: 14px 20px;
  background: rgba(59,130,246,0.06);
  border: 1px solid rgba(59,130,246,0.15);
  border-radius: 12px;
  margin-bottom: 24px;
  font-size: 14px;
  color: var(--text-secondary, #6b7280);
}
.tip-box .el-icon { color: var(--color-primary, #3b82f6); font-size: 18px; flex-shrink: 0; }

/* 社交标签 */
.social-icon-label {
  display: flex; align-items: center; gap: 6px;
  font-weight: 500;
}

/* 功能开关卡片 */
.feature-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.feature-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: var(--bg-elevated, #f9fafb);
  border: 2px solid var(--border-default, #e5e7eb);
  border-radius: 14px;
  transition: all 0.2s ease;
}
.feature-card.active { border-color: var(--color-primary, #3b82f6); background: rgba(59,130,246,0.04); }
.feature-icon {
  width: 48px; height: 48px;
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 22px;
  background: var(--bg-base, #fff);
  color: var(--text-secondary, #6b7280);
  flex-shrink: 0;
  transition: all 0.2s ease;
}
.feature-card.active .feature-icon { color: var(--color-primary, #3b82f6); background: rgba(59,130,246,0.1); }
.feature-info { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.feature-name { font-size: 15px; font-weight: 600; color: var(--text-primary, #1f2937); }
.feature-desc { font-size: 13px; color: var(--text-secondary, #6b7280); }
.feature-switch { cursor: pointer; }

/* 自定义开关 */
.switch-track {
  width: 48px; height: 26px;
  background: var(--border-default, #e5e7eb);
  border-radius: 13px;
  position: relative;
  transition: all 0.3s ease;
}
.switch-track.on { background: var(--color-primary, #3b82f6); }
.switch-thumb {
  width: 22px; height: 22px;
  background: #fff;
  border-radius: 50%;
  position: absolute;
  top: 2px; left: 2px;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}
.switch-track.on .switch-thumb { left: 24px; }

/* 收款码 */
.qrcode-section { margin-top: 8px; }
.qrcode-uploader :deep(.el-upload) { width: 160px; height: 160px; }
.qrcode-placeholder { width: 160px; height: 160px; }
.qrcode-img { width: 160px; height: 160px; }

/* 保存栏 */
.save-bar {
  display: flex;
  justify-content: flex-end;
  padding-top: 24px;
  margin-top: 24px;
  border-top: 1px solid var(--border-light, #f3f4f6);
}
.btn-save {
  display: flex; align-items: center; gap: 8px;
  padding: 0 28px; height: 44px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-save:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(59,130,246,0.4); }

/* 深色模式 */
[data-theme="dark"] .stat-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}
[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 15px var(--primary-glow);
}
[data-theme="dark"] .stat-value {
  color: var(--text-primary, #f9fafb);
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
}
[data-theme="dark"] .stat-label { color: var(--text-secondary, #9ca3af); }
[data-theme="dark"] .main-card {
  background: var(--bg-base, #1f2937);
  border-color: var(--border-default, #374151);
}
[data-theme="dark"] .tab-header {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(51, 65, 85, 0.7) 100%);
  border-color: var(--border-default, #374151);
}
[data-theme="dark"] .section-title {
  color: var(--text-primary, #f9fafb);
  border-color: var(--border-default, #374151);
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
[data-theme="dark"] .form-label { color: var(--text-primary, #f9fafb); }
[data-theme="dark"] .form-input :deep(.el-input__wrapper) {
  background: var(--bg-elevated, #374151);
  border-color: rgba(71, 85, 105, 0.5);
}
[data-theme="dark"] .form-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15), 0 0 12px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .form-input :deep(.el-textarea__inner) {
  background: var(--bg-elevated, #374151);
  color: var(--text-primary, #f9fafb);
}
[data-theme="dark"] .form-input :deep(.el-textarea__inner:focus) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15), 0 0 12px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .avatar-uploader :deep(.el-upload) {
  background: var(--bg-elevated, #374151);
  border-color: rgba(0, 212, 255, 0.2);
  transition: all 0.25s ease;
}
[data-theme="dark"] .avatar-uploader :deep(.el-upload:hover) {
  border-color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.15);
}
[data-theme="dark"] .radio-card {
  border-color: var(--border-default, #374151);
  color: var(--text-secondary, #9ca3af);
  transition: all 0.25s ease;
}
[data-theme="dark"] .radio-card.active {
  background: rgba(0, 212, 255, 0.08);
  border-color: rgba(0, 212, 255, 0.3);
  color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.15);
}
[data-theme="dark"] .tip-box {
  background: rgba(0, 212, 255, 0.06);
  border-color: rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .tip-box::before {
  background: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.4);
}
[data-theme="dark"] .feature-card {
  background: var(--bg-elevated, #374151);
  border-color: var(--border-default, #374151);
  transition: all 0.25s ease;
}
[data-theme="dark"] .feature-card:hover {
  border-color: rgba(0, 212, 255, 0.2);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.1);
}
[data-theme="dark"] .feature-icon {
  background: rgba(0, 212, 255, 0.1);
  color: var(--neon-blue, #00D4FF);
  box-shadow: 0 0 15px rgba(0, 212, 255, 0.2);
}
[data-theme="dark"] .switch-track {
  background: var(--border-default, #374151);
}
[data-theme="dark"] .switch-track.is-active {
  background: linear-gradient(135deg, var(--neon-green, #00FF88) 0%, #22C55E 100%);
  box-shadow: 0 0 12px rgba(0, 255, 136, 0.4);
}
[data-theme="dark"] .save-bar { border-color: var(--border-default, #374151); }
[data-theme="dark"] .btn-save {
  background: linear-gradient(135deg, var(--neon-blue, #00D4FF) 0%, var(--neon-purple, #BF5AF2) 100%);
  box-shadow: 0 4px 14px rgba(0, 212, 255, 0.4);
}
[data-theme="dark"] .btn-save:hover {
  box-shadow: 0 6px 20px rgba(0, 212, 255, 0.6);
}

/* 响应式 */
@media (max-width: 768px) {
  .stats-row { grid-template-columns: 1fr; }
  .form-grid { grid-template-columns: 1fr; }
  .form-grid .full-width { grid-column: span 1; }
  .form-grid.two-col-upload { grid-template-columns: 1fr; }
  .tab-header { padding: 0 12px; overflow-x: auto; }
  .tab-item { padding: 12px 16px; font-size: 13px; white-space: nowrap; }
  .tab-content { padding: 20px; }
  .save-bar { justify-content: center; }
  .btn-save { width: 100%; justify-content: center; }
  .qrcode-section .form-grid { grid-template-columns: 1fr; }
}
</style>

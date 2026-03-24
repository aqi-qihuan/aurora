<template>
  <div class="profile-page">
    <!-- 用户信息卡片 -->
    <div class="profile-header">
      <div class="header-bg">
        <div class="bg-pattern"></div>
      </div>
      <div class="profile-card">
        <div class="avatar-wrapper">
          <el-avatar v-if="avatar" :src="avatar" :size="96" class="user-avatar" />
          <div v-else class="avatar-placeholder">
            <el-icon><User /></el-icon>
          </div>
          <label class="avatar-edit">
            <input type="file" accept="image/*" hidden @change="handleAvatarUpload" />
            <el-icon><Camera /></el-icon>
          </label>
        </div>
        <div class="profile-info">
          <h2 class="profile-name">{{ infoForm.nickname || '未设置昵称' }}</h2>
          <p class="profile-desc" v-if="infoForm.intro">{{ infoForm.intro }}</p>
          <div class="profile-meta" v-if="infoForm.website">
            <el-icon><Link /></el-icon>
            <a :href="infoForm.website" target="_blank" rel="noopener">{{ infoForm.website }}</a>
          </div>
        </div>
      </div>
    </div>

    <!-- 设置选项卡 -->
    <div class="settings-card">
      <div class="tab-nav">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          :class="['tab-item', { active: activeName === tab.key }]"
          @click="activeName = tab.key"
        >
          <el-icon><component :is="tab.icon" /></el-icon>
          <span>{{ tab.label }}</span>
        </button>
      </div>

      <!-- 修改信息 -->
      <div class="tab-content" v-if="activeName === 'info'">
        <div class="section-header">
          <div class="section-icon">
            <el-icon><EditPen /></el-icon>
          </div>
          <div>
            <h3 class="section-title">基本信息</h3>
            <p class="section-desc">修改你的个人资料和展示信息</p>
          </div>
        </div>
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label">昵称 <span class="required">*</span></label>
            <input
              v-model="infoForm.nickname"
              class="form-input"
              placeholder="请输入昵称"
              maxlength="30"
            />
          </div>
          <div class="form-group">
            <label class="form-label">个人网站</label>
            <input
              v-model="infoForm.website"
              class="form-input"
              placeholder="https://yourwebsite.com"
            />
          </div>
          <div class="form-group full">
            <label class="form-label">个人简介</label>
            <textarea
              v-model="infoForm.intro"
              class="form-textarea"
              placeholder="介绍一下自己吧..."
              rows="3"
              maxlength="200"
            />
            <span class="char-count">{{ (infoForm.intro || '').length }}/200</span>
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="updateInfo" :disabled="!infoForm.nickname?.trim()">
            <el-icon><Check /></el-icon>保存修改
          </button>
        </div>
      </div>

      <!-- 修改密码 -->
      <div class="tab-content" v-if="activeName === 'password'">
        <div class="section-header">
          <div class="section-icon warning">
            <el-icon><Lock /></el-icon>
          </div>
          <div>
            <h3 class="section-title">修改密码</h3>
            <p class="section-desc">定期修改密码有助于保护账号安全</p>
          </div>
        </div>
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label">当前密码 <span class="required">*</span></label>
            <div class="input-wrapper">
              <input
                v-model="passwordForm.oldPassword"
                :type="showOldPwd ? 'text' : 'password'"
                class="form-input has-icon"
                placeholder="请输入当前密码"
                @keyup.enter="updatePassword"
              />
              <button class="input-toggle" @click="showOldPwd = !showOldPwd">
                <el-icon><View v-if="!showOldPwd" /><Hide v-else /></el-icon>
              </button>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">新密码 <span class="required">*</span></label>
            <div class="input-wrapper">
              <input
                v-model="passwordForm.newPassword"
                :type="showNewPwd ? 'text' : 'password'"
                class="form-input has-icon"
                placeholder="至少 6 个字符"
                @keyup.enter="updatePassword"
              />
              <button class="input-toggle" @click="showNewPwd = !showNewPwd">
                <el-icon><View v-if="!showNewPwd" /><Hide v-else /></el-icon>
              </button>
            </div>
            <div class="pwd-strength" v-if="passwordForm.newPassword">
              <div class="strength-bar">
                <span
                  v-for="i in 3"
                  :key="i"
                  :class="['strength-seg', { active: pwdStrength >= i }]"
                ></span>
              </div>
              <span :class="['strength-text', 'level-' + pwdStrength]">{{ pwdStrengthLabel }}</span>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">确认新密码 <span class="required">*</span></label>
            <div class="input-wrapper">
              <input
                v-model="passwordForm.confirmPassword"
                :type="showConfirmPwd ? 'text' : 'password'"
                class="form-input has-icon"
                :class="{ 'input-error': passwordForm.confirmPassword && passwordForm.confirmPassword !== passwordForm.newPassword }"
                placeholder="再次输入新密码"
                @keyup.enter="updatePassword"
              />
              <button class="input-toggle" @click="showConfirmPwd = !showConfirmPwd">
                <el-icon><View v-if="!showConfirmPwd" /><Hide v-else /></el-icon>
              </button>
            </div>
            <p v-if="passwordForm.confirmPassword && passwordForm.confirmPassword !== passwordForm.newPassword" class="field-error">
              <el-icon><WarningFilled /></el-icon>两次密码输入不一致
            </p>
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="updatePassword">
            <el-icon><Lock /></el-icon>更新密码
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import {
  User, Camera, EditPen, Lock, Check, Link,
  View, Hide, WarningFilled
} from '@element-plus/icons-vue'
import request from '@/utils/request'
import { useUserStore } from '@/stores/user'
import { getAuthHeaders } from '@/utils/auth'

const userStore = useUserStore()

const infoForm = reactive({
  nickname: userStore.userInfo?.nickname || '',
  intro: userStore.userInfo?.intro || '',
  website: userStore.userInfo?.website || ''
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const activeName = ref('info')
const avatar = computed(() => userStore.userInfo?.avatar)
const showOldPwd = ref(false)
const showNewPwd = ref(false)
const showConfirmPwd = ref(false)

const tabs = [
  { key: 'info', label: '基本信息', icon: 'EditPen' },
  { key: 'password', label: '安全设置', icon: 'Lock' }
]

const pwdStrength = computed(() => {
  const pwd = passwordForm.newPassword
  if (!pwd) return 0
  let score = 0
  if (pwd.length >= 6) score++
  if (/[A-Z]/.test(pwd) && /[a-z]/.test(pwd)) score++
  if (/[0-9]/.test(pwd) && /[^A-Za-z0-9]/.test(pwd)) score++
  return score
})

const pwdStrengthLabel = computed(() => {
  const labels = ['', '弱', '中', '强']
  return labels[pwdStrength.value] || ''
})

const handleAvatarUpload = (e) => {
  const file = e.target.files[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  const headers = getAuthHeaders()
  delete headers['Content-Type']
  request.post('/api/users/avatar', formData, { headers }).then(({ data }) => {
    if (data.flag) {
      ElMessage.success(data.message)
      userStore.updateAvatar(data.data)
    } else {
      ElMessage.error(data.message)
    }
  }).catch(() => {
    ElMessage.error('头像上传失败')
  })
}

const updateInfo = () => {
  if (!infoForm.nickname?.trim()) {
    ElMessage.error('昵称不能为空')
    return
  }
  request.put('/users/info', infoForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({ title: '成功', message: '修改成功' })
      userStore.updateUserInfo(infoForm)
    } else {
      ElNotification.error({ title: '失败', message: '修改失败' })
    }
  }).catch(() => {
    ElMessage.error('修改失败')
  })
}

const updatePassword = () => {
  if (!passwordForm.oldPassword || !passwordForm.newPassword || !passwordForm.confirmPassword) {
    ElMessage.error('请填写所有密码字段')
    return
  }
  if (passwordForm.newPassword.length < 6) {
    ElMessage.error('新密码不能少于 6 位')
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.error('两次密码输入不一致')
    return
  }
  request.put('/admin/users/password', passwordForm).then(({ data }) => {
    if (data.flag) {
      passwordForm.oldPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
      ElNotification.success({ title: '成功', message: '密码修改成功' })
    } else {
      ElNotification.error({ title: '失败', message: data.message || '修改失败' })
    }
  }).catch(() => {
    ElMessage.error('密码修改失败')
  })
}
</script>

<style scoped>
.profile-page {
  padding: 4px 0;
}

/* ========== 头部卡片 ========== */
.profile-header {
  position: relative;
  margin-bottom: 20px;
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid var(--border-color, #ebeef5);
}
.header-bg {
  height: 120px;
  background: linear-gradient(135deg, #1a73e8, #6366f1, #8b5cf6);
  position: relative;
  overflow: hidden;
}
.bg-pattern {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle at 20% 50%, rgba(255,255,255,0.1) 0%, transparent 50%),
                    radial-gradient(circle at 80% 20%, rgba(255,255,255,0.08) 0%, transparent 40%),
                    radial-gradient(circle at 60% 80%, rgba(255,255,255,0.06) 0%, transparent 45%);
}
.profile-card {
  display: flex;
  align-items: flex-end;
  gap: 20px;
  padding: 0 28px 24px;
  margin-top: -40px;
  position: relative;
  background: var(--bg-card, #fff);
}
.avatar-wrapper {
  position: relative;
  flex-shrink: 0;
}
.user-avatar {
  border: 4px solid var(--bg-card, #fff);
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
}
.avatar-placeholder {
  width: 96px;
  height: 96px;
  border-radius: 50%;
  background: linear-gradient(135deg, #e8f0fe, #d2e3fc);
  border: 4px solid var(--bg-card, #fff);
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-placeholder .el-icon { font-size: 36px; color: #1a73e8; }
.avatar-edit {
  position: absolute;
  bottom: 4px;
  right: 4px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #1a73e8;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 13px;
  box-shadow: 0 2px 8px rgba(26,115,232,0.35);
  transition: transform 0.2s;
}
.avatar-edit:hover { transform: scale(1.1); }
.profile-info { padding-bottom: 4px; min-width: 0; }
.profile-name {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary, #1f2937);
}
.profile-desc {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
  max-width: 400px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.profile-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 8px;
  font-size: 13px;
}
.profile-meta .el-icon { color: #1a73e8; font-size: 14px; }
.profile-meta a {
  color: #1a73e8;
  text-decoration: none;
  max-width: 300px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.profile-meta a:hover { text-decoration: underline; }

/* ========== 设置卡片 ========== */
.settings-card {
  background: var(--bg-card, #fff);
  border: 1px solid var(--border-color, #ebeef5);
  border-radius: 16px;
  overflow: hidden;
}

/* ========== 选项卡导航 ========== */
.tab-nav {
  display: flex;
  border-bottom: 1px solid var(--border-color, #ebeef5);
  padding: 0 24px;
  background: var(--bg-body, #f9fafb);
}
.tab-item {
  height: 52px;
  padding: 0 20px;
  border: none;
  background: none;
  border-bottom: 2px solid transparent;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}
.tab-item:hover { color: var(--text-primary, #1f2937); }
.tab-item.active {
  color: #1a73e8;
  border-bottom-color: #1a73e8;
}
.tab-item .el-icon { font-size: 16px; }

/* ========== 内容区域 ========== */
.tab-content {
  padding: 28px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
}
.section-icon {
  width: 42px;
  height: 42px;
  border-radius: 12px;
  background: linear-gradient(135deg, #e8f0fe, #d2e3fc);
  color: #1a73e8;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}
.section-icon.warning { background: linear-gradient(135deg, #fef3c7, #fde68a); color: #d97706; }
.section-title { margin: 0; font-size: 16px; font-weight: 600; color: var(--text-primary, #1f2937); }
.section-desc { margin: 3px 0 0; font-size: 13px; color: var(--text-secondary, #6b7280); }

/* ========== 表单 ========== */
.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  max-width: 640px;
}
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-group.full { grid-column: 1 / -1; }
.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary, #374151);
}
.required { color: #ef4444; }
.form-input {
  height: 40px;
  padding: 0 14px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 10px;
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff);
  outline: none;
  transition: all 0.2s;
}
.form-input:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1); }
.form-input::placeholder { color: var(--text-tertiary, #9ca3af); }
.form-input.has-icon { padding-right: 40px; }
.form-input.input-error { border-color: #ef4444; }
.form-input.input-error:focus { box-shadow: 0 0 0 3px rgba(239,68,68,0.1); }
.form-textarea {
  padding: 10px 14px;
  border: 1px solid var(--border-color, #d1d5db);
  border-radius: 10px;
  font-size: 14px;
  color: var(--text-primary, #1f2937);
  background: var(--bg-card, #fff);
  outline: none;
  resize: vertical;
  font-family: inherit;
  line-height: 1.5;
  transition: all 0.2s;
  box-sizing: border-box;
}
.form-textarea:focus { border-color: #1a73e8; box-shadow: 0 0 0 3px rgba(26,115,232,0.1); }
.form-textarea::placeholder { color: var(--text-tertiary, #9ca3af); }
.char-count {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  text-align: right;
  margin-top: -2px;
}

/* ========== 密码切换 ========== */
.input-wrapper { position: relative; }
.input-toggle {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-tertiary, #9ca3af);
  cursor: pointer;
  font-size: 16px;
  padding: 4px;
  display: flex;
  transition: color 0.2s;
}
.input-toggle:hover { color: var(--text-primary, #374151); }

/* ========== 密码强度 ========== */
.pwd-strength {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
}
.strength-bar {
  display: flex;
  gap: 4px;
}
.strength-seg {
  width: 32px;
  height: 4px;
  border-radius: 2px;
  background: var(--border-color, #e5e7eb);
  transition: background 0.3s;
}
.strength-seg.active.level-1 { background: #ef4444; }
.strength-seg.active.level-2 { background: #f59e0b; }
.strength-seg.active.level-3 { background: #10b981; }
.strength-seg.active { background: #10b981; }
.pwd-strength .strength-seg:nth-child(1).active { background: #ef4444; }
.pwd-strength .strength-seg:nth-child(2).active { background: #f59e0b; }
.pwd-strength .strength-seg:nth-child(3).active { background: #10b981; }
.strength-text { font-size: 12px; font-weight: 500; }
.strength-text.level-1 { color: #ef4444; }
.strength-text.level-2 { color: #f59e0b; }
.strength-text.level-3 { color: #10b981; }

/* ========== 错误提示 ========== */
.field-error {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #ef4444;
  margin-top: 2px;
}
.field-error .el-icon { font-size: 13px; }

/* ========== 表单操作 ========== */
.form-actions {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--border-color, #f0f0f0);
}
.btn {
  height: 38px;
  padding: 0 20px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-primary {
  background: linear-gradient(135deg, #1a73e8, #4285f4);
  color: #fff;
  box-shadow: 0 2px 8px rgba(26,115,232,0.25);
}
.btn-primary:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 14px rgba(26,115,232,0.35); }

/* ========== 暗色模式 ========== */
[data-theme="dark"] .profile-header { border-color: var(--border-color, #334155); }
[data-theme="dark"] .profile-card { background: var(--bg-card, #1e293b); }
[data-theme="dark"] .user-avatar { border-color: var(--bg-card, #1e293b); }
[data-theme="dark"] .avatar-placeholder { border-color: var(--bg-card, #1e293b); background: linear-gradient(135deg, #1e3a5f, #1e3a5f); }
[data-theme="dark"] .avatar-placeholder .el-icon { color: #60a5fa; }
[data-theme="dark"] .settings-card { background: var(--bg-card, #1e293b); border-color: var(--border-color, #334155); }
[data-theme="dark"] .tab-nav { background: var(--bg-body, #0f172a); border-color: var(--border-color, #334155); }
[data-theme="dark"] .tab-item { color: var(--text-tertiary, #94a3b8); }
[data-theme="dark"] .tab-item:hover { color: var(--text-primary, #e2e8f0); }
[data-theme="dark"] .tab-item.active { color: #60a5fa; border-bottom-color: #3b82f6; }
[data-theme="dark"] .form-input,
[data-theme="dark"] .form-textarea { background: var(--bg-body, #0f172a); border-color: var(--border-color, #334155); color: var(--text-primary, #e2e8f0); }
[data-theme="dark"] .form-input.input-error { border-color: #ef4444; }
[data-theme="dark"] .input-toggle { color: var(--text-tertiary, #64748b); }
[data-theme="dark"] .input-toggle:hover { color: var(--text-primary, #e2e8f0); }
[data-theme="dark"] .strength-seg { background: var(--border-color, #334155); }
[data-theme="dark"] .form-actions { border-top-color: var(--border-color, #334155); }
[data-theme="dark"] .profile-meta a { color: #60a5fa; }

/* ========== 响应式 ========== */
@media (max-width: 640px) {
  .header-bg { height: 100px; }
  .profile-card { flex-direction: column; align-items: center; padding: 0 20px 24px; margin-top: -32px; }
  .profile-info { text-align: center; display: flex; flex-direction: column; align-items: center; }
  .profile-desc { max-width: 100%; }
  .profile-meta a { max-width: 200px; }
  .tab-nav { padding: 0 16px; }
  .tab-item { padding: 0 14px; font-size: 13px; }
  .tab-content { padding: 20px; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>

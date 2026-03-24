<template>
  <div class="login-page">
    <!-- 左侧视觉区 -->
    <div class="login-visual">
      <div class="visual-overlay"></div>
      <div class="visual-content">
        <div class="visual-brand">
          <div class="brand-logo">
            <el-icon :size="36"><Sunny /></el-icon>
          </div>
          <h1 class="brand-name">Aurora</h1>
        </div>
        <p class="visual-slogan">记录技术，分享生活</p>
        <div class="visual-features">
          <div class="feature-item" v-for="(f, i) in features" :key="i" :style="{ '--i': i }">
            <span class="feature-dot"></span>
            <span>{{ f }}</span>
          </div>
        </div>
      </div>
      <div class="visual-particles">
        <span v-for="n in 6" :key="n" class="particle" :style="{ '--n': n }"></span>
      </div>
    </div>

    <!-- 右侧登录区 -->
    <div class="login-panel">
      <div class="login-form-wrapper">
        <!-- 品牌标识 (移动端) -->
        <div class="mobile-brand">
          <div class="brand-logo-sm">
            <el-icon :size="24"><Sunny /></el-icon>
          </div>
          <span class="brand-name-sm">Aurora</span>
        </div>

        <div class="login-header">
          <h2 class="login-title">欢迎回来</h2>
          <p class="login-subtitle">请输入您的账号信息登录管理后台</p>
        </div>

        <form class="login-form" @submit.prevent="handleLogin">
          <div class="form-group">
            <label class="form-label" for="username">用户名</label>
            <div class="input-wrapper">
              <span class="input-icon">
                <el-icon><User /></el-icon>
              </span>
              <input
                id="username"
                v-model="loginForm.username"
                type="text"
                autocomplete="username"
                placeholder="请输入用户名"
                @keyup.enter="$refs.passwordInput?.focus()"
                class="form-input"
              />
            </div>
            <p class="form-error" v-if="errors.username">{{ errors.username }}</p>
          </div>

          <div class="form-group">
            <label class="form-label" for="password">密码</label>
            <div class="input-wrapper">
              <span class="input-icon">
                <el-icon><Lock /></el-icon>
              </span>
              <input
                id="password"
                ref="passwordInput"
                v-model="loginForm.password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                placeholder="请输入密码"
                @keyup.enter="handleLogin"
                class="form-input"
              />
              <button
                type="button"
                class="password-toggle"
                @click="showPassword = !showPassword"
                aria-label="切换密码显示">
                <el-icon><View v-if="showPassword" /><Hide v-else /></el-icon>
              </button>
            </div>
            <p class="form-error" v-if="errors.password">{{ errors.password }}</p>
          </div>

          <button
            type="submit"
            class="login-btn"
            :disabled="loading">
            <span v-if="loading" class="btn-loading">
              <span class="spinner"></span>
              登录中...
            </span>
            <span v-else>登 录</span>
          </button>
        </form>

        <div class="login-footer">
          <span class="footer-text">Aurora Blog Admin</span>
          <span class="footer-divider">|</span>
          <span class="footer-text">Powered by Vue 3</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, View, Hide, Sunny } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import request from '@/utils/request'
import logger from '@/utils/logger'

const router = useRouter()
const userStore = useUserStore()

const loading = ref(false)
const showPassword = ref(false)
const passwordInput = ref(null)

const loginForm = reactive({
  username: '',
  password: ''
})

const errors = reactive({
  username: '',
  password: ''
})

const features = [
  ' Markdown 编辑器',
  ' 多主题切换',
  ' 标签云可视化',
  ' 访问数据分析'
]

const validate = () => {
  let valid = true
  errors.username = ''
  errors.password = ''

  if (!loginForm.username.trim()) {
    errors.username = '请输入用户名'
    valid = false
  }
  if (!loginForm.password.trim()) {
    errors.password = '请输入密码'
    valid = false
  }
  return valid
}

const handleLogin = () => {
  if (loading.value) return
  if (!validate()) return

  loading.value = true

  const params = new URLSearchParams()
  params.append('username', loginForm.username)
  params.append('password', loginForm.password)

  request.post('/users/login', params)
    .then(async ({ data }) => {
      logger.log('登录响应:', data)
      if (data.flag) {
        userStore.login(data.data)
        if (data.data.menus?.length > 0) {
          userStore.saveUserMenus(data.data.menus)
        }
        ElMessage.success('登录成功')
        await router.push({ path: '/home' })
      }
    })
    .catch((error) => {
      logger.error('登录失败:', error)
    })
    .finally(() => {
      loading.value = false
    })
}
</script>

<style scoped>
/* ===== 页面布局 ===== */
.login-page {
  display: flex;
  min-height: 100vh;
  min-height: 100dvh;
}

/* ===== 左侧视觉区 ===== */
.login-visual {
  flex: 1;
  position: relative;
  background:
    url(https://ws.aqi125.cn/aurora/articles/12f138cdc95482591d354088fe37ec32.jpg)
    center center / cover no-repeat;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.visual-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.75) 0%, rgba(30, 41, 59, 0.6) 50%, rgba(59, 130, 246, 0.3) 100%);
  z-index: 1;
}

.visual-content {
  position: relative;
  z-index: 2;
  text-align: center;
  color: #fff;
  padding: 40px;
  animation: fadeIn 0.8s ease;
}

.visual-brand {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  margin-bottom: 16px;
}

.brand-logo {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fbbf24;
  box-shadow: 0 4px 24px rgba(251, 191, 36, 0.25);
}

.brand-name {
  font-size: 36px;
  font-weight: 800;
  letter-spacing: 1px;
  margin: 0;
}

.visual-slogan {
  font-size: 16px;
  opacity: 0.8;
  margin: 0 0 40px;
  letter-spacing: 2px;
}

.visual-features {
  display: flex;
  flex-direction: column;
  gap: 14px;
  align-items: flex-start;
  max-width: 200px;
  margin: 0 auto;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  opacity: 0.9;
  animation: slideUp 0.5s ease both;
  animation-delay: calc(0.3s + var(--i) * 0.1s);
}

.feature-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #fbbf24;
  flex-shrink: 0;
  box-shadow: 0 0 8px rgba(251, 191, 36, 0.5);
}

/* 浮动粒子 */
.visual-particles {
  position: absolute;
  inset: 0;
  z-index: 1;
  pointer-events: none;
}

.particle {
  position: absolute;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  animation: float 6s ease-in-out infinite;
  animation-delay: calc(var(--n) * -1s);
}

.particle:nth-child(1) { left: 20%; top: 30%; animation-duration: 7s; }
.particle:nth-child(2) { left: 60%; top: 15%; animation-duration: 5s; width: 6px; height: 6px; }
.particle:nth-child(3) { left: 80%; top: 55%; animation-duration: 8s; }
.particle:nth-child(4) { left: 35%; top: 70%; animation-duration: 6s; width: 3px; height: 3px; }
.particle:nth-child(5) { left: 75%; top: 80%; animation-duration: 9s; }
.particle:nth-child(6) { left: 15%; top: 60%; animation-duration: 7.5s; width: 5px; height: 5px; }

/* ===== 右侧登录面板 ===== */
.login-panel {
  width: 460px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-elevated);
  padding: 40px;
  animation: slideInRight 0.5s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1));
}

.login-form-wrapper {
  width: 100%;
  max-width: 360px;
}

/* 移动端品牌 (桌面隐藏) */
.mobile-brand {
  display: none;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 32px;
}

.brand-logo-sm {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary);
}

.brand-name-sm {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

/* ===== 表单头部 ===== */
.login-header {
  margin-bottom: 36px;
}

.login-title {
  font-size: 26px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 8px;
}

.login-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

/* ===== 表单 ===== */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 14px;
  color: var(--text-muted);
  font-size: 16px;
  transition: color 0.2s ease;
  pointer-events: none;
  z-index: 1;
}

.form-input {
  width: 100%;
  height: 46px;
  padding: 0 14px 0 42px;
  border-radius: var(--radius-lg, 12px);
  border: 1.5px solid var(--border-default);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 14px;
  font-family: var(--font-sans);
  outline: none;
  transition: all 0.2s var(--ease-out);
}

.form-input::placeholder {
  color: var(--text-muted);
}

.form-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.form-input:focus ~ .input-icon,
.form-input:focus + .input-icon {
  color: var(--primary);
}

.input-wrapper:focus-within .input-icon {
  color: var(--primary);
}

.password-toggle {
  position: absolute;
  right: 12px;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md, 8px);
  border: none;
  background: transparent;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.password-toggle:hover {
  color: var(--text-secondary);
  background: var(--bg-hover);
}

.form-error {
  font-size: 12px;
  color: var(--danger);
  min-height: 0;
  margin: 0;
  transition: all 0.2s ease;
}

/* ===== 登录按钮 ===== */
.login-btn {
  width: 100%;
  height: 46px;
  border-radius: var(--radius-lg, 12px);
  border: none;
  background: var(--gradient-primary);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s var(--ease-out);
  margin-top: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.3);
}

.login-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(59, 130, 246, 0.4);
}

.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-loading {
  display: flex;
  align-items: center;
  gap: 8px;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

/* ===== 底部 ===== */
.login-footer {
  margin-top: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.footer-text {
  font-size: 12px;
  color: var(--text-muted);
}

.footer-divider {
  color: var(--border-default);
  font-size: 12px;
}

/* ===== 动画 ===== */
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideInRight {
  from { opacity: 0; transform: translateX(30px); }
  to { opacity: 1; transform: translateX(0); }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes float {
  0%, 100% { transform: translateY(0) scale(1); opacity: 0.3; }
  50% { transform: translateY(-20px) scale(1.2); opacity: 0.6; }
}

/* ===== 暗色模式 ===== */
[data-theme="dark"] .login-panel {
  background: var(--bg-base);
  border-left: 1px solid rgba(59, 130, 246, 0.15);
}

[data-theme="dark"] .form-input {
  background: var(--bg-deep);
  border-color: var(--border-default);
}

[data-theme="dark"] .form-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

[data-theme="dark"] .password-toggle:hover {
  background: var(--bg-surface);
}

[data-theme="dark"] .login-btn {
  box-shadow: 0 4px 20px rgba(59, 130, 246, 0.35);
}

[data-theme="dark"] .login-btn:hover:not(:disabled) {
  box-shadow: 0 6px 28px rgba(59, 130, 246, 0.5);
}

[data-theme="dark"] .brand-logo {
  background: rgba(59, 130, 246, 0.2);
  box-shadow: 0 4px 24px rgba(59, 130, 246, 0.3);
}

[data-theme="dark"] .brand-logo-sm {
  background: var(--primary-light);
}

/* ===== 响应式 ===== */
@media (max-width: 960px) {
  .login-visual {
    display: none;
  }

  .login-page {
    min-height: 100vh;
    min-height: 100dvh;
    background: var(--bg-base);
  }

  .login-panel {
    width: 100%;
    padding: 24px;
  }

  .mobile-brand {
    display: flex;
  }

  .login-form-wrapper {
    max-width: 400px;
  }

  .login-title {
    font-size: 22px;
  }

  .login-footer {
    margin-top: 36px;
  }
}

@media (max-width: 400px) {
  .login-panel {
    padding: 16px;
  }

  .form-input {
    height: 42px;
  }

  .login-btn {
    height: 42px;
  }
}
</style>

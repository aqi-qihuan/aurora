<template>
  <div class="error-container error-403">
    <div class="error-decoration">
      <span class="deco-ring deco-ring-1"></span>
      <span class="deco-ring deco-ring-2"></span>
    </div>
    <div class="error-content">
      <div class="error-icon">
        <el-icon :size="48"><Lock /></el-icon>
      </div>
      <div class="error-code">403</div>
      <div class="error-desc">
        <h2>无访问权限</h2>
        <p>抱歉，您没有权限访问此页面</p>
      </div>
      <div class="error-actions">
        <el-button type="primary" @click="goHome" class="action-btn">
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
        <el-button @click="goBack" class="action-btn action-btn-ghost">
          <el-icon><Back /></el-icon>
          返回上页
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { HomeFilled, Back, Lock } from '@element-plus/icons-vue'

const router = useRouter()

const goHome = () => {
  router.push('/home')
}

const goBack = () => {
  router.go(-1)
}
</script>

<style scoped>
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--gradient-bg);
  padding: 20px;
  position: relative;
  overflow: hidden;
  transition: background 0.3s ease;
}

[data-theme="dark"] .error-403 {
  background: var(--bg-deep);
}

/* 装饰元素 */
.error-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.deco-ring {
  position: absolute;
  border-radius: 50%;
  border: 1px solid var(--danger-light);
  opacity: 0.5;
  animation: ringFloat 8s ease-in-out infinite;
}

.deco-ring-1 {
  width: 250px;
  height: 250px;
  top: -60px;
  left: -60px;
}

.deco-ring-2 {
  width: 180px;
  height: 180px;
  bottom: -30px;
  right: -30px;
  animation-delay: 3s;
}

[data-theme="dark"] .deco-ring {
  border-color: rgba(239, 68, 68, 0.15);
  box-shadow: 0 0 15px rgba(239, 68, 68, 0.1);
}

@keyframes ringFloat {
  0%, 100% { transform: translate(0, 0) scale(1); opacity: 0.3; }
  50% { transform: translate(10px, -10px) scale(1.05); opacity: 0.6; }
}

.error-content {
  text-align: center;
  z-index: 1;
  animation: fadeInUp 0.6s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1));
}

.error-icon {
  margin-bottom: 16px;
  color: var(--danger);
  transition: color 0.3s ease;
}

[data-theme="dark"] .error-icon {
  color: #F87171;
  filter: drop-shadow(0 0 12px rgba(239, 68, 68, 0.4));
}

.error-code {
  font-size: 150px;
  font-weight: 800;
  line-height: 1;
  margin-bottom: 20px;
  background: linear-gradient(135deg, var(--danger) 0%, var(--neon-orange) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: -4px;
  user-select: none;
}

[data-theme="dark"] .error-code {
  background: linear-gradient(135deg, #F87171 0%, var(--neon-orange) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 0 20px rgba(239, 68, 68, 0.3));
}

.error-desc h2 {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 12px 0;
  transition: color 0.3s ease;
}

.error-desc p {
  font-size: 15px;
  color: var(--text-secondary);
  margin: 0 0 36px 0;
  transition: color 0.3s ease;
}

.error-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.action-btn {
  padding: 10px 24px;
  height: 42px;
  border-radius: var(--radius-md, 8px);
  font-weight: 500;
  font-size: 14px;
  transition: all var(--duration-normal, 250ms) var(--ease-out);
}

.action-btn:hover {
  transform: translateY(-2px);
}

.action-btn:active {
  transform: translateY(0);
}

.action-btn-ghost {
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  color: var(--text-primary);
}

[data-theme="dark"] .action-btn-ghost {
  background: var(--bg-surface);
  border-color: rgba(239, 68, 68, 0.3);
}

.action-btn-ghost:hover {
  border-color: var(--danger);
  color: var(--danger);
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 768px) {
  .error-container {
    flex-direction: column;
  }

  .error-content {
    margin-bottom: 30px;
  }

  .error-code {
    font-size: 100px;
    letter-spacing: -2px;
  }

  .error-desc h2 {
    font-size: 22px;
  }

  .error-desc p {
    font-size: 14px;
  }
}
</style>

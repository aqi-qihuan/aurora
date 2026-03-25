<template>
  <div class="reward-container" v-if="isRewardEnabled">
    <!-- 打赏按钮 -->
    <div class="reward-trigger" @click="toggleReward">
      <div class="reward-btn">
        <svg class="reward-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
        </svg>
        <span class="reward-text">赏</span>
      </div>
      <span class="reward-hint">如果文章对你有帮助，可以请我喝杯咖啡~</span>
    </div>

    <!-- 打赏面板 -->
    <transition name="fade-slide">
      <div class="reward-panel" v-if="showPanel">
        <div class="panel-header">
          <h3 class="panel-title">感谢支持</h3>
          <button class="close-btn" @click="showPanel = false">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <div class="panel-body">
          <!-- 二维码切换 -->
          <div class="qrcode-tabs">
            <button
              :class="['tab-btn', { active: activeTab === 'wechat' }]"
              @click="activeTab = 'wechat'"
              v-if="weiXinQRCode">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 01.213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 00.167-.054l1.903-1.114a.864.864 0 01.717-.098 10.16 10.16 0 002.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178A1.17 1.17 0 014.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178 1.17 1.17 0 01-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 01.598.082l1.584.926a.272.272 0 00.14.047c.134 0 .24-.111.24-.247 0-.06-.023-.12-.038-.177l-.327-1.233a.49.49 0 01.176-.554C23.237 18.254 24 16.667 24 14.904c0-3.282-3.066-5.97-7.062-6.046zm-2.745 2.986c.535 0 .969.44.969.982a.976.976 0 01-.969.983.976.976 0 01-.969-.983c0-.542.434-.982.97-.982zm4.842 0c.535 0 .969.44.969.982a.976.976 0 01-.969.983.976.976 0 01-.969-.983c0-.542.434-.982.97-.982z"/>
              </svg>
              <span>微信支付</span>
            </button>
            <button
              :class="['tab-btn', { active: activeTab === 'alipay' }]"
              @click="activeTab = 'alipay'"
              v-if="alipayQRCode">
              <svg viewBox="0 0 24 24" fill="currentColor">
                <path d="M21.422 15.358c-3.83-1.153-6.055-1.84-7.373-2.313.632-1.327 1.076-2.879 1.28-4.64h-3.372V6.903h4.255V5.86h-4.255V3.19H10.19v2.67H5.812v1.043h4.378V8.405H6.408v1.043h7.057c-.168 1.18-.469 2.25-.893 3.198-1.56-.467-3.13-.81-5.18-.81-3.01 0-4.6 1.415-4.6 3.206 0 1.79 1.59 3.206 4.6 3.206 2.41 0 4.34-.972 5.86-2.576 2.23 1.06 6.47 2.53 8.17 3.19V15.36zM7.392 17.21c-2.13 0-3.35-.82-3.35-1.98s1.22-1.98 3.35-1.98c1.64 0 3.03.28 4.33.71-1.12 1.39-2.62 3.25-4.33 3.25z"/>
              </svg>
              <span>支付宝</span>
            </button>
          </div>

          <!-- 二维码展示 -->
          <div class="qrcode-display">
            <div class="qrcode-wrapper">
              <img
                v-if="activeTab === 'wechat'"
                :src="weiXinQRCode"
                alt="微信收款码"
                class="qrcode-image" />
              <img
                v-else
                :src="alipayQRCode"
                alt="支付宝收款码"
                class="qrcode-image" />
              <p class="qrcode-label">
                {{ activeTab === 'wechat' ? '微信扫码打赏' : '支付宝扫码打赏' }}
              </p>
            </div>
          </div>

          <!-- 打赏列表 -->
          <div class="reward-footer">
            <p class="footer-text">您的支持是我最大的动力 ❤️</p>
          </div>
        </div>
      </div>
    </transition>

    <!-- 遮罩层 -->
    <transition name="fade">
      <div class="reward-overlay" v-if="showPanel" @click="showPanel = false"></div>
    </transition>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from 'vue'
import { useAppStore } from '@/stores/app'

export default defineComponent({
  name: 'Reward',
  setup() {
    const appStore = useAppStore()
    const showPanel = ref(false)
    const activeTab = ref<'wechat' | 'alipay'>('wechat')

    const websiteConfig = computed(() => appStore.websiteConfig || {})

    const isRewardEnabled = computed(() => {
      return websiteConfig.value.isReward === 1
    })

    const weiXinQRCode = computed(() => websiteConfig.value.weiXinQRCode)
    const alipayQRCode = computed(() => websiteConfig.value.alipayQRCode)

    const toggleReward = () => {
      if (weiXinQRCode.value || alipayQRCode.value) {
        // 默认显示第一个可用的
        if (!weiXinQRCode.value && alipayQRCode.value) {
          activeTab.value = 'alipay'
        } else {
          activeTab.value = 'wechat'
        }
        showPanel.value = true
      }
    }

    return {
      showPanel,
      activeTab,
      isRewardEnabled,
      weiXinQRCode,
      alipayQRCode,
      toggleReward
    }
  }
})
</script>

<style lang="scss" scoped>
.reward-container {
  @apply my-8;
}

// 打赏触发按钮
.reward-trigger {
  @apply flex flex-col items-center gap-3 py-6;
  cursor: pointer;
}

.reward-btn {
  @apply flex items-center justify-center;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #ff6b6b 0%, #ff8e53 100%);
  box-shadow: 0 4px 20px rgba(255, 107, 107, 0.4);
  transition: all 0.3s ease;

  &:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 30px rgba(255, 107, 107, 0.5);
  }

  &:active {
    transform: scale(0.95);
  }
}

.reward-icon {
  width: 28px;
  height: 28px;
  color: white;
  animation: heartbeat 1.5s ease-in-out infinite;
}

.reward-text {
  position: absolute;
  font-size: 16px;
  font-weight: 700;
  color: white;
  opacity: 0;
  transition: opacity 0.3s;
}

.reward-btn:hover {
  .reward-icon {
    opacity: 0;
  }
  .reward-text {
    opacity: 1;
  }
}

.reward-hint {
  @apply text-sm text-ob-dim opacity-70;
}

@keyframes heartbeat {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

// 打赏面板
.reward-panel {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  @apply bg-ob-deep-800 rounded-2xl;
  width: 360px;
  max-width: 90vw;
  z-index: 1001;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.08);
  overflow: hidden;
}

.panel-header {
  @apply flex items-center justify-between px-6 py-4;
  background: linear-gradient(135deg, rgba(255, 107, 107, 0.15) 0%, rgba(255, 142, 83, 0.1) 100%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.panel-title {
  @apply text-lg font-semibold;
  background: linear-gradient(135deg, #ff6b6b, #ff8e53);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.close-btn {
  @apply p-1.5 rounded-lg;
  background: rgba(255, 255, 255, 0.06);
  transition: all 0.2s;

  &:hover {
    background: rgba(255, 255, 255, 0.12);
  }

  svg {
    width: 18px;
    height: 18px;
    color: var(--text-dim);
  }
}

.panel-body {
  @apply p-6;
}

// 标签切换
.qrcode-tabs {
  @apply flex gap-3 mb-6;
}

.tab-btn {
  @apply flex-1 flex items-center justify-center gap-2 py-3 px-4 rounded-xl;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: all 0.2s;
  color: var(--text-dim);

  &.active {
    background: linear-gradient(135deg, rgba(255, 107, 107, 0.15) 0%, rgba(255, 142, 83, 0.1) 100%);
    border-color: rgba(255, 107, 107, 0.3);
    color: #ff6b6b;
  }

  &:hover:not(.active) {
    background: rgba(255, 255, 255, 0.08);
  }

  svg {
    width: 20px;
    height: 20px;
  }

  span {
    @apply text-sm font-medium;
  }
}

// 二维码展示
.qrcode-display {
  @apply flex justify-center;
}

.qrcode-wrapper {
  @apply flex flex-col items-center gap-3;
}

.qrcode-image {
  width: 200px;
  height: 200px;
  border-radius: 12px;
  background: white;
  padding: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
}

.qrcode-label {
  @apply text-sm text-ob-dim opacity-70;
}

// 底部
.reward-footer {
  @apply mt-6 pt-4;
  border-top: 1px dashed rgba(255, 255, 255, 0.08);
}

.footer-text {
  @apply text-center text-sm text-ob-dim opacity-60;
}

// 遮罩层
.reward-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 1000;
}

// 动画
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.3s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.9);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

// 响应式
@media (max-width: 640px) {
  .reward-panel {
    width: 95vw;
  }

  .qrcode-image {
    width: 180px;
    height: 180px;
  }
}
</style>

<template>
  <div class="oauth-background">
    <div class="oauth-content">
      <div class="loader-bar">
        <span></span>
        <span></span>
        <span></span>
        <span></span>
        <span></span>
      </div>
      <p class="oauth-hint">正在登录中...</p>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, getCurrentInstance } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useUserStore } from '@/stores/user'
import api from '@/api/api'
export default defineComponent({
  name: 'OauthLoginModel',
  setup() {
    const proxy: any = getCurrentInstance()?.appContext.config.globalProperties
    const userStore = useUserStore()
    const route = useRoute()
    const router = useRouter()
    if (route.path === '/oauth/login/qq') {
      //@ts-ignore
      if (QC.Login.check()) {
        //@ts-ignore
        QC.Login.getMe(function (openId, accessToken) {
          let params = {
            openId: openId,
            accessToken: accessToken
          }
          api.qqLogin(params).then(({ data }) => {
            if (data.flag) {
              userStore.userInfo = data.data
              userStore.token = data.data.token
              sessionStorage.setItem('token', data.data.token)
              proxy.$notify({
                title: 'Success',
                message: '登录成功',
                type: 'success'
              })
            }
          })
          if (userStore.currentUrl === '') {
            router.push({ path: '/' })
          } else {
            router.push({ path: userStore.currentUrl })
          }
        })
      }
    }
  }
})
</script>
<style lang="scss" scoped>
.oauth-background {
  position: fixed;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--background-primary);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.oauth-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.oauth-hint {
  font-size: 0.9rem;
  color: var(--text-dim);
  letter-spacing: 0.04em;
  animation: hint-pulse 1.5s ease-in-out infinite;
}

.loader-bar {
  position: relative;
  width: 56px;
  height: 32px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 3px;
}

.loader-bar span {
  display: block;
  width: 6px;
  height: 6px;
  border-radius: 3px;
  background: var(--text-accent);
  animation: loader-bounce 1.2s ease-in-out infinite;
}

.loader-bar span:nth-child(2) { animation-delay: 0.1s; }
.loader-bar span:nth-child(3) { animation-delay: 0.2s; }
.loader-bar span:nth-child(4) { animation-delay: 0.3s; }
.loader-bar span:nth-child(5) { animation-delay: 0.4s; }

@keyframes loader-bounce {
  0%, 100% {
    height: 6px;
    opacity: 0.4;
  }
  50% {
    height: 24px;
    opacity: 1;
  }
}

@keyframes hint-pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}
</style>

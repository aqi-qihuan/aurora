/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// QQ 登录 SDK 全局变量声明
declare global {
  const QC: {
    Login: {
      showPopup: (options: { appId: string; redirectURI: string }) => void
      check: () => boolean
      getMe: (callback: (openId: string, accessToken: string) => void) => void
    }
  }
}

export {}

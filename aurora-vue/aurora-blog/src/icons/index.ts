import SvgIcon from '@/components/SvgIcon/index.vue' // svg component
import { App } from 'vue'

// register globally
export const registerSvgIcon = (app: App): void => {
  app.component('svg-icon', SvgIcon)
  
  // Vite 使用 import.meta.glob 替代 require.context
  const modules = import.meta.glob('./svg/*.svg', { eager: true })
  // 导入所有 SVG 文件
  Object.values(modules)
}

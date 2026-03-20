import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import pinia from './stores'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import request from './utils/request'

// 导入自定义样式
import './styles/variables.css'
import './styles/components.css'

// 导入路由权限控制
import './permission'

// 导入自定义指令
import setupDirectives from './directives/permission'

// 创建应用实例
const app = createApp(App)

// 注册所有 Element Plus 图标
// 注意：Element Plus 图标支持 tree-shaking，只有实际使用的图标会被打包
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 注册自定义指令
setupDirectives(app)

// 全局属性
app.config.globalProperties.$axios = request
app.config.globalProperties.$moment = dayjs

// 使用插件
app.use(pinia)
app.use(router)
app.use(ElementPlus, { 
  locale: zhCn,
  zIndex: 3000
})

// 挂载应用
app.mount('#app')

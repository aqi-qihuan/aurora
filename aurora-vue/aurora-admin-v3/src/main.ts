import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import pinia from './stores'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
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

// 注册自定义指令
setupDirectives(app)

// 使用插件
app.use(pinia)
app.use(router)
app.use(ElementPlus, { 
  locale: zhCn,
  zIndex: 3000
})

// 挂载应用
app.mount('#app')

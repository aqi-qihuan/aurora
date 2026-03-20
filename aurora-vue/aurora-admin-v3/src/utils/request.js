import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = sessionStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = 'Bearer ' + token
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { data } = response
    
    // 处理特定错误码
    if (data.code !== 20000 && data.code !== 200) {
      // 显示后端返回的具体错误信息
      ElMessage.error(data.message || '操作失败')
      
      // 处理特定错误码
      if (data.code === 40001) {
        // Token 过期或无效
        sessionStorage.removeItem('token')
        router.push({ path: '/login' })
      }
    }
    
    return response
  },
  (error) => {
    // 网络错误或服务器错误
    const message = error.response?.data?.message || error.message || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default request

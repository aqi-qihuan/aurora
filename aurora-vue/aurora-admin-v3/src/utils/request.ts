import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

/**
 * 通用 API 响应类型
 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

/**
 * 扩展 AxiosRequestConfig，支持泛型响应类型
 */
export interface RequestConfig<T = unknown> extends AxiosRequestConfig {
  // 可以添加自定义配置
}

/**
 * 创建类型化的请求实例
 */
const request: AxiosInstance = axios.create({
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
  (response: AxiosResponse<ApiResponse>) => {
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

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { fileURLToPath, URL } from 'node:url'
import viteCompression from 'vite-plugin-compression'

// 获取项目根目录（兼容 Windows）
const projectRoot = fileURLToPath(new URL('.', import.meta.url))

export default defineConfig({
  plugins: [
    vue(),
    // Gzip 压缩
    viteCompression({
      algorithm: 'gzip',
      ext: '.gz',
      threshold: 10240, // 大于 10KB 才压缩
      deleteOriginFile: false
    })
  ],
  
  resolve: {
    alias: {
      '@': resolve(projectRoot, 'src')
    }
  },
  
  server: {
    port: 8080,
    host: true, // 允许外部访问
    open: false, // 不自动打开浏览器
    cors: true,
    proxy: {
      '/api': {
        target: 'https://www.aqi125.cn',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      }
    }
  },
  
  build: {
    outDir: 'dist',
    sourcemap: false,

    // 代码分割优化
    rollupOptions: {
      output: {
        // 手动分割代码块 - 解决循环依赖问题
        manualChunks: (id) => {
          // Vue 核心（必须优先处理，避免循环依赖）
          if (id.includes('node_modules/vue/') ||
              id.includes('node_modules/@vue/') ||
              id.includes('node_modules/vue-router/') ||
              id.includes('node_modules/pinia/')) {
            return 'vue-core'
          }
          // Element Plus（依赖 vue-core）
          if (id.includes('node_modules/element-plus/') ||
              id.includes('node_modules/@element-plus/')) {
            return 'element-plus'
          }
          // ECharts（独立，无循环依赖）
          if (id.includes('node_modules/echarts/') ||
              id.includes('node_modules/vue-echarts/') ||
              id.includes('node_modules/zrender/')) {
            return 'echarts'
          }
          // Axios
          if (id.includes('node_modules/axios/')) {
            return 'axios'
          }
          // 工具库
          if (id.includes('node_modules/lodash-es/') ||
              id.includes('node_modules/dayjs/') ||
              id.includes('node_modules/nprogress/')) {
            return 'utils'
          }
          // 其他 node_modules
          if (id.includes('node_modules')) {
            return 'vendor'
          }
        },

        // 资源文件命名
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]'
      }
    },

    // 压缩配置
    minify: 'terser',
    terserOptions: {
      compress: {
        // 生产环境移除 console
        drop_console: true,
        drop_debugger: true,
        // 移除无用代码
        dead_code: true,
        // 优化常量
        collapse_vars: true,
        reduce_vars: true
      }
    },

    // CSS 代码分割
    cssCodeSplit: true,

    // chunk 大小警告阈值
    chunkSizeWarningLimit: 1000,

    // 启用 Rollup 的模块预加载
    modulePreload: {
      polyfill: true
    }
  },
  
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "@/styles/element-plus.scss" as *;`
      }
    }
  },
  
  // 依赖优化
  optimizeDeps: {
    include: [
      'vue',
      'vue-router',
      'pinia',
      'axios',
      'element-plus',
      '@element-plus/icons-vue',
      'echarts',
      'vue-echarts'
    ],
    // 预构建排除
    exclude: []
  }
})

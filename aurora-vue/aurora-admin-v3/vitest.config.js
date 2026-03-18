import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { fileURLToPath, URL } from 'node:url'

// 获取项目根目录（兼容 Windows）
const projectRoot = fileURLToPath(new URL('.', import.meta.url))

export default defineConfig({
  plugins: [vue()],
  
  test: {
    // 测试环境
    environment: 'jsdom',
    
    // 全局变量（describe, it, expect 等）
    globals: true,
    
    // 测试文件匹配模式
    include: [
      'tests/**/*.{test,spec}.{mjs,cjs,js,ts}'
    ],
    
    // 排除文件
    exclude: [
      'node_modules',
      'dist',
      '.idea',
      '.git',
      '.cache'
    ],
    
    // 覆盖率配置
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      reportsDirectory: './coverage',
      include: [
        'src/**/*.{js,ts,vue}'
      ],
      exclude: [
        'src/main.js',
        'src/App.vue',
        'src/**/*.test.js',
        'src/**/*.spec.js'
      ],
      // 覆盖率阈值
      thresholds: {
        lines: 60,
        functions: 60,
        branches: 50,
        statements: 60
      }
    },
    
    // 测试超时时间（毫秒）
    testTimeout: 10000,
    
    // 钩子函数超时时间
    hookTimeout: 10000,
    
    // 并行执行
    threads: true,
    
    // 监听模式
    watch: false,
    
    // 报告器
    reporters: ['default'],
    
    // 设置文件
    // setupFiles: ['./tests/setup.js'],
    
    // 模拟日期
    fakeTimers: {
      toFake: ['setTimeout', 'clearTimeout', 'setInterval', 'clearInterval', 'Date']
    },
    
    // 池配置
    pool: 'threads',
    poolOptions: {
      threads: {
        singleThread: true
      }
    }
  },
  
  resolve: {
    alias: {
      '@': resolve(projectRoot, 'src')
    }
  },
  
  // 依赖优化
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'axios', 'element-plus']
  }
})

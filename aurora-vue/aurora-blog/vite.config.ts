import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'
import { prismjsPlugin } from 'vite-plugin-prismjs'

export default defineConfig({
  plugins: [
    vue(),
    // SVG sprite icon support (replaces svg-sprite-loader)
    createSvgIconsPlugin({
      iconDirs: [resolve(__dirname, 'src/icons/svg')],
      symbolId: 'icon-[name]'
    }),
    // PrismJS syntax highlighting
    prismjsPlugin({
      languages: ['javascript', 'css', 'sql', 'java', 'c', 'cpp', 'nginx', 'markup', 'shell', 'json'],
      plugins: ['line-numbers', 'toolbar', 'copy-to-clipboard'],
      theme: 'okaidia',
      css: true
    })
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    },
    // Vite requires explicit alias for vue-i18n (replaces chainWebpack alias)
    dedupe: ['vue']
  },
  // 修复 tocbot/prismjs 等 UMD 库在 ESM 环境下 global 未定义的问题
  define: {
    global: 'globalThis'
  },
  css: {
    preprocessorOptions: {
      scss: {
        // silence deprecation warnings from legacy Sass
        silenceDeprecations: ['legacy-js-api']
      }
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'https://www.aqi125.cn',
        changeOrigin: true,
        // 不要移除 /api 前缀，后端路由需要完整的路径
        // rewrite: (path) => path.replace(/^\/api/, '') // 错误：这会移除 /api 导致请求变成首页
      }
    }
  },
  build: {
    target: 'es2015',
    sourcemap: false,
    commonjsOptions: {
      // 忽略 UMD wrapper 中对 global/this 的依赖，让 Vite 正确处理
      ignoreTryCatch: false
    },
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes('node_modules')) {
            if (id.includes('vue') || id.includes('vue-router') || id.includes('pinia')) {
              return 'vue';
            }
            if (id.includes('element-plus')) {
              return 'elementPlus';
            }
            if (id.includes('markdown-it') || id.includes('mavon-editor')) {
              return 'markdown';
            }
            if (id.includes('prismjs')) {
              return 'prism';
            }
            if (id.includes('tocbot')) {
              return 'tocbot';
            }
          }
        }
      }
    }
  }
})

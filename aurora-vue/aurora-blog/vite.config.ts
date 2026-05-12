import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'
import { prismjsPlugin } from 'vite-plugin-prismjs'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    // Auto-import Vue and Element Plus components (enables tree-shaking)
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts'
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts'
    }),
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
          // Vue ecosystem - separate chunks for better caching
          if (id.includes('node_modules')) {
            // Vue core (reactivity, runtime-core, etc.)
            if (id.includes('vue/dist') || id.includes('@vue/')) {
              return 'vue-core';
            }
            // Vue Router
            if (id.includes('vue-router')) {
              return 'vue-router';
            }
            // Pinia
            if (id.includes('pinia')) {
              return 'pinia';
            }
            // Vue i18n
            if (id.includes('vue-i18n') || id.includes('@intlify/')) {
              return 'vue-i18n';
            }
            // Element Plus - separate chunk
            if (id.includes('element-plus')) {
              return 'element-plus';
            }
            // Markdown processing
            if (id.includes('markdown-it') || id.includes('mavon-editor') || id.includes('prismjs')) {
              return 'markdown';
            }
            // Tocbot
            if (id.includes('tocbot')) {
              return 'tocbot';
            }
            // Axios
            if (id.includes('axios')) {
              return 'axios';
            }
            // Other node_modules - group by first-level directory
            const match = id.match(/node_modules\/([^/]+)/);
            if (match) {
              return `vendor-${match[1]}`;
            }
          }
        }
      }
    }
  }
})

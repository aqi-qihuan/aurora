<template>
  <div class="copyright-statement">
    <!-- 标题 -->
    <div class="copyright-header">
      <svg class="copyright-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/>
        <path d="M12 16v-4M12 8h.01"/>
      </svg>
      <span class="copyright-title">版权声明</span>
    </div>

    <!-- 内容区 -->
    <div class="copyright-body">
      <!-- 作者 -->
      <div class="copyright-item">
        <div class="item-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
            <circle cx="12" cy="7" r="4"/>
          </svg>
        </div>
        <div class="item-content">
          <span class="item-label">本文作者</span>
          <a
            :href="article.author?.website || websiteConfig.authorWebsite || '#'"
            target="_blank"
            rel="noopener noreferrer"
            class="item-value item-link">
            {{ article.author?.nickname || websiteConfig.author || 'Author' }}
          </a>
        </div>
      </div>

      <!-- 链接 -->
      <div class="copyright-item">
        <div class="item-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
            <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
          </svg>
        </div>
        <div class="item-content">
          <span class="item-label">本文链接</span>
          <span class="item-value item-link copy-trigger" @click="copyLink">
            <span class="link-text">{{ articleLink }}</span>
            <svg class="copy-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
            </svg>
          </span>
        </div>
      </div>

      <!-- 许可协议 -->
      <div class="copyright-item">
        <div class="item-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
        </div>
        <div class="item-content">
          <span class="item-label">版权声明</span>
          <span class="item-value license-text">
            本站所有文章除特别声明外，均采用
            <a
              href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
              target="_blank"
              rel="noopener noreferrer"
              class="license-link">
              CC BY-NC-SA 4.0
            </a>
            许可协议。转载请注明文章出处！
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'
import { useAppStore } from '@/stores/app'

interface Author {
  nickname?: string
  avatar?: string
  website?: string
}

interface Article {
  id?: string | number
  articleTitle?: string
  author?: Author
}

export default defineComponent({
  name: 'CopyrightStatement',
  props: {
    article: {
      type: Object as PropType<Article>,
      required: true
    }
  },
  setup(props) {
    const appStore = useAppStore()

    const websiteConfig = computed(() => appStore.websiteConfig || {})

    const articleLink = computed(() => {
      const baseUrl = window.location.origin
      const articleId = props.article.id
      return articleId ? `${baseUrl}/articles/${articleId}` : window.location.href
    })

    const copyLink = async () => {
      try {
        await navigator.clipboard.writeText(articleLink.value)
      } catch (err) {
        const textArea = document.createElement('textarea')
        textArea.value = articleLink.value
        document.body.appendChild(textArea)
        textArea.select()
        document.execCommand('copy')
        document.body.removeChild(textArea)
      }
    }

    return {
      websiteConfig,
      articleLink,
      copyLink
    }
  }
})
</script>

<style lang="scss" scoped>
.copyright-statement {
  @apply bg-ob-deep-800 rounded-2xl my-8 overflow-hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

// 标题栏
.copyright-header {
  @apply flex items-center gap-3 px-6 py-4;
  background: linear-gradient(135deg, rgba(var(--text-accent-rgb, 99, 102, 241), 0.15) 0%, rgba(var(--text-accent-rgb, 99, 102, 241), 0.05) 100%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.copyright-icon {
  @apply flex-shrink-0;
  width: 20px;
  height: 20px;
  color: var(--text-accent, #6366f1);
}

.copyright-title {
  @apply text-sm font-semibold tracking-wide;
  color: var(--text-accent, #6366f1);
}

// 内容区
.copyright-body {
  @apply p-6;
}

.copyright-item {
  @apply flex items-start gap-4;
  padding: 12px 0;

  &:not(:last-child) {
    border-bottom: 1px dashed rgba(255, 255, 255, 0.08);
  }

  &:first-child {
    padding-top: 0;
  }

  &:last-child {
    padding-bottom: 0;
    border-bottom: none;
  }
}

// 图标
.item-icon {
  @apply flex-shrink-0 flex items-center justify-center;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: rgba(var(--text-accent-rgb, 99, 102, 241), 0.1);
  color: var(--text-accent, #6366f1);

  svg {
    width: 18px;
    height: 18px;
  }
}

// 内容
.item-content {
  @apply flex-1 flex flex-col gap-1 min-w-0;
}

.item-label {
  @apply text-xs font-medium uppercase tracking-wider opacity-60;
  color: var(--text-bright, #f3f4f6);
}

.item-value {
  @apply text-sm leading-relaxed;
  color: var(--text-normal, #d1d5db);
}

// 链接样式
.item-link {
  @apply cursor-pointer inline-flex items-center gap-2;
  color: var(--text-accent, #6366f1);
  text-decoration: none;
  transition: all 0.2s ease;

  &:hover {
    opacity: 0.8;
  }
}

.link-text {
  @apply flex-1;
  word-break: break-all;
}

.copy-icon {
  @apply flex-shrink-0 opacity-50;
  width: 14px;
  height: 14px;
  transition: opacity 0.2s;
}

.copy-trigger:hover .copy-icon {
  opacity: 1;
}

// 许可协议文本
.license-text {
  @apply leading-relaxed;
}

.license-link {
  color: var(--text-accent, #6366f1);
  text-decoration: none;
  font-weight: 500;
  padding: 0 2px;
  border-radius: 4px;
  transition: all 0.2s ease;

  &:hover {
    background: rgba(var(--text-accent-rgb, 99, 102, 241), 0.15);
  }
}

// 响应式
@media (max-width: 640px) {
  .copyright-header {
    @apply px-4 py-3;
  }

  .copyright-body {
    @apply p-4;
  }

  .copyright-item {
    @apply gap-3;
    padding: 10px 0;
  }

  .item-icon {
    width: 32px;
    height: 32px;

    svg {
      width: 16px;
      height: 16px;
    }
  }

  .item-content {
    @apply gap-0.5;
  }
}
</style>

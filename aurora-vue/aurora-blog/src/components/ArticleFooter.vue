<template>
  <div class="article-meta-footer" v-if="hasData">
    <div class="meta-divider">
      <span class="divider-line"></span>
      <svg class="divider-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span class="divider-line"></span>
    </div>
    <div class="meta-content">
      <div class="meta-item" v-if="article.createTime">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
          <line x1="16" y1="2" x2="16" y2="6" />
          <line x1="8" y1="2" x2="8" y2="6" />
          <line x1="3" y1="10" x2="21" y2="10" />
        </svg>
        <span class="meta-label">发布于</span>
        <span class="meta-value">{{ formatDate(article.createTime) }}</span>
      </div>
      <div class="meta-item" v-if="article.updateTime && article.updateTime !== article.createTime">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M1 4v6h6" />
          <path d="M3.51 15a9 9 0 102.13-9.36L1 10" />
        </svg>
        <span class="meta-label">更新于</span>
        <span class="meta-value">{{ formatDate(article.updateTime) }}</span>
      </div>
      <div class="meta-item" v-if="article.categoryName">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
        </svg>
        <span class="meta-label">分类</span>
        <span class="meta-value meta-link">{{ article.categoryName }}</span>
      </div>
      <div class="meta-item" v-if="article.tags && article.tags.length > 0">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z" />
          <line x1="7" y1="7" x2="7.01" y2="7" />
        </svg>
        <span class="meta-label">标签</span>
        <span class="meta-value">
          <span v-for="(tag, index) in article.tags" :key="tag.id" class="tag-badge">
            #{{ tag.tagName }}<template v-if="index < article.tags.length - 1"> </template>
          </span>
        </span>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from 'vue'

interface ArticleFooterArticle {
  id?: string | number
  createTime?: string
  updateTime?: string
  categoryName?: string
  tags?: Array<{ id: string | number; tagName: string }>
}

export default defineComponent({
  name: 'ArticleFooter',
  props: {
    article: {
      type: Object as PropType<ArticleFooterArticle>,
      required: true
    }
  },
  setup(props) {
    const hasData = computed(() => {
      return props.article && (props.article.createTime || props.article.categoryName)
    })

    const formatDate = (dateStr: string) => {
      if (!dateStr) return ''
      const d = new Date(dateStr)
      const year = d.getFullYear()
      const month = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
      return `${months[d.getMonth()]} ${day}, ${year}`
    }

    return { hasData, formatDate }
  }
})
</script>

<style lang="scss" scoped>
.article-meta-footer {
  @apply my-8;
}

.meta-divider {
  @apply flex items-center justify-center gap-4 mb-6;
}

.divider-line {
  @apply flex-1 h-px;
  background: linear-gradient(
    90deg,
    transparent,
    var(--text-accent) 20%,
    var(--text-accent) 80%,
    transparent
  );
  opacity: 0.25;
}

.divider-icon {
  @apply flex-shrink-0;
  width: 18px;
  height: 18px;
  color: var(--text-accent);
  opacity: 0.5;
}

.meta-content {
  @apply flex flex-wrap gap-x-6 gap-y-3;
  padding: 1rem 1.25rem;
  background: var(--bg-accent-05);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.04);
}

.meta-item {
  @apply flex items-center gap-2 text-sm;
  color: var(--text-dim);

  svg {
    @apply flex-shrink-0;
    width: 15px;
    height: 15px;
    opacity: 0.6;
  }
}

.meta-label {
  @apply text-xs uppercase tracking-wider font-medium;
  opacity: 0.5;
}

.meta-value {
  color: var(--text-normal);
}

.meta-link {
  color: var(--text-accent);
}

.tag-badge {
  color: var(--text-accent);
  opacity: 0.8;
  transition: opacity 0.2s;

  &:hover {
    opacity: 1;
  }
}

@media (max-width: 640px) {
  .meta-content {
    @apply gap-y-2;
    padding: 0.75rem 1rem;
  }

  .meta-item {
    @apply text-xs;
  }
}
</style>

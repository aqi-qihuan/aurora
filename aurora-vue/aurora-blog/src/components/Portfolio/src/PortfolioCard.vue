<template>
  <div class="portfolio-card group" @click="openRepo">
    <div class="portfolio-cover">
      <img v-if="data.cover" v-lazy="data.cover" :alt="data.name" />
      <div v-else class="cover-fallback" :style="coverStyle">
        <span class="cover-letter">{{ firstLetter }}</span>
        <span class="cover-lang">{{ data.language || 'Code' }}</span>
      </div>
      <span class="star-badge">
        <svg-icon icon-class="star" />
        {{ formatCount(data.stargazersCount) }}
      </span>
    </div>

    <div class="portfolio-content">
      <div class="portfolio-header">
        <svg-icon icon-class="github" class="repo-icon" />
        <h3 class="portfolio-name">{{ data.name || placeholder.name }}</h3>
        <span v-if="data.isFeatured === 1" class="featured-tag">
          <svg-icon icon-class="hot" /> {{ t('home.featured') }}
        </span>
      </div>

      <p class="portfolio-desc">{{ data.description || placeholder.desc }}</p>

      <ul v-if="topics && topics.length > 0" class="portfolio-topics">
        <li v-for="topic in topics.slice(0, 4)" :key="topic"># {{ topic }}</li>
      </ul>
      <ul v-else class="portfolio-topics">
        <li># {{ data.language || 'Code' }}</li>
      </ul>

      <div class="portfolio-meta">
        <span class="lang-dot" :style="{ background: langColor }"></span>
        <span class="meta-text">{{ data.language || t('home.unknown_lang') }}</span>
        <span class="meta-divider">·</span>
        <svg-icon icon-class="clock" class="meta-icon" />
        <span class="meta-text">{{ relativeTime }}</span>
      </div>

      <div class="portfolio-actions">
        <a :href="data.htmlUrl" target="_blank" rel="noopener" class="btn-primary" @click.stop>
          <svg-icon icon-class="github" />
          {{ t('home.view_repo') }}
        </a>
        <a
          v-if="data.homepage"
          :href="data.homepage"
          target="_blank"
          rel="noopener"
          class="btn-ghost"
          @click.stop>
          <svg-icon icon-class="external" />
          {{ t('home.demo') }}
        </a>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PortfolioItem } from './types'

const LANG_COLORS: Record<string, string> = {
  JavaScript: '#f1e05a',
  TypeScript: '#3178c6',
  Go: '#00ADD8',
  Java: '#b07219',
  Vue: '#41b883',
  Python: '#3572A5',
  HTML: '#e34c26',
  CSS: '#563d7c',
  Shell: '#89e051',
  C: '#555555',
  'C++': '#f34b7d',
  Rust: '#dea584',
  PHP: '#4F5D95',
  Ruby: '#701516',
  Kotlin: '#A97BFF',
  Swift: '#F05138',
  Dart: '#00B4AB',
  Lua: '#000080',
  Dockerfile: '#384d54',
  Makefile: '#427819'
}

export default defineComponent({
  name: 'PortfolioCard',
  props: {
    data: {
      type: Object as PropType<Partial<PortfolioItem>>,
      default: () => ({})
    }
  },
  setup(props) {
    const { t } = useI18n()

    const firstLetter = computed(() => {
      const name = props.data?.name || ''
      return name ? name.charAt(0).toUpperCase() : '?'
    })

    const langColor = computed(() => {
      const lang = props.data?.language || ''
      return LANG_COLORS[lang] || '#888888'
    })

    const coverStyle = computed(() => ({
      background: `linear-gradient(135deg, ${langColor.value}22 0%, ${langColor.value}44 100%)`,
      borderColor: langColor.value
    }))

    const topics = computed(() => props.data?.topics || [])

    const relativeTime = computed(() => {
      const t = props.data?.repoUpdatedAt
      if (!t) return ''
      const date = new Date(t)
      if (isNaN(date.getTime())) return ''
      const now = Date.now()
      const diff = now - date.getTime()
      const day = 24 * 3600 * 1000
      const days = Math.floor(diff / day)
      if (days < 1) return 'today'
      if (days < 30) return `${days}d ago`
      if (days < 365) return `${Math.floor(days / 30)}mo ago`
      return `${Math.floor(days / 365)}y ago`
    })

    const formatCount = (n: number | undefined) => {
      if (n === undefined || n === null) return '0'
      if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
      return String(n)
    }

    const openRepo = () => {
      if (props.data?.htmlUrl) {
        window.open(props.data.htmlUrl, '_blank', 'noopener')
      }
    }

    return {
      firstLetter,
      langColor,
      coverStyle,
      topics,
      relativeTime,
      formatCount,
      openRepo,
      t,
      placeholder: {
        name: 'Loading...',
        desc: 'Loading project description...'
      }
    }
  }
})
</script>

<style lang="scss" scoped>
.portfolio-card {
  display: flex;
  flex-direction: column;
  background: var(--background-secondary, #fff);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  height: 100%;
  border: 1px solid rgba(0, 0, 0, 0.04);

  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.12);
    .portfolio-cover img {
      transform: scale(1.05);
    }
    .cover-fallback {
      transform: scale(1.03);
    }
  }
}

.portfolio-cover {
  position: relative;
  width: 100%;
  height: 160px;
  overflow: hidden;
  background: #1a1a2e;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.4s ease;
  }

  .cover-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    transition: transform 0.4s ease;
    border-bottom: 3px solid;

    .cover-letter {
      font-size: 56px;
      font-weight: 700;
      color: #fff;
      text-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
      line-height: 1;
    }

    .cover-lang {
      font-size: 12px;
      letter-spacing: 0.1em;
      color: rgba(255, 255, 255, 0.7);
      text-transform: uppercase;
    }
  }

  .star-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    padding: 4px 10px;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(8px);
    border-radius: 999px;
    color: #fbbf24;
    font-size: 12px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 4px;

    svg {
      width: 12px;
      height: 12px;
    }
  }
}

.portfolio-content {
  flex: 1;
  padding: 18px 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.portfolio-header {
  display: flex;
  align-items: center;
  gap: 8px;

  .repo-icon {
    width: 18px;
    height: 18px;
    color: var(--text-secondary, #666);
    flex-shrink: 0;
  }

  .portfolio-name {
    font-size: 17px;
    font-weight: 600;
    color: var(--text-primary, #1a1a1a);
    margin: 0;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .featured-tag {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 8px;
    background: linear-gradient(135deg, #f97316, #ef4444);
    color: #fff;
    font-size: 11px;
    font-weight: 500;
    border-radius: 999px;
    flex-shrink: 0;

    svg {
      width: 10px;
      height: 10px;
    }
  }
}

.portfolio-desc {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary, #666);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 42px;
}

.portfolio-topics {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;

  li {
    font-size: 11px;
    color: var(--text-tertiary, #999);
    padding: 3px 8px;
    background: var(--background-tertiary, #f5f5f5);
    border-radius: 6px;
  }
}

.portfolio-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary, #888);
  margin-top: 2px;

  .lang-dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
  }

  .meta-text {
    color: var(--text-secondary, #888);
  }

  .meta-divider {
    color: var(--text-tertiary, #ccc);
  }

  .meta-icon {
    width: 12px;
    height: 12px;
  }
}

.portfolio-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
  padding-top: 8px;

  a {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    padding: 7px 14px;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    text-decoration: none;
    transition: all 0.2s ease;
    cursor: pointer;

    svg {
      width: 14px;
      height: 14px;
    }
  }

  .btn-primary {
    flex: 1;
    background: #24292e;
    color: #fff;

    &:hover {
      background: #1a1e22;
      transform: translateY(-1px);
    }
  }

  .btn-ghost {
    flex: 1;
    background: transparent;
    color: var(--text-primary, #333);
    border: 1px solid var(--border-color, #e5e7eb);

    &:hover {
      background: var(--background-tertiary, #f5f5f5);
      border-color: var(--text-tertiary, #ccc);
    }
  }
}

/* 深色主题适配 */
:global(html.dark) .portfolio-card {
  background: var(--background-secondary, #1f2937);
  border-color: rgba(255, 255, 255, 0.06);

  .portfolio-name {
    color: var(--text-primary, #f3f4f6);
  }
  .portfolio-desc {
    color: var(--text-secondary, #9ca3af);
  }
  .btn-ghost {
    color: var(--text-primary, #e5e7eb);
    border-color: rgba(255, 255, 255, 0.1);
  }
  .portfolio-topics li {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-tertiary, #9ca3af);
  }
}
</style>

<template>
  <div class="portfolio-card group" @click="openRepo">
    <!-- 切角装饰 -->
    <span class="corner corner-tl"></span>
    <span class="corner corner-br"></span>

    <!-- 封面 -->
    <div class="portfolio-cover">
      <img v-if="data.cover" v-lazy="data.cover" :alt="data.name" />
      <div v-else class="cover-fallback" :style="coverStyle">
        <span class="cover-letter">{{ firstLetter }}</span>
        <span class="cover-lang">{{ (data.language || 'CODE').toUpperCase() }}</span>
      </div>
      <!-- 渐变遮罩增强可读性 -->
      <div class="cover-overlay"></div>
      <!-- star 角标 -->
      <span class="star-badge">
        <svg-icon icon-class="star" />
        {{ formatCount(data.stargazersCount) }}
      </span>
      <!-- 语言色条 -->
      <span class="lang-bar" :style="{ background: langColor }"></span>
    </div>

    <!-- 内容区 -->
    <div class="portfolio-content">
      <div class="portfolio-header">
        <span class="header-line"></span>
        <span class="header-label">{{ t('home.project_label') || 'PROJECT' }}</span>
        <span v-if="data.isFeatured === 1" class="featured-tag">
          <svg-icon icon-class="hot" /> {{ t('home.featured') }}
        </span>
      </div>

      <h3 class="portfolio-name">{{ data.name || placeholder.name }}</h3>

      <p class="portfolio-desc">{{ data.description || placeholder.desc }}</p>

      <ul v-if="topics && topics.length > 0" class="portfolio-topics">
        <li v-for="topic in topics.slice(0, 4)" :key="topic"># {{ topic }}</li>
      </ul>
      <ul v-else class="portfolio-topics">
        <li># {{ data.language || 'Code' }}</li>
      </ul>

      <div class="portfolio-divider"></div>

      <div class="portfolio-meta">
        <span class="lang-dot" :style="{ background: langColor }"></span>
        <span class="meta-text">{{ data.language || t('home.unknown_lang') }}</span>
        <span class="meta-divider"></span>
        <svg-icon icon-class="clock" class="meta-icon" />
        <span class="meta-text">{{ relativeTime }}</span>
        <span class="meta-divider"></span>
        <svg-icon icon-class="github" class="meta-icon" />
        <span class="meta-text">{{ formatCount(data.forksCount) }} forks</span>
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
      background: `linear-gradient(135deg, ${langColor.value}18 0%, ${langColor.value}30 100%)`,
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
/* ============ HOK 风格变量 ============ */
$gold: #c9a961;
$gold-bright: #f4d47c;
$cyan: #00d4ff;
$cyan-dim: #0099cc;
$bg-deep: #0a0e1a;
$bg-card: #0f1424;
$bg-cover: #1a2240;
$border-dim: #1f2942;
$border-mid: #2a3454;
$text-primary: #e8ecf5;
$text-secondary: #8a93b0;
$text-tertiary: #5a6488;

.portfolio-card {
  position: relative;
  display: flex;
  flex-direction: column;
  background: $bg-card;
  border-radius: 2px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
  height: 100%;
  border: 0.5px solid rgba(201, 169, 97, 0.25);

  /* 切角装饰 */
  .corner {
    position: absolute;
    width: 14px;
    height: 14px;
    z-index: 3;
    pointer-events: none;

    &.corner-tl {
      top: 0;
      left: 0;
      background: $gold;
      clip-path: polygon(0 0, 100% 0, 0 100%);
    }

    &.corner-br {
      bottom: 0;
      right: 0;
      background: $gold;
      clip-path: polygon(100% 0, 100% 100%, 0 100%);
    }
  }

  &:hover {
    transform: translateY(-6px);
    border-color: $cyan;
    box-shadow:
      0 8px 32px rgba(0, 212, 255, 0.15),
      0 0 0 0.5px rgba(0, 212, 255, 0.4);

    .corner-tl,
    .corner-br {
      background: $cyan;
    }

    .portfolio-cover img {
      transform: scale(1.06);
    }
    .cover-fallback {
      transform: scale(1.03);
    }
    .lang-bar {
      height: 4px;
    }
  }
}

/* ============ 封面 ============ */
.portfolio-cover {
  position: relative;
  width: 100%;
  height: 170px;
  overflow: hidden;
  background: $bg-cover;

  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.5s ease;
  }

  .cover-fallback {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    transition: transform 0.5s ease;
    border-bottom: 0.5px solid;

    .cover-letter {
      font-size: 56px;
      font-weight: 700;
      color: #fff;
      line-height: 1;
      letter-spacing: 0.02em;
    }

    .cover-lang {
      font-size: 11px;
      letter-spacing: 0.2em;
      color: rgba(255, 255, 255, 0.65);
      font-family: ui-monospace, 'SF Mono', Menlo, monospace;
    }
  }

  /* 渐变遮罩 */
  .cover-overlay {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to bottom,
      rgba(10, 14, 26, 0) 0%,
      rgba(10, 14, 26, 0.3) 50%,
      rgba(15, 20, 36, 0.95) 100%
    );
    pointer-events: none;
  }

  /* star 角标 */
  .star-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    padding: 4px 10px;
    background: rgba(10, 14, 26, 0.85);
    border: 0.5px solid rgba(201, 169, 97, 0.4);
    border-radius: 2px;
    color: $gold-bright;
    font-size: 11px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 4px;
    backdrop-filter: blur(4px);

    svg {
      width: 11px;
      height: 11px;
    }
  }

  /* 底部语言色条 */
  .lang-bar {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 2px;
    transition: height 0.3s ease;
  }
}

/* ============ 内容区 ============ */
.portfolio-content {
  flex: 1;
  padding: 20px 22px 22px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.portfolio-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 2px;

  .header-line {
    width: 20px;
    height: 1.5px;
    background: $gold;
    flex-shrink: 0;
  }

  .header-label {
    font-size: 10px;
    letter-spacing: 0.2em;
    color: $gold;
    font-weight: 500;
    font-family: ui-monospace, 'SF Mono', Menlo, monospace;
    text-transform: uppercase;
  }

  .featured-tag {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 8px;
    background: linear-gradient(135deg, #f97316, #ef4444);
    color: #fff;
    font-size: 10px;
    font-weight: 500;
    border-radius: 2px;
    margin-left: auto;
    letter-spacing: 0.05em;

    svg {
      width: 10px;
      height: 10px;
    }
  }
}

.portfolio-name {
  font-size: 20px;
  font-weight: 700;
  color: $text-primary;
  margin: 0;
  letter-spacing: 0.01em;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.portfolio-desc {
  font-size: 13px;
  line-height: 1.6;
  color: $text-secondary;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 42px;
}

/* 描边标签（HOK 风格） */
.portfolio-topics {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;

  li {
    font-size: 11px;
    color: $cyan;
    padding: 3px 9px;
    background: transparent;
    border: 0.5px solid rgba(0, 212, 255, 0.35);
    border-radius: 1px;
    font-family: ui-monospace, 'SF Mono', Menlo, monospace;
    letter-spacing: 0.02em;
  }
}

.portfolio-divider {
  height: 0.5px;
  background: $border-dim;
  margin: 4px 0;
}

.portfolio-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: $text-secondary;
  font-family: ui-monospace, 'SF Mono', Menlo, monospace;

  .lang-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .meta-text {
    color: $text-secondary;
  }

  .meta-divider {
    width: 0.5px;
    height: 8px;
    background: $border-mid;
  }

  .meta-icon {
    width: 11px;
    height: 11px;
    color: $text-tertiary;
  }
}

/* 按钮区 */
.portfolio-actions {
  display: flex;
  gap: 8px;
  margin-top: auto;
  padding-top: 10px;

  a {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 9px 14px;
    border-radius: 1px;
    font-size: 12px;
    font-weight: 600;
    text-decoration: none;
    transition: all 0.25s ease;
    cursor: pointer;
    letter-spacing: 0.05em;
    text-transform: uppercase;

    svg {
      width: 13px;
      height: 13px;
    }
  }

  /* 主按钮：金色实心 */
  .btn-primary {
    flex: 1;
    background: linear-gradient(135deg, $gold 0%, darken($gold, 12%) 100%);
    color: $bg-deep;
    border: 0.5px solid $gold;

    &:hover {
      background: linear-gradient(135deg, $gold-bright 0%, $gold 100%);
      box-shadow: 0 0 16px rgba(201, 169, 97, 0.4);
      transform: translateY(-1px);
    }
  }

  /* 次按钮：描边幽灵 */
  .btn-ghost {
    flex: 1;
    background: transparent;
    color: $text-secondary;
    border: 0.5px solid $border-mid;

    &:hover {
      border-color: $cyan;
      color: $cyan;
      background: rgba(0, 212, 255, 0.05);
    }
  }
}

/* ============ 浅色主题适配 ============ */
:global(html:not(.dark)) .portfolio-card {
  background: #ffffff;
  border-color: rgba(201, 169, 97, 0.3);

  .portfolio-cover {
    background: #f5f7fa;

    .cover-overlay {
      background: linear-gradient(
        to bottom,
        rgba(255, 255, 255, 0) 0%,
        rgba(255, 255, 255, 0.2) 60%,
        rgba(255, 255, 255, 0.9) 100%
      );
    }

    .star-badge {
      background: rgba(255, 255, 255, 0.9);
      border-color: rgba(201, 169, 97, 0.4);
      color: #a88a4a;
    }
  }

  .portfolio-name {
    color: #1a1f36;
  }

  .portfolio-desc {
    color: #5a6488;
  }

  .portfolio-divider {
    background: #e5e8f0;
  }

  .portfolio-meta {
    color: #8a93b0;

    .meta-divider {
      background: #d5dae5;
    }
    .meta-icon {
      color: #aab2c8;
    }
  }

  .portfolio-actions {
    .btn-primary {
      color: #ffffff;
    }
    .btn-ghost {
      color: #5a6488;
      border-color: #d5dae5;

      &:hover {
        border-color: #0099cc;
        color: #0099cc;
        background: rgba(0, 153, 204, 0.04);
      }
    }
  }

  &:hover {
    border-color: #0099cc;
    box-shadow: 0 8px 32px rgba(0, 153, 204, 0.12);
  }
}
</style>

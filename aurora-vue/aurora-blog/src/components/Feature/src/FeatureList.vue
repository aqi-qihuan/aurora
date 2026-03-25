<template>
  <div class="inverted-main-grid py-8 gap-8 box-border">
    <div class="relative overflow-hidden h-56 lg:h-auto rounded-2xl bg-ob-deep-800 shadow-lg feature-list-hero">
      <div
        class="ob-gradient-plate opacity-90 relative z-10 bg-ob-deep-900 rounded-2xl flex justify-start items-end px-8 pb-10 shadow-md">
        <h2 class="text-3xl pb-8 lg:pb-16">
          <p :style="gradientText" class="feature-list-title">EDITOR'S SELECTION</p>
          <span class="relative text-2xl text-ob-bright font-semibold">
            <svg-icon class="inline-block" icon-class="hot" />
            {{ t('home.recommended') }}
          </span>
        </h2>
      </div>
      <span class="absolute top-0 w-full h-full z-0" :style="gradientBackground" />
    </div>

    <ul class="grid lg:grid-cols-2 gap-8">
      <template v-if="featuredArticles.length > 0">
        <li v-for="(article, index) in featuredArticles" :key="article.id">
          <ArticleCard class="home-featured-article" :data="article" :style="{ animationDelay: index * 0.08 + 's' }" />
        </li>
      </template>
      <template v-else>
        <li v-for="n in 2" :key="n">
          <ArticleCard :data="{}" />
        </li>
      </template>
    </ul>
  </div>
</template>

<script lang="ts">
// @ts-nocheck
import { useAppStore } from '@/stores/app'
import { useArticleStore } from '@/stores/article'
import { useI18n } from 'vue-i18n'
import { computed, defineComponent, toRef } from 'vue'
import { ArticleCard } from '@/components/ArticleCard'

export default defineComponent({
  name: 'FeatureList',
  components: {
    ArticleCard
  },
  setup() {
    const appStore = useAppStore()
    const articleStore = useArticleStore()
    const { t } = useI18n()
    return {
      gradientBackground: computed(() => {
        return { background: appStore.themeConfig.header_gradient_css }
      }),
      gradientText: appStore.themeConfig.background_gradient_style,
      featuredArticles: toRef(articleStore.$state, 'featuredArticles'),
      t
    }
  }
})
</script>

<style lang="scss">
.home-featured-article {
  .article-content {
    p {
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 4;
      -webkit-box-orient: vertical;
    }
    .article-footer {
      margin-top: 13px;
    }
  }
}
.feature-list-hero {
  transition: transform 0.3s ease, box-shadow 0.3s ease;
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
  }
}
.feature-list-title {
  letter-spacing: 0.08em;
}
</style>

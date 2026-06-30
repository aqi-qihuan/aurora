<template>
  <div class="portfolio-list-wrapper">
    <transition-group
      v-if="list && list.length > 0"
      name="stagger-fade"
      tag="ul"
      class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <li
        v-for="(item, index) in list"
        :key="item.id"
        :style="{ animationDelay: Number(index) * 0.08 + 's' }">
        <PortfolioCard :data="item" />
      </li>
    </transition-group>
    <ul v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <li v-for="n in 4" :key="n">
        <PortfolioCard :data="{}" />
      </li>
    </ul>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'
import PortfolioCard from './PortfolioCard.vue'
import type { PortfolioItem } from './types'

export default defineComponent({
  name: 'PortfolioList',
  components: { PortfolioCard },
  props: {
    list: {
      type: Array as PropType<PortfolioItem[]>,
      default: () => []
    }
  }
})
</script>

<style lang="scss" scoped>
.stagger-fade-enter-active {
  transition: all 0.35s cubic-bezier(0.22, 1, 0.36, 1);
  animation: cardFadeIn 0.4s ease both;
}
.stagger-fade-enter-from {
  opacity: 0;
  transform: translateY(20px);
}
@keyframes cardFadeIn {
  from {
    opacity: 0;
    transform: translateY(16px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>

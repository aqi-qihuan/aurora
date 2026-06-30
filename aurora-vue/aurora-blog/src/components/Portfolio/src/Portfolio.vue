<template>
  <div id="portfolio" class="portfolio-container">
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'Portfolio'
})
</script>

<style lang="scss" scoped>
.portfolio-container {
  margin-bottom: 2rem;
}
</style>

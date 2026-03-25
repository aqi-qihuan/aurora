<template>
  <div class="tag-list">
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
export default defineComponent({
  name: 'ObTagList'
})
</script>

<style lang="scss" scoped>
.tag-list {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 0.25rem 0;
  padding-top: 0.75rem;
}
</style>

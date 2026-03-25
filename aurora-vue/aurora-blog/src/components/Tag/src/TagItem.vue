<template>
  <div class="tag-chip">
    <router-link
      class="tag-name"
      :to="{ path: '/article-list/' + id, query: { tagName: name } }"
      :style="stylingTag()">
      <em class="tag-hash">#</em>
      {{ name }}
    </router-link>
    <span
      class="tag-count"
      :style="stylingTag()">
      {{ count }}
    </span>
  </div>
</template>

<script lang="ts">
import { defineComponent, toRefs } from 'vue'

export default defineComponent({
  name: 'ObTagItem',
  props: ['id', 'name', 'count', 'size'],
  setup(props) {
    const tagSize = toRefs(props).size
    const stylingTag = () => {
      if (tagSize.value === 'xs') {
        return { fontSize: '0.75rem', lineHeight: '1rem' }
      }
      if (tagSize.value === 'sm') {
        return { fontSize: '0.875rem', lineHeight: '1.25rem' }
      }
      if (tagSize.value === 'lg') {
        return { fontSize: '1.125rem', lineHeight: '1.75rem' }
      }
      if (tagSize.value === 'xl') {
        return { fontSize: '1.25rem', lineHeight: '1.75rem' }
      }
      if (tagSize.value === '2xl') {
        return { fontSize: '1.5rem', lineHeight: '2rem' }
      }
      return { fontSize: '1rem', lineHeight: '1.5rem' }
    }

    return { stylingTag }
  }
})
</script>

<style lang="scss" scoped>
.tag-chip {
  display: inline-flex;
  align-items: center;
  margin-right: 0.5rem;
  margin-bottom: 0.5rem;
  cursor: pointer;
  border-radius: 0.5rem;
  overflow: hidden;
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.25s ease,
              opacity 0.2s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.18);
  }

  &:active {
    transform: translateY(0) scale(0.97);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  }
}

.tag-name {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.75rem;
  background: var(--background-secondary);
  border-radius: 0.5rem 0 0 0.5rem;
  color: var(--text-normal);
  text-decoration: none;
  transition: color 0.2s ease, background 0.2s ease;
}

.tag-chip:hover .tag-name {
  color: var(--text-accent);
  background: var(--background-secondary);
}

.tag-hash {
  opacity: 0.4;
  margin-right: 0.15em;
  font-style: normal;
  transition: opacity 0.2s ease;
}

.tag-chip:hover .tag-hash {
  opacity: 0.7;
}

.tag-count {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem 0.5rem;
  background: var(--background-secondary);
  border-radius: 0 0.5rem 0.5rem 0;
  color: var(--text-dim);
  font-variant-numeric: tabular-nums;
  transition: color 0.2s ease, background 0.2s ease;
}

.tag-chip:hover .tag-count {
  color: var(--text-secondary);
}
</style>

<template>
  <div class="paginator">
    <ul>
      <li class="text-ob-bright" v-if="currentPage > 1" @click="pageChangeEmitter(currentPage - 1)">
        <svg-icon icon-class="arrow-left" />
        {{ t('settings.paginator.newer') }}
      </li>
      <li
        v-if="paginator.head !== 0"
        :class="{ active: currentPage === paginator.head }"
        @click="pageChangeEmitter(paginator.head)">
        {{ paginator.head }}
      </li>
      <li
        v-for="(page, key) in paginator.pages"
        :key="key"
        :class="{ active: currentPage === page }"
        @click="pageChangeEmitter(page)">
        {{ page }}
      </li>
      <li
        v-if="paginator.end !== 0"
        :class="{ active: currentPage === paginator.end }"
        @click="pageChangeEmitter(paginator.end)">
        {{ paginator.end }}
      </li>
      <li class="text-ob-bright" v-if="currentPage < pages" @click="pageChangeEmitter(currentPage + 1)">
        {{ t('settings.paginator.older') }}
        <svg-icon icon-class="arrow-right" />
      </li>
    </ul>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, toRefs } from 'vue'
import { useI18n } from 'vue-i18n'

export default defineComponent({
  name: 'Paginator',
  emits: ['pageChange'],
  props: ['pageTotal', 'pageSize', 'page'],
  setup(props, { emit }) {
    const { t } = useI18n()
    const pagination = toRefs(props)
    const pages = computed(() => {
      return Math.ceil(pagination.pageTotal.value / pagination.pageSize.value)
    })
    const paginator = computed(() => {
      if (pages.value <= 3) {
        const pageNumbers = []
        for (let i = 0; i < pages.value; i++) {
          pageNumbers.push(i + 1)
        }
        return {
          head: 0,
          pages: pageNumbers,
          end: 0
        }
      } else if (pagination.page.value >= 1 && pagination.page.value < 3) {
        return {
          head: 1,
          pages: [2, 3, '...'],
          end: pages.value
        }
      } else if (pagination.page.value >= 3 && pagination.page.value <= pages.value - 2) {
        return {
          head: 1,
          pages: ['...', pagination.page.value - 1, pagination.page.value, pagination.page.value + 1, '...'],
          end: pages.value
        }
      } else {
        return {
          head: 1,
          pages: ['...', pages.value - 2, pages.value - 1],
          end: pages.value
        }
      }
    })
    const pageChangeEmitter = (page: number | string) => {
      if (page === '...') return
      emit('pageChange', page)
    }
    return {
      currentPage: computed(() => {
        return pagination.page.value
      }),
      pageChangeEmitter,
      paginator,
      pages,
      t
    }
  }
})
</script>

<style lang="scss" scoped>
.paginator {
  @apply flex flex-row justify-center my-10;
  ul {
    @apply flex flex-row items-center gap-1;
    li {
      @apply flex flex-row items-center uppercase font-bold cursor-pointer px-3 py-1.5 rounded-lg text-sm;
      transition: all 0.25s cubic-bezier(0.22, 1, 0.36, 1);
      background: transparent;
      color: var(--text-dim);
      &:hover {
        background: var(--bg-accent-05);
        color: var(--text-normal);
        transform: translateY(-1px);
      }
      svg {
        @apply mx-1 font-bold;
        transition: transform 0.25s ease;
      }
      &:hover svg {
        transform: translateX(1px);
      }
    }
  }
  .active {
    background: var(--main-gradient);
    color: #fff;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    &:hover {
      background: var(--main-gradient);
      color: #fff;
      opacity: 0.9;
    }
  }
}
</style>

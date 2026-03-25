<template>
  <div class="flex flex-col">
    <div class="post-header">
      <Breadcrumb :current="t('menu.tags')" />
      <h1 class="post-title text-white uppercase">{{ t('menu.tags') }}</h1>
    </div>
    <div class="bg-ob-deep-800 px-14 py-16 rounded-2xl shadow-xl block tags-container">
      <TagList>
        <transition-group name="tag-pop" v-if="tags != '' && tags.length > 0">
          <TagItem
            v-for="(tag, index) in tags"
            :key="tag.id"
            :id="tag.id"
            :name="tag.tagName"
            :count="tag.count"
            size="xl"
            :style="{ animationDelay: index * 0.04 + 's' }"
          />
        </transition-group>
        <div v-else class="tags-empty">
          <svg-icon icon-class="tag" class="empty-icon" />
          <p class="empty-text">暂无标签</p>
        </div>
      </TagList>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, onUnmounted, toRef } from 'vue'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { useI18n } from 'vue-i18n'
import { useTagStore } from '@/stores/tag'
import { TagList, TagItem } from '@/components/Tag'
import { useCommonStore } from '@/stores/common'
import api from '@/api/api'

export default defineComponent({
  name: 'Tag',
  components: { Breadcrumb, TagList, TagItem },
  setup() {
    const commonStore = useCommonStore()
    const { t } = useI18n()
    const tagStore = useTagStore()
    onMounted(() => {
      fetchTags()
    })
    onUnmounted(() => {
      commonStore.resetHeaderImage()
    })
    const fetchTags = () => {
      api.getAllTags().then(({ data }) => {
        tagStore.tags = data.data
      })
    }
    return {
      tags: toRef(tagStore.$state, 'tags'),
      t
    }
  }
})
</script>

<style lang="scss" scoped>
.tags-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 0;
  color: var(--text-dim);
}

.empty-icon {
  font-size: 2.5rem;
  opacity: 0.3;
  margin-bottom: 1rem;
}

.empty-text {
  font-size: 0.9rem;
  opacity: 0.5;
}

.tag-pop-enter-active {
  transition: all 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.tag-pop-enter-from {
  opacity: 0;
  transform: scale(0.8) translateY(8px);
}

.tag-pop-leave-active {
  transition: all 0.2s ease;
  position: absolute;
}

.tag-pop-leave-to {
  opacity: 0;
  transform: scale(0.85);
}

.tag-pop-move {
  transition: transform 0.3s ease;
}
</style>

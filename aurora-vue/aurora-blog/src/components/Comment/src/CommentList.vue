<template>
  <div class="comment-list">
    <template v-if="comments && comments.length > 0">
      <transition-group name="comment-fade" tag="div" class="comment-list-inner">
        <CommentItem v-for="(comment, index) in comments" :key="comment.id" :comment="comment" :index="index" />
      </transition-group>
    </template>
    <div v-else class="empty-state">
      <svg-icon icon-class="message" class="empty-icon" />
      <p class="empty-text">还没有评论，来说点什么吧</p>
    </div>
    <button
      class="load-more-btn"
      v-if="haveMore"
      @click="loadMore">
      加载更多
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent, inject } from 'vue'
import CommentItem from './CommentItem.vue'
import { useCommentStore } from '@/stores/comment'
import emitter from '@/utils/mitt'

export default defineComponent({
  components: {
    CommentItem
  },
  setup() {
    const commentStore = useCommentStore()
    const loadMore = () => {
      switch (commentStore.type) {
        case 1:
          emitter.emit('articleLoadMore')
          break
        case 2:
          emitter.emit('messageLoadMore')
          break
        case 3:
          emitter.emit('aboutLoadMore')
          break
        case 4:
          emitter.emit('friendLinkLoadMore')
          break
        case 5:
          emitter.emit('talkLoadMore')
      }
    }
    return {
      comments: inject('comments'),
      haveMore: inject('haveMore'),
      loadMore
    }
  }
})
</script>
<style lang="scss" scoped>
.comment-list {
  margin-top: 1.5rem;
}

.comment-list-inner {
  display: flex;
  flex-direction: column;
}

.empty-state {
  text-align: center;
  padding: 2.5rem 1rem;
}

.empty-icon {
  width: 36px;
  height: 36px;
  color: var(--text-dim);
  opacity: 0.35;
  margin-bottom: 0.75rem;
}

.empty-text {
  font-size: 0.85rem;
  color: var(--text-dim);
  margin: 0;
}

.load-more-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 1.75rem auto 0;
  min-width: 6rem;
  padding: 0.55rem 1.25rem;
  border-radius: 0.5rem;
  background: var(--main-gradient);
  color: #fff;
  font-weight: 500;
  font-size: 0.85rem;
  border: none;
  outline: none;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.25s ease;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  }

  &:active {
    transform: translateY(0) scale(0.97);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  }
}
</style>

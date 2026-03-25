<template>
  <div class="comment-item">
    <div class="comment-item-row">
      <Avatar :url="comment.avatar" />
      <div class="comment-item-body">
        <div class="comment-bubble">
          <p class="comment-content" v-html="comment.commentContent.replaceAll('\n', '<br>')" />
          <div class="comment-meta">
            <span class="comment-author">{{ comment.nickname }} · {{ time }}</span>
            <button @click="clickOnReply" class="reply-btn">
              回复
            </button>
          </div>
        </div>
        <transition name="reply-slide">
          <CommentReplyForm
            v-show="show"
            :replyUserId="comment.userId"
            :initialContent="replyContent"
            @changeShow="changeShow" />
        </transition>
        <transition-group name="comment-fade">
          <CommentReplyItem
            v-for="reply in comment.replyDTOs"
            :key="reply.id"
            :reply="reply"
            :commentUserId="comment.userId" />
        </transition-group>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive, ref, toRefs, provide, computed, onMounted } from 'vue'
import Avatar from '@/components/Avatar.vue'
import CommentReplyItem from './CommentReplyItem.vue'
import CommentReplyForm from './CommentReplyForm.vue'

export default defineComponent({
  components: {
    Avatar,
    CommentReplyItem,
    CommentReplyForm
  },
  props: ['comment', 'index'],
  setup(props) {
    const formatTime = (time: any): string => {
      const date = new Date(time)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const minutes = Math.floor(diff / 60000)
      const hours = Math.floor(diff / 3600000)
      const days = Math.floor(diff / 86400000)
      if (minutes < 1) return '刚刚'
      if (minutes < 60) return minutes + ' 分钟前'
      if (hours < 24) return hours + ' 小时前'
      if (days < 30) return days + ' 天前'
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return year + '-' + month + '-' + day
    }
    const reactiveData = reactive({
      replyContent: '' as any,
      show: false as any
    })
    const time = computed(() => formatTime(props.comment.createTime))
    provide('parentId', computed(() => props.comment.id))
    provide('index', computed(() => props.index))
    const changeShow = () => {
      reactiveData.show = false
    }
    const clickOnReply = () => {
      reactiveData.replyContent = ''
      reactiveData.show = true
    }
    return {
      ...toRefs(reactiveData),
      time,
      clickOnReply,
      changeShow
    }
  }
})
</script>
<style lang="scss" scoped>
.comment-item {
  margin-top: 1.25rem;
  max-width: 100%;

  &:first-child {
    margin-top: 0;
  }
}

.comment-item-row {
  display: flex;
  gap: 0.75rem;
}

@media (min-width: 1280px) {
  .comment-item-row {
    gap: 1.25rem;
  }
}

.comment-item-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  min-width: 0;
}

.comment-bubble {
  position: relative;
  display: inline-block;
  max-width: 100%;
  padding: 1rem;
  border-radius: 0.625rem;
  background: var(--background-primary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  border: 1px solid transparent;
  transition: box-shadow 0.2s ease, border-color 0.2s ease;

  &::before {
    content: '';
    position: absolute;
    width: 0;
    height: 0;
    border-right: 8px solid var(--background-primary);
    border-top: 6px solid transparent;
    border-bottom: 6px solid transparent;
    left: -8px;
    top: 14px;
  }

  &:hover {
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
    border-color: rgba(var(--text-dim-rgb, 148, 163, 184), 0.15);
  }
}

.comment-content {
  line-height: 1.7;
  white-space: pre-line;
  word-wrap: break-word;
  word-break: break-all;
  color: var(--text-normal);
  font-size: 0.9rem;
}

.comment-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.75rem;
  font-size: 0.75rem;
  color: var(--text-dim);
  gap: 0.75rem;
}

.comment-author {
  white-space: nowrap;
}

.reply-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  cursor: pointer;
  color: var(--text-accent);
  font-weight: 500;
  font-size: 0.75rem;
  border: none;
  background: none;
  padding: 0.2rem 0.4rem;
  border-radius: 0.25rem;
  transition: opacity 0.2s ease, background 0.2s ease;
  white-space: nowrap;

  &:hover {
    opacity: 0.8;
    background: rgba(var(--text-accent-rgb, 100, 149, 237), 0.08);
  }

  &:active {
    transform: scale(0.96);
  }
}

</style>

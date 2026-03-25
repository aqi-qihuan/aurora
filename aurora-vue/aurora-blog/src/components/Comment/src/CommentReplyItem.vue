<template>
  <div class="reply-item">
    <div class="reply-item-row">
      <Avatar :url="reply.avatar" />
      <div class="reply-bubble">
        <p class="reply-content" v-html="commentContent.replaceAll('\n', '<br>')" />
        <div class="reply-meta">
          <span class="reply-author">{{ reply.nickname }} · {{ time }}</span>
          <button @click="clickOnSonReply" class="reply-btn">
            回复
          </button>
        </div>
      </div>
    </div>
    <transition name="reply-slide">
      <CommentReplyForm
        class="reply-form"
        v-show="show"
        :replyUserId="reply.userId"
        :initialContent="replyContent"
        @changeShow="changeShow" />
    </transition>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, reactive, toRefs } from 'vue'
import Avatar from '@/components/Avatar.vue'
import CommentReplyForm from './CommentReplyForm.vue'

export default defineComponent({
  components: {
    Avatar,
    CommentReplyForm
  },
  props: ['reply', 'commentUserId'],
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
    const time = computed(() => formatTime(props.reply.createTime))
    const clickOnSonReply = () => {
      reactiveData.replyContent = '@' + props.reply.nickname
      reactiveData.show = true
    }
    const changeShow = () => {
      reactiveData.show = false
    }
    const commentContent = computed(() => {
      if (props.reply.replyUserId !== props.commentUserId) {
        const nickname = props.reply.replyNickname || ''
        const website = props.reply.replyWebsite || ''
        if (website) {
          return `<a href="${website}" target="_blank" rel="noopener noreferrer" class="reply-link">@${nickname}</a> ` +
            props.reply.commentContent
        }
        return `<span class="reply-link">@${nickname}</span> ` + props.reply.commentContent
      }
      return props.reply.commentContent
    })
    return {
      ...toRefs(reactiveData),
      commentContent,
      time,
      clickOnSonReply,
      changeShow
    }
  }
})
</script>
<style lang="scss" scoped>
.reply-item {
  margin-top: 0.75rem;
}

.reply-item-row {
  display: flex;
  gap: 0.75rem;
}

@media (min-width: 1280px) {
  .reply-item-row {
    gap: 1.25rem;
  }
}

.reply-bubble {
  position: relative;
  display: inline-block;
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
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

.reply-content {
  line-height: 1.7;
  white-space: pre-line;
  word-wrap: break-word;
  word-break: break-all;
  color: var(--text-normal);
  font-size: 0.85rem;
}

.reply-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-dim);
  gap: 0.75rem;
}

.reply-author {
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

.reply-icon {
  width: 13px;
  height: 13px;
}

.reply-form {
  margin-top: 0.75rem;
}

.reply-link {
  color: var(--text-accent);
  font-weight: 500;
  text-decoration: none;
  transition: opacity 0.2s ease;

  &:hover {
    opacity: 0.8;
  }
}
</style>

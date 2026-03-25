<template>
  <div class="reply-form-wrap">
    <Avatar :url="avatar" />
    <div class="reply-form-body">
      <textarea
        v-model="commentContent"
        class="comment-textarea"
        :placeholder="initialContent || '回复...'"
        cols="30"
        rows="3" />
      <div class="form-actions">
        <button
          @click="CancelReply"
          class="cancel-btn">
          取消
        </button>
        <button
          @click="saveReply"
          class="submit-btn"
          :disabled="!commentContent.trim()">
          回复
        </button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, toRefs, reactive, getCurrentInstance, inject, computed } from 'vue'
import Avatar from '@/components/Avatar.vue'
import emitter from '@/utils/mitt'
import { useUserStore } from '@/stores/user'
import { useCommentStore } from '@/stores/comment'
import { useAppStore } from '@/stores/app'
import { useRoute } from 'vue-router'
import api from '@/api/api'

export default defineComponent({
  components: {
    Avatar
  },
  props: ['replyUserId', 'initialContent'],
  setup(props, { emit }) {
    const proxy: any = getCurrentInstance()?.appContext.config.globalProperties
    const userStore = useUserStore()
    const commentStore = useCommentStore()
    const appStore = useAppStore()
    const route = useRoute()
    const reactiveData = reactive({
      commentContent: '' as any
    })
    const parentId = inject('parentId')
    const index = inject('index')
    const saveReply = () => {
      if (userStore.userInfo === '') {
        proxy.$notify({
          title: 'Warning',
          message: '请登录后回复',
          type: 'warning'
        })
        return
      }
      if (reactiveData.commentContent.trim() === '') {
        proxy.$notify({
          title: 'Warning',
          message: '回复不能为空',
          type: 'warning'
        })
        return
      }
      const path = route.path
      const arr = path.split('/')
      const params: any = {
        type: commentStore.type,
        replyUserId: props.replyUserId,
        parentId: parentId,
        commentContent: reactiveData.commentContent
      }
      params.topicId = arr[2]
      api.saveComment(params).then(({ data }) => {
        if (data.flag) {
          emit('changeShow')
          fetchReplies()
          let isCommentReview = appStore.websiteConfig.isCommentReview
          if (isCommentReview) {
            proxy.$notify({
              title: 'Warning',
              message: '回复成功,正在审核中',
              type: 'warning'
            })
          } else {
            proxy.$notify({
              title: 'Success',
              message: '回复成功',
              type: 'success'
            })
          }
          reactiveData.commentContent = ''
        }
      }).catch(() => {
        proxy.$notify({ title: 'Error', message: '回复失败，请重试', type: 'error' })
      })
    }
    const fetchReplies = () => {
      switch (commentStore.type) {
        case 1:
          emitter.emit('articleFetchReplies', index)
          break
        case 2:
          emitter.emit('messageFetchReplies', index)
          break
        case 3:
          emitter.emit('aboutFetchReplies', index)
          break
        case 4:
          emitter.emit('friendLinkFetchReplies', index)
          break
        case 5:
          emitter.emit('talkFetchReplies', index)
      }
    }
    const CancelReply = () => {
      emit('changeShow')
    }
    return {
      ...toRefs(reactiveData),
      avatar: computed(() => userStore.userInfo.avatar),
      saveReply,
      CancelReply
    }
  }
})
</script>
<style lang="scss" scoped>
.reply-form-wrap {
  display: flex;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

@media (min-width: 1280px) {
  .reply-form-wrap {
    gap: 1.25rem;
  }
}

.reply-form-body {
  display: flex;
  flex-direction: column;
  flex-wrap: wrap-reverse;
  width: 100%;
}

.comment-textarea {
  width: 100%;
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  background: var(--background-primary);
  color: var(--text-normal);
  border: 1px solid transparent;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  resize: none;
  outline: none;
  font-size: 0.85rem;
  line-height: 1.6;
  transition: border-color 0.25s ease, box-shadow 0.25s ease;

  &::placeholder {
    color: var(--text-dim);
    opacity: 0.6;
  }

  &:focus {
    border-color: var(--text-accent);
    box-shadow: 0 0 0 2px rgba(var(--text-accent-rgb, 100, 149, 237), 0.12);
  }
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: 0.6rem;
  gap: 0.5rem;
}

.cancel-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 3.5rem;
  padding: 0.4rem 0.75rem;
  border-radius: 0.375rem;
  background: var(--background-secondary);
  color: var(--text-dim);
  font-weight: 500;
  font-size: 0.8rem;
  border: none;
  outline: none;
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease, transform 0.2s ease;

  &:hover {
    background: var(--background-tertiary, rgba(255, 255, 255, 0.1));
    color: var(--text-secondary);
  }

  &:active {
    transform: scale(0.97);
  }
}

.submit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 3.5rem;
  padding: 0.4rem 0.75rem;
  border-radius: 0.375rem;
  background: var(--main-gradient);
  color: #fff;
  font-weight: 500;
  font-size: 0.8rem;
  border: none;
  outline: none;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.25s ease,
              opacity 0.25s ease;

  &:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  }

  &:active:not(:disabled) {
    transform: translateY(0) scale(0.97);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}
</style>

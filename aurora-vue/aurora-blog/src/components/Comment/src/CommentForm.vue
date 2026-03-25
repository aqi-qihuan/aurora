<template>
  <div class="comment-form">
    <Avatar :url="avatar" />
    <div class="comment-form-body">
      <div class="textarea-wrap">
        <textarea
          v-model="commentContent"
          class="comment-textarea"
          placeholder="写下你的评论..."
          cols="30"
          rows="4"
          @input="onInput"
          @focus="isFocused = true"
          @blur="isFocused = false" />
        <span class="char-count" :class="{ 'is-limit': commentContent.length >= 500 }">
          {{ commentContent.length }} / 500
        </span>
      </div>
      <div class="form-actions">
        <span class="form-hint" v-if="!commentContent.length && !isFocused">Ctrl + Enter 快速发送</span>
        <button
          @click="saveComment"
          class="submit-btn"
          :disabled="!commentContent.trim()">
          <span class="btn-text">发送评论</span>
        </button>
      </div>
      <div class="form-divider"></div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, toRefs, reactive, getCurrentInstance, computed, ref } from 'vue'
import Avatar from '@/components/Avatar.vue'
import { useUserStore } from '@/stores/user'
import { useRoute } from 'vue-router'
import { useCommentStore } from '@/stores/comment'
import { useAppStore } from '@/stores/app'
import api from '@/api/api'
import emitter from '@/utils/mitt'

export default defineComponent({
  name: 'CommentForm',
  components: { Avatar },
  setup() {
    const proxy: any = getCurrentInstance()?.appContext.config.globalProperties
    const userStore = useUserStore()
    const commentStore = useCommentStore()
    const appStore = useAppStore()
    const route = useRoute()
    const isFocused = ref(false)
    const reactiveData = reactive({
      commentContent: '' as any
    })
    const onInput = () => {
      if (reactiveData.commentContent.length > 500) {
        reactiveData.commentContent = reactiveData.commentContent.slice(0, 500)
      }
    }
    const saveComment = () => {
      if (userStore.userInfo === '') {
        proxy.$notify({
          title: 'Warning',
          message: '请登录后评论',
          type: 'warning'
        })
        return
      }
      if (reactiveData.commentContent.trim() === '') {
        proxy.$notify({
          title: 'Warning',
          message: '评论不能为空',
          type: 'warning'
        })
        return
      }
      const path = route.path
      const arr = path.split('/')
      const params: any = {
        commentContent: reactiveData.commentContent,
        type: commentStore.type
      }
      params.topicId = arr[2]
      api.saveComment(params).then(({ data }) => {
        if (data.flag) {
          fetchComments()
          let isCommentReview = appStore.websiteConfig.isCommentReview
          if (isCommentReview) {
            proxy.$notify({
              title: 'Warning',
              message: '评论成功,正在审核中',
              type: 'warning'
            })
          } else {
            proxy.$notify({
              title: 'Success',
              message: '评论成功',
              type: 'success'
            })
          }
          reactiveData.commentContent = ''
        }
      }).catch(() => {
        proxy.$notify({ title: 'Error', message: '评论失败，请重试', type: 'error' })
      })
    }
    const fetchComments = () => {
      switch (commentStore.type) {
        case 1:
          emitter.emit('articleFetchComment')
          break
        case 2:
          emitter.emit('messageFetchComment')
          break
        case 3:
          emitter.emit('aboutFetchComment')
          break
        case 4:
          emitter.emit('friendLinkFetchComment')
          break
        case 5:
          emitter.emit('talkFetchComment')
      }
    }
    return {
      ...toRefs(reactiveData),
      avatar: computed(() => userStore.userInfo.avatar),
      isFocused,
      onInput,
      saveComment
    }
  }
})
</script>

<style lang="scss" scoped>
.comment-form {
  display: flex;
  gap: 0.75rem;
}

@media (min-width: 1280px) {
  .comment-form {
    gap: 1.25rem;
  }
}

.comment-form-body {
  display: flex;
  flex-direction: column;
  flex-wrap: wrap-reverse;
  width: 100%;
}

.textarea-wrap {
  position: relative;
}

.comment-textarea {
  width: 100%;
  padding: 1rem;
  border-radius: 0.625rem;
  background: var(--background-primary);
  color: var(--text-normal);
  border: 1px solid transparent;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  resize: none;
  outline: none;
  font-size: 0.9rem;
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

.char-count {
  position: absolute;
  right: 0.75rem;
  bottom: 0.5rem;
  font-size: 0.7rem;
  color: var(--text-dim);
  opacity: 0;
  transition: opacity 0.2s ease;

  .comment-textarea:focus ~ & {
    opacity: 1;
  }

  &.is-limit {
    color: #f56c6c;
    opacity: 1;
  }
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  margin-top: 0.75rem;
  gap: 0.75rem;
}

.form-hint {
  margin-right: auto;
  font-size: 0.75rem;
  color: var(--text-dim);
  opacity: 0;
  transition: opacity 0.25s ease;

  .comment-textarea:focus ~ .form-actions & {
    opacity: 0;
  }
}

.submit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
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

.btn-text {
  text-align: center;
  width: 100%;
}

.form-divider {
  width: 100%;
  margin-top: 1.5rem;
  border-bottom: 2px solid var(--text-normal);
  opacity: 0.15;
}
</style>

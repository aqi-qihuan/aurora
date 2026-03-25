<template>
  <div>
    <Breadcrumb :current="t('menu.message')" />
    <div class="flex flex-col">
      <div class="post-header">
        <h1 class="post-title text-white uppercase">{{ t('titles.message') }}</h1>
      </div>
      <div class="main-grid">
        <div class="relative">
          <transition name="msg-intro" appear>
            <div class="message-intro">
              <div class="intro-icon-wrap">
                <svg-icon icon-class="message" class="intro-icon" />
                <span class="intro-icon-ring"></span>
                <span class="intro-icon-ring ring-delay"></span>
              </div>
              <h2 class="message-welcome">这是一个留言板</h2>
              <p class="message-hint">欢迎大家前来留言</p>
              <div class="intro-divider">
                <span class="divider-line"></span>
                <svg-icon icon-class="dots" class="divider-dot" />
                <span class="divider-line"></span>
              </div>
            </div>
          </transition>
          <Comment />
        </div>
        <div class="col-span-1">
          <Sidebar>
            <Profile />
          </Sidebar>
        </div>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, reactive, toRefs, computed, provide } from 'vue'
import { useI18n } from 'vue-i18n'
import { Sidebar, Profile } from '../components/Sidebar'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { Comment } from '../components/Comment'
import { useCommentStore } from '@/stores/comment'
import api from '@/api/api'
import emitter from '@/utils/mitt'

export default defineComponent({
  name: 'Message',
  components: { Breadcrumb, Comment, Sidebar, Profile },
  setup() {
    const { t } = useI18n()
    const commentStore = useCommentStore()
    const reactiveData = reactive({
      comments: [] as any,
      haveMore: false as any,
      isReload: false as any
    })
    const pageInfo = reactive({
      current: 1,
      size: 7
    })
    commentStore.type = 2
    onMounted(() => {
      fetchComments()
    })
    provide(
      'comments',
      computed(() => reactiveData.comments)
    )

    provide(
      'haveMore',
      computed(() => reactiveData.haveMore)
    )
    emitter.on('messageFetchComment', () => {
      pageInfo.current = 1
      reactiveData.isReload = true
      fetchComments()
    })
    emitter.on('messageFetchReplies', (index) => {
      fetchReplies(index)
    })
    emitter.on('messageLoadMore', () => {
      fetchComments()
    })
    const fetchComments = () => {
      const params = {
        type: 2,
        topicId: null,
        current: pageInfo.current,
        size: pageInfo.size
      }
      api.getComments(params).then(({ data }) => {
        if (reactiveData.isReload) {
          reactiveData.comments = data.data.records
          reactiveData.isReload = false
        } else {
          reactiveData.comments.push(...data.data.records)
        }
        if (data.data.count <= reactiveData.comments.length) {
          reactiveData.haveMore = false
        } else {
          reactiveData.haveMore = true
        }
        pageInfo.current++
      })
    }
    const fetchReplies = (index: any) => {
      api.getRepliesByCommentId(reactiveData.comments[index].id).then(({ data }) => {
        reactiveData.comments[index].replyDTOs = data.data
      })
    }
    return {
      ...toRefs(reactiveData),
      t
    }
  }
})
</script>

<style lang="scss" scoped>
.message-intro {
  padding: 3rem 0 2rem;
  text-align: center;
  position: relative;
}

.intro-icon-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  margin-bottom: 1.25rem;
}

.intro-icon {
  width: 40px;
  height: 40px;
  color: var(--text-accent);
  position: relative;
  z-index: 2;
  animation: icon-bob 3s ease-in-out infinite;
  filter: drop-shadow(0 2px 8px rgba(var(--text-accent-rgb, 100, 149, 237), 0.3));
}

.intro-icon-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  border: 2px solid var(--text-accent);
  opacity: 0;
  animation: ring-pulse 2.5s ease-out infinite;
}

.ring-delay {
  animation-delay: 1.25s;
}

.message-welcome {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-bright);
  margin-bottom: 0.5rem;
  animation: text-fade-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) 0.15s both;
}

.message-hint {
  font-size: 0.9rem;
  color: var(--text-dim);
  letter-spacing: 0.04em;
  animation: text-fade-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) 0.3s both;
}

.intro-divider {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  margin-top: 1.5rem;
  animation: text-fade-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) 0.45s both;
}

.divider-line {
  width: 60px;
  height: 1px;
  background: linear-gradient(to right, transparent, var(--text-dim), transparent);
  opacity: 0.25;
}

.divider-dot {
  width: 16px;
  height: 16px;
  opacity: 0.2;
  color: var(--text-dim);
}

.msg-intro-enter-active {
  transition: all 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.msg-intro-enter-from {
  opacity: 0;
  transform: translateY(12px);
}

@keyframes icon-bob {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

@keyframes ring-pulse {
  0% {
    transform: scale(0.6);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.4);
    opacity: 0;
  }
}

@keyframes text-fade-in {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

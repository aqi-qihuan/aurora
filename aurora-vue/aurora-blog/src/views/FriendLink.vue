<template>
  <div>
    <Breadcrumb :current="t('menu.friends')" />
    <div class="flex flex-col">
      <div class="post-header">
        <h1 class="post-title text-white uppercase">{{ t('titles.friends') }}</h1>
      </div>
      <div class="main-grid">
        <div class="relative space-y-5">
          <div class="bg-ob-deep-800 p-4 lg:p-14 rounded-2xl shadow-xl mb-8 lg:mb-0">
            <el-row :gutter="36">
              <transition-group name="link-stagger">
                <el-col v-for="(link, index) in links" :key="link.id" :span="8" :xs="{ span: 20, offset: 2 }" class="mb-3" :style="{ animationDelay: index * 0.05 + 's' }">
                  <a :href="link.linkAddress" target="_blank" rel="noopener noreferrer" class="link-card">
                    <el-card shadow="never" class="link-inner">
                      <div class="link-avatar-wrap">
                        <el-avatar :size="56" :src="link.linkAvatar" class="link-avatar" />
                      </div>
                      <div class="link-info">
                        <div class="link-name">{{ link.linkName }}</div>
                        <div class="link-intro">{{ link.linkIntro }}</div>
                      </div>
                      <svg-icon icon-class="arrow-right" class="link-arrow" />
                    </el-card>
                  </a>
                </el-col>
              </transition-group>
            </el-row>
          </div>
          <transition name="msg-intro" appear>
            <div class="friend-apply">
              <svg-icon icon-class="link" class="apply-icon" />
              <p class="apply-text">需要交换友链的可在下方留言</p>
              <p class="apply-hint">友链信息展示需要，你的信息格式要包含：名称、头像、链接、介绍</p>
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
import { defineComponent, reactive, provide, computed, toRefs, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Sidebar, Profile } from '../components/Sidebar'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { Comment } from '../components/Comment'
import { useCommentStore } from '@/stores/comment'
import emitter from '@/utils/mitt'
import api from '@/api/api'

export default defineComponent({
  name: 'FriendLink',
  components: { Sidebar, Profile, Breadcrumb, Comment },
  setup() {
    const { t } = useI18n()
    const commentStore = useCommentStore()
    const reactiveData = reactive({
      links: '' as any,
      comments: [] as any,
      haveMore: false as any,
      isReload: false as any
    })
    const pageInfo = reactive({
      current: 1,
      size: 7
    })
    commentStore.type = 4
    onMounted(() => {
      fetchLinks()
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
    emitter.on('friendLinkFetchComment', () => {
      pageInfo.current = 1
      reactiveData.isReload = true
      fetchComments()
    })
    emitter.on('friendLinkFetchReplies', (index) => {
      fetchReplies(index)
    })
    emitter.on('friendLinkLoadMore', () => {
      fetchComments()
    })
    const fetchLinks = () => {
      api.getFriendLink().then(({ data }) => {
        reactiveData.links = data.data
      })
    }
    const fetchComments = () => {
      const params = {
        type: 4,
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
.link-card {
  display: block;
  text-decoration: none;
  cursor: pointer;
  border-radius: 0.625rem;
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.25s ease;

  &:hover {
    transform: translateY(-3px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  }

  &:active {
    transform: translateY(0) scale(0.98);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
}

.link-inner {
  background: var(--background-primary) !important;
  border-radius: 0.625rem !important;
  border: 1px solid transparent !important;
  display: flex !important;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem !important;
  transition: border-color 0.25s ease, background 0.25s ease;

  .link-card:hover & {
    border-color: var(--text-accent);
  }
}

.link-avatar-wrap {
  flex-shrink: 0;
}

.link-avatar {
  border-radius: 50% !important;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

  .link-card:hover & {
    transform: scale(1.08);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
  }
}

.link-info {
  flex: 1;
  min-width: 0;
}

.link-name {
  font-weight: 600;
  font-size: 1rem;
  color: var(--text-bright);
  margin-bottom: 0.25rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.2s ease;

  .link-card:hover & {
    color: var(--text-accent);
  }
}

.link-intro {
  font-size: 0.8rem;
  color: var(--text-dim);
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.link-arrow {
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  color: var(--text-dim);
  opacity: 0;
  transform: translateX(-4px);
  transition: opacity 0.2s ease, transform 0.25s cubic-bezier(0.22, 1, 0.36, 1), color 0.2s ease;

  .link-card:hover & {
    opacity: 1;
    transform: translateX(0);
    color: var(--text-accent);
  }
}

.friend-apply {
  padding: 2rem 0;
  text-align: center;
}

.apply-icon {
  width: 24px;
  height: 24px;
  color: var(--text-accent);
  opacity: 0.6;
  margin-bottom: 0.75rem;
}

.apply-text {
  font-size: 0.95rem;
  color: var(--text-secondary);
  margin-bottom: 0.4rem;
}

.apply-hint {
  font-size: 0.8rem;
  color: var(--text-dim);
  letter-spacing: 0.02em;
}

.link-stagger-enter-active {
  transition: all 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.link-stagger-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}

.link-stagger-leave-active {
  transition: all 0.2s ease;
  position: absolute;
}

.link-stagger-leave-to {
  opacity: 0;
  transform: scale(0.9);
}

.link-stagger-move {
  transition: transform 0.3s ease;
}

.msg-intro-enter-active {
  transition: all 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.msg-intro-enter-from {
  opacity: 0;
  transform: translateY(10px);
}
</style>

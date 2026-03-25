<template>
  <div>
    <Breadcrumb :current="t('menu.talks')" />
    <div class="flex flex-col">
      <div class="post-header">
        <h1 class="post-title text-white uppercase">{{ t('titles.talks') }}</h1>
      </div>
      <div class="main-grid">
        <div class="relative space-y-5">
          <transition-group name="talk-stagger" tag="div" class="space-y-5">
            <div
              class="bg-ob-deep-800 flex p-4 lg:p-8 rounded-2xl shadow-xl mb-0 talk-item"
              v-for="(item, index) in talks"
              :key="item.id"
              :style="{ animationDelay: index * 0.05 + 's' }"
              @click="toTalk(item.id)">
              <Avatar :url="item.avatar" />
              <div class="talk-info">
                <div class="user-nickname text-sm">
                  {{ item.nickname }}
                </div>
                <div class="time">
                  {{ t('settings.shared-on') }}
                  {{ formatTime(item.createTime) }},
                  {{ t(`settings.months[${new Date(item.createTime).getMonth()}]`) }}
                  {{ new Date(item.createTime).getDate() }}, {{ new Date(item.createTime).getFullYear() }}
                  <template v-if="item.isTop === 1">
                    <svg-icon icon-class="top" class="top-svg" /><span class="pin-badge">置顶</span>
                  </template>
                  <svg-icon icon-class="message" class="message-svg" />{{
                    item.commentCount == null ? 0 : item.commentCount
                  }}
                </div>
                <div class="talk-content" v-html="item.content" />
                <el-row class="talk-images" v-if="item.imgs">
                  <el-col :md="4" v-for="(img, index) of item.imgs" :key="index">
                    <el-image
                      class="images-items"
                      :src="img"
                      aspect-ratio="1"
                      max-height="200"
                      @click.stop="handlePreview(img)" />
                  </el-col>
                </el-row>
              </div>
            </div>
          </transition-group>
          <Paginator
            :pageSize="pagination.size"
            :pageTotal="pagination.total"
            :page="pagination.current"
            @pageChange="pageChangeHandler" />
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
import { defineComponent, onMounted, reactive, toRefs } from 'vue'
import { useI18n } from 'vue-i18n'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { Sidebar, Profile } from '../components/Sidebar'
import Paginator from '@/components/Paginator.vue'
import Avatar from '../components/Avatar.vue'
import { v3ImgPreviewFn } from 'v3-img-preview'
import { useRouter } from 'vue-router'
import api from '@/api/api'

export default defineComponent({
  name: 'talkList',
  components: { Breadcrumb, Sidebar, Profile, Paginator, Avatar },
  setup() {
    const { t } = useI18n()
    const router = useRouter()
    const pagination = reactive({
      size: 7,
      total: 0,
      current: 1
    })
    const reactiveData = reactive({
      images: [] as any,
      talks: '' as any
    })
    onMounted(() => {
      fetchTalks()
    })
    const handlePreview = (index: any) => {
      v3ImgPreviewFn({ images: reactiveData.images, index: reactiveData.images.indexOf(index) })
    }
    const fetchTalks = () => {
      const params = {
        current: pagination.current,
        size: pagination.size
      }
      api.getTalks(params).then(({ data }) => {
        reactiveData.talks = data.data.records
        pagination.total = data.data.count
        reactiveData.talks.forEach((item: any) => {
          if (item.imgs) {
            reactiveData.images.push(...item.imgs)
          }
        })
      })
    }
    const formatTime = (data: any): string => {
      let hours = new Date(data).getHours()
      let minutes = new Date(data).getMinutes()
      let seconds = new Date(data).getSeconds()
      return hours + ':' + minutes + ':' + seconds
    }
    const toPageTop = () => {
      window.scrollTo({
        top: 0
      })
    }
    const pageChangeHandler = (current: number) => {
      reactiveData.talks = ''
      toPageTop()
      pagination.current = current
      fetchTalks()
    }
    const toTalk = (id: any) => {
      router.push({ path: '/talks/' + id })
    }
    return {
      pagination,
      ...toRefs(reactiveData),
      formatTime,
      pageChangeHandler,
      handlePreview,
      toTalk,
      t
    }
  }
})
</script>

<style lang="scss" scoped>
.talk-item {
  cursor: pointer;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.3s ease, border-color 0.3s ease;
  border: 1px solid transparent;
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
    border-color: var(--text-accent);
  }
  &:active {
    transform: translateY(0);
  }
}
.pin-badge {
  color: #f21835;
  font-size: 12px;
  font-weight: 600;
  padding: 1px 6px;
  border-radius: 4px;
  background: rgba(242, 24, 53, 0.12);
  letter-spacing: 0.03em;
  transition: background 0.2s ease;
}
.talk-item:hover .pin-badge {
  background: rgba(242, 24, 53, 0.2);
}
.top-svg {
  margin-left: 5px;
  transition: transform 0.2s ease;
}
.talk-item:hover .top-svg {
  transform: scale(1.15);
}
.message-svg {
  margin-left: 5px;
  font-size: 15px;
  transition: transform 0.2s ease;
}
.talk-item:hover .message-svg {
  transform: scale(1.1);
}
.talk-info {
  flex: 1;
  margin-left: 12px;
}
.user-nickname {
  font-weight: 530;
  color: var(--text-bright);
  transition: color 0.2s ease;
}
.talk-item:hover .user-nickname {
  color: var(--text-accent);
}
.time {
  color: var(--text-dim);
  font-size: 13px;
  @media (min-width: 1280px) {
    margin-top: 4px;
  }
}
.talk-content {
  margin-top: 12px;
  font-size: 14px;
  line-height: 1.75;
  white-space: pre-line;
  word-wrap: break-word;
  word-break: break-all;
  color: var(--text-normal);
}
.talk-images {
  margin-top: 10px;
}
.images-items {
  cursor: zoom-in;
  border-radius: 10px;
  margin-right: 6px;
  margin-bottom: 6px;
  overflow: hidden;
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.3s ease;
  &:hover {
    transform: scale(1.03);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.18);
  }
}
.talk-stagger-enter-active {
  transition: all 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}
.talk-stagger-enter-from {
  opacity: 0;
  transform: translateY(16px);
}
.talk-stagger-leave-active {
  transition: all 0.25s ease;
  position: absolute;
  width: 100%;
}
.talk-stagger-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
.talk-stagger-move {
  transition: transform 0.35s ease;
}
</style>

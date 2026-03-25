<template>
  <div>
    <Breadcrumb :current="t('menu.album')" />
    <div class="flex flex-col">
      <div class="post-header">
        <h1 v-if="photoAlbumName != ''" class="post-title text-white uppercase">{{ photoAlbumName }}</h1>
        <ob-skeleton
          v-else
          class="post-title text-white uppercase"
          width="30%"
          height="clamp(1.2rem, calc(1rem + 3.5vw), 4rem)" />
      </div>
      <div class="main-grid">
        <div class="relative">
          <div class="post-html">
            <div
              class="list-lis"
              v-infinite-scroll="loadDataFromServer"
              :infinite-scroll-immediate-check="false"
              :infinite-scroll-disabled="noResult"
              infinite-scroll-watch-disabled="scrollDisabled"
              :infinite-scroll-distance="isMobile ? 0 : 30">
              <transition-group name="photo-fade" tag="div" class="photo-wrap">
                <img
                  v-for="(item, index) of photos"
                  class="photo"
                  :key="item"
                  :src="item"
                  :style="{ animationDelay: index * 0.04 + 's' }"
                  loading="lazy"
                  @click="handlePreview(index)" />
              </transition-group>
              <div v-if="noResult && photos.length === 0" class="photos-empty">
                <svg-icon icon-class="eye" class="empty-icon" />
                <p class="empty-text">暂无照片</p>
              </div>
            </div>
          </div>
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
import { computed, defineComponent, reactive, toRefs } from 'vue'
import { useI18n } from 'vue-i18n'
import { useCommonStore } from '@/stores/common'
import { useRoute, onBeforeRouteUpdate } from 'vue-router'
import { Sidebar, Profile } from '../components/Sidebar'
import Breadcrumb from '@/components/Breadcrumb.vue'
import { v3ImgPreviewFn } from 'v3-img-preview'
import api from '@/api/api'

export default defineComponent({
  name: 'Photos',
  components: { Breadcrumb, Sidebar, Profile },
  setup() {
    const { t } = useI18n()
    const route = useRoute()
    const commonStore = useCommonStore()
    const reactiveData = reactive({
      photoAlbumName: '' as any,
      noResult: false,
      photos: [] as any,
      current: 1,
      size: 10,
      albumId: route.params.albumId
    })
    onBeforeRouteUpdate((to) => {
      reactiveData.photoAlbumName = ''
      reactiveData.photos = []
      reactiveData.noResult = false
      reactiveData.current = 1
      reactiveData.albumId = to.params.albumId
      loadDataFromServer()
    })
    const handlePreview = (index: any) => {
      v3ImgPreviewFn({ images: reactiveData.photos, index })
    }
    const loadDataFromServer = () => {
      let params = {
        current: reactiveData.current,
        size: reactiveData.size
      }
      api.getPhotosByAlbumId(reactiveData.albumId, params).then(({ data }: { data: any }) => {
        if (data.data.photos.length > 0) {
          reactiveData.current++
          reactiveData.photoAlbumName = data.data.photoAlbumName
          reactiveData.photos.push(...data.data.photos)
        } else {
          reactiveData.noResult = true
        }
      })
    }
    return {
      ...toRefs(reactiveData),
      handlePreview,
      loadDataFromServer,
      isMobile: computed(() => commonStore.isMobile),
      t
    }
  }
})
</script>
<style lang="scss" scoped>
.photo-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.photo-wrap::after {
  content: '';
  display: block;
  flex-grow: 9999;
}

.photo {
  flex-grow: 1;
  height: 200px;
  object-fit: cover;
  border-radius: 8px;
  cursor: zoom-in;
  background: var(--background-secondary);
  transition: transform 0.3s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.3s ease,
              opacity 0.4s ease;

  &:hover {
    transform: scale(1.02);
    box-shadow: 0 8px 28px rgba(0, 0, 0, 0.25);
    z-index: 2;
  }

  &:active {
    transform: scale(0.99);
  }
}

.photos-empty {
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

.photo-fade-enter-active {
  transition: all 0.4s cubic-bezier(0.22, 1, 0.36, 1);
}

.photo-fade-enter-from {
  opacity: 0;
  transform: scale(0.92);
}

.photo-fade-leave-active {
  transition: all 0.25s ease;
  position: absolute;
}

.photo-fade-leave-to {
  opacity: 0;
  transform: scale(0.9);
}

.photo-fade-move {
  transition: transform 0.3s ease;
}

@media (max-width: 759px) {
  .photo-wrap {
    gap: 4px;
  }

  .photo {
    width: 100%;
    height: auto;
    aspect-ratio: 4 / 3;
    border-radius: 6px;
  }
}
</style>

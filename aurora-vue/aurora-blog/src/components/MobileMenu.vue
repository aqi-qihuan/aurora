<template>
  <div
    class="flex flex-col justify-center items-center"
    @touchstart="handleTouchStart"
    @touchmove="handleTouchMove"
    @touchend="handleTouchEnd">
    <img
      v-if="websiteConfig.authorAvatar !== ''"
      class="circle-avatar h-28 w-28 shadow-xl m-0"
      :src="websiteConfig.authorAvatar"
      alt="avatar" />
    <ob-skeleton v-else width="7rem" height="7rem" circle />

    <h2 class="text-center pt-4 text-4xl font-semibold text-ob-bright">
      <template v-if="websiteConfig.author">
        {{ websiteConfig.author }}
      </template>
      <ob-skeleton v-else height="2.25rem" width="7rem" />
    </h2>

    <span class="h-1 w-14 rounded-full mt-2" :style="gradientBackground" />

    <p
      v-if="websiteConfig.authorIntro"
      class="pt-6 px-2 w-full text-sm text-center text-ob-dim"
      v-html="websiteConfig.authorIntro" />
    <p v-else class="pt-6 px-10 w-full text-sm text-center flex flex-col gap-2">
      <ob-skeleton :count="2" height="20px" width="10rem" />
    </p>
    <Social />
    <ul class="grid grid-cols-3 pt-4 w-full px-2 text-lg">
      <li class="col-span-1 text-center">
        <span class="text-ob-bright">{{ articleCount }}</span>
        <p class="text-base text-ob-dim">{{ t('settings.articles') }}</p>
      </li>
      <li class="col-span-1 text-center">
        <span class="text-ob-bright">{{ categoryCount }}</span>
        <p class="text-base text-ob-dim">{{ t('settings.categories') }}</p>
      </li>
      <li class="col-span-1 text-center">
        <span class="text-ob-bright">{{ tagCount }}</span>
        <p class="text-base text-ob-dim">{{ t('settings.tags') }}</p>
      </li>
    </ul>
  </div>
  <ul class="flex flex-col justify-center items-center mt-8 w-full list-none text-ob-bright">
    <li class="pb-2 cursor-pointer" v-for="route in routes" :key="route.path">
      <div
        class="mobile-menu-item text-sm block px-1.5 py-0.5 rounded-md relative uppercase"
        @click="pushPage(route.path)"
        @touchstart="handleMenuItemTouchStart"
        @touchend="handleMenuItemTouchEnd"
        v-if="route.children && route.children.length === 0">
        <span class="relative z-50" v-if="$i18n.locale === 'cn' && route.i18n.cn">
          {{ route.i18n.cn }}
        </span>
        <span class="relative z-50" v-else-if="$i18n.locale === 'en' && route.i18n.en">
          {{ route.i18n.en }}
        </span>
        <span class="relative z-50" v-else>{{ route.name }}</span>
      </div>
      <Dropdown
        @command="pushPage"
        v-else
        class="flex flex-col justify-center items-center nav-link text-sm block px-1.5 py-0.5 rounded-md relative uppercase">
        <span class="relative z-50" v-if="$i18n.locale === 'cn' && route.i18n.cn">
          {{ route.i18n.cn }}
        </span>
        <span class="relative z-50" v-else-if="$i18n.locale === 'en' && route.i18n.en">
          {{ route.i18n.en }}
        </span>
        <span class="relative z-50" v-else>{{ route.name }}</span>
        <DropdownMenu expand>
          <DropdownItem v-for="sub in route.children" :key="sub.path" :name="sub.path">
            <span class="relative z-50" v-if="$i18n.locale === 'cn' && sub.i18n.cn">
              {{ sub.i18n.cn }}
            </span>
            <span class="relative z-50" v-else-if="$i18n.locale === 'en' && sub.i18n.en">
              {{ sub.i18n.en }}
            </span>
            <span class="relative z-50" v-else>{{ sub.name }}</span>
          </DropdownItem>
        </DropdownMenu>
      </Dropdown>
    </li>
    <li>
      <Dropdown
        class="flex flex-col justify-center items-center nav-link text-sm block px-1.5 py-0.5 rounded-md relative uppercase">
        <span class="relative z-50" v-if="$i18n.locale === 'cn'"> 相册 </span>
        <span class="relative z-50" v-else-if="$i18n.locale === 'en'"> PhotoAlbums </span>
        <DropdownMenu expand>
          <template v-for="item in albums" :key="item.id">
            <DropdownItem @click="pushPage(`/photos/${item.id}`)" :name="item.albumName">
              <span class="relative z-50">{{ item.albumName }}</span>
            </DropdownItem>
          </template>
        </DropdownMenu>
      </Dropdown>
    </li>
  </ul>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, reactive, toRefs, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { Dropdown, DropdownMenu, DropdownItem } from '@/components/Dropdown'
import { useRouter } from 'vue-router'
import { useNavigatorStore } from '@/stores/navigator'
import Social from '@/components/Social.vue'
import config from '@/config/config'
import api from '@/api/api'

export default defineComponent({
  name: 'ObMobileMenu',
  components: { Dropdown, DropdownMenu, DropdownItem, Social },
  setup() {
    const appStore = useAppStore()
    const router = useRouter()
    const navigatorStore = useNavigatorStore()
    const { t } = useI18n()
    const reactiveData = reactive({
      routes: '' as any,
      albums: [] as any
    })
    
    // 手势相关变量
    const touchStartX = ref(0)
    const touchStartY = ref(0)
    const swipeThreshold = 50 // 滑动阈值
    const isVerticalSwipe = ref(false)
    
    onMounted(() => {
      reactiveData.routes = config.routes
      fetchAblums()
    })
    
    const fetchAblums = () => {
      api.getAlbums().then(({ data }) => {
        reactiveData.albums = data.data
      }).catch((error) => {
        console.error('获取相册列表失败:', error)
      })
    }
    
    const pushPage = (path: string): void => {
      if (!path) return
      navigatorStore.toggleMobileMenu()
      navigatorStore.setOpenNavigator(false)
      if (path.match(/(http:\/\/|https:\/\/)((\w|=|\?|\.|\/|&|-)+)/g)) {
        window.location.href = path
      } else {
        router.push({
          path: path
        })
      }
    }
    
    // 手势处理函数
    const handleTouchStart = (e: TouchEvent) => {
      touchStartX.value = e.touches[0].clientX
      touchStartY.value = e.touches[0].clientY
      isVerticalSwipe.value = false
    }
    
    const handleTouchMove = (e: TouchEvent) => {
      const currentX = e.touches[0].clientX
      const currentY = e.touches[0].clientY
      const diffX = currentX - touchStartX.value
      const diffY = currentY - touchStartY.value
      
      // 判断是水平滑动还是垂直滑动
      if (Math.abs(diffX) > Math.abs(diffY)) {
        isVerticalSwipe.value = false
      } else {
        isVerticalSwipe.value = true
      }
    }
    
    const handleTouchEnd = (e: TouchEvent) => {
      if (isVerticalSwipe.value) return // 如果是垂直滑动,不处理
      
      const endX = e.changedTouches[0].clientX
      const diffX = endX - touchStartX.value
      
      // 向左滑动关闭菜单
      if (diffX < -swipeThreshold) {
        navigatorStore.toggleMobileMenu()
      }
    }
    
    // 菜单项触摸反馈
    const handleMenuItemTouchStart = (e: TouchEvent) => {
      const target = e.currentTarget as HTMLElement
      target.style.transform = 'scale(0.95)'
      target.style.opacity = '0.7'
    }
    
    const handleMenuItemTouchEnd = (e: TouchEvent) => {
      const target = e.currentTarget as HTMLElement
      target.style.transform = 'scale(1)'
      target.style.opacity = '1'
    }
    
    return {
      ...toRefs(reactiveData),
      themeConfig: computed(() => appStore.themeConfig),
      gradientBackground: computed(() => {
        return { background: appStore.themeConfig.header_gradient_css }
      }),
      pushPage,
      websiteConfig: computed(() => appStore.websiteConfig),
      articleCount: computed(() => appStore.articleCount),
      talkCount: computed(() => appStore.talkCount),
      categoryCount: computed(() => appStore.categoryCount),
      tagCount: computed(() => appStore.tagCount),
      t,
      handleTouchStart,
      handleTouchMove,
      handleTouchEnd,
      handleMenuItemTouchStart,
      handleMenuItemTouchEnd
    }
  }
})
</script>

<style lang="scss" scoped>
// 移动端菜单项触摸优化
.mobile-menu-item {
  min-height: 48px; // 符合移动端触摸目标最小尺寸
  min-width: 48px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 24px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  touch-action: manipulation; // 优化触摸性能
  user-select: none; // 防止文本选中
  -webkit-tap-highlight-color: transparent; // 移除iOS点击高亮
  position: relative;
  
  &::before {
    content: '';
    position: absolute;
    left: 0;
    bottom: 0;
    width: 0;
    height: 2px;
    background: var(--text-accent);
    transition: width 0.3s ease;
  }
  
  &:hover {
    color: var(--text-accent);
    
    &::before {
      width: 100%;
    }
  }
  
  &:active {
    background-color: rgba(255, 255, 255, 0.1);
    transform: scale(0.95);
  }
}

// 导航链接优化
.nav-link {
  min-height: 48px;
  min-width: 48px;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  
  &:active {
    opacity: 0.7;
    transform: scale(0.98);
  }
}

// 优化滚动性能
ul {
  -webkit-overflow-scrolling: touch;
  scroll-behavior: smooth;
}

// 优化头像区域
.circle-avatar {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  
  &:active {
    transform: scale(0.95);
  }
}

// 优化统计数据区域
.grid.grid-cols-3 {
  > li {
    padding: 12px 0;
    transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    
    &:active {
      transform: scale(0.95);
    }
  }
}
</style>

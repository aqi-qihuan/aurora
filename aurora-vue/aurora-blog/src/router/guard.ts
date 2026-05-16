import router from '@/router'
import { useAppStore } from '@/stores/app'

router.beforeEach(async (_: any, __: any, next: any) => {
  const appStore = useAppStore()
  appStore.startLoading()
  next()
})

router.afterEach(() => {
  const appStore = useAppStore()
  appStore.endLoading()
  document.getElementById('App-Container')?.focus()
})

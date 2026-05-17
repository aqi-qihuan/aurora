import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

export default pinia

// 导出所有 store
export { useAppStore } from './app'
export { useUserStore, type UserInfo, type LoginParams } from './user'
export { usePermissionStore, type PermissionState } from './permission'
export { usePageStateStore, type PageState, type PhotoPageState, type QuartzLogPageState } from './pageState'

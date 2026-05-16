import ObSkeleton from '@/components/LoadingSkeleton/src/Skeleton.vue'
import ObSkeletonTheme from '@/components/LoadingSkeleton/src/SkeletonTheme.vue'
import { App } from 'vue'

export const registerObSkeleton = (app: App): void => {
  if (ObSkeleton.name) {
    app.component(ObSkeleton.name, ObSkeleton)
  }
  if (ObSkeletonTheme.name) {
    app.component(ObSkeletonTheme.name, ObSkeletonTheme)
  }
}

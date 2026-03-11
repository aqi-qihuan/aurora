<template>
  <div
    class="ob-dropdown relative z-50"
    @click="toggle"
    @mouseover="hoverHandler"
    @mouseleave="leaveHandler"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    v-click-away="onClickAway">
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, reactive, ref, toRefs, watch } from 'vue'
import { useDropdownStore } from '@/stores/dropdown'
import { useCommonStore } from '@/stores/common'

export default defineComponent({
  emits: ['command'],
  name: 'Dropdown',
  props: {
    hover: {
      type: Boolean,
      default: false
    }
  },
  setup(props, { emit }) {
    const commonStore = useCommonStore()
    const mouseHover = toRefs(props).hover
    const dropdownStore = useDropdownStore()
    const eventId = ref(0)
    const touchStartTime = ref(0)
    let sharedState: { active: boolean } = reactive({
      active: false
    })
    provide('sharedState', sharedState)
    watch(
      () => dropdownStore.commandName,
      (newName, oldName) => {
        const name = newName ? newName : oldName
        if (eventId.value === dropdownStore.uid) emit('command', name)
      }
    )
    const toggle = () => {
      if (!sharedState.active) eventId.value = dropdownStore.setUid()
      if (!mouseHover.value) sharedState.active = !sharedState.active
    }
    const onClickAway = () => {
      if (!mouseHover.value && !commonStore.isMobile) {
        sharedState.active = false
        eventId.value = 0
      }
    }
    const hoverHandler = () => {
      if (!sharedState.active) eventId.value = dropdownStore.setUid()
      if (mouseHover.value) sharedState.active = true
    }
    const leaveHandler = () => {
      if (mouseHover.value) {
        sharedState.active = false
        eventId.value = 0
      }
    }
    
    // 移动端触摸处理
    const handleTouchStart = () => {
      touchStartTime.value = Date.now()
    }
    
    const handleTouchEnd = () => {
      const touchDuration = Date.now() - touchStartTime.value
      // 防止长按误触
      if (touchDuration < 500) {
        toggle()
      }
    }
    
    return { toggle, onClickAway, hoverHandler, leaveHandler, handleTouchStart, handleTouchEnd }
  }
})
</script>

<style lang="scss" scoped>
.ob-dropdown {
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
}
</style>

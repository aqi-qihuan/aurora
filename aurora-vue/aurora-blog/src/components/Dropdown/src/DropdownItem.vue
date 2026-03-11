<template>
  <div
    @click.stop.prevent="handleClick"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
    class="dropdown-item block cursor-pointer hover:bg-ob-trans my-1 px-4 py-2 font-medium hover:text-ob-bright">
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue'
import { useDropdownStore } from '@/stores/dropdown'

export default defineComponent({
  name: 'DropdownItem',
  props: {
    name: String
  },
  setup(props) {
    const dropdownStore = useDropdownStore()
    const touchStartTime = ref(0)
    
    const handleClick = () => {
      dropdownStore.setCommand(String(props.name))
    }
    
    const handleTouchStart = () => {
      touchStartTime.value = Date.now()
    }
    
    const handleTouchEnd = () => {
      const touchDuration = Date.now() - touchStartTime.value
      if (touchDuration < 500) {
        handleClick()
      }
    }
    
    return { handleClick, handleTouchStart, handleTouchEnd }
  }
})
</script>

<style lang="scss" scoped>
.dropdown-item {
  min-height: 44px; // 符合移动端触摸目标最小尺寸
  min-width: 44px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  touch-action: manipulation;
  -webkit-tap-highlight-color: transparent;
  user-select: none;
  transition: all 0.2s ease;
  
  &:active {
    background-color: rgba(255, 255, 255, 0.1);
    transform: scale(0.98);
  }
}
</style>

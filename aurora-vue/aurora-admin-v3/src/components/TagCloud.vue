<template>
  <div class="tag-cloud-3d-container" ref="containerRef">
    <div class="tag-cloud-3d" ref="cloudRef">
      <span
        v-for="(tag, index) in processedTags"
        :key="tag.id"
        class="tag-3d-item"
        :style="{
          color: tag.color,
          fontSize: tag.size + 'px',
          transform: `translate(-50%, -50%) translate3d(${tag.x}px, ${tag.y}px, ${tag.z}px)`,
          opacity: tag.opacity,
          zIndex: Math.round(tag.z)
        }"
        @mouseenter="pauseAnimation"
        @mouseleave="resumeAnimation"
      >
        {{ tag.name }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted, reactive } from 'vue'

const props = defineProps({
  tags: {
    type: Array,
    default: () => []
  },
  data: {
    type: Array,
    default: () => []
  }
})

const containerRef = ref(null)
const cloudRef = ref(null)
const isPaused = ref(false)
let animationId = null

// 颜色组
const colors = [
  '#3B82F6', '#8B5CF6', '#EC4899', '#F59E0B', 
  '#10B981', '#06B6D4', '#6366F1', '#F97316',
  '#EF4444', '#14B8A6', '#A855F7', '#F43F5E'
]

// 3D 标签数据
const processedTags = ref([])

// 球体参数
const radius = 180
const angleX = ref(0)
const angleY = ref(0)
const speed = 0.003

// 初始化标签位置（球形分布）
const initTags = () => {
  const tagData = props.tags || props.data || []
  const count = tagData.length
  
  if (count === 0) return
  
  const tags3D = tagData.map((tag, index) => {
    // 使用斐波那契球面分布算法
    const phi = Math.acos(-1 + (2 * index + 1) / count)
    const theta = Math.sqrt(count * Math.PI) * phi
    
    const x = radius * Math.cos(theta) * Math.sin(phi)
    const y = radius * Math.sin(theta) * Math.sin(phi)
    const z = radius * Math.cos(phi)
    
    return {
      ...tag,
      x,
      y,
      z,
      origX: x,
      origY: y,
      origZ: z,
      size: 12 + Math.random() * 10,
      color: colors[index % colors.length],
      opacity: 0.5 + (z + radius) / (2 * radius) * 0.5
    }
  })
  
  processedTags.value = tags3D
}

// 旋转矩阵
const rotateX = (x, y, z, angle) => {
  const cos = Math.cos(angle)
  const sin = Math.sin(angle)
  return {
    x: x,
    y: y * cos - z * sin,
    z: y * sin + z * cos
  }
}

const rotateY = (x, y, z, angle) => {
  const cos = Math.cos(angle)
  const sin = Math.sin(angle)
  return {
    x: x * cos - z * sin,
    y: y,
    z: x * sin + z * cos
  }
}

// 动画循环
const animate = () => {
  if (!isPaused.value) {
    angleY.value += speed
    
    processedTags.value = processedTags.value.map(tag => {
      // Y轴旋转
      const rotated = rotateY(tag.origX, tag.origY, tag.origZ, angleY.value)
      
      // 计算透视
      const perspective = 400
      const scale = perspective / (perspective - rotated.z)
      const opacity = 0.3 + (rotated.z + radius) / (2 * radius) * 0.7
      
      return {
        ...tag,
        x: rotated.x * scale,
        y: rotated.y * scale,
        z: rotated.z,
        opacity: Math.max(0.2, Math.min(1, opacity)),
        scale
      }
    })
    
    // 按 z 排序，远的先渲染
    processedTags.value.sort((a, b) => a.z - b.z)
  }
  
  animationId = requestAnimationFrame(animate)
}

// 暂停/恢复动画
const pauseAnimation = () => {
  isPaused.value = true
}

const resumeAnimation = () => {
  isPaused.value = false
}

// 鼠标交互
const handleMouseMove = (e) => {
  if (!containerRef.value) return
  
  const rect = containerRef.value.getBoundingClientRect()
  const centerX = rect.left + rect.width / 2
  const centerY = rect.top + rect.height / 2
  
  const mouseX = e.clientX - centerX
  const mouseY = e.clientY - centerY
  
  // 根据鼠标位置调整旋转速度
  const speedMultiplier = 0.00005
  angleX.value = mouseY * speedMultiplier
  angleY.value += mouseX * speedMultiplier
}

onMounted(() => {
  initTags()
  animate()
  
  // 添加鼠标移动监听
  document.addEventListener('mousemove', handleMouseMove)
})

onUnmounted(() => {
  if (animationId) {
    cancelAnimationFrame(animationId)
  }
  document.removeEventListener('mousemove', handleMouseMove)
})

// 监听数据变化
const stopWatch = computed(() => {
  initTags()
  return props.tags || props.data
})
</script>

<style scoped>
.tag-cloud-3d-container {
  width: 100%;
  height: 350px;
  position: relative;
  overflow: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
  perspective: 800px;
}

.tag-cloud-3d {
  position: relative;
  width: 400px;
  height: 350px;
  transform-style: preserve-3d;
}

.tag-3d-item {
  position: absolute;
  left: 50%;
  top: 50%;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.1s ease;
  text-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
  white-space: nowrap;
  will-change: transform, opacity;
  backface-visibility: hidden;
}

.tag-3d-item:hover {
  color: #3B82F6 !important;
  text-shadow: 0 0 20px rgba(59, 130, 246, 0.8);
  transform: translate(-50%, -50%) scale(1.3) !important;
}

/* 暗色主题 */
:root[data-theme='dark'] .tag-3d-item {
  text-shadow: 0 0 10px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .tag-3d-item:hover {
  text-shadow: 0 0 20px rgba(59, 130, 246, 0.8);
}

/* 响应式 */
@media (max-width: 768px) {
  .tag-cloud-3d-container {
    height: 280px;
  }
  
  .tag-cloud-3d {
    width: 300px;
    height: 280px;
  }
}
</style>

<template>
  <slot v-if="hasAuth" />
</template>

<script setup>
/**
 * 权限包装组件
 * 用于根据权限控制内容显示
 * 
 * 使用示例:
 * <AuthWrapper permission="/users">
 *   <el-button>用户管理</el-button>
 * </AuthWrapper>
 * 
 * <AuthWrapper :role="['admin', 'editor']">
 *   <el-button>管理员或编辑</el-button>
 * </AuthWrapper>
 */
import { computed } from 'vue'
import { hasPermission, hasRole } from '@/utils/auth'

const props = defineProps({
  // 权限标识
  permission: {
    type: [String, Array],
    default: null
  },
  // 角色标识
  role: {
    type: [String, Array],
    default: null
  },
  // 是否需要同时满足权限和角色
  requireAll: {
    type: Boolean,
    default: false
  }
})

const hasAuth = computed(() => {
  // 如果没有设置权限和角色,默认显示
  if (!props.permission && !props.role) {
    return true
  }
  
  const hasPermissionResult = props.permission ? hasPermission(props.permission) : true
  const hasRoleResult = props.role ? hasRole(props.role) : true
  
  // 如果需要同时满足
  if (props.requireAll) {
    return hasPermissionResult && hasRoleResult
  }
  
  // 只要满足一个即可
  return hasPermissionResult || hasRoleResult
})
</script>

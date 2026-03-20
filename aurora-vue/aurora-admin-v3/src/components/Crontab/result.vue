<template>
  <div class="crontab-result">
    <div class="result-item">
      <span class="label">秒:</span>
      <span class="value">{{ second || '*' }}</span>
    </div>
    <div class="result-item">
      <span class="label">分:</span>
      <span class="value">{{ minute || '*' }}</span>
    </div>
    <div class="result-item">
      <span class="label">时:</span>
      <span class="value">{{ hour || '*' }}</span>
    </div>
    <div class="result-item">
      <span class="label">日:</span>
      <span class="value">{{ day || '*' }}</span>
    </div>
    <div class="result-item">
      <span class="label">月:</span>
      <span class="value">{{ month || '*' }}</span>
    </div>
    <div class="result-item">
      <span class="label">周:</span>
      <span class="value">{{ week || '?' }}</span>
    </div>
    <div class="result-item">
      <span class="label">年:</span>
      <span class="value">{{ year || '*' }}</span>
    </div>
    <div class="cron-expression">
      <span class="label">Cron表达式:</span>
      <el-input v-model="cronExpression" readonly>
        <template #append>
          <el-button @click="copyCron">复制</el-button>
        </template>
      </el-input>
    </div>
    <div class="cron-description">
      <span class="label">执行说明:</span>
      <p>{{ description }}</p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  second: {
    type: String,
    default: '*'
  },
  minute: {
    type: String,
    default: '*'
  },
  hour: {
    type: String,
    default: '*'
  },
  day: {
    type: String,
    default: '*'
  },
  month: {
    type: String,
    default: '*'
  },
  week: {
    type: String,
    default: '?'
  },
  year: {
    type: String,
    default: '*'
  }
})

const cronExpression = computed(() => {
  return `${props.second} ${props.minute} ${props.hour} ${props.day} ${props.month} ${props.week} ${props.year}`
})

const description = computed(() => {
  // 简化的 Cron 表达式说明
  const parts = []
  
  if (props.second !== '*') parts.push(`秒: ${props.second}`)
  if (props.minute !== '*') parts.push(`分: ${props.minute}`)
  if (props.hour !== '*') parts.push(`时: ${props.hour}`)
  if (props.day !== '*') parts.push(`日: ${props.day}`)
  if (props.month !== '*') parts.push(`月: ${props.month}`)
  if (props.week !== '?') parts.push(`周: ${props.week}`)
  if (props.year !== '*') parts.push(`年: ${props.year}`)
  
  if (parts.length === 0) {
    return '每秒钟执行一次'
  }
  
  return `在指定的 ${parts.join(', ')} 时执行`
})

const copyCron = () => {
  navigator.clipboard.writeText(cronExpression.value).then(() => {
    ElMessage.success('Cron 表达式已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}
</script>

<style scoped>
.crontab-result {
  padding: 10px;
}

.result-item {
  display: flex;
  margin-bottom: 10px;
  align-items: center;
}

.label {
  font-weight: bold;
  margin-right: 10px;
  min-width: 80px;
}

.value {
  color: #409eff;
  font-family: monospace;
}

.cron-expression {
  margin-top: 20px;
  margin-bottom: 20px;
}

.cron-expression .label {
  display: block;
  margin-bottom: 10px;
}

.cron-description {
  margin-top: 20px;
}

.cron-description .label {
  display: block;
  margin-bottom: 10px;
}

.cron-description p {
  color: #606266;
  line-height: 1.6;
  margin: 0;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>

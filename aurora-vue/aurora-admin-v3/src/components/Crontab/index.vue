<template>
  <div class="crontab-container">
    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="秒" name="second">
        <Second v-model="second" />
      </el-tab-pane>
      <el-tab-pane label="分" name="minute">
        <Min v-model="minute" />
      </el-tab-pane>
      <el-tab-pane label="时" name="hour">
        <Hour v-model="hour" />
      </el-tab-pane>
      <el-tab-pane label="日" name="day">
        <Day v-model="day" />
      </el-tab-pane>
      <el-tab-pane label="月" name="month">
        <Month v-model="month" />
      </el-tab-pane>
      <el-tab-pane label="周" name="week">
        <Week v-model="week" />
      </el-tab-pane>
      <el-tab-pane label="年" name="year">
        <Year v-model="year" />
      </el-tab-pane>
    </el-tabs>
    
    <Result
      :second="second"
      :minute="minute"
      :hour="hour"
      :day="day"
      :month="month"
      :week="week"
      :year="year"
    />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import Second from './second.vue'
import Min from './min.vue'
import Hour from './hour.vue'
import Day from './day.vue'
import Month from './month.vue'
import Week from './week.vue'
import Year from './year.vue'
import Result from './result.vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: '* * * * * ? *'
  }
})

const emit = defineEmits(['update:modelValue'])

const activeTab = ref('minute')
const second = ref('*')
const minute = ref('*')
const hour = ref('*')
const day = ref('*')
const month = ref('*')
const week = ref('?')
const year = ref('*')

const parseCron = (cron) => {
  if (!cron) return
  
  const parts = cron.trim().split(/\s+/)
  if (parts.length >= 6) {
    second.value = parts[0] || '*'
    minute.value = parts[1] || '*'
    hour.value = parts[2] || '*'
    day.value = parts[3] || '*'
    month.value = parts[4] || '*'
    week.value = parts[5] || '?'
    year.value = parts[6] || '*'
  }
}

const generateCron = () => {
  return `${second.value} ${minute.value} ${hour.value} ${day.value} ${month.value} ${week.value} ${year.value}`
}

watch(() => props.modelValue, (newVal) => {
  parseCron(newVal)
}, { immediate: true })

watch([second, minute, hour, day, month, week, year], () => {
  emit('update:modelValue', generateCron())
})
</script>

<style scoped>
.crontab-container {
  width: 100%;
}

.el-tabs {
  margin-bottom: 20px;
}

:deep(.el-tabs__content) {
  padding: 20px;
}
</style>

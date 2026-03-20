<template>
  <div class="crontab-month">
    <el-radio-group v-model="type" @change="changeType">
      <div class="radio-item">
        <el-radio value="1">每月</el-radio>
      </div>
      <div class="radio-item">
        <el-radio value="2">周期从</el-radio>
        <el-input-number v-model="cycle.start" :min="1" :max="12" :disabled="type !== '2'" @change="changeCycle" />
        <span>到</span>
        <el-input-number v-model="cycle.end" :min="1" :max="12" :disabled="type !== '2'" @change="changeCycle" />
        <span>月</span>
      </div>
      <div class="radio-item">
        <el-radio value="3">从</el-radio>
        <el-input-number v-model="loop.start" :min="1" :max="12" :disabled="type !== '3'" @change="changeLoop" />
        <span>月开始,每</span>
        <el-input-number v-model="loop.end" :min="1" :max="12" :disabled="type !== '3'" @change="changeLoop" />
        <span>月执行一次</span>
      </div>
      <div class="radio-item">
        <el-radio value="4">指定月</el-radio>
        <el-select v-model="appoint" multiple :disabled="type !== '4'" placeholder="请选择" @change="changeAppoint">
          <el-option v-for="i in 12" :key="i" :label="i" :value="i.toString()" />
        </el-select>
      </div>
    </el-radio-group>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: '*'
  }
})

const emit = defineEmits(['update:modelValue'])

const type = ref('1')
const cycle = ref({ start: 1, end: 12 })
const loop = ref({ start: 1, end: 1 })
const appoint = ref([])

const parseValue = (value) => {
  if (value === '*') {
    type.value = '1'
  } else if (value.includes('-')) {
    type.value = '2'
    const [start, end] = value.split('-')
    cycle.value = { start: parseInt(start), end: parseInt(end) }
  } else if (value.includes('/')) {
    type.value = '3'
    const [start, end] = value.split('/')
    loop.value = { start: parseInt(start), end: parseInt(end) }
  } else {
    type.value = '4'
    appoint.value = value.split(',')
  }
}

const changeType = () => {
  switch (type.value) {
    case '1':
      emit('update:modelValue', '*')
      break
    case '2':
      changeCycle()
      break
    case '3':
      changeLoop()
      break
    case '4':
      changeAppoint()
      break
  }
}

const changeCycle = () => {
  emit('update:modelValue', `${cycle.value.start}-${cycle.value.end}`)
}

const changeLoop = () => {
  emit('update:modelValue', `${loop.value.start}/${loop.value.end}`)
}

const changeAppoint = () => {
  emit('update:modelValue', appoint.value.join(','))
}

watch(() => props.modelValue, (newVal) => {
  parseValue(newVal)
}, { immediate: true })
</script>

<style scoped>
.crontab-month {
  padding: 10px;
}

.radio-item {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.el-input-number {
  width: 100px;
}

.el-select {
  width: 200px;
}
</style>

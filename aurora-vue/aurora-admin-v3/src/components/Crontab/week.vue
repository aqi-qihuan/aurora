<template>
  <div class="crontab-week">
    <el-radio-group v-model="type" @change="changeType">
      <div class="radio-item">
        <el-radio value="1">每周</el-radio>
      </div>
      <div class="radio-item">
        <el-radio value="2">周期从</el-radio>
        <el-select v-model="cycle.start" :disabled="type !== '2'" placeholder="请选择" @change="changeCycle">
          <el-option v-for="item in weekOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <span>到</span>
        <el-select v-model="cycle.end" :disabled="type !== '2'" placeholder="请选择" @change="changeCycle">
          <el-option v-for="item in weekOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </div>
      <div class="radio-item">
        <el-radio value="3">第</el-radio>
        <el-input-number v-model="loop.start" :min="1" :max="4" :disabled="type !== '3'" @change="changeLoop" />
        <span>周的</span>
        <el-select v-model="loop.end" :disabled="type !== '3'" placeholder="请选择" @change="changeLoop">
          <el-option v-for="item in weekOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </div>
      <div class="radio-item">
        <el-radio value="4">指定周</el-radio>
        <el-select v-model="appoint" multiple :disabled="type !== '4'" placeholder="请选择" @change="changeAppoint">
          <el-option v-for="item in weekOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </div>
      <div class="radio-item">
        <el-radio value="5">本月最后一个</el-radio>
        <el-select v-model="last" :disabled="type !== '5'" placeholder="请选择" @change="changeLast">
          <el-option v-for="item in weekOptions" :key="item.value" :label="item.label" :value="item.value" />
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
    default: '?'
  }
})

const emit = defineEmits(['update:modelValue'])

const weekOptions = [
  { label: '周日', value: '1' },
  { label: '周一', value: '2' },
  { label: '周二', value: '3' },
  { label: '周三', value: '4' },
  { label: '周四', value: '5' },
  { label: '周五', value: '6' },
  { label: '周六', value: '7' }
]

const type = ref('1')
const cycle = ref({ start: '1', end: '7' })
const loop = ref({ start: 1, end: '1' })
const appoint = ref([])
const last = ref('1')

const parseValue = (value) => {
  if (value === '?' || value === '*') {
    type.value = '1'
  } else if (value.includes('-')) {
    type.value = '2'
    const [start, end] = value.split('-')
    cycle.value = { start, end }
  } else if (value.includes('#')) {
    type.value = '3'
    const [end, start] = value.split('#')
    loop.value = { start: parseInt(start), end }
  } else if (value.includes('L')) {
    type.value = '5'
    last.value = value.replace('L', '')
  } else {
    type.value = '4'
    appoint.value = value.split(',')
  }
}

const changeType = () => {
  switch (type.value) {
    case '1':
      emit('update:modelValue', '?')
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
    case '5':
      changeLast()
      break
  }
}

const changeCycle = () => {
  emit('update:modelValue', `${cycle.value.start}-${cycle.value.end}`)
}

const changeLoop = () => {
  emit('update:modelValue', `${loop.value.end}#${loop.value.start}`)
}

const changeAppoint = () => {
  emit('update:modelValue', appoint.value.join(','))
}

const changeLast = () => {
  emit('update:modelValue', `${last.value}L`)
}

watch(() => props.modelValue, (newVal) => {
  parseValue(newVal)
}, { immediate: true })
</script>

<style scoped>
.crontab-week {
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
  width: 150px;
}
</style>

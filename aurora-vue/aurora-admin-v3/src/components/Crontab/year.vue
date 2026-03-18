<template>
  <div class="crontab-year">
    <el-radio-group v-model="type" @change="changeType">
      <div class="radio-item">
        <el-radio value="1">每年</el-radio>
      </div>
      <div class="radio-item">
        <el-radio value="2">周期从</el-radio>
        <el-input-number v-model="cycle.start" :min="1970" :max="2099" :disabled="type !== '2'" @change="changeCycle" />
        <span>到</span>
        <el-input-number v-model="cycle.end" :min="1970" :max="2099" :disabled="type !== '2'" @change="changeCycle" />
        <span>年</span>
      </div>
      <div class="radio-item">
        <el-radio value="3">指定年</el-radio>
        <el-date-picker
          v-model="appoint"
          type="year"
          :disabled="type !== '3'"
          placeholder="选择年"
          format="YYYY"
          value-format="YYYY"
          @change="changeAppoint"
        />
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
const cycle = ref({ start: 2024, end: 2099 })
const appoint = ref('')

const parseValue = (value) => {
  if (value === '*') {
    type.value = '1'
  } else if (value.includes('-')) {
    type.value = '2'
    const [start, end] = value.split('-')
    cycle.value = { start: parseInt(start), end: parseInt(end) }
  } else {
    type.value = '3'
    appoint.value = value
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
      changeAppoint()
      break
  }
}

const changeCycle = () => {
  emit('update:modelValue', `${cycle.value.start}-${cycle.value.end}`)
}

const changeAppoint = () => {
  emit('update:modelValue', appoint.value)
}

watch(() => props.modelValue, (newVal) => {
  parseValue(newVal)
}, { immediate: true })
</script>

<style scoped>
.crontab-year {
  padding: 10px;
}

.radio-item {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.el-input-number {
  width: 120px;
}
</style>

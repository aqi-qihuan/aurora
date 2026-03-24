<template>
  <div
    ref="editor"
    class="edit-container"
    v-html="innerText"
    :placeholder="placeholder"
    :contenteditable="disable"
    @focus="onFocus"
    @blur="onBlur"
    @input="onInput" />
</template>

<script setup>
import { ref, watch, defineProps, defineEmits } from 'vue'
import DOMPurify from 'dompurify'

const sanitizeHtml = (html) => {
  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a', 'br', 'p', 'span', 'img', 'ul', 'ol', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'blockquote', 'pre', 'code', 'del', 'u', 'sup', 'sub', 'table', 'thead', 'tbody', 'tr', 'th', 'td'],
    ALLOWED_ATTR: ['href', 'title', 'target', 'class', 'src', 'alt', 'style']
  })
}

const props = defineProps({
  value: {
    type: String,
    default: ''
  },
  disable: {
    type: Boolean,
    default: true
  },
  placeholder: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['input', 'focus', 'blur'])

const editor = ref(null)
const innerText = ref(sanitizeHtml(props.value))
const isLocked = ref(false)
const range = ref(null)

watch(() => props.value, (newValue) => {
  if (!isLocked.value) {
    innerText.value = sanitizeHtml(newValue)
  }
})

const clear = () => {
  editor.value.innerHTML = ''
  emit('input', editor.value.innerHTML)
}

const onInput = () => {
  emit('input', editor.value.innerHTML)
}

const onFocus = () => {
  emit('focus', editor.value.innerHTML)
  isLocked.value = true
}

const onBlur = () => {
  if (window.getSelection) {
    const selection = window.getSelection()
    range.value = selection.getRangeAt(0)
  }
  emit('blur', editor.value.innerHTML)
  isLocked.value = false
}

const addText = (value) => {
  if (window.getSelection) {
    const selection = window.getSelection()
    selection.removeAllRanges()
    if (range.value == null) {
      editor.value.focus()
      range.value = selection.getRangeAt(0)
    }
    range.value.deleteContents()
    range.value.insertNode(range.value.createContextualFragment(sanitizeHtml(value)))
    range.value.collapse(false)
    selection.addRange(range.value)
    emit('input', editor.value.innerHTML)
  }
}

defineExpose({
  clear,
  addText
})
</script>

<style scoped>
.edit-container {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 8px;
  background: var(--bg-surface, #f0f1f4);
  font-size: 14px;
  line-height: 1.5;
  padding: 6px 12px;
  box-sizing: border-box;
  overflow: auto;
  word-break: break-all;
  outline: none;
  user-select: text;
  white-space: pre-wrap;
  text-align: left;
  -webkit-user-modify: read-write-plaintext-only;
}
.edit-container:empty::before {
  cursor: text;
  content: attr(placeholder);
  color: #999;
}
</style>

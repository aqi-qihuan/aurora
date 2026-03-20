/**
 * Aurora Admin V3 - 组件测试示例
 * 使用 Vitest 全局变量
 */

import { mount } from '@vue/test-utils'

// 示例按钮组件
const MyButton = {
  name: 'MyButton',
  template: `
    <button 
      class="my-button" 
      :class="{ 'is-disabled': disabled }"
      :disabled="disabled"
      @click="handleClick"
    >
      <slot>{{ text }}</slot>
    </button>
  `,
  props: {
    text: {
      type: String,
      default: 'Button'
    },
    disabled: {
      type: Boolean,
      default: false
    }
  },
  emits: ['click'],
  methods: {
    handleClick(event) {
      if (!this.disabled) {
        this.$emit('click', event)
      }
    }
  }
}

// 示例计数器组件
const Counter = {
  name: 'Counter',
  template: `
    <div class="counter">
      <span class="count">{{ count }}</span>
      <button class="increment" @click="increment">+1</button>
      <button class="decrement" @click="decrement">-1</button>
    </div>
  `,
  data() {
    return {
      count: 0
    }
  },
  methods: {
    increment() {
      this.count++
    },
    decrement() {
      this.count--
    }
  }
}

// 测试套件
describe('组件测试示例', () => {
  describe('MyButton 组件', () => {
    test('应该渲染默认文本', () => {
      const wrapper = mount(MyButton)
      
      expect(wrapper.text()).toBe('Button')
      expect(wrapper.find('button').exists()).toBe(true)
    })

    test('应该渲染自定义文本', () => {
      const wrapper = mount(MyButton, {
        props: {
          text: 'Click Me'
        }
      })
      
      expect(wrapper.text()).toBe('Click Me')
    })

    test('应该渲染插槽内容', () => {
      const wrapper = mount(MyButton, {
        slots: {
          default: 'Slot Content'
        }
      })
      
      expect(wrapper.text()).toBe('Slot Content')
    })

    test('应该触发 click 事件', async () => {
      const onClick = vi.fn()
      const wrapper = mount(MyButton, {
        props: {
          onClick
        }
      })
      
      await wrapper.find('button').trigger('click')
      
      expect(onClick).toHaveBeenCalled()
    })

    test('禁用状态下不应该触发 click 事件', async () => {
      const onClick = vi.fn()
      const wrapper = mount(MyButton, {
        props: {
          disabled: true,
          onClick
        }
      })
      
      await wrapper.find('button').trigger('click')
      
      expect(onClick).not.toHaveBeenCalled()
      expect(wrapper.find('button').attributes('disabled')).toBeDefined()
    })
  })

  describe('Counter 组件', () => {
    test('应该渲染初始计数', () => {
      const wrapper = mount(Counter)
      
      expect(wrapper.find('.count').text()).toBe('0')
    })

    test('点击 +1 应该增加计数', async () => {
      const wrapper = mount(Counter)
      
      await wrapper.find('.increment').trigger('click')
      
      expect(wrapper.find('.count').text()).toBe('1')
    })

    test('点击 -1 应该减少计数', async () => {
      const wrapper = mount(Counter)
      
      await wrapper.find('.decrement').trigger('click')
      
      expect(wrapper.find('.count').text()).toBe('-1')
    })

    test('多次点击应该正确计算', async () => {
      const wrapper = mount(Counter)
      
      await wrapper.find('.increment').trigger('click')
      await wrapper.find('.increment').trigger('click')
      await wrapper.find('.decrement').trigger('click')
      
      expect(wrapper.find('.count').text()).toBe('1')
    })
  })
})

/**
 * Aurora Admin V3 单元测试框架配置完成
 * 使用 Vitest 全局变量（globals: true）
 */

import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

describe('Vitest 测试环境验证', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  test('测试环境正常工作', () => {
    expect(true).toBe(true)
  })

  test('可以断言基本类型', () => {
    const num = 42
    const str = 'Hello Aurora'
    const arr = [1, 2, 3]
    const obj = { name: 'aurora' }

    expect(num).toBe(42)
    expect(str).toBe('Hello Aurora')
    expect(arr).toHaveLength(3)
    expect(arr).toContain(2)
    expect(obj).toHaveProperty('name')
    expect(obj.name).toBe('aurora')
  })

  test('可以使用 mock 函数', () => {
    const mockFn = vi.fn()
    mockFn('hello')
    
    expect(mockFn).toHaveBeenCalled()
    expect(mockFn).toHaveBeenCalledWith('hello')
    expect(mockFn).toHaveBeenCalledTimes(1)
  })

  test('可以模拟返回值', () => {
    const mockFn = vi.fn(() => 'mocked value')
    const result = mockFn()
    
    expect(result).toBe('mocked value')
  })

  test('可以测试异步函数', async () => {
    const asyncFn = vi.fn().mockResolvedValue('async result')
    const result = await asyncFn()
    
    expect(result).toBe('async result')
  })

  test('可以挂载 Vue 组件', () => {
    const TestComponent = {
      template: '<div class="test">Test Component</div>'
    }
    
    const wrapper = mount(TestComponent)
    
    expect(wrapper.html()).toContain('Test Component')
    expect(wrapper.find('.test').exists()).toBe(true)
  })

  test('可以测试 Vue 响应式数据', async () => {
    const TestComponent = {
      template: '<div>{{ message }}</div>',
      data() {
        return {
          message: 'Hello'
        }
      }
    }
    
    const wrapper = mount(TestComponent)
    
    expect(wrapper.text()).toBe('Hello')
    
    await wrapper.setData({ message: 'World' })
    
    expect(wrapper.text()).toBe('World')
  })
})

describe('工具函数测试示例', () => {
  test('日期格式化函数', () => {
    const formatDate = (date) => {
      const d = new Date(date)
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    }
    
    const result = formatDate('2026-03-19')
    expect(result).toBe('2026-03-19')
  })

  test('数组去重函数', () => {
    const unique = (arr) => [...new Set(arr)]
    
    const result = unique([1, 2, 2, 3, 3, 3])
    expect(result).toEqual([1, 2, 3])
  })

  test('深拷贝函数', () => {
    const deepClone = (obj) => JSON.parse(JSON.stringify(obj))
    
    const original = { a: 1, b: { c: 2 } }
    const cloned = deepClone(original)
    
    expect(cloned).toEqual(original)
    expect(cloned).not.toBe(original)
    expect(cloned.b).not.toBe(original.b)
  })
})

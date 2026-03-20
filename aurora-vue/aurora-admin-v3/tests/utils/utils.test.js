/**
 * Aurora Admin V3 - 工具函数测试示例
 * 使用 Vitest 全局变量
 */

/**
 * 验证邮箱格式
 */
export const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

/**
 * 验证手机号（中国大陆）
 */
export const isValidPhone = (phone) => {
  const phoneRegex = /^1[3-9]\d{9}$/
  return phoneRegex.test(phone)
}

/**
 * 格式化文件大小
 */
export const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 深度合并对象
 */
export const deepMerge = (target, source) => {
  const result = { ...target }
  for (const key in source) {
    if (source[key] instanceof Object && key in target) {
      result[key] = deepMerge(target[key], source[key])
    } else {
      result[key] = source[key]
    }
  }
  return result
}

// 测试套件
describe('工具函数测试', () => {
  describe('邮箱验证', () => {
    test('应该验证正确的邮箱格式', () => {
      expect(isValidEmail('test@example.com')).toBe(true)
      expect(isValidEmail('user.name@example.co.uk')).toBe(true)
      expect(isValidEmail('test+tag@example.com')).toBe(true)
    })

    test('应该拒绝错误的邮箱格式', () => {
      expect(isValidEmail('invalid-email')).toBe(false)
      expect(isValidEmail('test@')).toBe(false)
      expect(isValidEmail('@example.com')).toBe(false)
      expect(isValidEmail('test@example')).toBe(false)
    })
  })

  describe('手机号验证', () => {
    test('应该验证正确的手机号', () => {
      expect(isValidPhone('13800138000')).toBe(true)
      expect(isValidPhone('15012345678')).toBe(true)
      expect(isValidPhone('19900000000')).toBe(true)
    })

    test('应该拒绝错误的手机号', () => {
      expect(isValidPhone('12800138000')).toBe(false)
      expect(isValidPhone('1380013800')).toBe(false)
      expect(isValidPhone('138001380000')).toBe(false)
      expect(isValidPhone('abc12345678')).toBe(false)
    })
  })

  describe('文件大小格式化', () => {
    test('应该正确格式化字节', () => {
      expect(formatFileSize(0)).toBe('0 B')
      expect(formatFileSize(1024)).toBe('1 KB')
      expect(formatFileSize(1048576)).toBe('1 MB')
      expect(formatFileSize(1073741824)).toBe('1 GB')
      expect(formatFileSize(1536)).toBe('1.5 KB')
    })

    test('应该处理大文件', () => {
      expect(formatFileSize(1099511627776)).toBe('1 TB')
    })
  })

  describe('深度合并对象', () => {
    test('应该合并简单对象', () => {
      const target = { a: 1, b: 2 }
      const source = { b: 3, c: 4 }
      const result = deepMerge(target, source)
      
      expect(result).toEqual({ a: 1, b: 3, c: 4 })
    })

    test('应该深度合并嵌套对象', () => {
      const target = { a: { b: 1, c: 2 } }
      const source = { a: { c: 3, d: 4 } }
      const result = deepMerge(target, source)
      
      expect(result).toEqual({ a: { b: 1, c: 3, d: 4 } })
    })

    test('不应该修改原始对象', () => {
      const target = { a: 1 }
      const source = { b: 2 }
      deepMerge(target, source)
      
      expect(target).toEqual({ a: 1 })
    })
  })
})

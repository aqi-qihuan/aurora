// Vitest 全局变量测试
// globals: true 已在 vitest.config.js 中启用

test('基础测试 - 数学运算', () => {
  expect(1 + 1).toBe(2)
})

test('基础测试 - 字符串', () => {
  expect('hello').toBe('hello')
})

test('基础测试 - 数组', () => {
  const arr = [1, 2, 3]
  expect(arr).toHaveLength(3)
  expect(arr).toContain(2)
})

test('基础测试 - 对象', () => {
  const obj = { name: 'aurora' }
  expect(obj).toHaveProperty('name')
  expect(obj.name).toBe('aurora')
})

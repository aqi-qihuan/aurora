/**
 * Vitest 测试环境设置文件
 */

// 全局模拟 localStorage
if (typeof global.localStorage === 'undefined') {
  global.localStorage = {
    getItem: () => null,
    setItem: () => {},
    removeItem: () => {},
    clear: () => {},
    length: 0,
    key: () => null
  }
}

// 全局模拟 sessionStorage
if (typeof global.sessionStorage === 'undefined') {
  global.sessionStorage = {
    getItem: () => null,
    setItem: () => {},
    removeItem: () => {},
    clear: () => {},
    length: 0,
    key: () => null
  }
}

console.log('✅ Vitest 测试环境设置完成')

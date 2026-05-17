/**
 * Aurora Admin V3 - Store 测试示例
 * 使用 Vitest 全局变量
 */

import { setActivePinia, createPinia, defineStore } from 'pinia'

// 示例用户 Store
const useUserStore = defineStore('user', {
  state: () => ({
    userInfo: null,
    token: null,
    isLoggedIn: false
  }),
  
  getters: {
    userName: (state) => state.userInfo?.username || 'Guest',
    isAdmin: (state) => state.userInfo?.role === 'admin'
  },
  
  actions: {
    login(userInfo, token) {
      this.userInfo = userInfo
      this.token = token
      this.isLoggedIn = true
    },
    
    logout() {
      this.userInfo = null
      this.token = null
      this.isLoggedIn = false
    },
    
    updateUserInfo(newInfo) {
      if (this.userInfo) {
        this.userInfo = { ...this.userInfo, ...newInfo }
      }
    }
  }
})

// 示例计数器 Store
const useCounterStore = defineStore('counter', {
  state: () => ({
    count: 0
  }),
  
  getters: {
    doubleCount: (state) => state.count * 2,
    isPositive: (state) => state.count > 0
  },
  
  actions: {
    increment() {
      this.count++
    },
    
    decrement() {
      this.count--
    },
    
    reset() {
      this.count = 0
    }
  }
})

// 测试套件
describe('Store 测试示例', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('用户 Store', () => {
    test('应该有正确的初始状态', () => {
      const store = useUserStore()
      
      expect(store.userInfo).toBeNull()
      expect(store.token).toBeNull()
      expect(store.isLoggedIn).toBe(false)
    })

    test('应该正确登录', () => {
      const store = useUserStore()
      const mockUser = { id: 1, username: 'admin', role: 'admin' }
      const mockToken = 'test-token-123'
      
      store.login(mockUser, mockToken)
      
      expect(store.userInfo).toEqual(mockUser)
      expect(store.token).toBe(mockToken)
      expect(store.isLoggedIn).toBe(true)
    })

    test('应该正确登出', () => {
      const store = useUserStore()
      
      store.login({ id: 1, username: 'admin' }, 'token')
      store.logout()
      
      expect(store.userInfo).toBeNull()
      expect(store.token).toBeNull()
      expect(store.isLoggedIn).toBe(false)
    })

    test('应该正确获取用户名', () => {
      const store = useUserStore()
      
      expect(store.userName).toBe('Guest')
      
      store.login({ id: 1, username: 'testuser' }, 'token')
      expect(store.userName).toBe('testuser')
    })

    test('应该正确判断管理员权限', () => {
      const store = useUserStore()
      
      store.login({ id: 1, username: 'user', role: 'user' }, 'token')
      expect(store.isAdmin).toBe(false)
      
      store.login({ id: 2, username: 'admin', role: 'admin' }, 'token')
      expect(store.isAdmin).toBe(true)
    })

    test('应该正确更新用户信息', () => {
      const store = useUserStore()
      
      store.login({ id: 1, username: 'user', email: 'old@example.com' }, 'token')
      store.updateUserInfo({ email: 'new@example.com', nickname: 'New Name' })
      
      expect(store.userInfo.email).toBe('new@example.com')
      expect(store.userInfo.nickname).toBe('New Name')
      expect(store.userInfo.username).toBe('user')
    })
  })

  describe('计数器 Store', () => {
    test('应该有正确的初始状态', () => {
      const store = useCounterStore()
      
      expect(store.count).toBe(0)
      expect(store.doubleCount).toBe(0)
      expect(store.isPositive).toBe(false)
    })

    test('increment 应该增加计数', () => {
      const store = useCounterStore()
      
      store.increment()
      
      expect(store.count).toBe(1)
      expect(store.doubleCount).toBe(2)
      expect(store.isPositive).toBe(true)
    })

    test('decrement 应该减少计数', () => {
      const store = useCounterStore()
      
      store.decrement()
      
      expect(store.count).toBe(-1)
      expect(store.doubleCount).toBe(-2)
      expect(store.isPositive).toBe(false)
    })

    test('reset 应该重置计数', () => {
      const store = useCounterStore()
      
      store.increment()
      store.increment()
      store.reset()
      
      expect(store.count).toBe(0)
    })

    test('getters 应该是响应式的', () => {
      const store = useCounterStore()
      
      expect(store.doubleCount).toBe(0)
      
      store.count = 5
      
      expect(store.doubleCount).toBe(10)
    })
  })
})

/**
 * Aurora Admin V3 - API 测试示例
 * 使用 Vitest 全局变量
 */

import axios from 'axios'

// 模拟 axios
vi.mock('axios')

// API 请求函数
const api = {
  async getUser(id) {
    const response = await axios.get(`/api/users/${id}`)
    return response.data
  },

  async createUser(userData) {
    const response = await axios.post('/api/users', userData)
    return response.data
  },

  async updateUser(id, userData) {
    const response = await axios.put(`/api/users/${id}`, userData)
    return response.data
  },

  async deleteUser(id) {
    const response = await axios.delete(`/api/users/${id}`)
    return response.data
  },

  async getList(params) {
    const response = await axios.get('/api/list', { params })
    return response.data
  }
}

// 数据转换函数
const transformUserData = (data) => ({
  id: data.id,
  username: data.username,
  email: data.email,
  createdAt: new Date(data.created_at).toISOString()
})

const transformListData = (response) => ({
  list: response.data.list,
  total: response.data.count,
  page: response.data.pageNum,
  pageSize: response.data.pageSize
})

// 测试套件
describe('API 测试示例', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getUser API', () => {
    test('应该成功获取用户信息', async () => {
      const mockUser = { id: 1, username: 'admin', email: 'admin@example.com' }
      axios.get.mockResolvedValue({ data: mockUser })
      
      const result = await api.getUser(1)
      
      expect(axios.get).toHaveBeenCalledWith('/api/users/1')
      expect(result).toEqual(mockUser)
    })

    test('应该处理用户不存在的情况', async () => {
      axios.get.mockRejectedValue(new Error('User not found'))
      
      await expect(api.getUser(999)).rejects.toThrow('User not found')
    })
  })

  describe('createUser API', () => {
    test('应该成功创建用户', async () => {
      const newUser = { username: 'test', email: 'test@example.com' }
      const createdUser = { id: 1, ...newUser }
      
      axios.post.mockResolvedValue({ data: createdUser })
      
      const result = await api.createUser(newUser)
      
      expect(axios.post).toHaveBeenCalledWith('/api/users', newUser)
      expect(result).toEqual(createdUser)
    })
  })

  describe('updateUser API', () => {
    test('应该成功更新用户', async () => {
      const updateData = { email: 'newemail@example.com' }
      const updatedUser = { id: 1, username: 'admin', ...updateData }
      
      axios.put.mockResolvedValue({ data: updatedUser })
      
      const result = await api.updateUser(1, updateData)
      
      expect(axios.put).toHaveBeenCalledWith('/api/users/1', updateData)
      expect(result).toEqual(updatedUser)
    })
  })

  describe('deleteUser API', () => {
    test('应该成功删除用户', async () => {
      axios.delete.mockResolvedValue({ data: { success: true } })
      
      const result = await api.deleteUser(1)
      
      expect(axios.delete).toHaveBeenCalledWith('/api/users/1')
      expect(result).toEqual({ success: true })
    })
  })

  describe('getList API', () => {
    test('应该成功获取列表数据', async () => {
      const mockResponse = {
        data: {
          list: [{ id: 1 }, { id: 2 }],
          count: 2,
          pageNum: 1,
          pageSize: 10
        }
      }
      
      axios.get.mockResolvedValue(mockResponse)
      
      const result = await api.getList({ page: 1, size: 10 })
      
      expect(axios.get).toHaveBeenCalledWith('/api/list', {
        params: { page: 1, size: 10 }
      })
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('数据转换函数', () => {
    test('应该正确转换用户数据', () => {
      const input = {
        id: 1,
        username: 'admin',
        email: 'admin@example.com',
        created_at: '2026-03-19T10:00:00Z'
      }
      
      const result = transformUserData(input)
      
      expect(result).toHaveProperty('id', 1)
      expect(result).toHaveProperty('username', 'admin')
      expect(result).toHaveProperty('email', 'admin@example.com')
      expect(result).toHaveProperty('createdAt')
      expect(new Date(result.createdAt)).toBeInstanceOf(Date)
    })

    test('应该正确转换列表数据', () => {
      const input = {
        data: {
          list: [{ id: 1 }, { id: 2 }],
          count: 100,
          pageNum: 1,
          pageSize: 10
        }
      }
      
      const result = transformListData(input)
      
      expect(result).toEqual({
        list: [{ id: 1 }, { id: 2 }],
        total: 100,
        page: 1,
        pageSize: 10
      })
    })
  })
})

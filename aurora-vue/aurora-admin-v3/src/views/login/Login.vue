<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-title">管理员登录</div>
      <el-form
        status-icon
        :model="loginForm"
        :rules="rules"
        ref="ruleFormRef"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><UserFilled /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            show-password
            placeholder="密码"
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <el-button type="primary" @click="handleLogin" :loading="loading">登录</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { UserFilled, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import request from '@/utils/request'
import logger from '@/utils/logger'

const router = useRouter()
const userStore = useUserStore()

// 响应式数据
const ruleFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '用户名不能为空', trigger: 'blur' }],
  password: [{ required: true, message: '密码不能为空', trigger: 'blur' }]
}

// 登录方法
const handleLogin = async () => {
  if (!ruleFormRef.value) return
  
  await ruleFormRef.value.validate((valid) => {
    if (valid) {
      loading.value = true
      
      const params = new URLSearchParams()
      params.append('username', loginForm.username)
      params.append('password', loginForm.password)
      
      request.post('/users/login', params)
        .then(({ data }) => {
          logger.log('登录响应:', data)
          
          if (data.flag) {
            // 保存登录信息
            userStore.login(data.data)
            
            // 如果后端返回了菜单,保存菜单
            if (data.data.menus && data.data.menus.length > 0) {
              userStore.saveUserMenus(data.data.menus)
            }
            
            ElMessage.success('登录成功')
            
            // 跳转到首页,让路由守卫处理菜单加载
            router.push({ path: '/' })
          }
        })
        .catch((error) => {
          logger.error('登录失败:', error)
        })
        .finally(() => {
          loading.value = false
        })
    }
  })
}
</script>

<style scoped>
.login-container {
  position: absolute;
  top: 0;
  bottom: 0;
  right: 0;
  left: 0;
  background: url(https://ws.aqi125.cn/aurora/photos/525b7ec22916c978a5e06915cf6afbba.jpg) center center / cover
    no-repeat;
}

.login-card {
  position: absolute;
  top: 0;
  bottom: 0;
  right: 0;
  background: #fff;
  padding: 170px 60px 180px;
  width: 350px;
  overflow-y: auto;
}

.login-title {
  color: #303133;
  font-weight: bold;
  font-size: 1rem;
}

.login-form {
  margin-top: 1.2rem;
}

.login-card button {
  margin-top: 1rem;
  width: 100%;
}
</style>

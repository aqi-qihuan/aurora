<template>
  <div class="error-container">
    <div class="error-content">
      <div class="error-code">403</div>
      <div class="error-desc">
        <h2>无访问权限</h2>
        <p>抱歉,您没有权限访问此页面</p>
      </div>
      <div class="error-actions">
        <el-button type="primary" @click="goHome">
          <el-icon><HomeFilled /></el-icon>
          返回首页
        </el-button>
        <el-button @click="goBack">
          <el-icon><Back /></el-icon>
          返回上页
        </el-button>
      </div>
    </div>
    <div class="error-image">
      <el-icon :size="200" color="#F56C6C">
        <Lock />
      </el-icon>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { HomeFilled, Back, Lock } from '@element-plus/icons-vue'

const router = useRouter()

const goHome = () => {
  router.push('/home')
}

const goBack = () => {
  router.go(-1)
}
</script>

<style scoped>
.error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
  padding: 20px;
}

.error-content {
  text-align: center;
  color: white;
  margin-right: 60px;
}

.error-code {
  font-size: 150px;
  font-weight: bold;
  line-height: 1;
  margin-bottom: 20px;
  background: linear-gradient(to right, #fff 0%, #e3e3e3 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.error-desc h2 {
  font-size: 32px;
  margin: 0 0 15px 0;
}

.error-desc p {
  font-size: 16px;
  margin: 0 0 30px 0;
  opacity: 0.9;
}

.error-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
}

.error-image {
  opacity: 0.3;
}

@media (max-width: 768px) {
  .error-container {
    flex-direction: column;
  }

  .error-content {
    margin-right: 0;
    margin-bottom: 30px;
  }

  .error-code {
    font-size: 100px;
  }

  .error-desc h2 {
    font-size: 24px;
  }

  .error-image {
    display: none;
  }
}
</style>

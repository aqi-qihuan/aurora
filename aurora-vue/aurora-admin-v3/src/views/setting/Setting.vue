<template>
  <el-card class="main-card">
    <el-tabs v-model="activeName" @tab-click="handleTabClick">
      <el-tab-pane label="修改信息" name="info">
        <div class="info-container">
          <el-upload
            class="avatar-uploader"
            action="/api/users/avatar"
            :show-file-list="false"
            :headers="headers"
            :on-success="updateAvatar">
            <el-avatar v-if="avatar" :src="avatar" :size="120" />
            <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
          </el-upload>
          <el-form label-width="70px" :model="infoForm" style="width: 320px; margin-left: 3rem">
            <el-form-item label="昵称">
              <el-input v-model="infoForm.nickname" />
            </el-form-item>
            <el-form-item label="个人简介">
              <el-input v-model="infoForm.intro" />
            </el-form-item>
            <el-form-item label="个人网站">
              <el-input v-model="infoForm.website" />
            </el-form-item>
            <el-button @click="updateInfo" type="primary" style="margin-left: 4.375rem"> 修改 </el-button>
          </el-form>
        </div>
      </el-tab-pane>
      <el-tab-pane label="修改密码" name="password">
        <el-form label-width="70px" :model="passwordForm" style="width: 320px">
          <el-form-item label="旧密码">
            <el-input
              @keyup.enter="updatePassword"
              v-model="passwordForm.oldPassword"
              show-password />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input
              @keyup.enter="updatePassword"
              v-model="passwordForm.newPassword"
              show-password />
          </el-form-item>
          <el-form-item label="确认密码">
            <el-input
              @keyup.enter="updatePassword"
              v-model="passwordForm.confirmPassword"
              show-password />
          </el-form-item>
          <el-button type="primary" style="margin-left: 4.4rem" @click="updatePassword"> 修改 </el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </el-card>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import request from '@/utils/request'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const infoForm = reactive({
  nickname: userStore.userInfo?.nickname || '',
  intro: userStore.userInfo?.intro || '',
  website: userStore.userInfo?.website || ''
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const activeName = ref('info')

const headers = computed(() => ({ 
  Authorization: 'Bearer ' + sessionStorage.getItem('token') 
}))

const avatar = computed(() => userStore.userInfo?.avatar)

const handleTabClick = (tab) => {
  // Tab 切换处理
}

const updateAvatar = (response) => {
  if (response.flag) {
    ElMessage.success(response.message)
    userStore.updateAvatar(response.data)
  } else {
    ElMessage.error(response.message)
  }
}

const updateInfo = () => {
  if (!infoForm.nickname?.trim()) {
    ElMessage.error('昵称不能为空')
    return false
  }
  request.put('/users/info', infoForm).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: '修改成功'
      })
      userStore.updateUserInfo(infoForm)
    } else {
      ElNotification.error({
        title: '失败',
        message: '修改失败'
      })
    }
  })
}

const updatePassword = () => {
  if (!passwordForm.oldPassword?.trim()) {
    ElMessage.error('旧密码不能为空')
    return false
  }
  if (!passwordForm.newPassword?.trim()) {
    ElMessage.error('新密码不能为空')
    return false
  }
  if (passwordForm.newPassword.length < 6) {
    ElMessage.error('新密码不能少于6位')
    return false
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.error('两次密码输入不一致')
    return false
  }
  request.put('/admin/users/password', passwordForm).then(({ data }) => {
    if (data.flag) {
      passwordForm.oldPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
      ElNotification.success({
        title: '成功',
        message: '修改成功'
      })
    } else {
      ElNotification.error({
        title: '失败',
        message: '修改失败'
      })
    }
  })
}
</script>

<style scoped>
.avatar-uploader {
  display: inline-block;
}
.avatar-uploader :deep(.el-upload) {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
}
.avatar-uploader :deep(.el-upload:hover) {
  border-color: #409eff;
}
.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 120px;
  height: 120px;
  line-height: 120px;
  text-align: center;
}
.info-container {
  display: flex;
  align-items: center;
  margin-left: 20%;
  margin-top: 5rem;
}
</style>

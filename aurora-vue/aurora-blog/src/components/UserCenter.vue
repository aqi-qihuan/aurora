<template>
  <el-drawer
    v-model="visible"
    direction="rtl"
    :with-header="false"
    :before-close="handleClose"
    class="user-center-drawer"
    size="380px">
    <div class="drawer-body">
      <!-- Header -->
      <div class="drawer-header">
        <svg-icon icon-class="people" class="drawer-icon" />
        <h3 class="drawer-title">用户中心</h3>
        <p class="drawer-subtitle">你的信息将被严格保密</p>
      </div>

      <!-- Content -->
      <template v-if="userInfo !== ''">
        <div class="profile-section">
          <!-- Avatar -->
          <button id="pick-avatar" class="avatar-wrapper" @click="showCropper = true">
            <el-avatar :size="88" :src="userInfo.avatar" />
            <span class="avatar-overlay">
              <svg-icon icon-class="text-outline" />
            </span>
          </button>
          <avatar-cropper
            v-model="showCropper"
            @uploaded="handleSuccess"
            trigger="#pick-avatar"
            :request-options="options"
            upload-url="/api/users/avatar" />

          <!-- Form -->
          <el-form class="user-form">
            <el-form-item label="昵称">
              <el-input v-model="userInfo.nickname" placeholder="输入你的昵称" />
            </el-form-item>
            <el-form-item label="网址">
              <el-input v-model="userInfo.website" placeholder="https://example.com" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input v-model="userInfo.intro" placeholder="介绍一下自己吧" />
            </el-form-item>
            <el-form-item label="邮箱">
              <el-input disabled :placeholder="userInfo.email || '未绑定邮箱'">
                <template #append>
                  <span class="email-action" @click="changeEmailDialogVisible">
                    {{ userInfo.email ? '修改' : '绑定' }}
                  </span>
                </template>
              </el-input>
            </el-form-item>

            <!-- Subscribe toggle -->
            <div class="subscribe-row">
              <div class="subscribe-info">
                <span class="subscribe-label">邮件订阅</span>
                <span class="subscribe-desc">接收新文章通知</span>
              </div>
              <el-switch
                v-model="userInfo.isSubscribe"
                :loading="loading"
                :before-change="beforeChange"
                @change="changeSubscribe"
                :active-value="1"
                :inactive-value="0" />
            </div>

            <!-- Submit -->
            <div class="form-actions">
              <button type="button" class="btn-save" @click="commit">保存修改</button>
            </div>
          </el-form>
        </div>
      </template>
    </div>
  </el-drawer>

  <!-- Email Binding Dialog -->
  <el-dialog
    v-model="emailDialogVisible"
    width="380px"
    :fullscreen="false"
    class="email-dialog">
    <div class="dialog-header">
      <svg-icon icon-class="notice" class="dialog-icon" />
      <h3 class="dialog-title">{{ userInfo.email ? '修改邮箱' : '绑定邮箱' }}</h3>
      <p class="dialog-subtitle">输入邮箱并完成验证</p>
    </div>
    <el-form class="email-form">
      <el-form-item>
        <el-input v-model="email" placeholder="邮箱地址" />
      </el-form-item>
      <el-form-item>
        <el-input v-model="VerificationCode" placeholder="验证码">
          <template #append>
            <button type="button" class="code-btn" :disabled="codeCooldown > 0" @click="sendCode">
              {{ codeCooldown > 0 ? `${codeCooldown}s` : '发送' }}
            </button>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item>
        <button type="button" class="btn-save" @click="bingingEmail">确认绑定</button>
      </el-form-item>
    </el-form>
  </el-dialog>
</template>

<script lang="ts">
import { defineComponent, toRef, ref, reactive, toRefs, getCurrentInstance, computed, onMounted, onUnmounted } from 'vue'
import { useUserStore } from '@/stores/user'
import AvatarCropper from 'vue-avatar-cropper'
import api from '@/api/api'

export default defineComponent({
  name: 'UserCenter',
  components: { AvatarCropper },
  setup() {
    const proxy: any = getCurrentInstance()?.appContext.config.globalProperties
    const userStore = useUserStore()
    const reactiveData = reactive({
      emailDialogVisible: false,
      email: '' as any,
      VerificationCode: '' as any,
      loading: false,
      switchState: false,
      codeCooldown: 0
    })
    let showCropper = ref(false)
    let cooldownTimer: ReturnType<typeof setInterval> | null = null

    const handleClose = () => {
      userStore.userVisible = false
    }
    const changeEmailDialogVisible = () => {
      reactiveData.emailDialogVisible = true
    }
    const bingingEmail = () => {
      let params = {
        email: reactiveData.email,
        code: reactiveData.VerificationCode
      }
      api.bindingEmail(params).then(({ data }) => {
        if (data.flag) {
          proxy.$notify({ title: 'Success', message: '绑定成功', type: 'success' })
          userStore.userInfo.email = reactiveData.email
          reactiveData.emailDialogVisible = false
        }
      }).catch(() => {
        proxy.$notify({ title: 'Error', message: '绑定失败，请重试', type: 'error' })
      })
    }
    const handleSuccess = (data: any) => {
      data.response.json().then((data: any) => {
        if (data.flag) {
          userStore.userInfo.avatar = data.data
          proxy.$notify({ title: 'Success', message: '上传成功', type: 'success' })
        }
      }).catch(() => {
        proxy.$notify({ title: 'Error', message: '头像上传失败', type: 'error' })
      })
    }
    const changeSubscribe = () => {
      if (reactiveData.switchState) {
        let params = {
          userId: userStore.userInfo.userInfoId,
          isSubscribe: userStore.userInfo.isSubscribe
        }
        api.updateUserSubscribe(params).then(({ data }) => {
          if (data.flag) {
            proxy.$notify({ title: 'Success', message: '修改成功', type: 'success' })
          }
        }).catch(() => {
          proxy.$notify({ title: 'Error', message: '修改订阅状态失败', type: 'error' })
        })
      }
    }
    const commit = () => {
      let params = {
        nickname: userStore.userInfo.nickname,
        website: userStore.userInfo.website,
        intro: userStore.userInfo.intro
      }
      api.submitUserInfo(params).then(({ data }) => {
        if (data.flag) {
          proxy.$notify({ title: 'Success', message: '修改成功', type: 'success' })
        }
      }).catch(() => {
        proxy.$notify({ title: 'Error', message: '保存失败，请重试', type: 'error' })
      })
    }
    const sendCode = () => {
      if (reactiveData.codeCooldown > 0) return
      api.sendValidationCode(reactiveData.email).then(({ data }) => {
        if (data.flag) {
          proxy.$notify({ title: 'Success', message: '验证码已发送', type: 'success' })
          reactiveData.codeCooldown = 60
          cooldownTimer = setInterval(() => {
            reactiveData.codeCooldown--
            if (reactiveData.codeCooldown <= 0 && cooldownTimer) {
              clearInterval(cooldownTimer)
              cooldownTimer = null
            }
          }, 1000)
        }
      }).catch(() => {
        proxy.$notify({ title: 'Error', message: '发送失败，请重试', type: 'error' })
      })
    }
    const beforeChange = () => {
      reactiveData.switchState = true
      reactiveData.loading = true
      return new Promise((resolve, reject) => {
        if (userStore.userInfo.email === '' || userStore.userInfo.email === null) {
          reactiveData.loading = false
          proxy.$notify({ title: 'Warning', message: '邮箱未绑定,尽快绑定哦', type: 'warning' })
          return reject(new Error('Error'))
        } else {
          reactiveData.loading = false
          return resolve(true)
        }
      })
    }

    onUnmounted(() => {
      if (cooldownTimer) {
        clearInterval(cooldownTimer)
        cooldownTimer = null
      }
    })

    return {
      userInfo: toRef(userStore.$state, 'userInfo'),
      ...toRefs(reactiveData),
      visible: toRef(userStore.$state, 'userVisible'),
      showCropper,
      handleClose,
      bingingEmail,
      changeEmailDialogVisible,
      changeSubscribe,
      handleSuccess,
      sendCode,
      commit,
      beforeChange,
      options: computed(() => {
        return {
          method: 'POST',
          headers: {
            Authorization: 'Bearer ' + userStore.token
          }
        }
      })
    }
  }
})
</script>
<style lang="scss" scoped>
// Drawer layout
.drawer-body {
  padding: 2rem 1.75rem;
  display: flex;
  flex-direction: column;
  height: 100%;
}

// Drawer header
.drawer-header {
  text-align: center;
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid rgba(var(--text-dim-rgb, 148, 163, 184), 0.12);
}

.drawer-icon {
  width: 36px;
  height: 36px;
  color: var(--text-accent);
  opacity: 0.7;
}

.drawer-title {
  font-size: 1.35rem;
  font-weight: 600;
  color: var(--text-bright);
  margin: 0.75rem 0 0.25rem;
}

.drawer-subtitle {
  font-size: 0.8rem;
  color: var(--text-dim);
  margin: 0;
}

// Profile section
.profile-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  overflow-y: auto;
}

// Avatar
.avatar-wrapper {
  position: relative;
  display: inline-flex;
  border-radius: 50%;
  cursor: pointer;
  outline: none;
  background: none;
  border: 3px solid rgba(var(--text-accent-rgb, 100, 149, 237), 0.25);
  padding: 4px;
  transition: border-color 0.25s ease, transform 0.25s cubic-bezier(0.22, 1, 0.36, 1);

  &:hover {
    border-color: var(--text-accent);
    transform: scale(1.04);
  }

  &:active {
    transform: scale(0.97);
  }

  :deep(.el-avatar) {
    display: block;
  }
}

.avatar-overlay {
  position: absolute;
  inset: 3px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  opacity: 0;
  transition: opacity 0.2s ease;
  color: #fff;

  svg {
    width: 22px;
    height: 22px;
  }
}

.avatar-wrapper:hover .avatar-overlay {
  opacity: 1;
}

// Form
.user-form {
  width: 100%;
  margin-top: 2rem;

  :deep(.el-form-item) {
    margin-bottom: 1.25rem;
  }

  :deep(.el-form-item__label) {
    color: var(--text-normal);
    font-size: 0.85rem;
    font-weight: 500;
    width: 56px !important;
    text-align: left;
    padding-left: 0;
  }

  :deep(.el-input__wrapper) {
    background: var(--background-primary-alt) !important;
    border-radius: 0.5rem !important;
    box-shadow: none !important;
    border: 1px solid transparent !important;
    transition: border-color 0.25s ease, box-shadow 0.25s ease !important;
  }

  :deep(.el-input__wrapper:hover) {
    border-color: var(--text-dim) !important;
  }

  :deep(.el-input__wrapper.is-focus) {
    border-color: var(--text-accent) !important;
    box-shadow: 0 0 0 2px rgba(var(--text-accent-rgb, 100, 149, 237), 0.15) !important;
  }

  :deep(.el-input__inner) {
    color: var(--text-normal) !important;
    font-size: 0.9rem;
  }

  :deep(.el-input-group__append) {
    background: var(--background-primary-alt) !important;
    border: none;
    box-shadow: none;
  }
}

// Email action button
.email-action {
  color: var(--text-accent);
  cursor: pointer;
  font-size: 0.82rem;
  font-weight: 500;
  transition: opacity 0.2s ease;

  &:hover {
    opacity: 0.7;
  }
}

// Subscribe row
.subscribe-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.85rem 1rem;
  margin: 0.5rem 0 0.75rem;
  border-radius: 0.625rem;
  background: var(--background-primary-alt);
  border: 1px solid rgba(var(--text-dim-rgb, 148, 163, 184), 0.08);
  transition: border-color 0.25s ease;

  &:hover {
    border-color: rgba(var(--text-dim-rgb, 148, 163, 184), 0.18);
  }
}

.subscribe-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.subscribe-label {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-normal);
}

.subscribe-desc {
  font-size: 0.75rem;
  color: var(--text-dim);
}

// Save button
.form-actions {
  margin-top: 1.5rem;
  display: flex;
  justify-content: flex-end;
}

.btn-save {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.6rem 1.75rem;
  border-radius: 0.5rem;
  font-weight: 500;
  font-size: 0.9rem;
  border: none;
  outline: none;
  cursor: pointer;
  background: var(--main-gradient);
  color: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.25s ease;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  }

  &:active {
    transform: translateY(0) scale(0.98);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  }
}

// Code button
.code-btn {
  background: none;
  border: none;
  outline: none;
  cursor: pointer;
  color: var(--text-accent);
  font-size: 0.82rem;
  font-weight: 500;
  padding: 0 0.5rem;
  transition: opacity 0.2s ease;
  white-space: nowrap;

  &:hover:not(:disabled) {
    opacity: 0.7;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.4;
  }
}

// Email dialog
.email-dialog {
  :deep(.el-dialog) {
    border-radius: 1rem !important;
    overflow: hidden;
  }

  :deep(.el-dialog__header) {
    display: none;
  }

  :deep(.el-dialog__body) {
    padding: 2rem 2.25rem !important;
  }
}

.dialog-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.dialog-icon {
  width: 36px;
  height: 36px;
  color: var(--text-accent);
  opacity: 0.7;
}

.dialog-title {
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-bright);
  margin: 0.5rem 0 0.2rem;
}

.dialog-subtitle {
  font-size: 0.8rem;
  color: var(--text-dim);
  margin: 0;
}

.email-form {
  :deep(.el-form-item) {
    margin-bottom: 1rem;
  }

  :deep(.el-input__wrapper) {
    background: var(--background-primary-alt) !important;
    border-radius: 0.5rem !important;
    box-shadow: none !important;
    border: 1px solid transparent !important;
    transition: border-color 0.25s ease, box-shadow 0.25s ease !important;
  }

  :deep(.el-input__wrapper:hover) {
    border-color: var(--text-dim) !important;
  }

  :deep(.el-input__wrapper.is-focus) {
    border-color: var(--text-accent) !important;
    box-shadow: 0 0 0 2px rgba(var(--text-accent-rgb, 100, 149, 237), 0.15) !important;
  }

  :deep(.el-input__inner) {
    color: var(--text-normal) !important;
    font-size: 0.9rem;
  }

  :deep(.el-input-group__append) {
    background: var(--background-primary-alt) !important;
    border: none;
    box-shadow: none;
  }
}
</style>

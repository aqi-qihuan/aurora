<template>
  <div class="header-controls absolute top-10 right-0 flex flex-row" @keydown.k="handleOpenModel" tabindex="0">
    <span class="ob-drop-shadow" data-dia="search" @click="handleOpenModel">
      <svg-icon icon-class="search" />
    </span>
    <Dropdown v-if="multiLanguage === 1" @command="handleClick">
      <span class="ob-drop-shadow" data-dia="language">
        <svg-icon icon-class="globe" />
        <span v-if="$i18n.locale == 'cn'">中文</span>
        <span v-if="$i18n.locale == 'en'">EN</span>
      </span>
      <DropdownMenu>
        <DropdownItem name="en">English</DropdownItem>
        <DropdownItem name="cn">中文</DropdownItem>
      </DropdownMenu>
    </Dropdown>
    <template v-if="userInfo === ''">
      <span class="mr-3" @click="openLoginDialog">{{ t('settings.login') }}</span>
    </template>
    <template v-if="userInfo !== ''">
      <Dropdown hover>
        <span class="mr-2">
          <div class="flex-shrink-0">
            <div class="rounded-full ring-gray-100 overflow-hidden shaodw-lg w-9">
              <img class="avatar-img" :src="userInfo.avatar" alt="" />
            </div>
          </div>
        </span>
        <DropdownMenu>
          <template v-if="!isMobile">
            <DropdownItem @click="openUserCenter">{{ t('settings.personal-center') }}</DropdownItem>
          </template>
          <DropdownItem @click="logout">{{ t('settings.logout') }}</DropdownItem>
        </DropdownMenu>
      </Dropdown>
    </template>
    <span no-hover-effect class="ob-drop-shadow" data-dia="light-switch">
      <ThemeToggle />
    </span>
  </div>
  <el-dialog v-model="loginDialogVisible" width="30%" :fullscreen="isMobile" class="auth-dialog">
    <div class="dialog-header">
      <svg-icon icon-class="author" class="dialog-icon" />
      <h3 class="dialog-title">欢迎回来</h3>
      <p class="dialog-subtitle">登录你的账号</p>
    </div>
    <el-form @keyup.enter.native="login" class="auth-form">
      <el-form-item class="mt-5">
        <el-input v-model="loginInfo.username" placeholder="邮箱" />
      </el-form-item>
      <el-form-item class="mt-4">
        <el-input v-model="loginInfo.password" type="password" show-password placeholder="密码" />
      </el-form-item>
      <el-form-item>
        <button type="button" class="auth-btn auth-btn-primary" @click="login">登录</button>
      </el-form-item>
      <div class="dialog-divider">
        <span class="divider-line"></span>
        <span class="divider-text">或</span>
        <span class="divider-line"></span>
      </div>
      <el-form-item>
        <button type="button" class="auth-btn auth-btn-qq" @click="qqLogin">
          <svg-icon icon-class="qqLogin" class="qq-icon" />
          <span>QQ 登录</span>
        </button>
      </el-form-item>
      <div class="dialog-footer-links">
        <span class="auth-link" @click="openRegisterDialog">立即注册</span>
        <span class="auth-link" @click="openForgetPasswordDialog">忘记密码?</span>
      </div>
    </el-form>
  </el-dialog>
  <el-dialog v-model="registerDialogVisible" width="30%" :fullscreen="isMobile" class="auth-dialog">
    <div class="dialog-header">
      <svg-icon icon-class="people" class="dialog-icon" />
      <h3 class="dialog-title">创建账号</h3>
      <p class="dialog-subtitle">注册一个新账号</p>
    </div>
    <el-form class="auth-form">
      <el-form-item class="mt-5">
        <el-input v-model="loginInfo.username" placeholder="邮箱" />
      </el-form-item>
      <el-form-item class="mt-4">
        <el-input v-model="loginInfo.code" placeholder="验证码">
          <template #append>
            <span class="auth-link" @click="sendCode">发送</span>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item class="mt-4">
        <el-input v-model="loginInfo.password" type="password" show-password placeholder="密码" />
      </el-form-item>
      <el-form-item>
        <button type="button" class="auth-btn auth-btn-primary" @click="register">注册</button>
      </el-form-item>
      <div class="dialog-footer-links">
        <span class="auth-link" @click="returnLoginDialog">已有帐号? 登录</span>
      </div>
    </el-form>
  </el-dialog>
  <el-dialog v-model="forgetPasswordDialogVisible" width="30%" :fullscreen="isMobile" class="auth-dialog">
    <div class="dialog-header">
      <svg-icon icon-class="shield" class="dialog-icon" />
      <h3 class="dialog-title">重置密码</h3>
      <p class="dialog-subtitle">通过邮箱验证码重置</p>
    </div>
    <el-form class="auth-form">
      <el-form-item class="mt-5">
        <el-input v-model="loginInfo.username" placeholder="邮箱" />
      </el-form-item>
      <el-form-item class="mt-4">
        <el-input v-model="loginInfo.code" placeholder="验证码">
          <template #append>
            <span class="auth-link" @click="sendCode">发送</span>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item class="mt-4">
        <el-input v-model="loginInfo.password" type="password" show-password placeholder="新密码" />
      </el-form-item>
      <el-form-item>
        <button type="button" class="auth-btn auth-btn-primary" @click="updatePassword">确定</button>
      </el-form-item>
      <div class="dialog-footer-links">
        <span class="auth-link" @click="returnLoginDialog">返回登录</span>
      </div>
    </el-form>
  </el-dialog>
  <el-dialog v-model="articlePasswordDialogVisible" width="30%" :fullscreen="isMobile" class="auth-dialog">
    <div class="dialog-header">
      <svg-icon icon-class="lock" class="dialog-icon" />
      <h3 class="dialog-title">受密码保护</h3>
      <p class="dialog-subtitle">请输入文章密码</p>
    </div>
    <el-form @submit.native.prevent @keyup.enter.native="accessArticle" class="auth-form">
      <el-form-item class="mt-5">
        <el-input id="article-password-input" v-model="articlePassword" placeholder="文章受密码保护,请输入密码" />
      </el-form-item>
      <el-form-item>
        <button type="button" class="auth-btn auth-btn-primary" @click="accessArticle">校验密码</button>
      </el-form-item>
    </el-form>
  </el-dialog>
  <teleport to="body">
    <SearchModel />
  </teleport>
</template>

<script lang="ts">
import { computed, defineComponent, toRef, toRefs, reactive, getCurrentInstance, nextTick } from 'vue'
import { Dropdown, DropdownMenu, DropdownItem } from '@/components/Dropdown'
import { useAppStore } from '@/stores/app'
import { useCommonStore } from '@/stores/common'
import { useUserStore } from '@/stores/user'
import { useRoute, useRouter } from 'vue-router'
import ThemeToggle from '@/components/ToggleSwitch/ThemeToggle.vue'
import api from '@/api/api'
import SearchModel from '@/components/SearchModel.vue'
import { useSearchStore } from '@/stores/search'
import config from '@/config/config'
import { useI18n } from 'vue-i18n'
import emitter from '@/utils/mitt'

export default defineComponent({
  name: 'Controls',
  components: {
    Dropdown,
    DropdownMenu,
    DropdownItem,
    ThemeToggle,
    SearchModel
  },
  setup() {
    const { t } = useI18n()
    const proxy: any = getCurrentInstance()?.appContext.config.globalProperties
    const appStore = useAppStore()
    const commonStore = useCommonStore()
    const userStore = useUserStore()
    const searchStore = useSearchStore()
    const route = useRoute()
    const router = useRouter()
    const loginInfo = reactive({
      username: '' as any,
      password: '' as any,
      code: '' as any
    })
    const reactiveDate = reactive({
      loginDialogVisible: false,
      registerDialogVisible: false,
      forgetPasswordDialogVisible: false,
      articlePasswordDialogVisible: false,
      articlePassword: '',
      articleId: ''
    })
    emitter.on('changeArticlePasswordDialogVisible', (articleId: any) => {
      reactiveDate.articlePasswordDialogVisible = true
      reactiveDate.articlePassword = ''
      reactiveDate.articleId = articleId
      nextTick(() => {
        document.getElementById('article-password-input')?.focus()
      })
    })
    const handleClick = (name: string): void => {
      appStore.changeLocale(name)
    }
    const login = () => {
      if (loginInfo.username.trim().length == 0 || loginInfo.password.trim().length == 0) {
        proxy.$notify({
          title: 'Warning',
          message: '账号或者密码不能为空',
          type: 'warning'
        })
        return
      }
      let params = new URLSearchParams()
      params.append('username', loginInfo.username)
      params.append('password', loginInfo.password)
      api.login(params).then(({ data }) => {
        if (data.flag) {
          userStore.userInfo = data.data
          sessionStorage.setItem('token', data.data.token)
          userStore.token = data.data.token
          proxy.$notify({
            title: 'Success',
            message: '登录成功',
            type: 'success'
          })
          reactiveDate.loginDialogVisible = false
        }
      })
    }
    const logout = () => {
      api.logout().then(({ data }) => {
        if (data.flag) {
          userStore.userInfo = ''
          userStore.token = ''
          userStore.accessArticles = []
          sessionStorage.removeItem('token')
          proxy.$notify({
            title: 'Success',
            message: '登出成功',
            type: 'success'
          })
        }
      })
    }
    const openUserCenter = () => {
      userStore.userVisible = true
    }
    const openLoginDialog = () => {
      reactiveDate.loginDialogVisible = true
    }
    const openRegisterDialog = () => {
      loginInfo.code = ''
      reactiveDate.loginDialogVisible = false
      reactiveDate.registerDialogVisible = true
    }
    const returnLoginDialog = () => {
      reactiveDate.registerDialogVisible = false
      reactiveDate.forgetPasswordDialogVisible = false
      reactiveDate.loginDialogVisible = true
    }
    const openForgetPasswordDialog = () => {
      loginInfo.code = ''
      reactiveDate.loginDialogVisible = false
      reactiveDate.forgetPasswordDialogVisible = true
    }
    const sendCode = () => {
      api.sendValidationCode(loginInfo.username).then(({ data }) => {
        if (data.flag) {
          proxy.$notify({
            title: 'Success',
            message: '验证码已发送',
            type: 'success'
          })
        }
      })
    }
    const register = () => {
      let params = {
        code: loginInfo.code,
        username: loginInfo.username,
        password: loginInfo.password
      }
      api.register(params).then(({ data }) => {
        if (data.flag) {
          proxy.$notify({
            title: 'Success',
            message: '注册成功',
            type: 'success'
          })
          reactiveDate.registerDialogVisible = false
          reactiveDate.loginDialogVisible = true
        }
      })
    }
    const handleOpenModel: any = (status: boolean) => {
      searchStore.setOpenModal(status)
    }

    const qqLogin = () => {
      userStore.currentUrl = route.path
      reactiveDate.loginDialogVisible = false
      if (commonStore.isMobile) {
        //@ts-ignore
        QC.Login.showPopup({
          appId: config.qqLogin.QQ_APP_ID,
          redirectURI: config.qqLogin.QQ_REDIRECT_URI
        })
      } else {
        window.open(
          'https://graph.qq.com/oauth2.0/show?which=Login&display=pc&client_id=' +
            +config.qqLogin.QQ_APP_ID +
            '&response_type=token&scope=all&redirect_uri=' +
            config.qqLogin.QQ_REDIRECT_URI,
          '_self'
        )
      }
    }
    const updatePassword = () => {
      api.updatePassword(loginInfo).then(({ data }) => {
        if (data.flag) {
          proxy.$notify({
            title: 'Success',
            message: '修改成功',
            type: 'success'
          })
          reactiveDate.forgetPasswordDialogVisible = false
          reactiveDate.loginDialogVisible = true
        }
      })
    }
    const accessArticle = () => {
      if (reactiveDate.articlePassword.trim().length == 0) {
        proxy.$notify({
          title: 'Warning',
          message: '密码不能为空',
          type: 'warning'
        })
        return
      }
      api
        .accessArticle({
          articleId: reactiveDate.articleId,
          articlePassword: reactiveDate.articlePassword
        })
        .then(({ data }) => {
          if (data.flag) {
            reactiveDate.articlePasswordDialogVisible = false
            userStore.accessArticles.push(reactiveDate.articleId)
            router.push({ path: '/articles/' + reactiveDate.articleId })
          }
        })
    }
    return {
      handleOpenModel,
      loginInfo,
      ...toRefs(reactiveDate),
      userInfo: toRef(userStore.$state, 'userInfo'),
      isMobile: toRef(commonStore.$state, 'isMobile'),
      login,
      qqLogin,
      logout,
      handleClick,
      openUserCenter,
      openLoginDialog,
      openRegisterDialog,
      returnLoginDialog,
      sendCode,
      register,
      updatePassword,
      openForgetPasswordDialog,
      accessArticle,
      multiLanguage: computed(() => {
        let websiteConfig: any = appStore.websiteConfig
        return websiteConfig.multiLanguage
      }),
      t
    }
  }
})
</script>
<style lang="scss">
.auth-dialog .el-dialog {
  border-radius: 1rem !important;
  overflow: hidden;
}

.auth-dialog .el-dialog__header {
  display: none;
}

.auth-dialog .el-dialog__body {
  padding: 2rem 2.5rem !important;
}

.el-dialog__headerbtn {
  outline: none !important;
}

.el-input-group__append {
  background-color: var(--background-primary-alt) !important;
}

.el-form-item__label {
  text-align: left;
  width: 70px;
  color: var(--text-normal) !important;
}

.el-input__inner {
  color: var(--text-normal) !important;
  background-color: var(--background-primary-alt) !important;
  transition: border-color 0.25s ease !important;
}

.el-input__wrapper {
  background: var(--background-primary-alt) !important;
  border-radius: 0.5rem !important;
  box-shadow: none !important;
  border: 1px solid transparent !important;
  transition: border-color 0.25s ease, box-shadow 0.25s ease !important;
}

.el-input__wrapper:hover {
  border-color: var(--text-dim) !important;
}

.el-input__wrapper.is-focus {
  border-color: var(--text-accent) !important;
  box-shadow: 0 0 0 2px rgba(var(--text-accent-rgb, 100, 149, 237), 0.15) !important;
}
</style>
<style lang="scss" scoped>
.text {
  color: var(--text-normal);
  cursor: pointer;
}

.dialog-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.dialog-icon {
  width: 40px;
  height: 40px;
  color: var(--text-accent);
  opacity: 0.7;
  margin-bottom: 0.75rem;
}

.dialog-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-bright);
  margin: 0 0 0.25rem;
}

.dialog-subtitle {
  font-size: 0.85rem;
  color: var(--text-dim);
  margin: 0;
}

.auth-form {
  padding-top: 0.5rem;
}

.auth-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 0.65rem 1rem;
  border-radius: 0.5rem;
  font-weight: 500;
  font-size: 0.95rem;
  border: none;
  outline: none;
  cursor: pointer;
  transition: transform 0.25s cubic-bezier(0.22, 1, 0.36, 1),
              box-shadow 0.25s ease,
              opacity 0.2s ease;

  &:hover {
    transform: translateY(-1px);
  }

  &:active {
    transform: translateY(0) scale(0.98);
  }
}

.auth-btn-primary {
  background: var(--main-gradient);
  color: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.15);

  &:hover {
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
  }

  &:active {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
  }
}

.auth-btn-qq {
  background: var(--background-secondary);
  color: var(--text-normal);
  gap: 0.5rem;

  &:hover {
    background: var(--background-tertiary, rgba(255, 255, 255, 0.1));
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.1);
  }

  &:active {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }
}

.qq-icon {
  width: 18px;
  height: 18px;
}

.dialog-divider {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin: 1rem 0 0.75rem;
}

.divider-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(to right, transparent, var(--text-dim), transparent);
  opacity: 0.2;
}

.divider-text {
  font-size: 0.8rem;
  color: var(--text-dim);
  opacity: 0.5;
  white-space: nowrap;
}

.dialog-footer-links {
  display: flex;
  justify-content: space-between;
  margin-top: 1rem;
}

.auth-link {
  color: var(--text-dim);
  cursor: pointer;
  font-size: 0.85rem;
  transition: color 0.2s ease;

  &:hover {
    color: var(--text-accent);
  }
}

.header-controls {
  span {
    display: flex;
    justify-content: center;
    align-items: center;
    color: #fff;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    padding-right: 0.5rem;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
    &[no-hover-effect] {
      &:hover {
        opacity: 1;
      }
    }
    &:hover {
      opacity: 0.5;
    }
    &:active {
      transform: scale(0.95);
      opacity: 0.7;
    }
    .svg-icon {
      stroke: #fff;
      height: 2rem;
      width: 2rem;
      margin-right: 0.5rem;
      pointer-events: none;
    }
  }
  .search-bar {
    @apply bg-transparent flex flex-row px-0 mr-2 rounded-full;
    opacity: 0;
    width: 0;
    transition: 300ms all ease-out;
    &.active {
      @apply bg-ob-deep-800;
      opacity: 0.95;
      width: 200px;
      imput {
        width: initial;
      }
    }
    &:focus {
      appearance: none;
      outline: none;
    }
    input {
      @apply flex flex-1 bg-transparent text-ob-normal px-6 box-border;
      width: 0;
      appearance: none;
      outline: none;
    }
    svg {
      @apply float-right;
    }
  }
}

// 移动端优化
@media (max-width: 640px) {
  .header-controls {
    span {
      .svg-icon {
        height: 1.75rem;
        width: 1.75rem;
      }
    }
  }
}

.avatar-img {
  transition-property: transform;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 800ms;
  transform: rotate(-360deg);
}
.avatar-img:hover {
  transform: rotate(360deg);
}
</style>

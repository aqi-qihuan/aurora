<template>
  <el-card class="main-card user-management">
    <div class="title">
      <i class="el-icon-user-solid" />
      {{ this.$route.name }}
    </div>
    <div class="operation-container">
      <div class="filter-section">
        <el-select 
          clearable 
          v-model="loginType" 
          placeholder="请选择登录方式" 
          size="small" 
          class="filter-select">
          <el-option v-for="item in typeList" :key="item.type" :label="item.desc" :value="item.type" />
        </el-select>
        <el-input
          v-model="keywords"
          prefix-icon="el-icon-search"
          size="small"
          placeholder="请输入昵称"
          class="search-input"
          @keyup.enter.native="searchUsers" />
        <el-button type="primary" size="small" icon="el-icon-search" @click="searchUsers">
          搜索
        </el-button>
      </div>
    </div>
    <el-table 
      border 
      :data="userList" 
      v-loading="loading"
      class="user-table"
      :header-cell-style="{ background: 'var(--color-bg-hover)', color: 'var(--color-text)', fontWeight: '600' }">
      <el-table-column prop="linkAvatar" label="头像" align="center" width="80">
        <template slot-scope="scope">
          <el-avatar :src="scope.row.avatar" :size="40" class="user-avatar" />
        </template>
      </el-table-column>
      <el-table-column prop="nickname" label="昵称" align="center" min-width="120">
        <template slot-scope="scope">
          <span class="user-nickname">{{ scope.row.nickname }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="loginType" label="登录方式" align="center" width="100">
        <template slot-scope="scope">
          <el-tag 
            :type="getLoginType(scope.row.loginType).type" 
            effect="light"
            class="login-type-tag">
            <i :class="getLoginType(scope.row.loginType).icon" />
            {{ getLoginType(scope.row.loginType).label }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="roles" label="用户角色" align="center" min-width="150">
        <template slot-scope="scope">
          <div class="role-tags">
            <el-tag 
              v-for="(item, index) of scope.row.roles" 
              :key="index" 
              type="primary"
              effect="plain"
              size="mini"
              class="role-tag">
              {{ item.roleName }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="isDisable" label="状态" align="center" width="90">
        <template slot-scope="scope">
          <el-switch
            v-model="scope.row.isDisable"
            active-color="#F56C6C"
            inactive-color="#67C23A"
            :active-value="1"
            :inactive-value="0"
            @change="changeDisable(scope.row)"
            class="status-switch" />
          <div class="status-text" :class="scope.row.isDisable ? 'disabled' : 'enabled'">
            {{ scope.row.isDisable ? '已禁用' : '正常' }}
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="ipAddress" label="登录IP" align="center" width="130">
        <template slot-scope="scope">
          <span class="ip-address">{{ scope.row.ipAddress || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="ipSource" label="登录地址" align="center" min-width="120">
        <template slot-scope="scope">
          <span class="ip-source">{{ scope.row.ipSource || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="140" align="center">
        <template slot-scope="scope">
          <div class="time-info">
            <i class="el-icon-time" />
            <span>{{ scope.row.createTime | date }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="lastLoginTime" label="上次登录" width="140" align="center">
        <template slot-scope="scope">
          <div class="time-info">
            <i class="el-icon-time" />
            <span>{{ scope.row.lastLoginTime | date }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="100" fixed="right">
        <template slot-scope="scope">
          <el-button 
            type="primary" 
            size="mini" 
            icon="el-icon-edit"
            circle
            @click="openEditModel(scope.row)">
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      class="pagination-container"
      background
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      :page-sizes="[10, 20]"
      layout="total, sizes, prev, pager, next, jumper" />
    <el-dialog :visible.sync="isEdit" width="400px" class="user-edit-dialog">
      <div class="dialog-title-container" slot="title">
        <i class="el-icon-edit" />
        修改用户
      </div>
      <el-form label-width="70px" size="medium" :model="userForm" class="user-form">
        <el-form-item label="昵称">
          <el-input v-model="userForm.nickname" placeholder="请输入用户昵称" prefix-icon="el-icon-user" />
        </el-form-item>
        <el-form-item label="角色">
          <el-checkbox-group v-model="roleIds" class="role-checkbox-group">
            <el-checkbox 
              v-for="item of userRoles" 
              :key="item.id" 
              :label="item.id"
              border
              size="small">
              {{ item.roleName }}
            </el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="isEdit = false" size="medium">取 消</el-button>
        <el-button type="primary" @click="editUserRole" size="medium" icon="el-icon-check"> 确 定 </el-button>
      </div>
    </el-dialog>
  </el-card>
</template>

<script>
export default {
  created() {
    this.current = this.$store.state.pageState.user
    this.listUsers()
    this.listRoles()
  },
  data: function () {
    return {
      loading: true,
      isEdit: false,
      userForm: {
        userInfoId: null,
        nickname: ''
      },
      loginType: null,
      userRoles: [],
      roleIds: [],
      userList: [],
      typeList: [
        {
          type: 1,
          desc: '邮箱'
        },
        {
          type: 2,
          desc: 'QQ'
        },
        {
          type: 3,
          desc: '微博'
        }
      ],
      keywords: null,
      current: 1,
      size: 10,
      count: 0
    }
  },
  methods: {
    searchUsers() {
      this.current = 1
      this.listUsers()
    },
    sizeChange(size) {
      this.size = size
      this.listUsers()
    },
    currentChange(current) {
      this.current = current
      this.$store.commit('updateUserPageState', current)
      this.listUsers()
    },
    changeDisable(user) {
      this.axios.put('/api/admin/users/disable', {
        id: user.userInfoId,
        isDisable: user.isDisable
      })
    },
    openEditModel(user) {
      this.roleIds = []
      this.userForm = JSON.parse(JSON.stringify(user))
      this.userForm.roles.forEach((item) => {
        this.roleIds.push(item.id)
      })
      this.isEdit = true
    },
    editUserRole() {
      this.userForm.roleIds = this.roleIds
      this.axios.put('/api/admin/users/role', this.userForm).then(({ data }) => {
        if (data.flag) {
          this.$notify.success({
            title: '成功',
            message: data.message
          })
          this.listUsers()
        } else {
          this.$notify.error({
            title: '失败',
            message: data.message
          })
        }
        this.isEdit = false
      })
    },
    listUsers() {
      this.axios
        .get('/api/admin/users', {
          params: {
            current: this.current,
            size: this.size,
            keywords: this.keywords,
            loginType: this.loginType
          }
        })
        .then(({ data }) => {
          this.userList = data.data.records
          this.count = data.data.count
          this.loading = false
        })
    },
    listRoles() {
      this.axios.get('/api/admin/users/role').then(({ data }) => {
        this.userRoles = data.data
      })
      .catch(error => {
        this.$message.error('获取角色列表失败')
        console.error('API Error:', error)
      })
    }
  },
  watch: {
    loginType() {
      this.current = 1
      this.listUsers()
    }
  },
  computed: {
    getLoginType() {
      return function(type) {
        const types = {
          1: { label: '邮箱', type: 'success', icon: 'el-icon-message' },
          2: { label: 'QQ', type: 'primary', icon: 'el-icon-chat-dot-round' },
          3: { label: '微博', type: 'danger', icon: 'el-icon-share' }
        }
        return types[type] || { label: '未知', type: 'info', icon: 'el-icon-question' }
      }
    }
  }
}
</script>

<style scoped>
/* 用户管理页面样式 */
.user-management {
  animation: fadeIn 0.4s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--color-border);
}

.title i {
  color: var(--color-primary);
  font-size: 24px;
}

/* 筛选区域 */
.filter-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-select {
  width: 160px;
}

.search-input {
  width: 220px;
}

/* 用户表格 */
.user-table {
  margin-top: 20px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.user-table ::v-deep .el-table__header th {
  background: var(--color-bg-hover) !important;
  font-weight: 600;
  text-transform: uppercase;
  font-size: 12px;
  letter-spacing: 0.5px;
}

.user-table ::v-deep .el-table__row {
  transition: all 0.3s ease;
}

.user-table ::v-deep .el-table__row:hover {
  background-color: var(--color-bg-hover) !important;
  transform: translateX(4px);
}

/* 用户头像 */
.user-avatar {
  border: 2px solid var(--color-border);
  transition: all 0.3s ease;
}

.user-avatar:hover {
  transform: scale(1.1);
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px var(--color-primary-100);
}

/* 用户昵称 */
.user-nickname {
  font-weight: 500;
  color: var(--color-text);
}

/* 登录类型标签 */
.login-type-tag {
  border-radius: 20px;
  padding: 4px 12px;
  font-weight: 500;
}

.login-type-tag i {
  margin-right: 4px;
}

/* 角色标签 */
.role-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: center;
}

.role-tag {
  border-radius: 4px;
  font-weight: 500;
}

/* 状态开关 */
.status-switch {
  margin-bottom: 4px;
}

.status-text {
  font-size: 12px;
  font-weight: 500;
  margin-top: 4px;
}

.status-text.enabled {
  color: var(--color-success);
}

.status-text.disabled {
  color: var(--color-error);
}

/* IP 信息 */
.ip-address {
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--color-text-secondary);
  background: var(--color-bg-hover);
  padding: 2px 8px;
  border-radius: 4px;
}

.ip-source {
  font-size: 13px;
  color: var(--color-text-secondary);
}

/* 时间信息 */
.time-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.time-info i {
  color: var(--color-primary);
}

/* 分页 */
.pagination-container {
  float: right;
  margin-top: 24px;
}

/* 对话框 */
.dialog-title-container {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
}

.dialog-title-container i {
  color: var(--color-primary);
  font-size: 20px;
}

.user-form {
  padding: 20px 0;
}

.role-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.role-checkbox-group .el-checkbox {
  margin: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 响应式优化 */
@media (max-width: 768px) {
  .filter-section {
    flex-direction: column;
    width: 100%;
  }

  .filter-select,
  .search-input {
    width: 100% !important;
  }

  .user-table ::v-deep .el-table__header {
    display: none;
  }

  .user-table ::v-deep .el-table__row {
    display: flex;
    flex-direction: column;
    padding: 16px;
    margin-bottom: 12px;
    background: var(--color-bg-card);
    border-radius: 12px;
    border: 1px solid var(--color-border);
  }

  .user-table ::v-deep .el-table__row td {
    display: flex;
    align-items: center;
    padding: 8px 0 !important;
    border: none;
  }

  .user-table ::v-deep .el-table__row td::before {
    content: attr(data-label);
    font-weight: 600;
    color: var(--color-text-secondary);
    min-width: 80px;
    margin-right: 12px;
    font-size: 13px;
  }

  .user-table ::v-deep .el-table__row td .cell {
    flex: 1;
    text-align: left !important;
  }

  .role-tags {
    justify-content: flex-start;
  }

  .pagination-container {
    float: none;
    display: flex;
    justify-content: center;
    margin: 20px 0;
  }
}
</style>

/**
 * 现代化页面 Mixin
 * 为所有管理页面提供统一的现代化样式和功能
 */

export default {
  data() {
    return {
      // 页面加载状态
      pageLoading: false,
      // 搜索关键词
      searchKeywords: '',
      // 当前页码
      currentPage: 1,
      // 每页条数
      pageSize: 10,
      // 总条数
      totalCount: 0,
      // 选中的数据ID
      selectedIds: [],
      // 表格数据
      tableData: [],
      // 对话框可见性
      dialogVisible: false,
      // 当前操作类型
      currentAction: '',
    }
  },

  computed: {
    // 是否有选中项
    hasSelection() {
      return this.selectedIds.length > 0
    },
    
    // 选中项数量文本
    selectionText() {
      return this.selectedIds.length > 0 ? `已选择 ${this.selectedIds.length} 项` : ''
    },
    
    // 分页配置
    paginationConfig() {
      return {
        currentPage: this.currentPage,
        pageSize: this.pageSize,
        total: this.totalCount,
        pageSizes: [10, 20, 50, 100],
        layout: 'total, sizes, prev, pager, next, jumper'
      }
    }
  },

  methods: {
    /**
     * 处理搜索
     */
    handleSearch() {
      this.currentPage = 1
      this.fetchData()
    },

    /**
     * 处理重置
     */
    handleReset() {
      this.searchKeywords = ''
      this.currentPage = 1
      this.fetchData()
    },

    /**
     * 处理分页大小变化
     */
    handleSizeChange(size) {
      this.pageSize = size
      this.fetchData()
    },

    /**
     * 处理页码变化
     */
    handleCurrentChange(page) {
      this.currentPage = page
      this.fetchData()
    },

    /**
     * 处理表格选择变化
     */
    handleSelectionChange(selection) {
      this.selectedIds = selection.map(item => item.id)
    },

    /**
     * 打开对话框
     */
    openDialog(action, row = null) {
      this.currentAction = action
      this.dialogVisible = true
      if (row) {
        this.formData = { ...row }
      }
    },

    /**
     * 关闭对话框
     */
    closeDialog() {
      this.dialogVisible = false
      this.currentAction = ''
      if (this.$refs.form) {
        this.$refs.form.resetFields()
      }
    },

    /**
     * 显示成功消息
     */
    showSuccess(message) {
      this.$notify.success({
        title: '成功',
        message: message,
        duration: 2000
      })
    },

    /**
     * 显示错误消息
     */
    showError(message) {
      this.$notify.error({
        title: '错误',
        message: message,
        duration: 3000
      })
    },

    /**
     * 确认操作
     */
    confirmAction(message, callback) {
      this.$confirm(message, '确认操作', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        callback()
      }).catch(() => {})
    },

    /**
     * 获取数据（子类必须实现）
     */
    fetchData() {
      console.warn('fetchData method must be implemented')
    }
  }
}

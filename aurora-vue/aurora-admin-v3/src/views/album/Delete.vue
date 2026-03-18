<template>
  <el-card class="main-card">
    <div class="title">{{ route.meta.title || route.name || '回收站' }}</div>
    <div class="operation">
      <div class="all-check">
        <el-checkbox :indeterminate="isIndeterminate" v-model="checkAll" @change="handleCheckAllChange">
          全选
        </el-checkbox>
        <div class="check-count">已选择{{ selectPhotoIds.length }}张</div>
      </div>
      <el-button
        type="success"
        @click="updatePhotoDelete(null)"
        :disabled="selectPhotoIds.length === 0"
        size="small">
        <el-icon style="margin-right: 4px"><RefreshRight /></el-icon>
        批量恢复
      </el-button>
      <el-button
        type="danger"
        @click="batchDeletePhoto = true"
        :disabled="selectPhotoIds.length === 0"
        size="small">
        <el-icon style="margin-right: 4px"><Delete /></el-icon>
        批量删除
      </el-button>
    </div>
    <el-row class="photo-container" :gutter="10" v-loading="loading">
      <el-empty v-if="photos.length === 0" description="暂无照片" />
      <el-checkbox-group v-model="selectPhotoIds" @change="handleCheckedPhotoChange">
        <el-col :md="4" :sm="6" :xs="12" v-for="item of photos" :key="item.id">
          <el-checkbox :value="item.id">
            <div class="photo-item">
              <el-image fit="cover" class="photo-img" :src="item.photoSrc" :preview-src-list="photoSrcList" />
              <div class="photo-name">{{ item.photoName }}</div>
            </div>
          </el-checkbox>
        </el-col>
      </el-checkbox-group>
    </el-row>
    <el-pagination
      v-if="count > 0"
      :hide-on-single-page="true"
      class="pagination-container"
      @size-change="sizeChange"
      @current-change="currentChange"
      :current-page="current"
      :page-size="size"
      :total="count"
      layout="prev, pager, next" />
    
    <!-- 批量删除确认 -->
    <el-dialog v-model="batchDeletePhoto" width="30%">
      <template #header>
        <div class="dialog-title-container">
          <el-icon style="color: #ff9900; margin-right: 8px"><WarningFilled /></el-icon>提示
        </div>
      </template>
      <div style="font-size: 1rem">是否删除选中照片？</div>
      <template #footer>
        <el-button @click="batchDeletePhoto = false">取 消</el-button>
        <el-button type="primary" @click="deletePhotos"> 确 定 </el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElNotification } from 'element-plus'
import { RefreshRight, Delete, WarningFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'

const route = useRoute()

const loading = ref(true)
const batchDeletePhoto = ref(false)
const isIndeterminate = ref(false)
const checkAll = ref(false)
const photos = ref([])
const photoIds = ref([])
const selectPhotoIds = ref([])
const current = ref(1)
const size = ref(18)
const count = ref(0)

const photoSrcList = computed(() => photos.value.map(p => p.photoSrc))

onMounted(() => {
  listPhotos()
})

watch(photos, (newPhotos) => {
  photoIds.value = newPhotos.map(item => item.id)
})

const listPhotos = () => {
  request.get('/admin/photos', {
    params: {
      current: current.value,
      size: size.value,
      isDelete: 1
    }
  }).then(({ data }) => {
    if (data && data.data) {
      photos.value = data.data.records || []
      count.value = data.data.count || 0
    }
    loading.value = false
  }).catch(() => {
    loading.value = false
  })
}

const sizeChange = (val) => {
  size.value = val
  listPhotos()
}

const currentChange = (val) => {
  current.value = val
  listPhotos()
}

const updatePhotoDelete = (id) => {
  const param = id 
    ? { ids: [id], isDelete: 0 }
    : { ids: selectPhotoIds.value, isDelete: 0 }
  request.put('/admin/photos/delete', param).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listPhotos()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    batchDeletePhoto.value = false
  })
}

const deletePhotos = () => {
  request.delete('/admin/photos', { data: selectPhotoIds.value }).then(({ data }) => {
    if (data.flag) {
      ElNotification.success({
        title: '成功',
        message: data.message
      })
      listPhotos()
    } else {
      ElNotification.error({
        title: '失败',
        message: data.message
      })
    }
    batchDeletePhoto.value = false
  })
}

const handleCheckAllChange = (val) => {
  selectPhotoIds.value = val ? photoIds.value : []
  isIndeterminate.value = false
}

const handleCheckedPhotoChange = (value) => {
  const checkedCount = value.length
  checkAll.value = checkedCount === photoIds.value.length
  isIndeterminate.value = checkedCount > 0 && checkedCount < photoIds.value.length
}
</script>

<style scoped>
.operation {
  display: flex;
  justify-content: flex-end;
  margin-top: 2.25rem;
  margin-bottom: 2rem;
  align-items: center;
}
.all-check {
  display: inline-flex;
  align-items: center;
  margin-right: 1rem;
}
.check-count {
  margin-left: 1rem;
  font-size: 12px;
}
.photo-item {
  position: relative;
  cursor: pointer;
  margin-bottom: 1rem;
}
.photo-img {
  width: 100%;
  height: 7rem;
  border-radius: 4px;
}
.photo-name {
  font-size: 14px;
  margin-top: 0.3rem;
  text-align: center;
}
</style>

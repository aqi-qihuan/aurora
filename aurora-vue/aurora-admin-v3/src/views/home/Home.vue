<template>
  <div class="dashboard-container">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6" :lg="6" 
              v-for="(stat, index) in stats" 
              :key="stat.title"
              :class="`stagger-${index + 1}`">
        <el-card class="stat-card hover-lift" shadow="hover">
          <div class="card-content">
            <div class="card-icon-wrapper" :style="{ background: stat.gradient }">
              <el-icon :size="32">
                <component :is="stat.icon" />
              </el-icon>
            </div>
            <div class="card-desc">
              <div class="card-title">{{ stat.title }}</div>
              <div class="card-count">{{ formatNumber(stat.value) }}</div>
              <div class="card-trend" :class="{ 'trend-down': stat.trend < 0 }">
                <el-icon>
                  <Top v-if="stat.trend >= 0" />
                  <Bottom v-else />
                </el-icon>
                <span>较上周 {{ stat.trend >= 0 ? '+' : '' }}{{ stat.trend }}%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 一周访问量图表 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <el-icon><TrendCharts /></el-icon>
              一周访问量
            </div>
            <el-button type="primary" :icon="Refresh" @click="refreshData">刷新</el-button>
          </div>
          <div class="chart-container" v-loading="loading">
            <v-chart :option="viewCountOption" autoresize style="height: 300px" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 文章排行和分类 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <el-icon><DataAnalysis /></el-icon>
              文章浏览量排行
            </div>
          </div>
          <div class="chart-container" v-loading="loading">
            <v-chart :option="articleRankOption" autoresize style="height: 300px" />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <el-icon><PieChart /></el-icon>
              文章分类统计
            </div>
          </div>
          <div class="chart-container pie-chart" v-loading="loading">
            <v-chart :option="categoryOption" autoresize style="height: 300px" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 用户地域分布和标签 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <el-icon><Location /></el-icon>
              用户地域分布
            </div>
            <el-radio-group v-model="userType" size="small">
              <el-radio-button :value="1">用户</el-radio-button>
              <el-radio-button :value="2">游客</el-radio-button>
            </el-radio-group>
          </div>
          <div class="chart-container" v-loading="mapLoading">
            <v-chart v-if="mapReady" :option="userAreaOption" autoresize style="height: 400px" />
            <el-empty v-else description="地图加载中..." :image-size="100" />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <el-icon><CollectionTag /></el-icon>
              文章标签统计
            </div>
          </div>
          <div class="tag-cloud-container" v-loading="loading">
            <TagCloud :tags="tagDTOs" v-if="tagDTOs.length > 0" />
            <el-empty v-else description="暂无标签数据" :image-size="100" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { 
  Top, Bottom, TrendCharts, DataAnalysis, PieChart, Location, CollectionTag, Refresh,
  User, Document, ChatDotRound, View
} from '@element-plus/icons-vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart, PieChart as EchartsPie, MapChart } from 'echarts/charts'
import { 
  GridComponent, TooltipComponent, LegendComponent, VisualMapComponent, GeoComponent 
} from 'echarts/components'
import TagCloud from '@/components/TagCloud.vue'
import request from '@/utils/request'
import { registerChinaMap, provinceCoordinates, convertToScatterData } from '@/utils/chinaMap'

// 注册 ECharts 组件
use([
  CanvasRenderer, LineChart, BarChart, EchartsPie, MapChart,
  GridComponent, TooltipComponent, LegendComponent, VisualMapComponent, GeoComponent
])

// 响应式数据
const loading = ref(true)
const mapLoading = ref(false)
const mapReady = ref(false)
const userType = ref(1)

// 主题状态
const isDark = ref(false)

// 监听主题变化
const observeTheme = () => {
  const checkTheme = () => {
    const theme = document.documentElement.getAttribute('data-theme')
    isDark.value = theme === 'dark'
  }
  
  // 初始检查
  checkTheme()
  
  // 监听 data-theme 属性变化
  const observer = new MutationObserver(() => {
    checkTheme()
  })
  
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-theme']
  })
}

// 动态图表颜色配置
const chartColors = computed(() => {
  if (isDark.value) {
    // 深色主题 - 极客风
    return {
      // 文字颜色
      textPrimary: '#F8FAFC',
      textSecondary: '#E2E8F0',
      textMuted: '#94A3B8',
      
      // 边框和线条
      border: '#64748B',
      splitLine: '#475569',
      
      // Tooltip
      tooltipBg: 'rgba(27, 35, 54, 0.95)',
      tooltipBorder: '#475569',
      tooltipText: '#F8FAFC',
      
      // 主色
      primary: '#00D4FF',
      primaryLight: '#3B82F6',
      
      // 地图
      mapArea: '#1B2336',
      mapBorder: '#475569',
      mapEmphasis: '#272F42',
      
      // 渐变色
      gradient1: '#60A5FA',
      gradient2: '#3B82F6',
      
      // 面积图
      areaOpacity1: 'rgba(0, 212, 255, 0.4)',
      areaOpacity2: 'rgba(0, 212, 255, 0.05)',
      
      // 饼图颜色
      pieColors: ['#00D4FF', '#3B82F6', '#8B5CF6', '#BF5AF2', '#FF9F0A', '#22C55E', '#EF4444', '#F59E0B']
    }
  } else {
    // 浅色主题 - 清爽明亮
    return {
      // 文字颜色
      textPrimary: '#1F2937',
      textSecondary: '#4B5563',
      textMuted: '#6B7280',
      
      // 边框和线条
      border: '#D1D5DB',
      splitLine: '#E5E7EB',
      
      // Tooltip
      tooltipBg: 'rgba(255, 255, 255, 0.95)',
      tooltipBorder: '#D1D5DB',
      tooltipText: '#1F2937',
      
      // 主色
      primary: '#3B82F6',
      primaryLight: '#60A5FA',
      
      // 地图
      mapArea: '#F3F4F6',
      mapBorder: '#D1D5DB',
      mapEmphasis: '#E5E7EB',
      
      // 渐变色
      gradient1: '#60A5FA',
      gradient2: '#3B82F6',
      
      // 面积图
      areaOpacity1: 'rgba(59, 130, 246, 0.3)',
      areaOpacity2: 'rgba(59, 130, 246, 0.05)',
      
      // 饼图颜色
      pieColors: ['#3B82F6', '#60A5FA', '#8B5CF6', '#A78BFA', '#F97316', '#10B981', '#EF4444', '#F59E0B']
    }
  }
})

// 统计卡片数据
const stats = ref([
  {
    title: '访问量',
    value: 0,
    icon: View,
    gradient: 'linear-gradient(135deg, #3B82F6 0%, #2563EB 100%)',
    trend: 12
  },
  {
    title: '用户量',
    value: 0,
    icon: User,
    gradient: 'linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%)',
    trend: 8
  },
  {
    title: '文章量',
    value: 0,
    icon: Document,
    gradient: 'linear-gradient(135deg, #F97316 0%, #EA580C 100%)',
    trend: 15
  },
  {
    title: '留言量',
    value: 0,
    icon: ChatDotRound,
    gradient: 'linear-gradient(135deg, #10B981 0%, #059669 100%)',
    trend: -3
  }
])

// 标签数据
const tagDTOs = ref([])

// 一周访问量图表配置
const viewCountOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'cross' },
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText }
  },
  color: [chartColors.value.primary],
  grid: {
    left: '3%', right: '4%', bottom: '3%', top: '10%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: viewCountData.value.xAxis,
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 12, fontWeight: 500 },
    axisLine: { lineStyle: { color: chartColors.value.border, width: 2 } },
    axisTick: { lineStyle: { color: chartColors.value.border } }
  },
  yAxis: {
    type: 'value',
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 12, fontWeight: 500 },
    splitLine: { lineStyle: { color: chartColors.value.splitLine, type: 'dashed' } },
    axisLine: { show: true, lineStyle: { color: chartColors.value.border, width: 2 } }
  },
  series: [{
    name: '访问量',
    type: 'line',
    data: viewCountData.value.series,
    smooth: true,
    symbol: 'circle',
    symbolSize: 8,
    lineStyle: { width: 3, color: chartColors.value.primary },
    areaStyle: { 
      opacity: 0.3,
      color: {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: chartColors.value.areaOpacity1 },
          { offset: 1, color: chartColors.value.areaOpacity2 }
        ]
      }
    }
  }]
}))

// 文章排行图表配置
const articleRankOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText }
  },
  color: [chartColors.value.primaryLight],
  grid: {
    left: '3%', right: '4%', bottom: '10%', top: '5%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: articleRankData.value.xAxis,
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 11, fontWeight: 500, rotate: 30, interval: 0 },
    axisLine: { lineStyle: { color: chartColors.value.border, width: 2 } },
    axisTick: { lineStyle: { color: chartColors.value.border } }
  },
  yAxis: {
    type: 'value',
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 12, fontWeight: 500 },
    splitLine: { lineStyle: { color: chartColors.value.splitLine, type: 'dashed' } },
    axisLine: { show: true, lineStyle: { color: chartColors.value.border, width: 2 } }
  },
  series: [{
    name: '浏览量',
    type: 'bar',
    data: articleRankData.value.series,
    barWidth: '60%',
    itemStyle: {
      color: {
        type: 'linear',
        x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: chartColors.value.gradient1 },
          { offset: 1, color: chartColors.value.gradient2 }
        ]
      },
      borderRadius: [4, 4, 0, 0]
    }
  }]
}))

// 文章分类图表配置
const categoryOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText }
  },
  legend: {
    orient: 'vertical',
    left: 'left',
    textStyle: { color: chartColors.value.textSecondary, fontSize: 12 }
  },
  color: chartColors.value.pieColors,
  series: [{
    name: '文章分类',
    type: 'pie',
    radius: ['40%', '70%'],
    center: ['60%', '50%'],
    data: categoryData.value,
    label: {
      show: true,
      formatter: '{b}: {c}',
      color: chartColors.value.textPrimary
    },
    emphasis: {
      label: {
        show: true,
        fontSize: 14,
        fontWeight: 'bold'
      }
    }
  }]
}))

// 用户地域分布图表配置
const userAreaOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText },
    formatter: function(params) {
      if (params.value) {
        return params.name + ': ' + params.value + '人'
      }
      return params.name + ': 0人'
    }
  },
  visualMap: {
    min: 0,
    max: 100,
    left: 'left',
    top: 'bottom',
    text: ['高', '低'],
    calculable: true,
    textStyle: { color: chartColors.value.textMuted },
    pieces: [
      { gt: 100, label: '100人以上', color: '#EF4444' },
      { gte: 51, lte: 100, label: '51-100人', color: '#22C55E' },
      { gte: 21, lte: 50, label: '21-50人', color: '#F59E0B' },
      { gt: 0, lte: 20, label: '1-20人', color: '#3B82F6' }
    ],
    show: true
  },
  geo: {
    map: 'china',
    zoom: 1.2,
    roam: true,
    layoutCenter: ['50%', '50%'],
    itemStyle: {
      areaColor: chartColors.value.mapArea,
      borderColor: chartColors.value.mapBorder
    },
    emphasis: {
      itemStyle: {
        areaColor: chartColors.value.mapEmphasis,
        shadowOffsetX: 0,
        shadowOffsetY: 0,
        borderWidth: 0
      }
    }
  },
  series: [{
    name: '用户人数',
    type: 'map',
    geoIndex: 0,
    data: userAreaData.value,
    areaColor: '#0FB8F0'
  }]
}))

// 图表数据
const viewCountData = ref({ xAxis: [], series: [] })
const articleRankData = ref({ xAxis: [], series: [] })
const categoryData = ref([])
const userAreaData = ref([])

// 注册中国地图（异步）
const initChinaMap = async () => {
  try {
    await registerChinaMap()
    mapReady.value = true
    console.log('✅ 地图初始化完成')
  } catch (error) {
    console.error('❌ 地图初始化失败:', error)
    mapReady.value = true // 即使失败也设置为 true，使用简化版
  }
}

// 监听用户类型变化
watch(userType, () => {
  listUserArea()
})

// 获取首页数据
const getData = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin')
    
    if (data && data.data) {
      // 更新统计卡片
      stats.value[0].value = data.data.viewsCount || 0
      stats.value[1].value = data.data.userCount || 0
      stats.value[2].value = data.data.articleCount || 0
      stats.value[3].value = data.data.messageCount || 0
      
      // 一周访问量
      if (data.data.uniqueViewDTOs && data.data.uniqueViewDTOs.length > 0) {
        viewCountData.value.xAxis = data.data.uniqueViewDTOs.map(item => item.day)
        viewCountData.value.series = data.data.uniqueViewDTOs.map(item => item.viewsCount)
      }
      
      // 文章分类
      if (data.data.categoryDTOs && data.data.categoryDTOs.length > 0) {
        categoryData.value = data.data.categoryDTOs.map(item => ({
          value: item.articleCount,
          name: item.categoryName
        }))
      }
      
      // 文章排行
      if (data.data.articleRankDTOs && data.data.articleRankDTOs.length > 0) {
        articleRankData.value.xAxis = data.data.articleRankDTOs.map(item => item.articleTitle)
        articleRankData.value.series = data.data.articleRankDTOs.map(item => item.viewsCount)
      }
      
      // 标签数据
      if (data.data.tagDTOs && data.data.tagDTOs.length > 0) {
        tagDTOs.value = data.data.tagDTOs.map(item => ({
          id: item.id,
          name: item.tagName
        }))
      }
    }
  } catch (error) {
    console.error('获取数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 获取用户地域分布
const listUserArea = async () => {
  try {
    mapLoading.value = true
    const { data } = await request.get('/admin/users/area', {
      params: { type: userType.value }
    })
    
    if (data && data.data) {
      const processedData = data.data.map(item => {
        let provinceName = item.province || item.name || ''
        
        if (provinceName && !provinceName.match(/(市|省|自治区|特别行政区)$/)) {
          const municipalities = ['北京', '天津', '上海', '重庆']
          const specialRegions = ['香港', '澳门']
          const autonomousRegions = ['内蒙古', '广西', '西藏', '宁夏', '新疆']
          
          if (municipalities.includes(provinceName)) {
            provinceName = provinceName + '市'
          } else if (specialRegions.includes(provinceName)) {
            provinceName = provinceName + '特别行政区'
          } else if (autonomousRegions.includes(provinceName)) {
            if (provinceName === '内蒙古') {
              provinceName = '内蒙古自治区'
            } else if (provinceName === '广西') {
              provinceName = '广西壮族自治区'
            } else if (provinceName === '西藏') {
              provinceName = '西藏自治区'
            } else if (provinceName === '宁夏') {
              provinceName = '宁夏回族自治区'
            } else if (provinceName === '新疆') {
              provinceName = '新疆维吾尔自治区'
            }
          } else {
            provinceName = provinceName + '省'
          }
        }
        
        return {
          name: provinceName,
          value: item.count || item.value || 0
        }
      })
      
      userAreaData.value = processedData
      console.log('✅ 用户地域数据:', processedData)
    }
  } catch (error) {
    console.error('获取用户地域分布失败:', error)
  } finally {
    mapLoading.value = false
  }
}

// 刷新数据
const refreshData = () => {
  loading.value = true
  viewCountData.value = { xAxis: [], series: [] }
  articleRankData.value = { xAxis: [], series: [] }
  categoryData.value = []
  userAreaData.value = []
  tagDTOs.value = []
  
  getData()
  listUserArea()
}

// 格式化数字
const formatNumber = (num) => {
  if (!num) return 0
  return num.toLocaleString()
}

// 生命周期
onMounted(() => {
  observeTheme()
  initChinaMap()
  getData()
  listUserArea()
})
</script>

<style scoped>
/* ===== 首页样式 - 双主题支持 ===== */
.dashboard-container {
  padding: 24px;
}

.stat-row {
  margin-bottom: 20px;
}

/* 统计卡片 */
.stat-card {
  margin-bottom: 20px;
  border: 1px solid var(--border-light) !important;
  border-radius: 12px !important;
  background: var(--bg-base) !important;
  transition: all 0.25s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1)) !important;
}

.stat-card:hover {
  border-color: var(--primary) !important;
  box-shadow: var(--shadow-lg) !important;
}

.card-content {
  display: flex;
  align-items: center;
}

.card-icon-wrapper {
  width: 72px;
  height: 72px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20px;
  color: #fff;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-card:hover .card-icon-wrapper {
  transform: scale(1.1) rotate(5deg);
}

.card-desc {
  flex: 1;
}

.card-title {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 8px;
  font-weight: 500;
}

.card-count {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
  font-family: var(--font-mono, 'JetBrains Mono', monospace);
}

.card-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--success);
  font-weight: 500;
}

.card-trend.trend-down {
  color: var(--danger);
}

.chart-row {
  margin-bottom: 20px;
}

/* 图表卡片 */
.chart-card {
  margin-bottom: 20px;
  border: 1px solid var(--border-light) !important;
  border-radius: 12px !important;
  background: var(--bg-base) !important;
  transition: all 0.25s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1)) !important;
}

.chart-card:hover {
  border-color: var(--primary) !important;
  box-shadow: var(--shadow-lg) !important;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.chart-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-container {
  width: 100%;
}

.pie-chart {
  display: flex;
  justify-content: center;
}

.tag-cloud-container {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.hover-lift {
  transition: all 0.3s ease;
}

.hover-lift:hover {
  transform: translateY(-4px);
}

/* 深色主题特殊效果 */
[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.5) !important;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4),
              0 0 20px rgba(59, 130, 246, 0.2) !important;
}

[data-theme="dark"] .card-icon-wrapper {
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.3);
}

[data-theme="dark"] .chart-card:hover {
  border-color: rgba(59, 130, 246, 0.5) !important;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4),
              0 0 20px rgba(59, 130, 246, 0.2) !important;
}

@media (max-width: 768px) {
  .dashboard-container {
    padding: 16px;
  }
  
  .card-icon-wrapper {
    width: 56px;
    height: 56px;
  }
  
  .card-count {
    font-size: 24px;
  }
}
</style>

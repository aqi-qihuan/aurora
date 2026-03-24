<template>
  <div class="dashboard-page">
    <!-- 问候横幅 -->
    <div class="greeting-banner">
      <div class="greeting-content">
        <div class="greeting-text">
          <h1 class="greeting-title">{{ greetingText }}，{{ userName }}</h1>
          <p class="greeting-desc">{{ greetingDesc }}</p>
        </div>
        <div class="greeting-date">
          <div class="date-day">{{ currentDay }}</div>
          <div class="date-info">
            <span class="date-weekday">{{ currentWeekday }}</span>
            <span class="date-month">{{ currentMonth }}</span>
          </div>
        </div>
      </div>
      <div class="banner-decoration">
        <span class="deco-dot" v-for="i in 5" :key="i" :style="{ '--i': i }"></span>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div
        v-for="(stat, index) in stats"
        :key="stat.title"
        class="stat-card"
        :style="{ '--delay': index * 0.08 + 's' }">
        <div class="stat-bg-icon" :style="{ background: stat.gradient }">
          <el-icon :size="28"><component :is="stat.icon" /></el-icon>
        </div>
        <div class="stat-info">
          <span class="stat-label">{{ stat.title }}</span>
          <span class="stat-value">{{ formatNumber(stat.value) }}</span>
          <span class="stat-trend" :class="{ down: stat.trend < 0 }">
            <el-icon :size="14">
              <Top v-if="stat.trend >= 0" />
              <Bottom v-else />
            </el-icon>
            {{ stat.trend >= 0 ? '+' : '' }}{{ stat.trend }}%
          </span>
        </div>
      </div>
    </div>

    <!-- 一周访问量 -->
    <div class="card">
      <div class="card-header">
        <div class="card-title">
          <span class="title-icon" style="background: var(--primary-light)">
            <el-icon><TrendCharts /></el-icon>
          </span>
          一周访问趋势
        </div>
        <button class="icon-btn" @click="refreshData" title="刷新数据">
          <el-icon><Refresh /></el-icon>
        </button>
      </div>
      <div class="card-body" v-loading="loading">
        <v-chart :option="viewCountOption" autoresize class="chart-full" />
      </div>
    </div>

    <!-- 文章排行 + 分类统计 -->
    <div class="grid-2col">
      <div class="card col-wide">
        <div class="card-header">
          <div class="card-title">
            <span class="title-icon" style="background: var(--warning-light)">
              <el-icon><DataAnalysis /></el-icon>
            </span>
            文章浏览量排行
          </div>
        </div>
        <div class="card-body" v-loading="loading">
          <v-chart :option="articleRankOption" autoresize class="chart-full" />
        </div>
      </div>
      <div class="card col-narrow">
        <div class="card-header">
          <div class="card-title">
            <span class="title-icon" style="background: var(--info-light)">
              <el-icon><PieChart /></el-icon>
            </span>
            分类统计
          </div>
        </div>
        <div class="card-body" v-loading="loading">
          <v-chart :option="categoryOption" autoresize class="chart-full" />
        </div>
      </div>
    </div>

    <!-- 地域分布 + 标签云 -->
    <div class="grid-2col">
      <div class="card col-wide">
        <div class="card-header">
          <div class="card-title">
            <span class="title-icon" style="background: var(--success-light)">
              <el-icon><Location /></el-icon>
            </span>
            用户地域分布
          </div>
          <div class="toggle-group">
            <button
              v-for="opt in [{v:1,l:'用户'},{v:2,l:'游客'}]"
              :key="opt.v"
              :class="['toggle-btn', { active: userType === opt.v }]"
              @click="userType = opt.v">
              {{ opt.l }}
            </button>
          </div>
        </div>
        <div class="card-body" v-loading="mapLoading">
          <v-chart v-if="mapReady" :option="userAreaOption" autoresize class="chart-map" />
          <div v-else class="loading-placeholder">
            <el-icon :size="40" class="spin"><Loading /></el-icon>
            <span>地图数据加载中...</span>
          </div>
        </div>
      </div>
      <div class="card col-narrow">
        <div class="card-header">
          <div class="card-title">
            <span class="title-icon" style="background: rgba(236, 72, 153, 0.08)">
              <el-icon><CollectionTag /></el-icon>
            </span>
            热门标签
          </div>
        </div>
        <div class="card-body tag-cloud-body" v-loading="loading">
          <TagCloud :tags="tagDTOs" v-if="tagDTOs.length > 0" />
          <div v-else class="empty-state">
            <el-icon :size="36" color="var(--text-muted)"><CollectionTag /></el-icon>
            <p>暂无标签数据</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch, onBeforeUnmount } from 'vue'
import {
  Top, Bottom, TrendCharts, DataAnalysis, PieChart, Location, CollectionTag, Refresh,
  User, Document, ChatDotRound, View, Loading
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
import { registerChinaMap } from '@/utils/chinaMap'
import { useAppStore } from '@/stores/app'

// 注册 ECharts 组件
use([
  CanvasRenderer, LineChart, BarChart, EchartsPie, MapChart,
  GridComponent, TooltipComponent, LegendComponent, VisualMapComponent, GeoComponent
])

const appStore = useAppStore()

// 响应式数据
const loading = ref(true)
const mapLoading = ref(false)
const mapReady = ref(false)
const userType = ref(1)
const isDark = ref(false)

// 问候语
const userName = computed(() => appStore.userInfo?.nickname || appStore.userInfo?.username || '管理员')

const greetingText = computed(() => {
  const h = new Date().getHours()
  if (h < 6) return '夜深了'
  if (h < 12) return '上午好'
  if (h < 14) return '中午好'
  if (h < 18) return '下午好'
  return '晚上好'
})

const greetingDesc = computed(() => {
  const descs = [
    '今天也是充满活力的一天',
    '欢迎回到 Aurora 管理后台',
    '祝你有美好的一天',
    '让我们一起管理你的博客世界'
  ]
  const idx = Math.floor(new Date().getDay() + new Date().getHours() / 6) % descs.length
  return descs[idx]
})

const now = new Date()
const currentDay = computed(() => now.getDate())
const currentWeekday = computed(() => ['周日','周一','周二','周三','周四','周五','周六'][now.getDay()])
const currentMonth = computed(() => `${now.getFullYear()}年${now.getMonth() + 1}月`)

// 主题监听
let themeObserver = null
const observeTheme = () => {
  const checkTheme = () => {
    isDark.value = document.documentElement.getAttribute('data-theme') === 'dark'
  }
  checkTheme()
  themeObserver = new MutationObserver(checkTheme)
  themeObserver.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-theme']
  })
}

// 图表颜色
const chartColors = computed(() => {
  if (isDark.value) {
    return {
      textPrimary: '#F8FAFC',
      textSecondary: '#94A3B8',
      textMuted: '#64748B',
      border: '#475569',
      splitLine: '#334155',
      tooltipBg: 'rgba(15, 23, 42, 0.95)',
      tooltipBorder: '#334155',
      tooltipText: '#F8FAFC',
      primary: '#3B82F6',
      primaryLight: '#60A5FA',
      mapArea: '#1E293B',
      mapBorder: '#334155',
      mapEmphasis: '#272F42',
      gradient1: '#60A5FA',
      gradient2: '#2563EB',
      areaOpacity1: 'rgba(59, 130, 246, 0.35)',
      areaOpacity2: 'rgba(59, 130, 246, 0.02)',
      pieColors: ['#3B82F6','#60A5FA','#8B5CF6','#A78BFA','#F97316','#22C55E','#EF4444','#F59E0B','#06B6D4','#EC4899']
    }
  }
  return {
    textPrimary: '#1E293B',
    textSecondary: '#64748B',
    textMuted: '#94A3B8',
    border: '#E2E8F0',
    splitLine: '#F1F5F9',
    tooltipBg: 'rgba(255, 255, 255, 0.96)',
    tooltipBorder: '#E2E8F0',
    tooltipText: '#1E293B',
    primary: '#3B82F6',
    primaryLight: '#93C5FD',
    mapArea: '#F8FAFC',
    mapBorder: '#E2E8F0',
    mapEmphasis: '#EFF6FF',
    gradient1: '#60A5FA',
    gradient2: '#3B82F6',
    areaOpacity1: 'rgba(59, 130, 246, 0.25)',
    areaOpacity2: 'rgba(59, 130, 246, 0.01)',
    pieColors: ['#3B82F6','#60A5FA','#8B5CF6','#A78BFA','#F97316','#10B981','#EF4444','#F59E0B','#06B6D4','#EC4899']
  }
})

// 统计卡片
const stats = ref([
  { title: '访问量', value: 0, icon: View, gradient: 'linear-gradient(135deg, #3B82F6, #2563EB)', trend: 12 },
  { title: '用户量', value: 0, icon: User, gradient: 'linear-gradient(135deg, #8B5CF6, #7C3AED)', trend: 8 },
  { title: '文章量', value: 0, icon: Document, gradient: 'linear-gradient(135deg, #F97316, #EA580C)', trend: 15 },
  { title: '留言量', value: 0, icon: ChatDotRound, gradient: 'linear-gradient(135deg, #10B981, #059669)', trend: -3 }
])

// 标签数据
const tagDTOs = ref([])

// 图表数据
const viewCountData = ref({ xAxis: [], series: [] })
const articleRankData = ref({ xAxis: [], series: [] })
const categoryData = ref([])
const userAreaData = ref([])

// 一周访问量
const viewCountOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'cross' },
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText, fontSize: 13 },
    padding: [10, 14],
    borderRadius: 8
  },
  color: [chartColors.value.primary],
  grid: { left: '2%', right: '3%', bottom: '3%', top: '8%', containLabel: true },
  xAxis: {
    type: 'category',
    data: viewCountData.value.xAxis,
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 12 },
    axisLine: { lineStyle: { color: chartColors.value.border } },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value',
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 12 },
    splitLine: { lineStyle: { color: chartColors.value.splitLine, type: 'dashed' } },
    axisLine: { show: false },
    axisTick: { show: false }
  },
  series: [{
    name: '访问量',
    type: 'line',
    data: viewCountData.value.series,
    smooth: true,
    symbol: 'circle',
    symbolSize: 7,
    lineStyle: { width: 3, color: chartColors.value.primary },
    itemStyle: { color: chartColors.value.primary, borderWidth: 2, borderColor: '#fff' },
    areaStyle: {
      color: {
        type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: chartColors.value.areaOpacity1 },
          { offset: 1, color: chartColors.value.areaOpacity2 }
        ]
      }
    }
  }]
}))

// 文章排行
const articleRankOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText, fontSize: 13 },
    padding: [10, 14],
    borderRadius: 8
  },
  color: [chartColors.value.primaryLight],
  grid: { left: '2%', right: '3%', bottom: '12%', top: '4%', containLabel: true },
  xAxis: {
    type: 'category',
    data: articleRankData.value.xAxis,
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 11, rotate: 25, interval: 0 },
    axisLine: { lineStyle: { color: chartColors.value.border } },
    axisTick: { show: false }
  },
  yAxis: {
    type: 'value',
    axisLabel: { color: chartColors.value.textSecondary, fontSize: 12 },
    splitLine: { lineStyle: { color: chartColors.value.splitLine, type: 'dashed' } },
    axisLine: { show: false },
    axisTick: { show: false }
  },
  series: [{
    name: '浏览量',
    type: 'bar',
    data: articleRankData.value.series,
    barWidth: '55%',
    itemStyle: {
      color: {
        type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
        colorStops: [
          { offset: 0, color: chartColors.value.gradient1 },
          { offset: 1, color: chartColors.value.gradient2 }
        ]
      },
      borderRadius: [6, 6, 0, 0]
    }
  }]
}))

// 分类统计
const categoryOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText, fontSize: 13 },
    padding: [10, 14],
    borderRadius: 8
  },
  legend: {
    orient: 'vertical',
    left: 'left',
    top: 'center',
    textStyle: { color: chartColors.value.textSecondary, fontSize: 12 },
    itemWidth: 10,
    itemHeight: 10,
    itemGap: 12,
    icon: 'circle'
  },
  color: chartColors.value.pieColors,
  series: [{
    name: '文章分类',
    type: 'pie',
    radius: ['45%', '72%'],
    center: ['60%', '50%'],
    data: categoryData.value,
    label: { show: false },
    emphasis: {
      label: { show: true, fontSize: 14, fontWeight: 'bold', color: chartColors.value.textPrimary },
      itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0,0,0,0.15)' }
    },
    itemStyle: { borderRadius: 6, borderColor: 'transparent', borderWidth: 2 }
  }]
}))

// 地域分布
const userAreaOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: chartColors.value.tooltipBg,
    borderColor: chartColors.value.tooltipBorder,
    borderWidth: 1,
    textStyle: { color: chartColors.value.tooltipText, fontSize: 13 },
    padding: [10, 14],
    borderRadius: 8,
    formatter: (p) => p.value ? `${p.name}: ${p.value}人` : `${p.name}: 0人`
  },
  visualMap: {
    min: 0, max: 100, left: 'left', top: 'bottom',
    text: ['高', '低'],
    calculable: true,
    textStyle: { color: chartColors.value.textMuted, fontSize: 11 },
    itemWidth: 12, itemHeight: 100,
    pieces: [
      { gt: 100, label: '100+', color: '#EF4444' },
      { gte: 51, lte: 100, label: '51-100', color: '#22C55E' },
      { gte: 21, lte: 50, label: '21-50', color: '#F59E0B' },
      { gt: 0, lte: 20, label: '1-20', color: '#3B82F6' }
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
      borderColor: chartColors.value.mapBorder,
      borderWidth: 1
    },
    emphasis: {
      itemStyle: {
        areaColor: chartColors.value.mapEmphasis,
        shadowOffsetX: 0, shadowOffsetY: 0, borderWidth: 0
      }
    }
  },
  series: [{
    name: '用户人数',
    type: 'map',
    geoIndex: 0,
    data: userAreaData.value
  }]
}))

// 注册中国地图
const initChinaMap = async () => {
  try {
    await registerChinaMap()
    mapReady.value = true
  } catch {
    mapReady.value = true
  }
}

// 监听用户类型
watch(userType, () => { listUserArea() })

// 获取首页数据
const getData = async () => {
  try {
    loading.value = true
    const { data } = await request.get('/admin')
    if (data?.data) {
      stats.value[0].value = data.data.viewsCount || 0
      stats.value[1].value = data.data.userCount || 0
      stats.value[2].value = data.data.articleCount || 0
      stats.value[3].value = data.data.messageCount || 0

      if (data.data.uniqueViewDTOs?.length) {
        viewCountData.value.xAxis = data.data.uniqueViewDTOs.map(i => i.day)
        viewCountData.value.series = data.data.uniqueViewDTOs.map(i => i.viewsCount)
      }
      if (data.data.categoryDTOs?.length) {
        categoryData.value = data.data.categoryDTOs.map(i => ({
          value: i.articleCount, name: i.categoryName
        }))
      }
      if (data.data.articleRankDTOs?.length) {
        articleRankData.value.xAxis = data.data.articleRankDTOs.map(i => i.articleTitle)
        articleRankData.value.series = data.data.articleRankDTOs.map(i => i.viewsCount)
      }
      if (data.data.tagDTOs?.length) {
        tagDTOs.value = data.data.tagDTOs.map(i => ({ id: i.id, name: i.tagName }))
      }
    }
  } catch { /* ignore */ }
  finally { loading.value = false }
}

// 获取用户地域分布
const listUserArea = async () => {
  try {
    mapLoading.value = true
    const { data } = await request.get('/admin/users/area', { params: { type: userType.value } })
    if (data?.data) {
      const municipalities = ['北京','天津','上海','重庆']
      const specialRegions = ['香港','澳门']
      const autonomousRegions = {
        '内蒙古': '内蒙古自治区', '广西': '广西壮族自治区', '西藏': '西藏自治区',
        '宁夏': '宁夏回族自治区', '新疆': '新疆维吾尔自治区'
      }
      userAreaData.value = data.data.map(item => {
        let name = item.province || item.name || ''
        if (!name.match(/(市|省|自治区|特别行政区)$/)) {
          if (municipalities.includes(name)) name += '市'
          else if (specialRegions.includes(name)) name += '特别行政区'
          else if (autonomousRegions[name]) name = autonomousRegions[name]
          else name += '省'
        }
        return { name, value: item.count || item.value || 0 }
      })
    }
  } catch { /* ignore */ }
  finally { mapLoading.value = false }
}

// 刷新
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

const formatNumber = (num) => (!num ? 0 : num.toLocaleString())

onMounted(() => {
  observeTheme()
  initChinaMap()
  getData()
  listUserArea()
})

onBeforeUnmount(() => {
  themeObserver?.disconnect()
})
</script>

<style scoped>
/* ===== 页面容器 ===== */
.dashboard-page {
  padding: var(--page-padding, 20px);
  max-width: 1440px;
  margin: 0 auto;
  animation: fadeIn 0.5s ease;
}

/* ===== 问候横幅 ===== */
.greeting-banner {
  position: relative;
  background: var(--gradient-primary);
  border-radius: var(--radius-xl, 16px);
  padding: 28px 32px;
  margin-bottom: 24px;
  overflow: hidden;
  color: #fff;
}

.greeting-content {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.greeting-title {
  font-size: 22px;
  font-weight: 700;
  margin: 0 0 6px;
}

.greeting-desc {
  font-size: 14px;
  opacity: 0.85;
  margin: 0;
}

.greeting-date {
  text-align: center;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border-radius: var(--radius-lg, 12px);
  padding: 12px 20px;
  min-width: 80px;
}

.date-day {
  font-size: 32px;
  font-weight: 800;
  line-height: 1.1;
}

.date-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 12px;
  opacity: 0.85;
}

.banner-decoration {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.deco-dot {
  position: absolute;
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
  right: calc(180px - var(--i) * 50px);
  top: calc(-20px + var(--i) * 15px);
}

.deco-dot:nth-child(2) { width: 80px; height: 80px; }
.deco-dot:nth-child(3) { width: 50px; height: 50px; }
.deco-dot:nth-child(4) { width: 30px; height: 30px; }
.deco-dot:nth-child(5) { width: 16px; height: 16px; }

/* ===== 统计卡片 ===== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  background: var(--bg-base);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg, 12px);
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  transition: all 0.3s var(--ease-out, cubic-bezier(0.16, 1, 0.3, 1));
  animation: slideUp 0.5s ease both;
  animation-delay: var(--delay);
  cursor: default;
}

.stat-card:hover {
  border-color: var(--primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-3px);
}

.stat-bg-icon {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg, 12px);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  transition: transform 0.3s var(--ease-spring, cubic-bezier(0.34, 1.56, 0.64, 1));
}

.stat-card:hover .stat-bg-icon {
  transform: scale(1.08) rotate(3deg);
}

.stat-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 800;
  color: var(--text-primary);
  font-family: var(--font-mono, monospace);
  line-height: 1.2;
}

.stat-trend {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  color: var(--success);
  font-weight: 600;
  margin-top: 4px;
}

.stat-trend.down {
  color: var(--danger);
}

/* ===== 通用卡片 ===== */
.card {
  background: var(--bg-base);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg, 12px);
  margin-bottom: 20px;
  transition: all 0.3s var(--ease-out);
}

.card:hover {
  border-color: var(--border-default);
  box-shadow: var(--shadow-sm);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px 0;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.title-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md, 8px);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: var(--primary);
}

.card-body {
  padding: 16px 22px 22px;
  position: relative;
}

/* ===== 双列布局 ===== */
.grid-2col {
  display: grid;
  grid-template-columns: 1.6fr 1fr;
  gap: 20px;
}

.col-wide, .col-narrow {
  margin-bottom: 20px;
}

.col-wide { grid-column: 1; }
.col-narrow { grid-column: 2; }

/* ===== 图表高度 ===== */
.chart-full {
  width: 100%;
  height: 300px;
}

.chart-map {
  width: 100%;
  height: 380px;
}

/* ===== 工具按钮 ===== */
.icon-btn {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-md, 8px);
  border: 1px solid var(--border-default);
  background: var(--bg-surface);
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s var(--ease-out);
}

.icon-btn:hover {
  color: var(--primary);
  border-color: var(--primary);
  background: var(--primary-light);
}

/* ===== 切换按钮组 ===== */
.toggle-group {
  display: flex;
  gap: 0;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md, 8px);
  overflow: hidden;
}

.toggle-btn {
  padding: 5px 16px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--bg-surface);
  border: none;
  cursor: pointer;
  transition: all 0.2s var(--ease-out);
}

.toggle-btn + .toggle-btn {
  border-left: 1px solid var(--border-default);
}

.toggle-btn.active {
  background: var(--primary);
  color: #fff;
}

.toggle-btn:hover:not(.active) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* ===== 标签云容器 ===== */
.tag-cloud-body {
  height: 340px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ===== 空状态 ===== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  height: 100%;
  color: var(--text-muted);
  font-size: 13px;
}

/* ===== 加载占位 ===== */
.loading-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  height: 380px;
  color: var(--text-muted);
  font-size: 13px;
}

.spin {
  animation: spin 1s linear infinite;
}

/* ===== 暗色模式 ===== */
[data-theme="dark"] .greeting-banner {
  background: linear-gradient(135deg, #1E293B 0%, #0F172A 50%, #1E1B4B 100%);
  border: 1px solid rgba(59, 130, 246, 0.15);
}

[data-theme="dark"] .greeting-date {
  background: rgba(255, 255, 255, 0.08);
}

[data-theme="dark"] .stat-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3), 0 0 15px var(--primary-glow);
}

[data-theme="dark"] .card:hover {
  border-color: rgba(59, 130, 246, 0.3);
}

/* ===== 动画 ===== */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ===== 响应式 ===== */
@media (max-width: 1023px) {
  .grid-2col {
    grid-template-columns: 1fr;
  }
  .col-wide, .col-narrow {
    grid-column: 1;
  }
}

@media (max-width: 767px) {
  .dashboard-page {
    padding: var(--page-padding-mobile, 12px);
  }

  .greeting-banner {
    padding: 20px;
    border-radius: var(--radius-lg, 12px);
  }

  .greeting-content {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
  }

  .greeting-title {
    font-size: 18px;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }

  .stat-card {
    padding: 16px;
  }

  .stat-bg-icon {
    width: 44px;
    height: 44px;
  }

  .stat-value {
    font-size: 22px;
  }

  .card-header {
    padding: 14px 16px 0;
  }

  .card-body {
    padding: 12px 16px 16px;
  }

  .chart-map {
    height: 280px;
  }
}
</style>

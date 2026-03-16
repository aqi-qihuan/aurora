<template>
  <div class="dashboard-container">
    <!-- 统计卡片 - 现代化设计 -->
    <el-row :gutter="20" class="stat-row">
      <el-col :xs="12" :sm="12" :md="6" :lg="6" 
              v-for="(stat, index) in stats" 
              :key="stat.title"
              :class="`stagger-${index + 1}`">
        <el-card class="stat-card hover-lift" shadow="hover">
          <div class="card-content">
            <div class="card-icon-wrapper" :style="{ background: stat.gradient }">
              <i :class="stat.icon" />
            </div>
            <div class="card-desc">
              <div class="card-title">{{ stat.title }}</div>
              <div class="card-count">
                <count-to :start-val="0" :end-val="stat.value" :duration="2000" />
              </div>
              <div class="card-trend" :class="{ 'trend-down': stat.trend < 0 }">
                <i :class="stat.trend >= 0 ? 'el-icon-top' : 'el-icon-bottom'" />
                <span>较上周 {{ stat.trend >= 0 ? '+' : '' }}{{ stat.trend }}%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :sm="24" :md="24" :lg="24">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-data-line" />
              一周访问量
            </div>
            <el-button type="text" icon="el-icon-refresh" @click="refreshData" class="refresh-btn">
              刷新
            </el-button>
          </div>
          <div class="chart-container">
            <v-chart :options="viewCount" v-loading="loading" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 文章贡献统计 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :sm="24" :md="24" :lg="24">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-date" />
              文章贡献统计
            </div>
            <el-tooltip content="展示每日文章发布情况" placement="top">
              <i class="el-icon-question info-icon" />
            </el-tooltip>
          </div>
          <div class="heatmap-container" v-loading="loading">
            <calendar-heatmap :end-date="new Date()" :values="articleStatisticsDTOs" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 文章排行和分类 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :sm="24" :md="16" :lg="16">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-s-data" />
              文章浏览量排行
            </div>
          </div>
          <div class="chart-container">
            <v-chart :options="ariticleRank" v-loading="loading" />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="8" :lg="8">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-pie-chart" />
              文章分类统计
            </div>
          </div>
          <div class="chart-container pie-chart">
            <v-chart :options="category" v-loading="loading" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 用户分布和标签 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :xs="24" :sm="24" :md="16" :lg="16">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-map-location" />
              用户地域分布
            </div>
            <div class="area-wrapper">
              <el-radio-group v-model="type" size="small">
                <el-radio-button :label="1">用户</el-radio-button>
                <el-radio-button :label="2">游客</el-radio-button>
              </el-radio-group>
            </div>
          </div>
          <div class="chart-container map-container" v-loading="loading">
            <v-chart :options="userAreaMap" />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="8" :lg="8">
        <el-card class="chart-card hover-lift" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-collection-tag" />
              文章标签统计
            </div>
          </div>
          <div class="chart-container tag-container" v-loading="loading">
            <tag-cloud :data="tagDTOs" />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import '@/assets/js/china'
import { getCurrentTheme, getMapTheme, watchThemeChange } from '@/assets/js/echarts-theme'

export default {
  name: 'Home',
  created() {
    this.listUserArea()
    this.getData()
    this.initThemeWatcher()
  },
  mounted() {
    this.updateChartTheme()
  },
  data() {
    return {
      loading: true,
      type: 1,
      viewsCount: 0,
      messageCount: 0,
      userCount: 0,
      articleCount: 0,
      articleStatisticsDTOs: [],
      tagDTOs: [],
      stats: [
        {
          title: '访问量',
          value: 0,
          icon: 'iconfont el-icon-myfangwenliang',
          gradient: 'linear-gradient(135deg, #3B82F6 0%, #2563EB 100%)',
          trend: 12
        },
        {
          title: '用户量',
          value: 0,
          icon: 'iconfont el-icon-myuser',
          gradient: 'linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%)',
          trend: 8
        },
        {
          title: '文章量',
          value: 0,
          icon: 'iconfont el-icon-mywenzhang-copy',
          gradient: 'linear-gradient(135deg, #F97316 0%, #EA580C 100%)',
          trend: 15
        },
        {
          title: '留言量',
          value: 0,
          icon: 'el-icon-s-comment',
          gradient: 'linear-gradient(135deg, #10B981 0%, #059669 100%)',
          trend: -3
        }
      ],
      viewCount: {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          },
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          borderColor: '#E2E8F0',
          borderWidth: 1,
          textStyle: {
            color: '#1E293B'
          }
        },
        color: ['#2563EB'],
        legend: {
          data: ['访问量'],
          textStyle: {
            color: '#64748B'
          }
        },
        grid: {
          left: '2%',
          right: '2%',
          bottom: '2%',
          top: '12%',
          containLabel: true
        },
        xAxis: {
          data: [],
          axisLine: {
            lineStyle: {
              color: '#E2E8F0'
            }
          },
          axisLabel: {
            color: '#64748B'
          }
        },
        yAxis: {
          axisLine: {
            show: false
          },
          splitLine: {
            lineStyle: {
              color: '#F1F5F9'
            }
          },
          axisLabel: {
            color: '#64748B'
          }
        },
        series: [
          {
            name: '访问量',
            type: 'line',
            data: [],
            smooth: true,
            symbol: 'circle',
            symbolSize: 8,
            lineStyle: {
              width: 3,
              shadowColor: 'rgba(37, 99, 235, 0.3)',
              shadowBlur: 10,
              shadowOffsetY: 5
            },
            areaStyle: {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(37, 99, 235, 0.3)' },
                  { offset: 1, color: 'rgba(37, 99, 235, 0.05)' }
                ]
              }
            }
          }
        ]
      },
      ariticleRank: {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'shadow'
          },
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          borderColor: '#E2E8F0',
          borderWidth: 1,
          textStyle: {
            color: '#1E293B'
          }
        },
        color: ['#3B82F6'],
        grid: {
          left: '2%',
          right: '2%',
          bottom: '5%',
          top: '5%',
          containLabel: true
        },
        xAxis: {
          data: [],
          axisLabel: {
            color: '#64748B',
            interval: 0,
            rotate: 30
          },
          axisLine: {
            lineStyle: {
              color: '#E2E8F0'
            }
          }
        },
        yAxis: {
          axisLine: {
            show: false
          },
          splitLine: {
            lineStyle: {
              color: '#F1F5F9'
            }
          },
          axisLabel: {
            color: '#64748B'
          }
        },
        series: [
          {
            name: '浏览量',
            type: 'bar',
            data: [],
            barWidth: '60%',
            itemStyle: {
              borderRadius: [4, 4, 0, 0],
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: '#3B82F6' },
                  { offset: 1, color: '#60A5FA' }
                ]
              }
            }
          }
        ]
      },
      category: {
        color: ['#2563EB', '#3B82F6', '#60A5FA', '#93C5FD', '#F97316', '#FB923C', '#FDBA74'],
        legend: {
          data: [],
          bottom: 'bottom',
          textStyle: {
            color: '#64748B'
          }
        },
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          borderColor: '#E2E8F0',
          borderWidth: 1,
          textStyle: {
            color: '#1E293B'
          }
        },
        series: [
          {
            name: '文章分类',
            type: 'pie',
            roseType: 'radius',
            radius: ['30%', '70%'],
            center: ['50%', '45%'],
            data: [],
            itemStyle: {
              borderRadius: 6,
              borderColor: '#fff',
              borderWidth: 2
            },
            label: {
              color: '#64748B'
            }
          }
        ]
      },
      userAreaMap: {
        tooltip: {
          formatter: function (e) {
            var value = e.value ? e.value : 0
            return e.seriesName + '<br />' + e.name + '：' + value
          },
          backgroundColor: 'rgba(255, 255, 255, 0.95)',
          borderColor: '#E2E8F0',
          borderWidth: 1,
          textStyle: {
            color: '#1E293B'
          }
        },
        visualMap: {
          min: 0,
          max: 1000,
          right: 26,
          bottom: 40,
          showLabel: true,
          textStyle: {
            color: '#64748B'
          },
          pieces: [
            {
              gt: 100,
              label: '100人以上',
              color: '#DC2626'
            },
            {
              gte: 51,
              lte: 100,
              label: '51-100人',
              color: '#059669'
            },
            {
              gte: 21,
              lte: 50,
              label: '21-50人',
              color: '#D97706'
            },
            {
              label: '1-20人',
              gt: 0,
              lte: 20,
              color: '#2563EB'
            }
          ],
          show: true
        },
        geo: {
          map: 'china',
          zoom: 1.2,
          layoutCenter: ['50%', '50%'],
          itemStyle: {
            normal: {
              borderColor: 'rgba(0, 0, 0, 0.2)',
              areaColor: '#F1F5F9'
            },
            emphasis: {
              areaColor: '#BFDBFE',
              shadowOffsetX: 0,
              shadowOffsetY: 0,
              borderWidth: 0
            }
          }
        },
        series: [
          {
            name: '用户人数',
            type: 'map',
            geoIndex: 0,
            data: [],
            areaColor: '#0FB8F0'
          }
        ]
      }
    }
  },
  watch: {
    type() {
      this.listUserArea()
    }
  },
  methods: {
    getData() {
      this.axios.get('/api/admin').then(({ data }) => {
        this.viewsCount = data.data.viewsCount
        this.messageCount = data.data.messageCount
        this.userCount = data.data.userCount
        this.articleCount = data.data.articleCount
        
        // 更新统计数据
        this.stats[0].value = data.data.viewsCount
        this.stats[1].value = data.data.userCount
        this.stats[2].value = data.data.articleCount
        this.stats[3].value = data.data.messageCount
        
        this.articleStatisticsDTOs = data.data.articleStatisticsDTOs
        if (data.data.uniqueViewDTOs != null) {
          data.data.uniqueViewDTOs.forEach((item) => {
            this.viewCount.xAxis.data.push(item.day)
            this.viewCount.series[0].data.push(item.viewsCount)
          })
        }

        if (data.data.categoryDTOs != null) {
          data.data.categoryDTOs.forEach((item) => {
            this.category.series[0].data.push({
              value: item.articleCount,
              name: item.categoryName
            })
            this.category.legend.data.push(item.categoryName)
          })
        }

        if (data.data.articleRankDTOs != null) {
          data.data.articleRankDTOs.forEach((item) => {
            this.ariticleRank.series[0].data.push(item.viewsCount)
            this.ariticleRank.xAxis.data.push(item.articleTitle)
          })
        }

        if (data.data.tagDTOs != null) {
          data.data.tagDTOs.forEach((item) => {
            this.tagDTOs.push({
              id: item.id,
              name: item.tagName
            })
          })
        }

        this.loading = false
        
        // 数据加载完成后应用主题
        this.$nextTick(() => {
          this.updateChartTheme()
        })
      })
    },
    listUserArea() {
      this.axios
        .get('/api/admin/users/area', {
          params: {
            type: this.type
          }
        })
        .then(({ data }) => {
          this.userAreaMap.series[0].data = data.data
        })
    },
    refreshData() {
      this.loading = true
      // 清空数据
      this.viewCount.xAxis.data = []
      this.viewCount.series[0].data = []
      this.category.series[0].data = []
      this.category.legend.data = []
      this.ariticleRank.series[0].data = []
      this.ariticleRank.xAxis.data = []
      this.tagDTOs = []
      
      this.getData()
      this.listUserArea()
    },
    initThemeWatcher() {
      this.themeObserver = watchThemeChange((isDark) => {
        this.updateChartTheme()
      })
    },
    updateChartTheme() {
      const theme = getCurrentTheme()
      const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
      
      // 更新访问量图表主题
      this.viewCount = {
        ...this.viewCount,
        tooltip: {
          ...this.viewCount.tooltip,
          backgroundColor: theme.tooltip.backgroundColor,
          borderColor: theme.tooltip.borderColor,
          textStyle: { color: theme.tooltip.textStyle.color }
        },
        legend: {
          ...this.viewCount.legend,
          textStyle: { color: theme.legend.textStyle.color }
        },
        xAxis: {
          ...this.viewCount.xAxis,
          axisLine: { lineStyle: { color: theme.xAxis.axisLine.lineStyle.color } },
          axisLabel: { color: theme.xAxis.axisLabel.color }
        },
        yAxis: {
          ...this.viewCount.yAxis,
          splitLine: { lineStyle: { color: theme.yAxis.splitLine.lineStyle.color } },
          axisLabel: { color: theme.yAxis.axisLabel.color }
        }
      }
      
      // 更新文章排行图表主题
      this.ariticleRank = {
        ...this.ariticleRank,
        tooltip: {
          ...this.ariticleRank.tooltip,
          backgroundColor: theme.tooltip.backgroundColor,
          borderColor: theme.tooltip.borderColor,
          textStyle: { color: theme.tooltip.textStyle.color }
        },
        xAxis: {
          ...this.ariticleRank.xAxis,
          axisLine: { lineStyle: { color: theme.xAxis.axisLine.lineStyle.color } },
          axisLabel: { color: theme.xAxis.axisLabel.color }
        },
        yAxis: {
          ...this.ariticleRank.yAxis,
          splitLine: { lineStyle: { color: theme.yAxis.splitLine.lineStyle.color } },
          axisLabel: { color: theme.yAxis.axisLabel.color }
        }
      }
      
      // 更新分类图表主题
      this.category = {
        ...this.category,
        legend: {
          ...this.category.legend,
          textStyle: { color: theme.legend.textStyle.color }
        },
        tooltip: {
          ...this.category.tooltip,
          backgroundColor: theme.tooltip.backgroundColor,
          borderColor: theme.tooltip.borderColor,
          textStyle: { color: theme.tooltip.textStyle.color }
        },
        series: [{
          ...this.category.series[0],
          itemStyle: {
            ...this.category.series[0].itemStyle,
            borderColor: isDark ? '#1E293B' : '#fff'
          },
          label: {
            color: theme.textStyle.color
          }
        }]
      }
      
      // 更新地图主题
      const mapTheme = getMapTheme(isDark)
      this.userAreaMap = {
        ...this.userAreaMap,
        tooltip: {
          ...this.userAreaMap.tooltip,
          backgroundColor: mapTheme.tooltip.backgroundColor,
          borderColor: mapTheme.tooltip.borderColor,
          textStyle: { color: mapTheme.tooltip.textStyle.color }
        },
        visualMap: {
          ...this.userAreaMap.visualMap,
          textStyle: { color: mapTheme.visualMap.textStyle.color }
        },
        geo: {
          ...this.userAreaMap.geo,
          itemStyle: mapTheme.geo.itemStyle
        }
      }
    }
  },
  beforeDestroy() {
    // 清理主题监听器
    if (this.themeObserver) {
      this.themeObserver.disconnect()
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  padding: var(--space-6);
  animation: fadeInUp 0.6s var(--ease-out);
}

/* 移动端padding优化 */
@media (max-width: 768px) {
  .dashboard-container {
    padding: var(--space-4);
  }
}

/* 统计卡片样式 - 现代化设计 */
.stat-row {
  margin-bottom: var(--space-4);
}

.stat-card {
  margin-bottom: var(--space-5);
  border: none;
}

.stat-card :deep(.el-card__body) {
  padding: var(--space-5);
}

.card-content {
  display: flex;
  align-items: center;
}

.card-icon-wrapper {
  width: 72px;
  height: 72px;
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: var(--space-5);
  box-shadow: var(--shadow-md);
  transition: all var(--duration-base) var(--ease-spring);
}

.card-icon-wrapper i {
  font-size: 2rem;
  color: #fff;
}

.stat-card:hover .card-icon-wrapper {
  transform: scale(1.1) rotate(5deg);
  box-shadow: var(--shadow-lg);
}

/* 移动端图标缩小 */
@media (max-width: 768px) {
  .card-icon-wrapper {
    width: 56px;
    height: 56px;
    margin-right: var(--space-3);
  }

  .card-icon-wrapper i {
    font-size: 1.5rem;
  }

  .card-count {
    font-size: var(--text-xl) !important;
  }

  .card-title {
    font-size: var(--text-xs) !important;
  }

  .card-trend {
    font-size: 11px !important;
  }
}

.card-desc {
  flex: 1;
}

.card-title {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
  margin-bottom: var(--space-2);
  font-weight: var(--font-medium);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.card-count {
  font-size: var(--text-2xl);
  font-weight: var(--font-bold);
  color: var(--color-text);
  margin-bottom: var(--space-2);
  font-family: var(--font-sans);
  line-height: var(--leading-tight);
}

.card-trend {
  font-size: var(--text-xs);
  color: var(--color-success);
  display: flex;
  align-items: center;
  font-weight: var(--font-medium);
}

.card-trend i {
  margin-right: var(--space-1);
}

.card-trend.trend-down {
  color: var(--color-error);
}

/* 图表卡片样式 */
.chart-row {
  margin-bottom: var(--space-4);
}

.chart-card {
  margin-bottom: var(--space-5);
  border: none;
}

.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-5);
  padding-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border);
}

.chart-title {
  font-size: var(--text-lg);
  font-weight: var(--font-semibold);
  color: var(--color-text);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.chart-title i {
  font-size: var(--text-xl);
  color: var(--color-primary);
}

.info-icon {
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color var(--duration-fast) var(--ease-out);
}

.info-icon:hover {
  color: var(--color-primary);
}

.refresh-btn {
  color: var(--color-primary) !important;
  font-weight: var(--font-medium);
}

.refresh-btn:hover {
  color: var(--color-primary-light) !important;
}

.area-wrapper {
  display: flex;
  align-items: center;
}

.chart-container {
  height: 350px;
  position: relative;
  width: 100%;
}

.chart-container.pie-chart {
  height: 320px;
}

/* 移动端图表高度优化 */
@media (max-width: 768px) {
  .chart-container {
    height: 280px;
  }

  .chart-container.pie-chart {
    height: 260px;
  }

  .chart-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--space-3);
  }

  .chart-title {
    font-size: var(--text-base);
  }
}

.map-container {
  height: 450px;
}

@media (max-width: 768px) {
  .map-container {
    height: 350px;
  }
}

.heatmap-container {
  padding: var(--space-4) 0;
  min-height: 200px;
}

.tag-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-5);
  min-height: 300px;
}

/* ECharts样式优化 */
:deep(.echarts) {
  width: 100%;
  height: 100%;
}

/* 动画延迟类 */
.stagger-1 { animation-delay: 0.1s; }
.stagger-2 { animation-delay: 0.2s; }
.stagger-3 { animation-delay: 0.3s; }
.stagger-4 { animation-delay: 0.4s; }

/* 悬停提升效果 */
.hover-lift {
  transition: transform var(--duration-base) var(--ease-out),
              box-shadow var(--duration-base) var(--ease-out);
}

.hover-lift:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-hover) !important;
}

/* 淡入动画 */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>

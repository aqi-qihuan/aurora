<template>
  <div class="dashboard-container">
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :xs="12" :sm="12" :md="6" :lg="6">
        <el-card class="stat-card" shadow="hover">
          <div class="card-content">
            <div class="card-icon-wrapper" style="background: linear-gradient(135deg, #36d1dc 0%, #5b86e5 100%)">
              <i class="iconfont el-icon-myfangwenliang" />
            </div>
            <div class="card-desc">
              <div class="card-title">访问量</div>
              <div class="card-count">
                <count-to :start-val="0" :end-val="viewsCount" :duration="2000" />
              </div>
              <div class="card-trend">
                <i class="el-icon-top" />
                <span>较上周 +12%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6" :lg="6">
        <el-card class="stat-card" shadow="hover">
          <div class="card-content">
            <div class="card-icon-wrapper" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)">
              <i class="iconfont el-icon-myuser" />
            </div>
            <div class="card-desc">
              <div class="card-title">用户量</div>
              <div class="card-count">
                <count-to :start-val="0" :end-val="userCount" :duration="2000" />
              </div>
              <div class="card-trend">
                <i class="el-icon-top" />
                <span>较上周 +8%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6" :lg="6">
        <el-card class="stat-card" shadow="hover">
          <div class="card-content">
            <div class="card-icon-wrapper" style="background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)">
              <i class="iconfont el-icon-mywenzhang-copy" />
            </div>
            <div class="card-desc">
              <div class="card-title">文章量</div>
              <div class="card-count">
                <count-to :start-val="0" :end-val="articleCount" :duration="2000" />
              </div>
              <div class="card-trend">
                <i class="el-icon-top" />
                <span>较上周 +15%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6" :lg="6">
        <el-card class="stat-card" shadow="hover">
          <div class="card-content">
            <div class="card-icon-wrapper" style="background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)">
              <i class="el-icon-s-comment" />
            </div>
            <div class="card-desc">
              <div class="card-title">留言量</div>
              <div class="card-count">
                <count-to :start-val="0" :end-val="messageCount" :duration="2000" />
              </div>
              <div class="card-trend">
                <i class="el-icon-bottom" />
                <span>较上周 -3%</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20">
      <el-col :xs="24" :sm="24" :md="24" :lg="24">
        <el-card class="chart-card" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-data-line" />
              一周访问量
            </div>
            <el-button type="text" icon="el-icon-refresh" @click="refreshData">刷新</el-button>
          </div>
          <div class="chart-container">
            <v-chart :options="viewCount" v-loading="loading" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20">
      <el-col :xs="24" :sm="24" :md="24" :lg="24">
        <el-card class="chart-card" shadow="hover">
          <div class="chart-header">
            <div class="chart-title">
              <i class="el-icon-date" />
              文章贡献统计
            </div>
            <el-tooltip content="展示每日文章发布情况" placement="top">
              <i class="el-icon-question" />
            </el-tooltip>
          </div>
          <div class="heatmap-container" v-loading="loading">
            <calendar-heatmap :end-date="new Date()" :values="articleStatisticsDTOs" />
          </div>
        </el-card>
      </el-col>
    </el-row>

  <el-row :gutter="20">
    <el-col :xs="24" :sm="24" :md="16" :lg="16">
      <el-card>
        <div class="e-title">文章浏览量排行</div>
        <div class="chart-container">
          <v-chart :options="ariticleRank" v-loading="loading" />
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="24" :md="8" :lg="8">
      <el-card>
        <div class="e-title">文章分类统计</div>
        <div class="chart-container">
          <v-chart :options="category" v-loading="loading" />
        </div>
      </el-card>
    </el-col>
  </el-row>

  <el-row :gutter="20">
    <el-col :xs="24" :sm="24" :md="16" :lg="16">
      <el-card>
        <div class="e-title">用户地域分布</div>
        <div class="chart-container" v-loading="loading">
          <div class="area-wrapper">
            <el-radio-group v-model="type">
              <el-radio :label="1">用户</el-radio>
              <el-radio :label="2">游客</el-radio>
            </el-radio-group>
          </div>
          <v-chart :options="userAreaMap" />
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="24" :md="8" :lg="8">
      <el-card>
        <div class="e-title">文章标签统计</div>
        <div class="chart-container" v-loading="loading">
          <tag-cloud :data="tagDTOs" />
        </div>
      </el-card>
    </el-col>
  </el-row>
</div>
</template>

<script>
import '@/assets/js/china'
export default {
  created() {
    this.listUserArea()
    this.getData()
  },
  data: function () {
    return {
      loading: true,
      type: 1,
      viewsCount: 0,
      messageCount: 0,
      userCount: 0,
      articleCount: 0,
      articleStatisticsDTOs: [],
      tagDTOs: [],
      viewCount: {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        color: ['#3888fa'],
        legend: {
          data: ['访问量']
        },
        grid: {
          left: '0%',
          right: '0%',
          bottom: '0%',
          top: '10%',
          containLabel: true
        },
        xAxis: {
          data: [],
          axisLine: {
            lineStyle: {
              color: '#666'
            }
          }
        },
        yAxis: {
          axisLine: {
            lineStyle: {
              color: '#048CCE'
            }
          }
        },
        series: [
          {
            name: '访问量',
            type: 'line',
            data: [],
            smooth: true
          }
        ]
      },
      ariticleRank: {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross'
          }
        },
        color: ['#58AFFF'],
        grid: {
          left: '0%',
          right: '0%',
          bottom: '0%',
          top: '10%',
          containLabel: true
        },
        xAxis: {
          data: []
        },
        yAxis: {},
        series: [
          {
            name: ['浏览量'],
            type: 'bar',
            data: []
          }
        ]
      },
      category: {
        color: ['#7EC0EE', '#FF9F7F', '#FFD700', '#C9C9C9', '#E066FF', '#36dc59', '#C0FF3E'],
        legend: {
          data: [],
          bottom: 'bottom'
        },
        tooltip: {
          trigger: 'item'
        },
        series: [
          {
            name: '文章分类',
            type: 'pie',
            roseType: 'radius',
            data: []
          }
        ]
      },
      userAreaMap: {
        tooltip: {
          formatter: function (e) {
            var value = e.value ? e.value : 0
            return e.seriesName + '<br />' + e.name + '：' + value
          }
        },
        visualMap: {
          min: 0,
          max: 1000,
          right: 26,
          bottom: 40,
          showLabel: !0,
          pieces: [
            {
              gt: 100,
              label: '100人以上',
              color: '#ED5351'
            },
            {
              gte: 51,
              lte: 100,
              label: '51-100人',
              color: '#59D9A5'
            },
            {
              gte: 21,
              lte: 50,
              label: '21-50人',
              color: '#F6C021'
            },
            {
              label: '1-20人',
              gt: 0,
              lte: 20,
              color: '#6DCAEC'
            }
          ],
          show: !0
        },
        geo: {
          map: 'china',
          zoom: 1.2,
          layoutCenter: ['50%', '50%'],
          itemStyle: {
            normal: {
              borderColor: 'rgba(0, 0, 0, 0.2)'
            },
            emphasis: {
              areaColor: '#F5DEB3',
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
  methods: {
    getData() {
      this.axios.get('/api/admin').then(({ data }) => {
        this.viewsCount = data.data.viewsCount
        this.messageCount = data.data.messageCount
        this.userCount = data.data.userCount
        this.articleCount = data.data.articleCount
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
      this.getData()
      this.listUserArea()
    }
  },
  watch: {
    type() {
      this.listUserArea()
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}

/* 移动端padding优化 */
@media (max-width: 768px) {
  .dashboard-container {
    padding: 12px;
  }
}

/* 统计卡片样式 */
.stat-card {
  transition: all 0.3s ease;
  border: none;
  margin-bottom: 20px;
}

.stat-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 20px rgba(0, 0, 0, 0.12) !important;
}

.card-content {
  display: flex;
  align-items: center;
  padding: 10px;
}

.card-icon-wrapper {
  width: 80px;
  height: 80px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

/* 移动端图标缩小 */
@media (max-width: 768px) {
  .card-icon-wrapper {
    width: 60px;
    height: 60px;
    margin-right: 12px;
  }

  .card-icon-wrapper i {
    font-size: 2rem;
  }

  .card-count {
    font-size: 24px;
  }

  .card-title {
    font-size: 12px;
  }

  .card-trend {
    font-size: 11px;
  }
}

.card-icon-wrapper i {
  font-size: 2.5rem;
  color: #fff;
}

.stat-card:hover .card-icon-wrapper {
  transform: scale(1.1) rotate(5deg);
}

.card-desc {
  flex: 1;
}

.card-title {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
  margin-bottom: 8px;
  font-weight: 500;
}

.card-count {
  font-size: 28px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
}

.card-trend {
  font-size: 12px;
  color: #67c23a;
  display: flex;
  align-items: center;
}

.card-trend i {
  margin-right: 4px;
}

.card-trend.el-icon-bottom {
  color: #f56c6c;
}

/* 图表卡片样式 */
.chart-card {
  transition: all 0.3s ease;
  border: none;
  margin-bottom: 20px;
}

.chart-card:hover {
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1) !important;
}

.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #f0f0f0;
}

.chart-title {
  font-size: 16px;
  font-weight: bold;
  color: #202a34;
  display: flex;
  align-items: center;
}

.chart-title i {
  margin-right: 8px;
  font-size: 18px;
  color: #409eff;
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

/* 移动端图表高度优化 */
@media (max-width: 768px) {
  .chart-container {
    height: 280px;
  }

  .chart-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .chart-title {
    font-size: 14px;
  }
}

.map-container {
  height: 450px;
}

.heatmap-container {
  padding: 20px 0;
  min-height: 200px;
}

.tag-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

/* 优化滚动条 */
.dashboard-container ::v-deep .el-scrollbar__wrap {
  overflow-x: hidden;
}

/* 动画效果 */
.dashboard-container {
  animation: fadeIn 0.6s ease-in;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* ECharts响应式 */
.echarts {
  width: 100%;
  height: 100%;
}
</style>

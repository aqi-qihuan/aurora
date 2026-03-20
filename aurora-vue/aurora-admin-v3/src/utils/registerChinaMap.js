// 动态注册中国地图数据
import * as echarts from 'echarts'

// 省份坐标映射
export const provinceCoordinates = {
  '北京': [116.4, 39.9],
  '天津': [117.2, 39.1],
  '河北': [114.5, 38.0],
  '山西': [112.5, 37.8],
  '内蒙古': [111.7, 40.8],
  '辽宁': [123.4, 41.8],
  '吉林': [125.3, 43.9],
  '黑龙江': [126.6, 45.8],
  '上海': [121.5, 31.2],
  '江苏': [118.8, 32.1],
  '浙江': [120.2, 30.3],
  '安徽': [117.3, 31.9],
  '福建': [119.3, 26.1],
  '江西': [115.9, 28.7],
  '山东': [117.0, 36.7],
  '河南': [113.7, 34.8],
  '湖北': [114.3, 30.6],
  '湖南': [113.0, 28.2],
  '广东': [113.3, 23.1],
  '广西': [108.3, 22.8],
  '海南': [110.3, 20.0],
  '重庆': [106.5, 29.6],
  '四川': [104.1, 30.7],
  '贵州': [106.7, 26.6],
  '云南': [102.7, 25.0],
  '西藏': [91.1, 29.6],
  '陕西': [109.0, 34.3],
  '甘肃': [103.8, 36.1],
  '青海': [101.8, 36.6],
  '宁夏': [106.3, 38.5],
  '新疆': [87.6, 43.8],
  '台湾': [121.5, 25.0],
  '香港': [114.2, 22.3],
  '澳门': [113.5, 22.2]
}

// 中国地图 GeoJSON 数据（完整版，从静态文件加载）
let chinaMapRegistered = false

export async function registerChinaMap() {
  if (chinaMapRegistered) return
  
  try {
    // 使用动态 import 加载 china.js
    // china.js 会自动调用 echarts.registerMap
    await import('@/assets/js/china.js')
    chinaMapRegistered = true
    console.log('✅ 中国地图注册成功')
  } catch (error) {
    console.warn('⚠️ 完整地图加载失败，使用简化版:', error.message)
    // 如果加载失败，注册简化版
    registerSimplifiedChinaMap()
  }
}

// 注册简化版中国地图（省份多边形）
function registerSimplifiedChinaMap() {
  const features = Object.entries(provinceCoordinates).map(([name, coords]) => ({
    type: 'Feature',
    properties: { name },
    geometry: {
      type: 'Point',
      coordinates: coords
    }
  }))
  
  echarts.registerMap('china', {
    type: 'FeatureCollection',
    features
  })
  chinaMapRegistered = true
  console.log('✅ 简化版中国地图注册成功')
}

export default { registerChinaMap, provinceCoordinates }


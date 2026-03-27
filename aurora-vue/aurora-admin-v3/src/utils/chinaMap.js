/**
 * 中国地图 ES Module
 * 优先使用本地打包的地图数据，确保云端可用
 * 
 * 数据源: https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json
 */

import * as echarts from 'echarts'
// 直接导入本地地图数据 JSON (会被 Vite 打包)
import localChinaMapData from '@/assets/json/china-map.json'

// 省份坐标映射（用于散点图）
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

// 地图数据缓存 (用于远程加载的缓存)
let remoteMapDataCache = null
let chinaMapRegistered = false

/**
 * 从阿里云 DataV 加载中国地图 GeoJSON 数据
 * @returns {Promise<Object>} GeoJSON 数据
 */
async function fetchChinaMapData() {
  if (remoteMapDataCache) {
    return remoteMapDataCache
  }
  
  try {
    const response = await fetch('https://geo.datav.aliyun.com/areas_v3/bound/100000_full.json')
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    remoteMapDataCache = await response.json()
    return remoteMapDataCache
  } catch (error) {
    console.error('加载中国地图数据失败:', error)
    throw error
  }
}

/**
 * 注册中国地图到 ECharts
 * 直接使用本地打包的 JSON 数据，确保云端可用
 * @returns {Promise<boolean>} 是否注册成功
 */
export async function registerChinaMap() {
  if (chinaMapRegistered) {
    return true
  }
  
  try {
    // 直接使用导入的本地地图数据 (优先，确保云端可用)
    echarts.registerMap('china', localChinaMapData)
    chinaMapRegistered = true
    console.log('✅ 中国地图注册成功 (本地数据)')
    return true
  } catch (localError) {
    console.warn('⚠️ 本地地图加载失败，尝试在线加载:', localError.message)
    
    try {
      // 降级方案: 从在线加载完整的省份边界数据
      const geoJSON = await fetchChinaMapData()
      
      // 注册到 ECharts
      echarts.registerMap('china', geoJSON)
      chinaMapRegistered = true
      
      console.log('✅ 中国地图注册成功 - 在线数据')
      return true
    } catch (onlineError) {
      console.warn('⚠️ 在线地图加载失败，使用简化版:', onlineError.message)
      
      // 最终降级: 使用简化版地图
      registerSimplifiedChinaMap()
      return true
    }
  }
}

/**
 * 注册简化版中国地图（省份多边形）
 * 用于网络失败时的降级方案
 */
function registerSimplifiedChinaMap() {
  const features = Object.entries(provinceCoordinates).map(([name, coords]) => ({
    type: 'Feature',
    properties: { 
      name,
      cp: coords
    },
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

/**
 * 获取省份坐标
 * @param {string} provinceName 省份名称
 * @returns {Array<number>} [经度, 纬度]
 */
export function getProvinceCoordinate(provinceName) {
  // 处理省份名称（移除"省"、"市"、"自治区"等后缀）
  const normalizedName = provinceName
    .replace(/省|市|自治区|特别行政区|壮族|回族|维吾尔/g, '')
  
  // 直接匹配
  if (provinceCoordinates[provinceName]) {
    return provinceCoordinates[provinceName]
  }
  
  // 标准化后匹配
  if (provinceCoordinates[normalizedName]) {
    return provinceCoordinates[normalizedName]
  }
  
  // 模糊匹配
  for (const [name, coords] of Object.entries(provinceCoordinates)) {
    if (name.includes(normalizedName) || normalizedName.includes(name)) {
      return coords
    }
  }
  
  console.warn(`未找到省份坐标: ${provinceName}`)
  return null
}

/**
 * 将用户地域数据转换为散点图数据
 * @param {Array<{province: string, count: number}>} areaData 用户地域数据
 * @returns {Array<{name: string, value: Array<number|number>}>} 散点图数据
 */
export function convertToScatterData(areaData) {
  return areaData
    .map(item => {
      const coords = getProvinceCoordinate(item.province)
      if (coords) {
        return {
          name: item.province,
          value: [...coords, item.count]
        }
      }
      return null
    })
    .filter(Boolean)
}

export default {
  registerChinaMap,
  provinceCoordinates,
  getProvinceCoordinate,
  convertToScatterData
}

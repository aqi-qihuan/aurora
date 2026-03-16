/**
 * ECharts 主题配置
 * 支持浅色和深色模式
 */

// 浅色模式主题
export const lightTheme = {
  // 背景颜色
  backgroundColor: 'transparent',
  
  // 颜色板
  color: ['#2563EB', '#3B82F6', '#60A5FA', '#93C5FD', '#F97316', '#FB923C', '#FDBA74', '#10B981', '#34D399', '#6EE7B7'],
  
  // 文本样式
  textStyle: {
    color: '#1E293B'
  },
  
  // 标题
  title: {
    textStyle: {
      color: '#1E293B'
    },
    subtextStyle: {
      color: '#64748B'
    }
  },
  
  // 图例
  legend: {
    textStyle: {
      color: '#64748B'
    }
  },
  
  // 提示框
  tooltip: {
    backgroundColor: 'rgba(255, 255, 255, 0.95)',
    borderColor: '#E2E8F0',
    borderWidth: 1,
    textStyle: {
      color: '#1E293B'
    },
    extraCssText: 'box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); border-radius: 8px;'
  },
  
  // 坐标轴
  xAxis: {
    axisLine: {
      lineStyle: {
        color: '#E2E8F0'
      }
    },
    axisLabel: {
      color: '#64748B'
    },
    splitLine: {
      lineStyle: {
        color: '#F1F5F9'
      }
    }
  },
  
  yAxis: {
    axisLine: {
      lineStyle: {
        color: '#E2E8F0'
      }
    },
    axisLabel: {
      color: '#64748B'
    },
    splitLine: {
      lineStyle: {
        color: '#F1F5F9'
      }
    }
  }
}

// 深色模式主题
export const darkTheme = {
  // 背景颜色
  backgroundColor: 'transparent',
  
  // 颜色板 - 更鲜艳的配色以适应深色背景
  color: ['#3B82F6', '#60A5FA', '#93C5FD', '#F97316', '#FB923C', '#FDBA74', '#10B981', '#34D399', '#6EE7B7', '#A78BFA'],
  
  // 文本样式
  textStyle: {
    color: '#F8FAFC'
  },
  
  // 标题
  title: {
    textStyle: {
      color: '#F8FAFC'
    },
    subtextStyle: {
      color: '#94A3B8'
    }
  },
  
  // 图例
  legend: {
    textStyle: {
      color: '#94A3B8'
    }
  },
  
  // 提示框
  tooltip: {
    backgroundColor: 'rgba(30, 41, 59, 0.95)',
    borderColor: '#334155',
    borderWidth: 1,
    textStyle: {
      color: '#F8FAFC'
    },
    extraCssText: 'box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3); border-radius: 8px;'
  },
  
  // 坐标轴
  xAxis: {
    axisLine: {
      lineStyle: {
        color: '#334155'
      }
    },
    axisLabel: {
      color: '#94A3B8'
    },
    splitLine: {
      lineStyle: {
        color: '#1E293B'
      }
    }
  },
  
  yAxis: {
    axisLine: {
      lineStyle: {
        color: '#334155'
      }
    },
    axisLabel: {
      color: '#94A3B8'
    },
    splitLine: {
      lineStyle: {
        color: '#1E293B'
      }
    }
  }
}

/**
 * 获取当前主题配置
 * @returns {Object} 当前主题配置
 */
export function getCurrentTheme() {
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  return isDark ? darkTheme : lightTheme
}

/**
 * 应用主题到图表配置
 * @param {Object} option - ECharts 配置
 * @param {Object} theme - 主题配置
 * @returns {Object} 合并后的配置
 */
export function applyTheme(option, theme = null) {
  const currentTheme = theme || getCurrentTheme()
  
  return {
    ...option,
    backgroundColor: currentTheme.backgroundColor,
    textStyle: {
      ...currentTheme.textStyle,
      ...option.textStyle
    }
  }
}

/**
 * 为地图图表获取主题配置
 * @param {boolean} isDark - 是否为深色模式
 * @returns {Object} 地图主题配置
 */
export function getMapTheme(isDark = false) {
  if (isDark) {
    return {
      tooltip: {
        backgroundColor: 'rgba(30, 41, 59, 0.95)',
        borderColor: '#334155',
        textStyle: {
          color: '#F8FAFC'
        }
      },
      visualMap: {
        textStyle: {
          color: '#94A3B8'
        }
      },
      geo: {
        itemStyle: {
          normal: {
            borderColor: 'rgba(255, 255, 255, 0.2)',
            areaColor: '#334155'
          },
          emphasis: {
            areaColor: '#3B82F6',
            shadowOffsetX: 0,
            shadowOffsetY: 0,
            borderWidth: 0
          }
        }
      }
    }
  }
  
  return {
    tooltip: {
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#E2E8F0',
      textStyle: {
        color: '#1E293B'
      }
    },
    visualMap: {
      textStyle: {
        color: '#64748B'
      }
    },
    geo: {
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
    }
  }
}

/**
 * 监听主题变化
 * @param {Function} callback - 主题变化回调函数
 */
export function watchThemeChange(callback) {
  // 使用 MutationObserver 监听 data-theme 属性变化
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'attributes' && mutation.attributeName === 'data-theme') {
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
        callback(isDark)
      }
    })
  })
  
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-theme']
  })
  
  return observer
}

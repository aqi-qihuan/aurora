/**
 * 日志工具类
 * 开发环境输出日志，生产环境静默
 */

const isDev = import.meta.env.DEV

export const logger = {
  log: (...args) => isDev && console.log(...args),
  warn: (...args) => isDev && console.warn(...args),
  error: (...args) => console.error(...args), // 错误始终输出
  debug: (...args) => isDev && console.debug(...args),
  info: (...args) => isDev && console.info(...args),
  
  // 分组日志
  group: (label) => isDev && console.group(label),
  groupEnd: () => isDev && console.groupEnd(),
  
  // 表格日志
  table: (data) => isDev && console.table(data),
  
  // 时间追踪
  time: (label) => isDev && console.time(label),
  timeEnd: (label) => isDev && console.timeEnd(label)
}

export default logger

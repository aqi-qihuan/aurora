/**
 * 日志工具类
 * 开发环境输出日志，生产环境静默。
 */

const isDev: boolean = import.meta.env.DEV

export const logger = {
  log: (...args: unknown[]): void => { isDev && console.log(...args) },
  warn: (...args: unknown[]): void => { isDev && console.warn(...args) },
  error: (...args: unknown[]): void => { console.error(...args) }, // 错误始终输出
  debug: (...args: unknown[]): void => { isDev && console.debug(...args) },
  info: (...args: unknown[]): void => { isDev && console.info(...args) },

  // 分组日志
  group: (label: string): void => { isDev && console.group(label) },
  groupEnd: (): void => { isDev && console.groupEnd() },

  // 表格日志
  table: (data: unknown): void => { isDev && console.table(data) },

  // 时间追踪
  time: (label: string): void => { isDev && console.time(label) },
  timeEnd: (label: string): void => { isDev && console.timeEnd(label) }
}

export default logger

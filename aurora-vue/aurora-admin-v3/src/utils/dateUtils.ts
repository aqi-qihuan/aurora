import dayjs from 'dayjs'

/**
 * 格式化日期为 YYYY-MM-DD
 * @param date - 日期（字符串、时间戳或 Date 对象）
 * @returns 格式化后的日期字符串
 */
export const formatDate = (date: string | number | Date | null | undefined): string => {
  return date ? dayjs(date).format('YYYY-MM-DD') : ''
}

/**
 * 格式化日期时间为 YYYY-MM-DD HH:mm:ss
 * @param date - 日期（字符串、时间戳或 Date 对象）
 * @returns 格式化后的日期时间字符串
 */
export const formatDateTime = (date: string | number | Date | null | undefined): string => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : '-'
}

export default {
  formatDate,
  formatDateTime
}

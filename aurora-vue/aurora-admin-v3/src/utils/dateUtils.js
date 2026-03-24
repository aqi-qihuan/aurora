import dayjs from 'dayjs'

/**
 * 格式化日期为 YYYY-MM-DD
 */
export const formatDate = (date) => date ? dayjs(date).format('YYYY-MM-DD') : ''

/**
 * 格式化日期时间为 YYYY-MM-DD HH:mm:ss
 */
export const formatDateTime = (date) => date ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : '-'

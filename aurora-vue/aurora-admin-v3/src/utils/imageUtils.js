import * as imageConversion from 'image-conversion'

/**
 * 创建图片上传前压缩处理器
 * @param {number} maxSizeKB - 最大文件大小(KB)，默认500KB
 * @returns {Function} el-upload 的 beforeUpload 处理函数
 */
export const createBeforeUploadHandler = (maxSizeKB = 500) => {
  return (file) => {
    return new Promise((resolve) => {
      if (file.size / 1024 < maxSizeKB) {
        resolve(file)
      } else {
        imageConversion.compressAccurately(file, maxSizeKB).then(resolve)
      }
    })
  }
}

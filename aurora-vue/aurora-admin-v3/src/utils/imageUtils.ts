import * as imageConversion from 'image-conversion'

/**
 * 创建图片上传前压缩处理器
 * @param maxSizeKB - 最大文件大小(KB)，默认 500KB
 * @returns el-upload 的 beforeUpload 处理函数
 */
export function createBeforeUploadHandler(maxSizeKB = 500): any {
  return (file: any) => {
    return new Promise((resolve) => {
      if (file.size / 1024 < maxSizeKB) {
        resolve(file)
      } else {
        imageConversion.compressAccurately(file, maxSizeKB).then(resolve)
      }
    })
  }
}

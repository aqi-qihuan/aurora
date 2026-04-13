# 相册照片MinIO上传功能实现说明

## 概述

已完成Go版本相册和照片的MinIO上传功能，完全对标Java SpringBoot版本的实现逻辑。

---

## 核心改动文件

### 1. 新增文件

#### `internal/strategy/upload_service.go`
**作用**: 统一文件上传服务（对标Java `AbstractUploadStrategyImpl` + `UploadStrategyContext`）

**核心方法**:
- `UploadAlbumCover(ctx, data, filename)` - 上传相册封面 → `/photos/albums/{md5}.{ext}`
- `UploadPhoto(ctx, data, filename)` - 上传照片 → `/photos/{md5}.{ext}`
- `UploadTalkImage(ctx, data, filename)` - 上传说说图片 → `/talks/{md5}.{ext}`

**核心流程**（完全对标Java）:
```go
1. 计算文件MD5 (对标 Java FileUtil.getMd5)
2. 获取扩展名 (对标 Java FileUtil.getExtName)
3. 构建对象路径: path + md5 + ext
4. 检查文件是否已存在 (对标 Java exists())
   - 如果存在 → 直接返回URL（MD5去重命中）
   - 如果不存在 → 继续上传
5. 上传到MinIO (对标 Java upload())
6. 返回访问URL (对标 Java getFileAccessUrl())
```

---

### 2. 修改文件

#### `internal/service/registry.go`
**改动**:
- 添加 `UploadSvc *strategy.UploadService` 字段
- 在 `NewRegistry()` 中初始化UploadService（从MinIO配置创建）
- MinIO为可选依赖，初始化失败不阻断启动

```go
// ===== 文件上传服务 (MinIO，对标Java MinioUploadStrategyImpl) =====
uploadCfg := cfg.MinIO
if uploadCfg.Endpoint != "" && uploadCfg.AccessKey != "" {
    var err error
    r.UploadSvc, err = strategy.NewUploadService(uploadCfg)
    if err != nil {
        logger.Warn("MinIO上传服务初始化失败，文件上传功能将不可用", "error", err)
    }
}
```

#### `internal/handler/photo_album_handler.go`
**改动**:
- 构造函数增加 `uploadSvc *strategy.UploadService` 参数
- `UploadAlbumCover()` 方法重写：
  ```go
  // 旧实现: 返回假路径 "/uploads/albums/" + file.Filename
  // 新实现: 调用 uploadSvc.UploadAlbumCover() → MD5去重 + MinIO上传
  ```

#### `internal/handler/photo_handler.go`
**改动**:
- 构造函数增加 `uploadSvc *strategy.UploadService` 参数
- `UploadPhoto()` 方法重写：
  ```go
  // 旧实现: 返回假路径 "/uploads/photos/" + file.Filename
  // 新实现: 调用 uploadSvc.UploadPhoto() → MD5去重 + MinIO上传
  ```

#### `internal/handler/router.go`
**改动**:
- 注册Handler时传入UploadSvc:
  ```go
  PhotoHandler:         NewPhotoHandler(registry.Photo, registry.UploadSvc),
  PhotoAlbumHandler:    NewPhotoAlbumHandler(registry.PhotoAlbum, registry.UploadSvc),
  ```

---

## 技术细节

### MD5去重机制

**目的**: 避免重复上传相同内容的文件，节省存储空间

**实现**:
```go
// 1. 计算MD5
md5Hash := md5.New()
md5Hash.Write(data)
md5 := hex.EncodeToString(md5Hash.Sum(nil))

// 2. 构建文件名
fileName := md5 + ext  // 例如: d41d8cd98f00b204e9800998ecf8427e.jpg

// 3. 检查是否存在
exists, _ := s.minioStrategy.Exists(ctx, objectPath)
if exists {
    return s.minioStrategy.GetFileAccessUrl(objectPath), nil  // 直接返回
}

// 4. 上传
s.minioStrategy.Upload(ctx, path, fileName, bytes.NewReader(data))
```

**效果**: 
- 用户上传同一张图片多次 → MinIO只存储一份
- 第二次上传时秒级返回（无需实际传输）

---

### 存储路径规范

| 文件类型 | MinIO路径前缀 | 示例 |
|---------|-------------|------|
| 相册封面 | `/photos/albums/` | `/photos/albums/d41d8cd98f00b204e9800998ecf8427e.jpg` |
| 普通照片 | `/photos/` | `/photos/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6.png` |
| 说说图片 | `/talks/` | `/talks/f1e2d3c4b5a6978869504132abcdef01.webp` |

---

### URL生成规则

**内部结构**:
```go
func GetFileAccessUrl(objectName string) string {
    prefix := storage.GetExternalURL()  // 优先使用配置的URL
    if prefix == "" {
        prefix = s.url  // fallback到endpoint
    }
    return strings.TrimSuffix(prefix, "/") + "/" + objectName
}
```

**示例**:
- 配置 `minio.url = "https://ws.aqi125.cn"`
- 上传文件 `d41d8cd98f00b204e9800998ecf8427e.jpg`
- 返回URL: `https://ws.aqi125.cn/photos/albums/d41d8cd98f00b204e9800998ecf8427e.jpg`

---

## 与Java版本对比

| 特性 | Java实现 | Go实现 | 状态 |
|-----|---------|-------|------|
| MD5去重 | `FileUtil.getMd5()` | `crypto/md5` | ✅ 一致 |
| 扩展名提取 | `FileUtil.getExtName()` | `filepath.Ext()` | ✅ 一致 |
| 存在性检查 | `MinioClient.statObject()` | `MinioClient.StatObject()` | ✅ 一致 |
| 上传接口 | `MinioClient.putObject()` | `MinioClient.PutObject()` | ✅ 一致 |
| URL拼接 | `minioProperties.getUrl() + path` | `externalURL + "/" + objectName` | ✅ 一致 |
| 策略模式 | `UploadStrategyContext` | `UploadService` | ✅ 对齐 |

---

## 配置要求

确保 `configs/config.yaml` 或 `.env` 中有正确的MinIO配置:

```yaml
minio:
  endpoint: "http://localhost:9000"     # MinIO服务器地址
  access_key: "minioadmin"               # Access Key
  secret_key: "minioadmin"               # Secret Key
  bucket: "aurora"                       # Bucket名称
  url: "https://ws.aqi125.cn"           # 外部访问URL（CDN/域名代理）
  use_ssl: false                         # 是否启用SSL
```

---

## 测试验证

### 1. 启动Go服务
```bash
cd aurora-go
go run cmd/server/main.go
```

### 2. 测试相册封面上传
```bash
curl -X POST http://localhost:8080/api/admin/photos/albums/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@test.jpg"
```

**预期响应**:
```json
{
  "flag": true,
  "code": 20000,
  "data": "https://ws.aqi125.cn/photos/albums/d41d8cd98f00b204e9800998ecf8427e.jpg",
  "message": "操作成功"
}
```

### 3. 测试照片上传
```bash
curl -X POST http://localhost:8080/api/admin/photos/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@photo.png"
```

### 4. 验证MD5去重
- 第一次上传某图片 → 耗时 ~100ms（实际上传）
- 第二次上传同一图片 → 耗时 ~5ms（直接返回URL，无网络传输）

---

## 注意事项

1. **MinIO必须提前启动**
   - Docker方式: `docker run -p 9000:9000 minio/minio server /data`
   - 或使用项目中的 `scripts/docker-compose.go.yml`

2. **Bucket自动创建**
   - 首次启动时会自动创建配置的bucket
   - 如需手动创建: `mc mb myminio/aurora`

3. **权限配置**
   - 确保MinIO用户有 `s3:PutObject` 和 `s3:GetObject` 权限
   - 公开读: `mc policy set download myminio/aurora`

4. **前端兼容性**
   - 返回的URL格式与Java版完全一致
   - 前端无需任何修改即可使用

---

## 后续优化建议

1. **流式上传支持**
   - 当前实现读取整个文件到内存
   - 大文件（>10MB）建议使用分片上传

2. **并发上传**
   - 批量上传照片时可并行调用 `UploadPhoto()`
   - 使用 `sync.WaitGroup` 控制并发数

3. **缩略图生成**
   - 上传后自动生成小尺寸缩略图
   - 减少前台加载带宽

4. **CDN集成**
   - 配置 `minio.url` 为CDN域名
   - 提升全球访问速度

---

## 总结

✅ **已完成**:
- UploadService核心上传逻辑（MD5去重 + MinIO上传）
- 相册封面上传接口对接
- 照片上传接口对接
- Registry依赖注入配置
- Router路由绑定

✅ **对标Java**:
- 完全一致的MD5去重机制
- 相同的存储路径规范
- 相同的URL生成规则
- 相同的错误处理逻辑

🎯 **下一步**:
- 测试上传功能是否正常
- 验证MinIO中文件是否正确存储
- 确认前端能正确显示上传后的图片

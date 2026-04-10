# ============================================
# Aurora Go 交叉编译 & 部署脚本
# 支持: linux-amd64 / linux-arm64 / darwin / windows
# ============================================

set -e

APP_NAME="aurora-go"
VERSION="1.0.0"
BUILD_DIR="./build"
DIST_DIR="./dist"
MAIN_PKG="./cmd/server"
LDFLAGS="-w -s -X main.version=${VERSION}"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "============================================"
echo " Aurora Go Build Script v${VERSION}"
echo " Time: ${BUILD_TIME}"
echo " Commit: ${GIT_COMMIT}"
echo "============================================"

# 创建目录
mkdir -p ${BUILD_DIR}
mkdir -p ${DIST_DIR}

# ========== 编译函数 ==========
build() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT=$3

    echo ">>> Building for ${GOOS}/${GOARCH}..."
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags="${LDFLAGS}" \
        -o "${OUTPUT}" \
        "${MAIN_PKG}"

    local SIZE=$(du -h "${OUTPUT}" | cut -f1)
    echo "<<< Built: ${OUTPUT} (${SIZE})"
}

# ========== Linux AMD64 (主流服务器) ==========
build linux amd64 "${BUILD_DIR}/${APP_NAME}-linux-amd64"

# ========== Linux ARM64 (树莓派/ARM服务器) ==========
build linux arm64 "${BUILD_DIR}/${APP_NAME}-linux-arm64"

# ========== macOS AMD64 ==========
build darwin amd64 "${BUILD_DIR}/${APP_NAME}-darwin-amd64"

# ========== macOS ARM64 (Apple Silicon) ==========
build darwin arm64 "${BUILD_DIR}/${APP_NAME}-darwin-arm64"

# ========== Windows AMD64 ==========
build windows amd64 "${BUILD_DIR}/${APP_NAME}-windows-amd64.exe"

# ========== 打包发布 ==========
echo ""
echo ">>> Creating distribution packages..."

cd ${BUILD_DIR}

# Linux AMD64 package
tar -czvf "${DIST_DIR}/${APP_NAME}-v${VERSION}-linux-amd64.tar.gz" \
    "${APP_NAME}-linux-amd64" ../configs/ ../scripts/docker-compose.go.yml

# Linux ARM64 package
tar -czvf "${DIST_DIR}/${APP_NAME}-v${VERSION}-linux-arm64.tar.gz" \
    "${APP_NAME}-linux-arm64" ../configs/ ../scripts/docker-compose.go.yml

echo ""
echo "============================================"
echo " Build Complete!"
echo ""
echo " Artifacts:"
ls -lh ${DIST_DIR}/
echo ""
echo " Binaries:"
ls -lh *.exe 2>/dev/null || ls -lh aurora-go-*
echo "============================================"

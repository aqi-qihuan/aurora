.PHONY: all build run test lint clean docker-build docker-up docker-down help

# 变量
APP_NAME := aurora-go
VERSION := 1.0.0
BUILD_DIR := ./build
MAIN := ./cmd/server
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION)"

# Go 参数
GOCMD := go
GOBUILD := $(GOCMD) build $(LDFLAGS)
GOTEST := $(GOCMD) test -v -race
GOLINT := golangci-lint run

# 默认目标
all: lint test build

# 构建
build:
	@echo "=== 构建 Aurora Go ==="
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 运行 (开发模式, debug日志)
run: build
	@echo "=== 启动开发服务器 ==="
	./$(BUILD_DIR)/$(APP_NAME) --config configs/config.yaml

# 测试
test:
	@echo "=== 运行测试 ==="
	$(GOTEST) ./...
	@echo "测试通过"

# 基准测试
bench:
	@echo "=== 运行基准测试 ==="
	$(GOCMD) test -bench=. -benchmem ./...

# Lint检查
lint:
	@echo "=== Lint检查 ==="
	$(GOLINT) ./...

# 格式化代码
fmt:
	@echo "=== 格式化代码 ===
	gofmt -w .
	goimports -w .

# 清理
clean:
	@echo "=== 清理构建产物 ==="
	rm -rf $(BUILD_DIR)
	rm -rf dist/

# Docker 构建镜像
docker-build:
	@echo "=== Docker构建 ==="
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Docker Compose 编排启动 (含所有中间件)
docker-up:
	@echo "=== Docker Compose 启动 ==="
	docker compose -f scripts/docker-compose.go.yml up -d

# Docker Compose 停止
docker-down:
	@echo "=== Docker Compose 停止 ==="
	docker compose -f scripts/docker-compose.go.yml down

# 交叉编译 (Linux amd64)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN)

# 交叉编译 (Linux arm64 / 树莓派/ARM服务器)
build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN)

# Swagger 文档生成
docs:
	@echo "=== 生成 Swagger 文档 ==="
	swag init -g cmd/server/main.go -o api/docs --parseDependency --parseInternal

# 帮助信息
help:
	@echo "Aurora Go 构建命令:"
	@echo "  make build         - 构建二进制"
	@echo "  make run           - 运行开发服务器"
	@echo "  make test          - 运行测试"
	@echo "  make bench         - 运行基准测试"
	@echo "  make lint          - Lint检查"
	@echo "  make fmt           - 格式化代码"
	@echo "  make clean         - 清理"
	@echo "  make docker-build  - Docker构建镜像"
	@echo "  make docker-up     - Docker Compose启动"
	@echo "  make docker-down   - Docker Compose停止"
	@echo "  make docs          - 生成Swagger文档"

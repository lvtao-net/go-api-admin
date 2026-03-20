.PHONY: all build build-embed run clean dev frontend-build frontend-dev test embed release

# 变量定义
APP_NAME := go-gin-api-admin
VERSION := 1.0.0
BUILD_DIR := build
MAIN_PATH := cmd/server/main.go
FRONTEND_DIR := web
EMBED_DIR := internal/embed/dist

# Go 编译参数
GOCMD := go
GOBUILD := go build
GOCLEAN := go clean
GOTEST := go test
GOMOD := go mod

# 版本信息（从 git tag 获取，如果没有则使用默认值）
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -s -w"

all: clean deps build-local

# 下载依赖
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# 复制前端构建产物到 embed 目录
prepare-embed: frontend-build
	@echo "Preparing embed directory..."
	@rm -rf $(EMBED_DIR)
	@mkdir -p $(EMBED_DIR)
	@cp -r $(FRONTEND_DIR)/dist/* $(EMBED_DIR)/
	@echo "Static files prepared for embedding!"

# 构建当前平台（开发模式，不嵌入静态文件）
build-local:
	@echo "Building for current platform (development mode)..."
	@if [ ! -d "$(BUILD_DIR)" ]; then mkdir -p $(BUILD_DIR); fi
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@cp config.yaml $(BUILD_DIR)/config.yaml 2>/dev/null || true
	@echo "Build complete! Binary in $(BUILD_DIR)/"

# 构建当前平台（嵌入静态文件）
build-embed: prepare-embed
	@echo "Building for current platform (with embedded static files)..."
	@if [ ! -d "$(BUILD_DIR)" ]; then mkdir -p $(BUILD_DIR); fi
	$(GOBUILD) $(LDFLAGS) -tags embed -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@cp config.yaml.example $(BUILD_DIR)/config.yaml 2>/dev/null || true
	@echo "Build complete! Binary in $(BUILD_DIR)/"

# 构建生产版本（多平台，嵌入静态文件）
build: prepare-embed
	@echo "Building production binaries with embedded static files..."
	@if [ ! -d "$(BUILD_DIR)" ]; then mkdir -p $(BUILD_DIR); fi
	@echo "Building for darwin/amd64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -tags embed -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "Building for darwin/arm64..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -tags embed -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Building for linux/amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -tags embed -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "Building for linux/arm64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -tags embed -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Building for windows/amd64..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -tags embed -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Build complete! Binaries in $(BUILD_DIR)/"

# 开发模式运行
run:
	@echo "Running in development mode..."
	$(GOCMD) run $(MAIN_PATH) -config config.yaml

# 生产模式运行（需要先构建）
run-prod:
	@echo "Running in production mode..."
	./$(BUILD_DIR)/$(APP_NAME) -config config.yaml

# 清理
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(EMBED_DIR)
	@echo "Clean complete!"

# 前端开发
frontend-dev:
	@echo "Starting frontend development server..."
	cd $(FRONTEND_DIR) && npm run dev

# 构建前端
frontend-build:
	@echo "Building frontend..."
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo "Installing frontend dependencies..."; \
		cd $(FRONTEND_DIR) && npm install; \
	fi
	cd $(FRONTEND_DIR) && npm run build
	@echo "Frontend build complete!"

# 初始化前端
frontend-init:
	@echo "Initializing frontend..."
	cd $(FRONTEND_DIR) && npm install

# 测试
test:
	$(GOTEST) -v ./...

# 打包发布（创建压缩包）
release: deps build
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/release
	@echo "Creating darwin-amd64 archive..."
	@cd $(BUILD_DIR) && tar -czf release/$(APP_NAME)-darwin-amd64.tar.gz $(APP_NAME)-darwin-amd64
	@echo "Creating darwin-arm64 archive..."
	@cd $(BUILD_DIR) && tar -czf release/$(APP_NAME)-darwin-arm64.tar.gz $(APP_NAME)-darwin-arm64
	@echo "Creating linux-amd64 archive..."
	@cd $(BUILD_DIR) && tar -czf release/$(APP_NAME)-linux-amd64.tar.gz $(APP_NAME)-linux-amd64
	@echo "Creating linux-arm64 archive..."
	@cd $(BUILD_DIR) && tar -czf release/$(APP_NAME)-linux-arm64.tar.gz $(APP_NAME)-linux-arm64
	@echo "Creating windows-amd64 archive..."
	@cd $(BUILD_DIR) && zip -q release/$(APP_NAME)-windows-amd64.zip $(APP_NAME)-windows-amd64.exe
	@cp config.yaml.example $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/ 2>/dev/null || true
	@echo "=========================================="
	@echo "Release packages ready in $(BUILD_DIR)/release/"
	@echo "=========================================="
	@ls -lh $(BUILD_DIR)/release/

# 创建 git tag 并推送
tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Tag $(VERSION) created and pushed!"
	@echo "GitHub Actions will automatically build and create a release."

# 开发环境设置
dev-setup: deps frontend-init
	@echo "Development environment setup complete!"
	@echo "Run 'make run' to start the server"

# 数据库迁移
migrate:
	@echo "Running database migrations..."
	$(GOCMD) run $(MAIN_PATH) -config config.yaml -migrate

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

# Docker 运行
docker-run:
	@echo "Running Docker container..."
	docker run -p 8099:8099 $(APP_NAME):$(VERSION)

# 显示版本信息
version:
	@echo "Version: $(VERSION)"

# 帮助
help:
	@echo "可用命令:"
	@echo "  make build-local   - 构建当前平台（开发模式）"
	@echo "  make build-embed   - 构建当前平台（嵌入静态文件）"
	@echo "  make build         - 构建多平台生产版本"
	@echo "  make run           - 开发模式运行"
	@echo "  make run-prod      - 生产模式运行"
	@echo "  make clean         - 清理构建产物"
	@echo "  make frontend-dev  - 前端开发服务器"
	@echo "  make frontend-build- 构建前端"
	@echo "  make test          - 运行测试"
	@echo "  make release       - 打包发布（创建压缩包）"
	@echo "  make tag VERSION=v1.0.0 - 创建并推送 git tag"
	@echo "  make dev-setup     - 初始化开发环境"
	@echo "  make version       - 显示当前版本"

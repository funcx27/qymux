# Qymux v1.0.0 Makefile
# 纯传输层库

.PHONY: all build test clean fmt help

# 变量定义
GO_MODULE := github.com/funcx27/qymux
GO_TEST_FLAGS := -v -race -cover -coverprofile=coverage.out

# 默认目标
all: build

# 构建整个项目
build:
	@echo "构建 Qymux v1.0.0..."
	go build ./...
	@echo "构建完成"

# 运行测试
test:
	@echo "运行测试..."
	go test $(GO_TEST_FLAGS) ./...

# 格式化代码
fmt:
	@echo "格式化Go代码..."
	gofmt -w -s .
	@echo "格式化完成"

# 清理构建产物
clean:
	@echo "清理构建产物..."
	go clean
	rm -f coverage.out
	rm -f coverage.html
	@echo "清理完成"

# 创建测试证书
cert:
	@echo "创建测试证书..."
	cd pkg/cert && go run -exec "go run ." cert.go 2>&1 | grep -v "go: downloading"

# 安装依赖
deps:
	@echo "安装Go依赖..."
	go mod download
	go mod tidy

# 代码覆盖率报告
coverage: test
	@echo "生成覆盖率报告..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 运行静态分析
lint:
	@echo "运行静态分析..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "警告: golangci-lint未安装, 跳过静态分析"; \
		echo "安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 显示帮助信息
help:
	@echo "Qymux v1.0.0 Makefile 命令:"
	@echo ""
	@echo "构建命令:"
	@echo "  make build        - 构建整个项目"
	@echo "  make clean        - 清理构建产物"
	@echo ""
	@echo "测试命令:"
	@echo "  make test         - 运行单元测试"
	@echo "  make coverage     - 生成代码覆盖率报告"
	@echo ""
	@echo "其他命令:"
	@echo "  make fmt          - 格式化Go代码"
	@echo "  make lint         - 运行静态分析"
	@echo "  make cert         - 创建测试证书"
	@echo "  make deps         - 安装依赖"
	@echo "  make help         - 显示此帮助信息"
	@echo ""

# 文件：OnlineOj/Makefile

# =========================
# 基础目录
# =========================
ROOT_DIR        := $(CURDIR)
BUILD_DIR       := $(ROOT_DIR)/build#编译产物
PID_DIR         := $(BUILD_DIR)/pids#后台任务信息
LOG_DIR         := $(ROOT_DIR)/logs#日志存储路径
TEMP_DIR        := $(ROOT_DIR)/temp#可执行程序存储路径

# =========================
# 二进制文件
# =========================
GATEWAY_BIN     := $(BUILD_DIR)/gateway
JUDGE_BIN       := $(BUILD_DIR)/judge

# =========================
# 配置文件
# =========================
GATEWAY_CONFIG          ?= $(ROOT_DIR)/gateway/config/gateway.yaml
GATEWAY_LOCAL_CONFIG    ?= $(ROOT_DIR)/gateway/config/gateway.local.yaml
JUDGE_CONFIG            ?= $(ROOT_DIR)/judge/config/judge.yaml
JUDGE_LOCAL_CONFIG      ?= $(ROOT_DIR)/judge/config/judge.local.yaml

# =========================
# 可配置参数
# =========================
# 默认调试模式。运行模式：prod
MODE            ?= debug

# 网关监听 host
GATEWAY_HOST    ?= 127.0.0.1
# 网关监听 port
GATEWAY_PORT    ?= 9000

# 判题服务监听 host
JUDGE_HOST      ?= 127.0.0.1
# 判题服务监听 port
JUDGE_BASE_PORT ?= 10000
# 启动的判题服务数量。支持通过 make 命令行覆盖变量（如 make run JUDGE_COUNT=3）
JUDGE_COUNT     ?= 1
# 动态生成所有 Judge 地址（逗号分隔，例如 "127.0.0.1:10000,127.0.0.1:10001"）
JUDGE_ADDRS = $(shell \
	start=$(JUDGE_BASE_PORT); \
	count=$(JUDGE_COUNT); \
	end=$$((start + count - 1)); \
	seq $$start $$end | sed 's/.*/$(JUDGE_HOST):&/' | tr '\n' ',' | sed 's/,$$//' \
)

# =========================
# 固定参数
# =========================
PROD            := prod#生产模式

# 为目标
.PHONY: all prepare gateway judge run-gateway start-gateway run-judge run-judges run stop stop-gateway stop-judges clean help

all: gateway judge

prepare:
	@echo "$(LOG_DIR)"
	@echo "$(TEMP_DIR)"
	@mkdir -p "$(BUILD_DIR)" "$(PID_DIR)" "$(LOG_DIR)" "$(TEMP_DIR)"

# =========================
# 编译 gateway（进入子模块编译）
# =========================
gateway: prepare
	@echo "Building gateway..."
	@cd gateway && go build -o ../build/gateway ./cmd/server
# 在模块中编译
# 模块目录$go build + 目录自动找到main.go编译整个包
# go build + 单独文件：只能编译单个文件

# =========================
# 编译 judge（进入子模块编译）
# =========================
judge: prepare
	@echo "Building judge..."
	@cd judge && go build -o ../build/judge ./cmd/server

# =========================
# 前台运行 gateway
# =========================
run-gateway: gateway
	@echo "Running gateway on $(GATEWAY_HOST):$(GATEWAY_PORT)..."
	@"$(GATEWAY_BIN)" \
		-config "$(GATEWAY_CONFIG)" \
		-config-local "$(GATEWAY_LOCAL_CONFIG)" \
		-m "$(MODE)" \
		-h "$(GATEWAY_HOST)" \
		-p "$(GATEWAY_PORT)" \
		-name "gateway" \
		-id 1 \
		-judge-addrs '$(JUDGE_ADDRS)'

# =========================
# 后台运行 gateway
# =========================
start-gateway: gateway
	@echo "Starting gateway on $(GATEWAY_HOST):$(GATEWAY_PORT)..."
	@"$(GATEWAY_BIN)" \
		-config "$(GATEWAY_CONFIG)" \
		-config-local "$(GATEWAY_LOCAL_CONFIG)" \
		-m "$(PROD)" \
		-h "$(GATEWAY_HOST)" \
		-p "$(GATEWAY_PORT)" \
		-name "gateway" \
		-id 1 \
		-judge-addrs '$(JUDGE_ADDRS)' > "$(LOG_DIR)/gateway.stdout.log" 2>&1 & \
	echo $$! > "$(PID_DIR)/gateway.pid"
# &1 = 标准输出（stdout）的文件描述符
# 2>&1 把标准错误（2）重定向 到 标准输出（1）所在的位置
# [文件描述符] > &[目标文件描述符]
# $$! Makefile 会把 $$ 变成 $ 最终传给 Shell 的是：$!
# $! Shell 内置变量：上一个后台运行进程的 PID（进程号）

# =========================
# 前台运行单个 judge
# =========================
run-judge: judge
	@echo "Running judge on $(JUDGE_HOST):$(JUDGE_BASE_PORT)..."
	@"$(JUDGE_BIN)" \
		-config "$(JUDGE_CONFIG)" \
		-config-local "$(JUDGE_LOCAL_CONFIG)" \
		-m "$(MODE)" \
		-h "$(JUDGE_HOST)" \
		-p "$(JUDGE_BASE_PORT)" \
		-name "judge" \
		-id "$(JUDGE_BASE_PORT)"

# =========================
# 后台运行多个 judge
# =========================
start-judges: judge
	@echo "Starting $(JUDGE_COUNT) judge instance(s)..."
	@i=0; \
	while [ $$i -lt $(JUDGE_COUNT) ]; do \
		port=$$(( $(JUDGE_BASE_PORT) + $$i )); \
		id=$$(( $$i + 1 )); \
		echo "Starting judge-$$id on $(JUDGE_HOST):$$port"; \
		"$(JUDGE_BIN)" \
			-config "$(JUDGE_CONFIG)" \
			-config-local "$(JUDGE_LOCAL_CONFIG)" \
			-m "$(PROD)" \
			-h "$(JUDGE_HOST)" \
			-p "$$port" \
			-name "judge-$$id" \
			-id "$$id" > "$(LOG_DIR)/judge-$$id.stdout.log" 2>&1 & \
		echo $$! > "$(PID_DIR)/judge-$$id.pid"; \
		i=$$(( $$i + 1 )); \
	done

# =========================
# 一键启动
# =========================
run: gateway judge
	@$(MAKE) start-judges JUDGE_COUNT=$(JUDGE_COUNT) JUDGE_BASE_PORT=$(JUDGE_BASE_PORT) JUDGE_HOST=$(JUDGE_HOST) MODE=$(PROD)
	@sleep 1
	@$(MAKE) start-gateway \
		GATEWAY_HOST=$(GATEWAY_HOST) \
		GATEWAY_PORT=$(GATEWAY_PORT) \
		JUDGE_ADDRS='$(JUDGE_ADDRS)' \
		MODE=$(PROD)
	@echo "Gateway started at http://$(GATEWAY_HOST):$(GATEWAY_PORT)"
	@echo "Judge addresses: $(JUDGE_ADDRS)"

stop-gateway:
	@if [ -f "$(PID_DIR)/gateway.pid" ]; then \
		kill "$$(cat "$(PID_DIR)/gateway.pid")" 2>/dev/null || true; \
		rm -f "$(PID_DIR)/gateway.pid"; \
		echo "Gateway stopped."; \
	else \
		echo "Gateway is not running."; \
	fi

stop-judges:
	@found=0; \
	for f in "$(PID_DIR)"/judge-*.pid; do \
		if [ -f "$$f" ]; then \
			kill "$$(cat "$$f")" 2>/dev/null || true; \
			rm -f "$$f"; \
			found=1; \
		fi; \
	done; \
	if [ $$found -eq 1 ]; then \
		echo "All judge instances stopped."; \
	else \
		echo "No judge instances are running."; \
	fi

stop: stop-gateway stop-judges
	@rm -fr build
	@rm -fr temp

clean: stop
	@rm -rf "$(BUILD_DIR)"

help:
	@echo "==================== 构建与运行指南 ===================="
	@echo ""
	@echo "【编译相关】"
	@echo "  make gateway          - 仅编译 gateway"
	@echo "  make judge            - 仅编译 judge"
	@echo "  make all              - 编译 gateway + judge"
	@echo ""
	@echo "【前台运行（调试用）】"
	@echo "  make run-gateway      - 前台运行 gateway（日志输出到终端，Ctrl+C 停止）"
	@echo "  make run-judge        - 前台运行单个 judge（日志输出到终端）。Id被固定为1"
	@echo ""
	@echo "【后台运行（生产用）】"
	@echo "  make start-gateway    - 后台启动 gateway（Gorm/Gin日志写入 logs/gateway.stdout.log）"
	@echo "  make start-judges     - 后台启动多个 judge（数量由 JUDGE_COUNT 控制）"
	@echo "  make run              - 一键后台启动 gateway + 所有 judge（推荐）"
	@echo ""
	@echo "【服务管理】"
	@echo "  make stop-gateway     - 停止后台 gateway"
	@echo "  make stop-judges      - 停止所有后台 judge"
	@echo "  make stop             - 停止全部服务"
	@echo "  make clean            - 停止服务 + 删除 build/ 目录"
	@echo ""
	@echo "==================== 可配置变量 ===================="
	@echo "  MODE              - 前台运行的模式 (debug/prod)        默认: debug"
	@echo "  GATEWAY_HOST      - Gateway 监听地址                   默认: 127.0.0.1"
	@echo "  GATEWAY_PORT      - Gateway HTTP 端口                  默认: 9000"
	@echo "  JUDGE_HOST        - Judge 监听地址                     默认: 127.0.0.1"
	@echo "  JUDGE_BASE_PORT   - 首个 Judge 的 gRPC 端口            默认: 10000"
	@echo "  JUDGE_COUNT       - Judge 实例数量                     默认: 1"
	@echo "  JUDGE_ADDRS       - Judge 地址列表（逗号分隔），自动根据 JUDGE_HOST/JUDGE_BASE_PORT/COUNT 生成, 基于 JUDGE_BASE_PORT 逐增1"
	@echo "                       （通常不需要手动设置，由 Makefile 自动计算）"
	@echo ""
	@echo "==================== 使用示例 ===================="
	@echo "  # 开发调试（前台运行）"
	@echo "  make run-gateway"
	@echo "  make run-judge JUDGE_BASE_PORT=10001"
	@echo ""
	@echo "  # 生产启动（后台运行，2个Judge）"
	@echo "  make run JUDGE_COUNT=2" 
	@echo "              自动启动 prod 生产模式"
	@echo ""
	@echo "  # 停止所有服务"
	@echo "  make stop"
	@echo ""
	@echo "  # 清理构建产物"
	@echo "  make clean"
	@echo "==================== 注意 ====================""
	@echo "	          后台运行的服务默认是生产模式!"
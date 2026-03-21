MAKEFLAGS += --no-print-directory

# 可执行文件的名称
BINARY_NAME=idb-agent

# 版本号
VERSION=1.0.0

# 从 Git 中获取版本号
GIT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo $(VERSION))

# 源代码目录
SRC_DIR=.

# 安装目录
BIN_DIR=/var/lib/idb-agent
CONFIG_DIR=/etc/idb-agent
DATA_DIR=/var/lib/idb-agent/data
LOG_DIR=/var/log/idb-agent
RUN_DIR=/run/idb-agent

# 系统服务目录
SYSTEM_SERVICE_DIR=/etc/systemd/system/

# 配置文件路径
CONFIG_FILE=idb-agent.conf

# 注册服务文件路径
SERVICE_FILE=idb-agent.service

# 日志文件路径
LOG_FILE=/var/log/idb-agent/idb-agent.log

define step
	@printf "  -> %s\n" "$(1)"
endef

# 构建命令
build:
	$(call step,编译 agent)
	$(eval KEY := $(shell cat ../.build.key 2>/dev/null || echo "default_key"))
	@sed -i "s/secret_key=.*/secret_key=$(KEY)/" idb-agent.conf
	@GOOS=linux GOARCH=amd64 go build \
		-ldflags "-s -w \
		-X 'github.com/sensdata/idb/agent/global.Version=$(GIT_VERSION)'" \
		-o $(BINARY_NAME) $(SRC_DIR)

# 安装命令
install: build
	$(call step,准备 agent 目录与服务)
	@sudo mkdir -p $(BIN_DIR); \
	sudo mkdir -p $(CONFIG_DIR); \
	sudo mkdir -p $(DATA_DIR); \
	sudo mkdir -p $(LOG_DIR); \
	sudo mkdir -p $(RUN_DIR); \
	sudo systemctl stop $(BINARY_NAME).service 2>/dev/null || true; \
	sudo cp $(BINARY_NAME) $(BIN_DIR); \
	sudo cp $(CONFIG_FILE) $(CONFIG_DIR); \
	sudo cp $(SERVICE_FILE) $(SYSTEM_SERVICE_DIR); \
	sudo chmod +x $(BIN_DIR)/$(BINARY_NAME); \
	sudo touch $(LOG_FILE); \
	sudo systemctl daemon-reload; \
	sudo systemctl enable $(BINARY_NAME).service; \
	sudo systemctl restart $(BINARY_NAME).service; \
	sudo ln -sf /var/lib/idb-agent/idb-agent /usr/local/bin/idb-agent
	$(call step,agent 安装完成)
	@$(MAKE) clean

# 卸载命令
uninstall:
	sudo systemctl stop $(BINARY_NAME).service; \
	sudo systemctl disable $(BINARY_NAME).service; \
	sudo rm -f $(SYSTEM_SERVICE_DIR)/$(SERVICE_FILE); \
	sudo rm -f $(BIN_DIR)/$(BINARY_NAME); \
	sudo rm -rf $(CONFIG_DIR); \
	sudo rm -rf $(DATA_DIR); \
	sudo rm -rf $(LOG_DIR); \
	sudo rm -rf $(RUN_DIR); \
	sudo systemctl daemon-reload; \

# 清理构建文件
clean:
	@rm -f $(BINARY_NAME)
	
.PHONY: build install clean

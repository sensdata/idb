# 可执行文件的名称
BINARY_NAME=idb-center

# 版本号
VERSION=1.0.0

# 源代码目录
SRC_DIR=.

# 安装目录
BIN_DIR=/usr/local/bin
CONFIG_DIR=/etc/idb
DATA_DIR=/var/lib/idb
LOG_DIR =/var/log/idb

# 系统服务目录
SYSTEM_SERVICE_DIR=/etc/systemd/system/

# 配置文件路径
CONFIG_FILE=idb-center.json

# 注册服务文件路径
SERVICE_FILE=idb-center.service

# 日志文件路径
LOG_FILE=/var/log/idb/center-run.log

# 构建命令
build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) $(SRC_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-macos $(SRC_DIR)

# 安装命令
install: build
	@if [ "$(shell uname)" = "Linux" ]; then \
		sudo mkdir -p $(BIN_DIR); \
		sudo mkdir -p $(CONFIG_DIR); \
		sudo mkdir -p $(DATA_DIR); \
		sudo mkdir -p $(LOG_DIR); \
		sudo cp $(BINARY_NAME) $(BIN_DIR); \
		sudo cp $(CONFIG_FILE) $(CONFIG_DIR); \
		sudo cp $(SERVICE_FILE) $(SYSTEM_SERVICE_DIR); \
		sudo chmod +x $(BIN_DIR)/$(BINARY_NAME); \
		sudo touch $(LOG_FILE); \
		sudo systemctl daemon-reload; \
		sudo systemctl enable $(BINARY_NAME).service; \
		sudo systemctl start $(BINARY_NAME).service; \
	fi
	@if [ "$(shell uname)" = "Darwin" ]; then \
		sudo mkdir -p $(BIN_DIR); \
		sudo mkdir -p $(CONFIG_DIR); \
		sudo mkdir -p $(DATA_DIR); \
		sudo mkdir -p $(LOG_DIR); \
		sudo cp $(BINARY_NAME) $(BIN_DIR); \
		sudo cp $(CONFIG_FILE) $(CONFIG_DIR); \
		sudo chmod +x $(BIN_DIR)/$(BINARY_NAME); \
		sudo touch $(LOG_FILE); \
	fi
	$(MAKE) clean  # 在安装后调用clean目标

# 卸载命令
uninstall:
	@if [ "$(shell uname)" = "Linux" ]; then \
		sudo systemctl stop $(BINARY_NAME).service; \
		sudo systemctl disable $(BINARY_NAME).service; \
		sudo rm -f $(SYSTEM_SERVICE_DIR)/$(SERVICE_FILE); \
		sudo rm -f $(BIN_DIR)/$(BINARY_NAME); \
		sudo rm -rf $(CONFIG_DIR); \
		sudo rm -rf $(DATA_DIR); \
		sudo rm -rf $(LOG_DIR); \
		sudo systemctl daemon-reload; \
	fi
	@if [ "$(shell uname)" = "Darwin" ]; then \
		sudo rm -f $(BIN_DIR)/$(BINARY_NAME); \
		sudo rm -rf $(CONFIG_DIR); \
		sudo rm -rf $(DATA_DIR); \
		sudo rm -rf $(LOG_DIR); \
	fi

# 清理构建文件
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-macos

.PHONY: build install uninstall clean

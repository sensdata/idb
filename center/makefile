# 可执行文件的名称
BINARY_NAME=idb

# 版本号
VERSION=1.0.0

# 从 Git 中获取版本号
GIT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo $(VERSION))

# 源代码目录
SRC_DIR=.
# 前端项目源码目录
FRONTEND_DIR=../frontend

# 安装目录
BIN_DIR=/var/lib/idb
HOME_DIR=/var/lib/idb/home
CONFIG_DIR=/etc/idb
DATA_DIR=/var/lib/idb/data
LOG_DIR=/var/log/idb
RUN_DIR=/run/idb

# 系统服务目录
SYSTEM_SERVICE_DIR=/etc/systemd/system/

# 配置文件路径
CONFIG_FILE=idb.conf

# 注册服务文件路径
SERVICE_FILE=idb.service

# 日志文件路径
LOG_FILE=/var/log/idb/idb.log

# 证书文件路径
CERT_FILE=./global/certs/cert.pem
KEY_FILE=./global/certs/key.pem

# 构建命令
build:
	# 读取密钥
	$(eval KEY = $(shell cat ../.build.key 2>/dev/null || echo "default_key"))
	# 编译
	GOOS=linux GOARCH=amd64 go build \
		-ldflags "-s -w \
		-X 'github.com/sensdata/idb/center/global.Version=$(GIT_VERSION)' \
		-X 'github.com/sensdata/idb/center/global.DefaultKey=$(KEY)'" \
		-o $(BINARY_NAME) $(SRC_DIR)

# 安装命令
install: build
	# 创建必要的目录
	sudo mkdir -p $(BIN_DIR); \
	sudo mkdir -p $(CONFIG_DIR); \
	sudo mkdir -p $(DATA_DIR); \
	sudo mkdir -p $(LOG_DIR); \
	sudo mkdir -p $(RUN_DIR); \
	sudo mkdir -p $(HOME_DIR); \

	# 拷贝前端编译结果
	sudo cp -r $(FRONTEND_DIR)/dist/* $(HOME_DIR); \

	# 拷贝证书文件
	sudo cp ${CERT_FILE} $(BIN_DIR); \
	sudo cp ${KEY_FILE} $(BIN_DIR); \

	# 拷贝其他必要的文件
	sudo cp $(BINARY_NAME) $(BIN_DIR); \
	sudo cp $(CONFIG_FILE) $(CONFIG_DIR); \
	sudo cp $(SERVICE_FILE) $(SYSTEM_SERVICE_DIR); \
	sudo chmod +x $(BIN_DIR)/$(BINARY_NAME); \
	sudo touch $(LOG_FILE); \
	sudo systemctl daemon-reload; \
	sudo systemctl enable $(BINARY_NAME).service; \
	sudo systemctl start $(BINARY_NAME).service; \
	sudo ln -sf /var/lib/idb/idb /usr/local/bin/idb
	$(MAKE) clean  # 在安装后调用clean目标

# 卸载命令
uninstall:
	sudo systemctl stop $(BINARY_NAME).service; \
	sudo systemctl disable $(BINARY_NAME).service; \
	sudo rm -f $(SYSTEM_SERVICE_DIR)/$(SERVICE_FILE); \
	sudo rm -rf $(BIN_DIR); \
	sudo rm -rf $(CONFIG_DIR); \
	sudo rm -rf $(DATA_DIR); \
	sudo rm -rf $(LOG_DIR); \
	sudo rm -rf $(RUN_DIR); \
	sudo systemctl daemon-reload; \

# 清理构建文件
clean:
	rm -f $(BINARY_NAME)

.PHONY: build install uninstall clean

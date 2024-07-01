# 可执行文件的名称
BINARY_NAME=idbagent

# 版本号
VERSION=1.0.0

# 源代码目录
SRC_DIR=.

# 安装目录
INSTALL_DIR=/usr/local/bin

# 配置文件路径
CONFIG_FILE=config.json

# 构建命令
build:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) $(SRC_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME)-macos $(SRC_DIR)
	
# 安装命令
install: build
	@if [ "$(shell uname)" = "Linux" ]; then \
		mkdir -p $(INSTALL_DIR); \
		cp $(BINARY_NAME) $(INSTALL_DIR); \
		cp $(CONFIG_FILE) $(INSTALL_DIR); \
		chmod +x $(INSTALL_DIR)/$(BINARY_NAME); \
	fi
	@if [ "$(shell uname)" = "Darwin" ]; then \
		mkdir -p $(INSTALL_DIR); \
		cp $(BINARY_NAME)-macos $(INSTALL_DIR)/$(BINARY_NAME); \
		cp $(CONFIG_FILE) $(INSTALL_DIR); \
		chmod +x $(INSTALL_DIR)/$(BINARY_NAME); \
	fi
	$(MAKE) clean  # 在安装后调用clean目标

# 清理构建文件
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-macos
	
.PHONY: build install clean

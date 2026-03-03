# 证书文件路径
CERT_FILE=cert.pem
KEY_FILE=key.pem

# 目录路径
AGENT_CERT_DIR=agent/global/certs
CENTER_CERT_DIR=center/global/certs
FRONTEND_DIR=frontend
NPM_INSTALL_FLAGS=--prefer-offline --no-audit --no-fund --progress=false --ignore-scripts

BUILD_KEY=.build.key
BUILD_JWT_KEY=.build.jwt.key
AGENT_CERT_FILE=$(AGENT_CERT_DIR)/cert.pem
AGENT_KEY_FILE=$(AGENT_CERT_DIR)/key.pem
CENTER_CERT_FILE=$(CENTER_CERT_DIR)/cert.pem
CENTER_KEY_FILE=$(CENTER_CERT_DIR)/key.pem

deploy: build-frontend generate-key generate-jwt-key generate-certs
	$(MAKE) -C center install
	$(MAKE) -C agent install

build-frontend:
	cd $(FRONTEND_DIR) && (npm ci $(NPM_INSTALL_FLAGS) || npm install $(NPM_INSTALL_FLAGS)) && IDB_ENABLE_IMAGEMIN=0 npm run build

# 生成密钥（仅首次）
generate-key:
	@if [ ! -s $(BUILD_KEY) ]; then \
		KEY=$$(openssl rand -base64 64 | tr -dc 'a-z0-9' | head -c 24); \
		echo "$$KEY" > $(BUILD_KEY); \
		echo "Generated $(BUILD_KEY)"; \
	else \
		echo "Keep existing $(BUILD_KEY)"; \
	fi

# 生成JWT密钥（仅首次，14位字母数字）
generate-jwt-key:
	@if [ ! -s $(BUILD_JWT_KEY) ]; then \
		JWT_KEY=$$(openssl rand -base64 64 | tr -dc 'a-z0-9' | head -c 14); \
		echo "$$JWT_KEY" > $(BUILD_JWT_KEY); \
		echo "Generated $(BUILD_JWT_KEY)"; \
	else \
		echo "Keep existing $(BUILD_JWT_KEY)"; \
	fi

# 生成证书（仅首次）并分发
generate-certs:
	mkdir -p $(AGENT_CERT_DIR)
	mkdir -p $(CENTER_CERT_DIR)
	@if [ -s $(AGENT_CERT_FILE) ] && [ -s $(AGENT_KEY_FILE) ] && [ -s $(CENTER_CERT_FILE) ] && [ -s $(CENTER_KEY_FILE) ]; then \
		echo "Keep existing certs"; \
	else \
		echo "Generating certs"; \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout $(KEY_FILE) -out $(CERT_FILE) -config ssl.cnf -extensions v3_ca; \
		cp -f $(CERT_FILE) $(AGENT_CERT_DIR)/; \
		cp -f $(KEY_FILE) $(AGENT_CERT_DIR)/; \
		cp -f $(CERT_FILE) $(CENTER_CERT_DIR)/; \
		cp -f $(KEY_FILE) $(CENTER_CERT_DIR)/; \
		$(MAKE) clean; \
	fi

# 清理临时生成的根目录证书文件
clean:
	rm -f $(CERT_FILE) $(KEY_FILE)

.PHONY: deploy build-frontend generate-certs generate-key generate-jwt-key clean

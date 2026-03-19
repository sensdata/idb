MAKEFLAGS += --no-print-directory

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

define step
	@printf "\n==> %s\n" "$(1)"
endef

deploy: build-frontend generate-key generate-jwt-key generate-certs
	$(call step,安装中心服务)
	@$(MAKE) -C center install
	$(call step,安装 Agent 服务)
	@$(MAKE) -C agent install
	$(call step,部署完成)

build-frontend:
	$(call step,构建前端)
	@cd $(FRONTEND_DIR) && \
		( npm ci $(NPM_INSTALL_FLAGS) --loglevel=error || npm install $(NPM_INSTALL_FLAGS) --loglevel=error ) && \
		IDB_ENABLE_IMAGEMIN=0 npm run build --silent

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
	$(call step,准备证书)
	@mkdir -p $(AGENT_CERT_DIR)
	@mkdir -p $(CENTER_CERT_DIR)
	@if [ -s $(AGENT_CERT_FILE) ] && [ -s $(AGENT_KEY_FILE) ] && [ -s $(CENTER_CERT_FILE) ] && [ -s $(CENTER_KEY_FILE) ]; then \
		echo "Keep existing certs"; \
	else \
		echo "Generating certs"; \
		openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout $(KEY_FILE) -out $(CERT_FILE) -config ssl.cnf -extensions v3_ca >/dev/null 2>&1 || { echo "Failed to generate certs"; exit 1; }; \
		cp -f $(CERT_FILE) $(AGENT_CERT_DIR)/; \
		cp -f $(KEY_FILE) $(AGENT_CERT_DIR)/; \
		cp -f $(CERT_FILE) $(CENTER_CERT_DIR)/; \
		cp -f $(KEY_FILE) $(CENTER_CERT_DIR)/; \
		$(MAKE) clean; \
	fi

# 清理临时生成的根目录证书文件
clean:
	@rm -f $(CERT_FILE) $(KEY_FILE)

.PHONY: deploy build-frontend generate-certs generate-key generate-jwt-key clean

# 证书文件路径
CERT_FILE=cert.pem
KEY_FILE=key.pem

# 目录路径
AGENT_DIR=agent/global/certs
CENTER_DIR=center/global/certs

# 生成密钥的命令
generate-key:
	$(eval KEY := $(shell openssl rand -base64 64 | tr -dc 'a-z0-9' | head -c 24))
	@echo "$(KEY)" > .build.key

# 生成证书的命令
generate-certs:
	# 生成证书
	openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout $(KEY_FILE) -out $(CERT_FILE) -config ssl.cnf -extensions v3_ca

	# 创建证书目录
	mkdir -p $(AGENT_DIR)
	mkdir -p $(CENTER_DIR)

	# 拷贝证书
	cp -f $(CERT_FILE) $(AGENT_DIR)/
	cp -f $(KEY_FILE) $(AGENT_DIR)/
	cp -f $(CERT_FILE) $(CENTER_DIR)/
	cp -f $(KEY_FILE) $(CENTER_DIR)/
	$(MAKE) clean

# 清理证书文件
clean:
	rm -f $(CERT_FILE) $(KEY_FILE)

.PHONY: generate-certs generate-key clean
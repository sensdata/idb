# 证书文件路径
CERT_FILE=cert.pem
KEY_FILE=key.pem

# 目录路径
AGENT_DIR=agent/global/certs
CENTER_DIR=center/global/certs

# 生成证书的命令
generate-certs:
	# 生成证书
	openssl req -x509 -nodes -days 730 -newkey rsa:2048 -keyout $(KEY_FILE) -out $(CERT_FILE) -config ssl.cnf \

	# 拷贝证书
	sudo cp -f $(CERT_FILE) $(AGENT_DIR); \
	sudo cp -f $(KEY_FILE) $(AGENT_DIR); \
	sudo cp -f $(CERT_FILE) $(CENTER_DIR); \
	sudo cp -f $(KEY_FILE) $(CENTER_DIR); \
	$(MAKE) clean  # 在拷贝后调用clean目标

# 清理证书文件
clean:
	rm -f $(CERT_FILE) $(KEY_FILE)

.PHONY: generate-certs clean
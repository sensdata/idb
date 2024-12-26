# 证书文件
CERT_FILE=cert.pem
KEY_FILE=key.pem

# 目录路径
AGENT_DIR=agent/agent/
CENTER_DIR=center/core/conn/

# 生成证书的命令
generate-certs:
	openssl req -x509 -newkey rsa:2048 -keyout $(KEY_FILE) -out $(CERT_FILE) -days 365 -nodes -subj "/CN=localhost"

# 更新证书：生成并拷贝到指定目录
update-certs: generate-certs
	sudo cp $(CERT_FILE) $(AGENT_DIR); \
	sudo cp $(KEY_FILE) $(AGENT_DIR); \
	sudo cp $(CERT_FILE) $(CENTER_DIR); \

# 清理证书文件
clean:
	rm -f $(CERT_FILE) $(KEY_FILE)

.PHONY: generate-certs update-certs clean
package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sensdata/idb/center/core/api/middleware"
	"github.com/sensdata/idb/center/core/api/router"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"github.com/sensdata/idb/core/utils"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var API ApiServer = ApiServer{
	Router: gin.Default(),
}

type ApiServer struct {
	Router *gin.Engine
	server *http.Server
	ln     net.Listener
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (s *ApiServer) InitRouter() {
	// 注册 API 路由
	s.setUpDefaultRouters()
}

func (s *ApiServer) Start() error {
	// 初始化validator
	global.LOG.Info("Init validator")
	global.VALID = utils.InitValidator()

	// 获取连接配置
	settings, err := s.getServerSettings()
	if err != nil {
		global.LOG.Error("Failed to get server settings: %v", err)
		return err
	}

	global.LOG.Info("Server Settings: %v", settings)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", settings.BindIP, settings.BindPort),
		Handler: s.Router,
	}
	tcpItem := "tcp4"
	ln, err := net.Listen(tcpItem, server.Addr)
	if err != nil {
		global.LOG.Error("Failed to listen to %s", server.Addr)
		return err
	}

	s.server = server
	s.ln = ln

	if settings.Https == "yes" {
		var cert tls.Certificate
		var certPath string
		var keyPath string
		var err error
		if settings.HttpsCertType == "default" {
			certPath, keyPath, err = s.checkCertAndKey(settings.BindDomain)
			if err != nil {
				global.LOG.Error("Failed to check cert files")
				return err
			}
		} else {
			certPath = settings.HttpsCertPath
			keyPath = settings.HttpsKeyPath
		}
		certificate, err := os.ReadFile(certPath)
		if err != nil {
			global.LOG.Error("Failed to read cert file %s : %v", settings.HttpsCertPath, err)
			return err
		}
		global.CertPem = certificate
		key, err := os.ReadFile(keyPath)
		if err != nil {
			global.LOG.Error("Failed to read key file %s : %v", settings.HttpsKeyPath, err)
			return err
		}
		global.KeyPem = key
		cert, err = tls.X509KeyPair(certificate, key)
		if err != nil {
			global.LOG.Error("Failed to create tls cert pair")
			return err
		}
		server.TLSConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			MinVersion:         tls.VersionTLS13, // 设置最小 TLS 版本
			InsecureSkipVerify: true,
		}
		go func() {
			global.LOG.Info("listen at https://%s:%d [%s]", settings.BindIP, settings.BindPort, tcpItem)
			if err := server.ServeTLS(tcpKeepAliveListener{ln.(*net.TCPListener)}, certPath, keyPath); err != nil {
				global.LOG.Error("Listen at https://%s:%d [%s] Failed: %v", settings.BindIP, settings.BindPort, tcpItem, err)
				return
			}
		}()
	} else {
		go func() {
			global.LOG.Info("listen at http://%s:%d [%s]", settings.BindIP, settings.BindPort, tcpItem)
			if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
				global.LOG.Error("Listen at http://%s:%d [%s] Failed: %v", settings.BindIP, settings.BindPort, tcpItem, err)
				return
			}
		}()
	}
	return nil
}

func (s *ApiServer) checkCertAndKey(domain string) (string, string, error) {
	var certPath, keyPath string
	if domain == "" {
		certPath = filepath.Join(constant.CenterBinDir, "tls_cert.pem")
		keyPath = filepath.Join(constant.CenterBinDir, "tls_key.pem")
	} else {
		certPath = filepath.Join(constant.CenterBinDir, domain+"_cert.pem")
		keyPath = filepath.Join(constant.CenterBinDir, domain+"_key.pem")
	}

	// 如果证书和私钥已存在，则跳过生成
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			return "", "", nil
		}
	}

	// 加载CA证书和私钥
	rootCertPath := filepath.Join(constant.CenterBinDir, "cert.pem")
	rootKeyPath := filepath.Join(constant.CenterBinDir, "key.pem")
	caCertPEM, err := os.ReadFile(rootCertPath)
	if err != nil {
		return "", "", fmt.Errorf("读取 CA 证书失败: %v", err)
	}
	caKeyPEM, err := os.ReadFile(rootKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("读取 CA 私钥失败: %v", err)
	}

	// 解析CA证书
	caBlock, _ := pem.Decode(caCertPEM)
	if caBlock == nil {
		return "", "", errors.New("无法解析 CA 证书 PEM")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return "", "", fmt.Errorf("解析 CA 证书失败: %v", err)
	}

	// 解析CA私钥
	keyBlock, _ := pem.Decode(caKeyPEM)
	if keyBlock == nil {
		return "", "", errors.New("无法解析 CA 私钥 PEM")
	}

	var parsedKey any
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		// PKCS#1 格式
		parsedKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		// PKCS#8 格式
		parsedKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	default:
		return "", "", fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
	}
	if err != nil {
		return "", "", fmt.Errorf("解析 CA 私钥失败: %v", err)
	}

	// 确保是 *rsa.PrivateKey 类型
	caPrivateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return "", "", errors.New("CA 私钥不是 RSA 类型")
	}

	// 生成新密钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("生成私钥失败: %v", err)
	}

	// 生成序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("生成序列号失败: %v", err)
	}

	// === 构建 SAN 信息 ===
	var dnsNames []string
	var ipAddrs []net.IP

	// 必备 IP（通用）
	ipAddrs = append(ipAddrs, net.ParseIP("127.0.0.1"))
	ipAddrs = append(ipAddrs, net.ParseIP("::1"))

	// 内网IP
	if ip := getLocalIP(); ip != nil {
		ipAddrs = append(ipAddrs, ip)
	}

	if domain == "" {
		// 自签证书模式
		dnsNames = append(dnsNames, "localhost")
		if ip := net.ParseIP(global.Host); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		}
	} else {
		// 绑定域名模式
		dnsNames = append(dnsNames, domain)
	}

	// === 构建 Subject ===
	subject := pkix.Name{
		Country:            []string{"CN"},
		Organization:       []string{"Sensdata"},
		OrganizationalUnit: []string{"iDB AutoCert"},
		CommonName:         domain,
	}
	if domain == "" {
		subject.CommonName = "localhost"
	}

	// === 构建证书模板 ===
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now().Add(-10 * time.Minute),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  ipAddrs,
		DNSNames:     dnsNames,
	}

	global.LOG.Info("生成证书: %s\nDNS: %v\nIP: %v\n", certPath, dnsNames, ipAddrs)

	// 签发证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &priv.PublicKey, caPrivateKey)
	if err != nil {
		return "", "", fmt.Errorf("签发证书失败: %v", err)
	}

	// 写入证书
	if err := writePemFile(certPath, "CERTIFICATE", certDER); err != nil {
		return "", "", err
	}

	// 写入私钥
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := writePemFile(keyPath, "RSA PRIVATE KEY", privBytes); err != nil {
		return "", "", err
	}

	return certPath, keyPath, nil
}

func getLocalIP() net.IP {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.To4() != nil {
				return ip
			}
		}
	}
	return nil
}

func writePemFile(path, pemType string, derBytes []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	return pem.Encode(file, &pem.Block{Type: pemType, Bytes: derBytes})
}

func (s *ApiServer) Stop() error {
	global.LOG.Info("正在停止 API 服务器...")

	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭监听器
	if s.ln != nil {
		if err := s.ln.Close(); err != nil {
			global.LOG.Error("监听器关闭失败: %v", err)
		}
	}

	// 优雅关闭 HTTP 服务器
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			global.LOG.Error("HTTP 服务器关闭失败: %v", err)
			return err
		}
	}

	global.LOG.Info("API 服务器已停止")
	return nil
}

func (s *ApiServer) getServerSettings() (*model.SettingInfo, error) {
	settingRepo := repo.NewSettingsRepo()
	bindIP, err := settingRepo.Get(settingRepo.WithByKey("BindIP"))
	if err != nil {
		return nil, err
	}
	bindPort, err := settingRepo.Get(settingRepo.WithByKey("BindPort"))
	if err != nil {
		return nil, err
	}
	bindPortValue, err := strconv.Atoi(bindPort.Value)
	if err != nil {
		return nil, err
	}
	https, err := settingRepo.Get(settingRepo.WithByKey("Https"))
	if err != nil {
		return nil, err
	}
	httpsCertType, err := settingRepo.Get(settingRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return nil, err
	}
	httpsCertPath, err := settingRepo.Get(settingRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return nil, err
	}
	httpsKeyPath, err := settingRepo.Get(settingRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return nil, err
	}

	return &model.SettingInfo{
		BindIP:        bindIP.Value,
		BindPort:      bindPortValue,
		Https:         https.Value,
		HttpsCertType: httpsCertType.Value,
		HttpsCertPath: httpsCertPath.Value,
		HttpsKeyPath:  httpsKeyPath.Value,
	}, nil
}

// SetupRouter sets up the API routes
func (s *ApiServer) setUpDefaultRouters() {
	global.LOG.Info("register router - api")
	apiGroup := s.Router.Group("api/v1")

	// 添加全局日志中间件
	apiGroup.Use(func(c *gin.Context) {
		// 记录请求信息
		global.LOG.Info("Request: %s %s", c.Request.Method, c.Request.URL.Path)

		// 根据请求方法打印不同的信息
		if c.Request.Method == "GET" {
			global.LOG.Info("Query: %s", c.Request.URL.Query())
		} else if c.Request.Method == "POST" {
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)                 // 读取请求体
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新设置请求体
			}
			global.LOG.Info("Body: %s", string(bodyBytes))
		}
		c.Next() // 继续处理请求
		// 记录响应信息
		global.LOG.Info("Response: %d", c.Writer.Status())
	})

	// 绑定域名过滤
	apiGroup.Use(middleware.BindDomain())
	// 初始化路由
	for _, router := range router.RouterGroups {
		router.InitRouter(apiGroup)
	}

	// 添加 Swagger 路由到 apiGroup
	global.LOG.Info("register router - swagger")
	swaggerGroup := apiGroup.Group("swagger")
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 处理未匹配的请求，重定向到 /var/lib/idb/home
	s.Router.NoRoute(func(c *gin.Context) {
		// 这里可以使用 c.FileServer 来处理目录下的所有请求
		c.File("/var/lib/idb/home/index.html") // 默认返回 index.html
	})

	// 设置静态文件路由，确保可以访问 assets 目录
	s.Router.Static("/assets", "/var/lib/idb/home/assets") // 处理 assets 目录
}

// SetUpPluginRouters sets up routers from plugins
func (s *ApiServer) SetUpPluginRouters(group string, routes []plugin.PluginRoute) {
	global.LOG.Info("register router - %s", group)
	pluginGroup := s.Router.Group("api/v1/" + group)
	pluginGroup.Use(middleware.NewJWT().JWTAuth())
	for _, route := range routes {
		switch route.Method {
		case "GET":
			pluginGroup.GET(route.Path, route.Handler)
		case "POST":
			pluginGroup.POST(route.Path, route.Handler)
		case "DELETE":
			pluginGroup.DELETE(route.Path, route.Handler)
		case "PUT":
			pluginGroup.PUT(route.Path, route.Handler)
		}
	}

	// 设置 git 路由
	gitGroup := s.Router.Group("api/v1/git/" + group)
	repoPath := fmt.Sprintf("/var/lib/idb/data/%s/global", group)
	// 处理 git clone/pull 请求
	gitGroup.GET("/*path", s.handleGitRoute(repoPath))
	gitGroup.POST("/*path", s.handleGitRoute(repoPath))
}

func (s *ApiServer) handleGitRoute(repoPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := git.PlainOpen(repoPath)
		if err != nil {
			global.LOG.Error("Failed to open repo %s, %v", repoPath, err)
			c.String(http.StatusInternalServerError, "Failed to open repository")
			return
		}
		s.handleGitRequest(c, r)
	}
}

func (s *ApiServer) handleGitRequest(c *gin.Context, repo *git.Repository) {
	path := c.Param("path")
	global.LOG.Info("Git request: %s %s", c.Request.Method, path)

	switch {
	case strings.HasSuffix(path, "/info/refs"):
		global.LOG.Info("Handling info/refs request")
		s.handleGitInfoRefs(c, repo)
	case strings.HasSuffix(path, "/git-upload-pack"):
		global.LOG.Info("Handling git-upload-pack request")
		s.handleGitUploadPack(c, repo)
	case strings.HasSuffix(path, "/git-receive-pack"):
		global.LOG.Info("Push operation not allowed")
		c.String(http.StatusForbidden, "Push operation not allowed")
	default:
		global.LOG.Info("Unknown git operation: %s", path)
		c.String(http.StatusNotFound, "Not Found")
	}
}

func (s *ApiServer) handleGitInfoRefs(c *gin.Context, repo *git.Repository) {
	c.Header("Content-Type", "application/x-git-upload-pack-advertisement")

	// 按照 Git 协议格式发送服务能力声明
	serviceLine := "# service=git-upload-pack\n"
	pktLine := fmt.Sprintf("%04x%s", len(serviceLine)+4, serviceLine)
	c.Writer.Write([]byte(pktLine))
	c.Writer.Write([]byte("0000")) // flush-pkt

	// 获取并发送引用信息
	refs, err := repo.References()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get references")
		return
	}

	// 首先发送 HEAD 引用，添加 multi_ack_detailed 支持
	head, err := repo.Head()
	if err == nil {
		capabilities := "multi_ack_detailed multi_ack thin-pack side-band side-band-64k ofs-delta shallow deepen-since deepen-not allow-tip-sha1-in-want allow-reachable-sha1-in-want no-progress include-tag"
		line := fmt.Sprintf("%s HEAD\x00%s\n", head.Hash(), capabilities)
		c.Writer.Write([]byte(fmt.Sprintf("%04x%s", len(line)+4, line)))
	}

	// 发送其他引用
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() || ref.Name().IsTag() {
			line := fmt.Sprintf("%s %s\n", ref.Hash(), ref.Name())
			c.Writer.Write([]byte(fmt.Sprintf("%04x%s", len(line)+4, line)))
		}
		return nil
	})

	c.Writer.Write([]byte("0000")) // 结束标记
}

func (s *ApiServer) handleGitUploadPack(c *gin.Context, repo *git.Repository) {
	c.Header("Content-Type", "application/x-git-upload-pack-result")
	global.LOG.Info("Starting git upload-pack process")

	// 读取并解析客户端的请求
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		global.LOG.Error("Failed to read request body: %v", err)
		c.String(http.StatusBadRequest, "Failed to read request")
		return
	}
	global.LOG.Info("Received request body length: %d bytes", len(body))

	// 解析 want/have 行
	var wantHashes []plumbing.Hash
	data := string(body)
	for len(data) > 0 {
		if len(data) < 4 {
			break
		}
		length, err := strconv.ParseInt(data[:4], 16, 32)
		if err != nil || length == 0 {
			data = data[4:]
			continue
		}
		if int(length) > len(data) {
			break
		}
		line := data[4:length]
		if strings.HasPrefix(line, "want ") {
			hashStr := strings.TrimPrefix(line, "want ")
			hashStr = strings.Split(hashStr, " ")[0]
			hash := plumbing.NewHash(hashStr)
			wantHashes = append(wantHashes, hash)
			global.LOG.Info("Parsed want hash: %s", hashStr)
		}
		data = data[length:]
	}

	if len(wantHashes) == 0 {
		global.LOG.Error("No want lines received in request")
		c.String(http.StatusBadRequest, "No want lines received")
		return
	}
	global.LOG.Info("Total want hashes: %d", len(wantHashes))

	// 首先发送 NAK
	nakLine := "NAK\n"
	c.Writer.Write([]byte(fmt.Sprintf("%04x%s", len(nakLine)+4, nakLine)))
	global.LOG.Info("Sent NAK response")

	// 使用 side-band 协议包装 writer
	writer := &bandWriter{
		w: c.Writer,
	}

	// 发送进度信息
	s.writePktLine(c, "\x02Counting objects\n")

	// 收集所有需要的对象
	objectSet := make(map[plumbing.Hash]struct{})
	for _, hash := range wantHashes {
		// 获取提交对象
		commit, err := repo.CommitObject(hash)
		if err != nil {
			global.LOG.Error("Failed to get commit object %s: %v", hash, err)
			continue
		}

		// 递归收集所有父提交
		queue := []*object.Commit{commit}
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			// 添加当前提交
			objectSet[current.Hash] = struct{}{}
			global.LOG.Info("Added commit object: %s", current.Hash)

			// 添加父提交到队列
			for _, parent := range current.ParentHashes {
				parentCommit, err := repo.CommitObject(parent)
				if err != nil {
					global.LOG.Error("Failed to get parent commit %s: %v", parent, err)
					continue
				}
				queue = append(queue, parentCommit)
			}

			// 获取并添加树对象
			tree, err := current.Tree()
			if err != nil {
				global.LOG.Error("Failed to get tree for commit %s: %v", current.Hash, err)
				continue
			}
			objectSet[tree.Hash] = struct{}{}
			global.LOG.Info("Added tree object: %s", tree.Hash)

			// 遍历所有树对象
			trees := []*object.Tree{tree}
			for len(trees) > 0 {
				currentTree := trees[0]
				trees = trees[1:]

				for _, entry := range currentTree.Entries {
					objectSet[entry.Hash] = struct{}{}
					global.LOG.Info("Added entry object: %s", entry.Hash)

					// 如果是子树，添加到遍历队列
					if entry.Mode == filemode.Dir {
						subTree, err := repo.TreeObject(entry.Hash)
						if err != nil {
							global.LOG.Error("Failed to get subtree %s: %v", entry.Hash, err)
							continue
						}
						trees = append(trees, subTree)
					}
				}
			}
		}
	}

	// 转换为切片
	var objects []plumbing.Hash
	for hash := range objectSet {
		objects = append(objects, hash)
	}
	global.LOG.Info("Total unique objects collected: %d", len(objects))

	// 发送 packfile
	s.writePktLine(c, "\x02Preparing packfile\n")

	// 创建 packfile encoder
	encoder := packfile.NewEncoder(writer.DataBand(), repo.Storer, false)

	// 编码所有对象
	packHash, err := encoder.Encode(objects, 10)
	if err != nil {
		global.LOG.Error("Failed to encode packfile: %v", err)
		s.writePktLine(c, "\x02Error creating packfile\n")
		return
	}
	global.LOG.Info("Successfully created packfile with hash: %s", packHash)

	// 发送完成信息
	s.writePktLine(c, "\x02Done\n")
	s.writePktLine(c, "")
	global.LOG.Info("Git upload-pack process completed successfully")
}

// writePktLine 写入一个 pkt-line
func (s *ApiServer) writePktLine(c *gin.Context, data string) {
	if data == "" {
		c.Writer.Write([]byte("0000"))
		return
	}
	pktLine := fmt.Sprintf("%04x%s", len(data)+4, data)
	c.Writer.Write([]byte(pktLine))
}

// bandWriter 实现 side-band 协议
type bandWriter struct {
	w http.ResponseWriter
}

func (w *bandWriter) DataBand() io.Writer {
	return &bandWriterBand{w: w.w, band: 1}
}

type bandWriterBand struct {
	w    http.ResponseWriter
	band byte
}

func (w *bandWriterBand) Write(p []byte) (int, error) {
	totalWritten := 0
	originalLen := len(p)

	for len(p) > 0 {
		chunk := p
		if len(chunk) > 8192 {
			chunk = chunk[:8192]
		}

		// 计算 pkt-line 头部
		pktLine := fmt.Sprintf("%04x", len(chunk)+5)

		// 写入头部
		n, err := w.w.Write([]byte(pktLine))
		if err != nil {
			global.LOG.Error("Failed to write pkt-line header: %v", err)
			return totalWritten, err
		}

		// 写入 band 标识
		n, err = w.w.Write([]byte{w.band})
		if err != nil {
			global.LOG.Error("Failed to write band identifier: %v", err)
			return totalWritten, err
		}

		// 写入数据块
		n, err = w.w.Write(chunk)
		if err != nil {
			global.LOG.Error("Failed to write chunk data: %v", err)
			return totalWritten, err
		}
		totalWritten += n

		// 更新剩余数据
		p = p[len(chunk):]
	}

	// 返回原始数据长度，表示全部写入成功
	return originalLen, nil
}

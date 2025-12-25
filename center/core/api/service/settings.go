package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sensdata/idb/center/core/conn"
	db "github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/message"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type SettingsService struct{}

type ISettingsService interface {
	About() (*model.About, error)
	IPs() (*model.AvailableIps, error)
	Timezones(req model.SearchPageInfo) (*model.PageResult, error)
	Settings() (*model.SettingInfo, error)
	Update(req model.UpdateSettingRequest) (*model.UpdateSettingResponse, error)
	Upgrade() error
}

func NewISettingsService() ISettingsService {
	return &SettingsService{}
}

func (s *SettingsService) About() (*model.About, error) {
	var about model.About

	about.Version = global.Version

	// 获取新版本信息
	about.NewVersion = getLatestVersion()

	return &about, nil
}

func getLatestVersion() string {
	cmd := fmt.Sprintf("curl -sSL %s", conn.CONFMAN.GetConfig().Latest)
	global.LOG.Info("Getting latest version: %s", cmd)
	latest, err := utils.Exec(cmd)
	if err != nil {
		global.LOG.Error("Failed to get latest version: %v", err)
		return ""
	}
	global.LOG.Info("Got latest version: %s", latest)
	return strings.TrimSpace(latest)
}

func (s *SettingsService) IPs() (*model.AvailableIps, error) {
	var availableIps model.AvailableIps
	availableIps.IPs = make([]model.BindIp, 0)

	// 添加几项ip：
	// 所有IP - 0.0.0.0
	availableIps.IPs = append(availableIps.IPs, model.BindIp{IP: "0.0.0.0", Name: "All IP"})
	// 127.0.0.1 - 127.0.0.1
	availableIps.IPs = append(availableIps.IPs, model.BindIp{IP: "127.0.0.1", Name: "127.0.0.1"})
	// ::1 - ::1
	availableIps.IPs = append(availableIps.IPs, model.BindIp{IP: "::1", Name: "::1"})
	// Link-Local Address
	interfaces, err := net.Interfaces()
	if err != nil {
		return &availableIps, nil
	}
	for _, iface := range interfaces {
		// 只要 eth0
		if iface.Name != "eth0" {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			// 获取 Link-Local 地址
			if ipNet.IP.IsLinkLocalUnicast() {
				availableIps.IPs = append(
					availableIps.IPs,
					model.BindIp{IP: ipNet.IP.String(), Name: ipNet.IP.String()},
				)
			}
		}
	}

	return &availableIps, nil
}

func (s *SettingsService) Timezones(req model.SearchPageInfo) (*model.PageResult, error) {
	var (
		result    model.PageResult
		backDatas []db.Timezone
	)

	// 获取所有时区
	timezones, err := TimezoneRepo.GetList()
	if err != nil {
		return &result, err
	}
	// 分页
	total, start, end := len(timezones), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]db.Timezone, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = timezones[start:end]
	}

	result.Total = int64(total)
	result.Items = backDatas

	return &result, nil
}
func (s *SettingsService) Settings() (*model.SettingInfo, error) {
	bindIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return nil, err
	}
	bindPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
	if err != nil {
		return nil, err
	}
	bindPortValue, err := strconv.Atoi(bindPort.Value)
	if err != nil {
		return nil, err
	}
	bindDomain, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindDomain"))
	if err != nil {
		return nil, err
	}
	https, err := SettingsRepo.Get(SettingsRepo.WithByKey("Https"))
	if err != nil {
		return nil, err
	}
	httpsCertType, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertType"))
	if err != nil {
		return nil, err
	}
	httpsCertPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsCertPath"))
	if err != nil {
		return nil, err
	}
	httpsKeyPath, err := SettingsRepo.Get(SettingsRepo.WithByKey("HttpsKeyPath"))
	if err != nil {
		return nil, err
	}

	return &model.SettingInfo{
		BindIP:        bindIP.Value,
		BindPort:      bindPortValue,
		BindDomain:    bindDomain.Value,
		Https:         https.Value,
		HttpsCertType: httpsCertType.Value,
		HttpsCertPath: httpsCertPath.Value,
		HttpsKeyPath:  httpsKeyPath.Value,
	}, nil
}

func (s *SettingsService) Update(req model.UpdateSettingRequest) (*model.UpdateSettingResponse, error) {
	var response model.UpdateSettingResponse
	var scheme string
	switch req.Https {
	case "no":
		scheme = "http"
	case "yes":
		scheme = "https"
		// 检查证书
		if err := s.checkCertAndKey(req); err != nil {
			global.LOG.Error("Failed to check cert and key: %v", err)
			return &response, err
		}
	default:
		global.LOG.Error("Invalid https value: %s", req.Https)
		return &response, errors.New("invalid https value")
	}

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return &response, err
	}

	// 开始事务
	tx := global.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			global.LOG.Error("Transaction failed: %v  - rollback", r)
		} else if err != nil {
			tx.Rollback() // 如果发生错误，回滚事务
			global.LOG.Error("Error Happend - rollback")
		}
	}()

	if err = s.updateBindIP(req.BindIP); err != nil {
		global.LOG.Error("Failed to save BindIP to %s: %v", req.BindIP, err)
		return &response, err
	}

	if err = s.updateBindPort(req.BindPort); err != nil {
		global.LOG.Error("Failed to save BindPort to %d: %v", req.BindPort, err)
		return &response, err
	}

	if err = s.updateBindDomain(req.BindDomain); err != nil {
		global.LOG.Error("Failed to save BindDomain to %s: %v", req.BindDomain, err)
		return &response, err
	}

	if err = s.updateHttps(req); err != nil {
		global.LOG.Error("Failed to save Https settings: %v", err)
		return &response, err
	}

	// 提交事务
	tx.Commit()

	var url string
	if len(req.BindDomain) == 0 {
		url = fmt.Sprintf("%s://%s:%d/manage/settings", scheme, host.Addr, req.BindPort)
	} else {
		url = fmt.Sprintf("%s://%s:%d/manage/settings", scheme, req.BindDomain, req.BindPort)
	}
	response.RedirectUrl = url

	go func() {
		// 发送 SIGTERM 信号给主进程，触发容器重启
		if err := syscall.Kill(1, syscall.SIGTERM); err != nil {
			global.LOG.Error("Failed to send termination signal: %v", err)
		}
	}()

	return &response, nil
}

func (s *SettingsService) checkCertAndKey(req model.UpdateSettingRequest) error {
	switch req.HttpsCertType {
	// 默认证书，看看是否签发
	case "default":
		if err := s.checkDefaultCert(req.BindDomain); err != nil {
			global.LOG.Error("Failed to check default cert: %v", err)
			return err
		}
	// 自定义证书，需要获取内容
	case "custom":
		if len(req.HttpsCertPath) == 0 || len(req.HttpsKeyPath) == 0 {
			global.LOG.Error("Invalid cert path or key path: %s, %s", req.HttpsCertPath, req.HttpsKeyPath)
			return errors.New("invalid cert path or key path")
		}
		if err := s.checkCustomCert(req.HttpsCertPath, req.HttpsKeyPath); err != nil {
			global.LOG.Error("Failed to check custom cert: %v", err)
			return err
		}
	default:
		global.LOG.Error("Invalid cert type: %s", req.HttpsCertType)
		return errors.New("invalid cert type")
	}
	return nil
}

func (s *SettingsService) checkDefaultCert(domain string) error {
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
			global.LOG.Info("Default cert and key already exist: %s, %s", certPath, keyPath)

			// 保存到db
			if err := s.saveCertToDB(certPath, keyPath); err != nil {
				return err
			}
			return nil
		}
	}

	// 加载CA证书和私钥
	rootCertPath := filepath.Join(constant.CenterBinDir, "cert.pem")
	rootKeyPath := filepath.Join(constant.CenterBinDir, "key.pem")
	caCertPEM, err := os.ReadFile(rootCertPath)
	if err != nil {
		global.LOG.Error("Failed to read CA cert: %v", err)
		return fmt.Errorf("读取 CA 证书失败: %v", err)
	}
	caKeyPEM, err := os.ReadFile(rootKeyPath)
	if err != nil {
		global.LOG.Error("Failed to read CA key: %v", err)
		return fmt.Errorf("读取 CA 私钥失败: %v", err)
	}

	// 解析CA证书
	caBlock, _ := pem.Decode(caCertPEM)
	if caBlock == nil {
		global.LOG.Error("Failed to decode CA cert PEM")
		return fmt.Errorf("无法解析 CA 证书 PEM")
	}
	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		global.LOG.Error("Failed to parse CA cert: %v", err)
		return fmt.Errorf("解析 CA 证书失败: %v", err)
	}

	// 解析CA私钥
	keyBlock, _ := pem.Decode(caKeyPEM)
	if keyBlock == nil {
		global.LOG.Error("Failed to decode CA key PEM")
		return fmt.Errorf("无法解析 CA 私钥 PEM")
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
		global.LOG.Error("Unsupported private key type: %s", keyBlock.Type)
		return fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
	}
	if err != nil {
		global.LOG.Error("Failed to parse CA key: %v", err)
		return fmt.Errorf("解析 CA 私钥失败: %v", err)
	}

	// 确保是 *rsa.PrivateKey 类型
	caPrivateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		global.LOG.Error("CA private key is not RSA type")
		return fmt.Errorf("CA 私钥不是 RSA 类型")
	}

	// 生成新密钥
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		global.LOG.Error("Failed to generate private key: %v", err)
		return fmt.Errorf("生成私钥失败: %v", err)
	}

	// 生成序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		global.LOG.Error("Failed to generate serial number: %v", err)
		return fmt.Errorf("生成序列号失败: %v", err)
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
		global.LOG.Error("Failed to create certificate: %v", err)
		return fmt.Errorf("签发证书失败: %v", err)
	}

	// 写入证书
	if err := writePemFile(certPath, "CERTIFICATE", certDER); err != nil {
		global.LOG.Error("Failed to write certificate: %v", err)
		return err
	}

	// 写入私钥
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := writePemFile(keyPath, "RSA PRIVATE KEY", privBytes); err != nil {
		global.LOG.Error("Failed to write private key: %v", err)
		return err
	}

	// 保存到db
	if err := s.saveCertToDB(certPath, keyPath); err != nil {
		return err
	}
	return nil
}

func (s *SettingsService) checkCustomCert(certPath, keyPath string) error {
	// 获取cert文件内容
	fileInfo, err := s.getFileContent(certPath)
	if err != nil {
		return err
	}

	// 检查key文件内容
	keyInfo, err := s.getFileContent(keyPath)
	if err != nil {
		return err
	}

	// 保存到db
	if err := SettingsRepo.Update("HttpsCertData", fileInfo.Content); err != nil {
		return err
	}

	if err := SettingsRepo.Update("HttpsKeyData", keyInfo.Content); err != nil {
		return err
	}

	return nil
}

func (s *SettingsService) getFileContent(path string) (*model.FileInfo, error) {
	var fileInfo model.FileInfo

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return &fileInfo, err
	}

	req := model.FileContentReq{
		Path:   path,
		Expand: true,
	}
	data, err := utils.ToJSONString(req)
	if err != nil {
		return &fileInfo, err
	}

	actionRequest := model.HostAction{
		HostID: host.ID,
		Action: model.Action{
			Action: model.File_Content,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		return &fileInfo, err
	}

	if !actionResponse.Result {
		global.LOG.Error("action failed")
		if strings.Contains(actionResponse.Data, "no such file or directory") {
			return &fileInfo, constant.ErrFileNotExist
		}
		return &fileInfo, fmt.Errorf("failed to get file content")
	}

	err = utils.FromJSONString(actionResponse.Data, &fileInfo)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to file content: %v", err)
		return &fileInfo, fmt.Errorf("json err: %v", err)
	}
	return &fileInfo, nil
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

func (s *SettingsService) saveCertToDB(certPath, keyPath string) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		global.LOG.Error("Failed to read cert file: %v", err)
		return fmt.Errorf("读取cert证书失败: %v", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		global.LOG.Error("Failed to read key file: %v", err)
		return fmt.Errorf("读取key证书失败: %v", err)
	}
	if err := SettingsRepo.Upsert("HttpsCertData", string(certPEM)); err != nil {
		global.LOG.Error("Failed to update cert data: %v", err)
		return err
	}

	if err := SettingsRepo.Upsert("HttpsKeyData", string(keyPEM)); err != nil {
		global.LOG.Error("Failed to update key data: %v", err)
		return err
	}
	return nil
}

func (s *SettingsService) updateBindIP(newIP string) error {
	if len(newIP) == 0 {
		return errors.New("invalid bind ip")
	}

	oldIP, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindIP"))
	if err != nil {
		return err
	}
	if newIP == oldIP.Value {
		return nil
	}

	return SettingsRepo.Update("BindIP", newIP)
}

func (s *SettingsService) updateBindPort(newPort int) error {
	if newPort <= 0 || newPort > 65535 {
		return errors.New("server port must between 1 - 65535")
	}
	oldPort, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindPort"))
	if err != nil {
		return err
	}
	newPortStr := strconv.Itoa(newPort)
	if newPortStr == oldPort.Value {
		return nil
	}

	if common.ScanPort(newPort) {
		return errors.New(constant.ErrPortInUsed)
	}

	// TODO: 处理port的更换（调用nftables）

	return SettingsRepo.Update("BindPort", newPortStr)
}

func (s *SettingsService) updateBindDomain(newDomain string) error {
	domain := newDomain
	if newDomain == "empty" {
		domain = ""
	}
	oldDomain, err := SettingsRepo.Get(SettingsRepo.WithByKey("BindDomain"))
	if err != nil {
		return err
	}
	if domain == oldDomain.Value {
		return nil
	}
	return SettingsRepo.Update("BindDomain", domain)
}

func (s *SettingsService) updateHttps(req model.UpdateSettingRequest) error {
	if err := SettingsRepo.Update("Https", req.Https); err != nil {
		return err
	}

	if req.Https == "yes" {
		if err := SettingsRepo.Update("HttpsCertType", req.HttpsCertType); err != nil {
			return err
		}

		if req.HttpsCertType == "custom" {
			if err := SettingsRepo.Update("HttpsCertPath", req.HttpsCertPath); err != nil {
				return err
			}

			if err := SettingsRepo.Update("HttpsKeyPath", req.HttpsKeyPath); err != nil {
				return err
			}
		} else {
			if err := SettingsRepo.Update("HttpsCertPath", ""); err != nil {
				return err
			}

			if err := SettingsRepo.Update("HttpsKeyPath", ""); err != nil {
				return err
			}
		}
	} else {
		if err := SettingsRepo.Update("HttpsCertType", "default"); err != nil {
			return err
		}

		if err := SettingsRepo.Update("HttpsCertPath", ""); err != nil {
			return err
		}

		if err := SettingsRepo.Update("HttpsKeyPath", ""); err != nil {
			return err
		}
	}

	return nil
}

func (s *SettingsService) Upgrade() error {
	newVersion := getLatestVersion()
	if len(newVersion) == 0 {
		return errors.New("failed to get latest version")
	}

	if global.Version == newVersion {
		return errors.New("already latest version")
	}

	// 找宿主机host
	host, err := HostRepo.Get(HostRepo.WithByDefault())
	if err != nil {
		global.LOG.Error("Failed to get default host")
		return err
	}

	agentConn, err := conn.CENTER.GetAgentConn(&host)
	if err != nil {
		global.LOG.Error("Failed to get agent connection")
		return err
	}

	// 创建消息
	cmd := fmt.Sprintf("curl -sSL https://static.sensdata.com/idb/release/upgrade.sh -o /tmp/upgrade.sh && bash /tmp/upgrade.sh %s", newVersion)

	msgID := utils.GenerateMsgId()
	msg, err := message.CreateMessage(
		msgID,
		cmd,
		host.AgentKey,
		utils.GenerateNonce(16),
		global.Version,
		message.CmdMessage,
	)
	if err != nil {
		global.LOG.Error("Failed to create command message: %v", err)
		return err
	}
	err = message.SendMessage(*agentConn, msg)
	if err != nil {
		global.LOG.Error("Failed to send command message: %v", err)
		return err
	}

	return nil
}

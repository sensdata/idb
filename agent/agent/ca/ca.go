package ca

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type CaService struct {
	rootCertMap         map[string]*x509.Certificate
	intermediateCertMap map[string]*x509.Certificate
	mu                  sync.Mutex // 保证并发安全
}

type ICaService interface {
	GetCertificateGroups() (*model.PageResult, error)
	GetPrivateKeyInfo(req model.GroupPkRequest) (*model.PrivateKeyInfo, error)
	GetCSRInfo(req model.GroupPkRequest) (*model.CSRInfo, error)
	GenerateCertificate(req model.CreateGroupRequest) error
	RemoveCertificateGroup(req model.DeleteGroupRequest) error
	GenerateSelfSignedCertificate(req model.SelfSignedRequest) error
	GetCertificateInfo(req model.CertificateInfoRequest) (*model.CertificateInfo, error)
	CompleteCertificateChain(req model.CertificateInfoRequest) error
	RemoveCertificate(req model.DeleteCertificateRequest) error
	ImportCertificate(req model.ImportCertificateRequest) error
	UpdateCertificate(req model.UpdateCertificateRequest) error
}

func NewICaService() ICaService {
	return &CaService{}
}

func (s *CaService) GenerateCertificate(req model.CreateGroupRequest) error {

	// 1. 生成存储目录路径
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)
	if err := utils.EnsurePaths([]string{certificateDir}); err != nil {
		global.LOG.Error("Failed to create dir %s, %v", certificateDir, err)
		return err
	}

	// 2. 根据密钥算法生成私钥并保存
	privateKeyPath := filepath.Join(certificateDir, req.Alias+".key")
	var privateKey []byte
	var err error
	switch req.KeyAlgorithm {
	case "RSA 2048":
		privateKey, err = generateRSAKey(2048)
	case "RSA 3072":
		privateKey, err = generateRSAKey(3072)
	case "RSA 4096":
		privateKey, err = generateRSAKey(4096)
	case "EC 256":
		privateKey, err = generateECDSAKey("P-256")
	case "EC 384":
		privateKey, err = generateECDSAKey("P-384")
	default:
		return fmt.Errorf("unsupported key algorithm: %s", req.KeyAlgorithm)
	}
	if err != nil {
		return err
	}

	// 保存私钥到文件
	err = savePrivateKey(privateKeyPath, privateKey)
	if err != nil {
		return err
	}

	// 3. 生成CSR文件
	csrPath := filepath.Join(certificateDir, req.Alias+".csr")
	csrBytes, err := generateCSR(req, privateKey)
	if err != nil {
		return err
	}

	// 保存CSR文件
	err = saveCSR(csrPath, csrBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *CaService) GenerateSelfSignedCertificate(req model.SelfSignedRequest) error {
	// 校验有效期
	var expireDuration time.Duration
	if req.ExpireUnit == "year" {
		expireDuration = time.Duration(req.ExpireValue) * 365 * 24 * time.Hour
	} else if req.ExpireUnit == "day" {
		expireDuration = time.Duration(req.ExpireValue) * 24 * time.Hour
	} else {
		return fmt.Errorf("invalid expire unit: %s", req.ExpireUnit)
	}

	// 校验域名
	var domains []string
	if req.AltDomains != "" {
		domainArray := strings.Split(req.AltDomains, "\n")
		for _, domain := range domainArray {
			if !common.IsValidDomain(domain) {
				return fmt.Errorf("invalid domain: %s", domain)
			} else {
				domains = append(domains, domain)
			}
		}
	}

	// 校验 IP
	var ips []net.IP
	if req.AltIPs != "" {
		ipArray := strings.Split(req.AltIPs, "\n")
		for _, ip := range ipArray {
			if ipAddr := net.ParseIP(ip); ipAddr == nil {
				return fmt.Errorf("invalid ip: %s", ip)
			} else {
				ips = append(ips, ipAddr)
			}
		}
	}

	// 根据 Alias 查找目录下的 .csr 和 .key 文件
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)
	csrPath := filepath.Join(certificateDir, req.Alias+".csr")
	keyPath := filepath.Join(certificateDir, req.Alias+".key")

	// 检查文件是否存在
	if _, err := os.Stat(csrPath); os.IsNotExist(err) {
		return fmt.Errorf("CSR file not found: %s", csrPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("Key file not found: %s", keyPath)
	}

	// 读取私钥
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}
	privateBlock, _ := pem.Decode(privateKeyBytes)
	if privateBlock == nil {
		return fmt.Errorf("failed to decode private key PEM block")
	}

	// 读取 CSR
	csrBytes, err := os.ReadFile(csrPath)
	if err != nil {
		return fmt.Errorf("failed to read CSR file: %v", err)
	}
	csrBlock, _ := pem.Decode(csrBytes)
	if csrBlock == nil || csrBlock.Type != "CERTIFICATE REQUEST" {
		return fmt.Errorf("failed to decode CSR PEM block")
	}

	// 解析私钥（支持多种格式）
	var privateKey interface{}
	switch privateBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(privateBlock.Bytes)
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
	case "ENCRYPTED PRIVATE KEY":
		return fmt.Errorf("encrypted private key not supported yet, please provide password")
	default:
		return fmt.Errorf("unsupported private key type: %s", privateBlock.Type)
	}
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	// 提取公钥
	var publicKey interface{}
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		publicKey = &k.PublicKey
	case *ecdsa.PrivateKey:
		publicKey = &k.PublicKey
	case ed25519.PrivateKey:
		publicKey = k.Public().(ed25519.PublicKey)
	default:
		return fmt.Errorf("unsupported private key type %T", k)
	}

	// 解析 CSR
	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CSR: %v", err)
	}

	// 证书模板
	certTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               csr.Subject,
		EmailAddresses:        csr.EmailAddresses,
		Issuer:                csr.Subject, // 自签：签发者 = 自己
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(expireDuration),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  req.IsCA,
		BasicConstraintsValid: true,
		DNSNames:              domains,
		IPAddresses:           ips,
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// 保存证书文件
	timestamp := time.Now().Unix()
	certPath := filepath.Join(certificateDir, fmt.Sprintf("%d.crt", timestamp))
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %v", err)
	}
	defer certFile.Close()

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return fmt.Errorf("failed to encode certificate to PEM: %v", err)
	}

	return nil
}

func (s *CaService) GetPrivateKeyInfo(req model.GroupPkRequest) (*model.PrivateKeyInfo, error) {
	var privateKeyInfo model.PrivateKeyInfo

	// 根据 alias 构建文件路径
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)
	keyPath := fmt.Sprintf("%s/%s.key", certificateDir, req.Alias)

	// 检查 .key 文件是否存在
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return &privateKeyInfo, fmt.Errorf("key file not found: %s", keyPath)
	}

	// 读取私钥文件
	privKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return &privateKeyInfo, fmt.Errorf("failed to read private key: %v", err)
	}

	// 解码 PEM 格式
	block, _ := pem.Decode(privKeyBytes)
	if block == nil {
		return &privateKeyInfo, fmt.Errorf("failed to decode PEM block")
	}

	var (
		parsedPrivKey interface{}
		keyAlgorithm  string
		keySize       int
	)

	switch block.Type {
	case "RSA PRIVATE KEY":
		// PKCS#1 (RSA)
		parsedPrivKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return &privateKeyInfo, fmt.Errorf("failed to parse RSA private key: %v", err)
		}
		keyAlgorithm = "RSA"
		keySize = parsedPrivKey.(*rsa.PrivateKey).N.BitLen()

	case "EC PRIVATE KEY":
		// SEC1 (EC)
		parsedPrivKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return &privateKeyInfo, fmt.Errorf("failed to parse EC private key: %v", err)
		}
		keyAlgorithm = "ECDSA"
		keySize = parsedPrivKey.(*ecdsa.PrivateKey).Params().BitSize

	case "PRIVATE KEY":
		// PKCS#8 (un-encrypted)
		parsedPrivKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return &privateKeyInfo, fmt.Errorf("failed to parse PKCS#8 private key: %v", err)
		}
		switch k := parsedPrivKey.(type) {
		case *rsa.PrivateKey:
			keyAlgorithm = "RSA"
			keySize = k.N.BitLen()
		case *ecdsa.PrivateKey:
			keyAlgorithm = "ECDSA"
			keySize = k.Params().BitSize
		case ed25519.PrivateKey:
			keyAlgorithm = "Ed25519"
			keySize = 256
		default:
			return &privateKeyInfo, fmt.Errorf("unsupported private key type %T", k)
		}

	case "ENCRYPTED PRIVATE KEY":
		// PKCS#8 (encrypted)
		// 需要密码才能解密，UI 层应该提示用户输入
		return &privateKeyInfo, fmt.Errorf("encrypted private key is not supported yet, please provide password")

	default:
		return &privateKeyInfo, fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	// 填充返回结构体
	privateKeyInfo = model.PrivateKeyInfo{
		Alias:        req.Alias,
		KeyAlgorithm: keyAlgorithm,
		KeySize:      keySize,
		Pem:          string(privKeyBytes), // 保留原始 PEM
		Source:       keyPath,
	}

	return &privateKeyInfo, nil
}

func (s *CaService) GetCSRInfo(req model.GroupPkRequest) (*model.CSRInfo, error) {
	var csrInfo model.CSRInfo

	// 根据 alias 构建文件路径
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)
	csrPath := fmt.Sprintf("%s/%s.csr", certificateDir, req.Alias)

	// 检查 .csr 文件是否存在
	if _, err := os.Stat(csrPath); os.IsNotExist(err) {
		return &csrInfo, fmt.Errorf("CSR file not found: %s", csrPath)
	}

	// 读取 CSR 文件
	csrBytes, err := os.ReadFile(csrPath)
	if err != nil {
		return &csrInfo, fmt.Errorf("failed to read CSR file: %v", err)
	}

	// 解码 PEM 格式
	csrBlock, _ := pem.Decode(csrBytes)
	if csrBlock == nil || csrBlock.Type != "CERTIFICATE REQUEST" {
		return &csrInfo, fmt.Errorf("failed to decode CSR PEM block")
	}

	// 解析 CSR
	csr, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		return &csrInfo, fmt.Errorf("failed to parse CSR: %v", err)
	}

	// 填充返回结构体
	csrInfo = model.CSRInfo{
		CommonName:     csr.Subject.CommonName,
		Country:        strings.Join(csr.Subject.Country, ", "),
		Organization:   strings.Join(csr.Subject.Organization, ", "),
		EmailAddresses: csr.EmailAddresses,
		Pem:            string(csrBytes), // 直接使用原始 CSR 字节
	}

	return &csrInfo, nil
}

func (s *CaService) GetCertificateInfo(req model.CertificateInfoRequest) (*model.CertificateInfo, error) {
	var certInfo model.CertificateInfo

	// 检查 .crt 文件是否存在
	if _, err := os.Stat(req.Source); os.IsNotExist(err) {
		return &certInfo, fmt.Errorf("Certificate file not found: %s", req.Source)
	}

	// 读取证书文件
	certBytes, err := os.ReadFile(req.Source)
	if err != nil {
		return &certInfo, fmt.Errorf("failed to read certificate file: %v", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return &certInfo, fmt.Errorf("failed to decode certificate file: %s", req.Source)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return &certInfo, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// 获取证书信息
	certInfo.Domain = cert.Subject.CommonName
	certInfo.AltDomains = append(cert.DNSNames, convertIPsToStrings(cert.IPAddresses)...)
	certInfo.NotBefore = cert.NotBefore
	certInfo.NotAfter = cert.NotAfter
	certInfo.Country = strings.Join(cert.Subject.Country, ", ")
	certInfo.Organization = strings.Join(cert.Subject.Organization, ", ")
	certInfo.IsCA = cert.IsCA

	// 获取密钥算法和位数
	if cert.PublicKeyAlgorithm == x509.RSA {
		certInfo.KeyAlgorithm = "RSA"
		certInfo.KeySize = cert.PublicKey.(*rsa.PublicKey).N.BitLen()
	} else if cert.PublicKeyAlgorithm == x509.ECDSA {
		certInfo.KeyAlgorithm = "ECDSA"
		certInfo.KeySize = cert.PublicKey.(*ecdsa.PublicKey).Curve.Params().BitSize
	} else {
		certInfo.KeyAlgorithm = "Unknown"
	}

	// 获取签发机构信息
	certInfo.IssuerCN = cert.Issuer.CommonName
	certInfo.IssuerCountry = strings.Join(cert.Issuer.Country, ", ")
	certInfo.IssuerOrganization = strings.Join(cert.Issuer.Organization, ", ")

	// 获取证书的 PEM 格式
	certInfo.Pem = string(certBytes)

	return &certInfo, nil
}

func (s *CaService) CompleteCertificateChain(req model.CertificateInfoRequest) error {

	// 检查 Mozilla CA 是否已准备好
	ok, err := s.loadMozillaCAStore()
	if err != nil || !ok {
		return fmt.Errorf("failed to load Mozilla CA store: %v", err)
	}

	// 补齐链
	fullChain, err := s.completeCertificateChain(req.Source)
	if err != nil {
		return fmt.Errorf("error to complete chain: %v", err)
	}
	if fullChain == "" {
		return fmt.Errorf("failed to complete chain: %v", err)
	}

	// 将补齐的证书链写回覆盖原文件
	if err := os.WriteFile(req.Source, []byte(fullChain), 0600); err != nil {
		return fmt.Errorf("failed to write full chain to file: %v", err)
	}

	return nil
}

func (s *CaService) GetCertificateGroups() (*model.PageResult, error) {
	var result model.PageResult

	// 扫描根目录下所有子目录
	baseDir := filepath.Join(constant.CenterDataDir, "certificates")
	if err := utils.EnsurePaths([]string{baseDir}); err != nil {
		global.LOG.Error("Failed to create dir %s, %v", baseDir, err)
		return nil, err
	}

	dirs, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %v", err)
	}

	var groups []model.CertificateGroup
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		var certGroup model.CertificateGroup

		// 获取项目节点名
		alias := dir.Name()
		certGroup.Alias = alias

		// 检查 alias.csr 并提取 CommonName 作为申请者
		csrName := alias + ".csr"
		csrPath := filepath.Join(baseDir, alias, csrName)
		if _, err := os.Stat(csrPath); err == nil {
			csrBytes, err := os.ReadFile(csrPath)
			if err == nil {
				certGroup.Requester = extractCommonNameFromCSR(csrBytes)
			}
		}

		// 扫描所有 *.crt 文件
		crtFiles, err := filepath.Glob(filepath.Join(baseDir, alias, "*.crt"))
		if err != nil {
			return nil, fmt.Errorf("failed to scan crt files in %s: %v", filepath.Join(baseDir, alias), err)
		}

		for _, crtFile := range crtFiles {
			certInfo, err := extractCertificateInfo(crtFile)
			if err != nil {
				return nil, fmt.Errorf("failed to extract certificate info from %s: %v", crtFile, err)
			}
			certGroup.Certificates = append(certGroup.Certificates, *certInfo)
		}

		groups = append(groups, certGroup)
	}

	result.Total = int64(len(groups))
	result.Items = groups

	return &result, nil
}

func (s *CaService) RemoveCertificateGroup(req model.DeleteGroupRequest) error {
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)

	// 检查目录是否存在
	if _, err := os.Stat(certificateDir); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", certificateDir)
	}

	// 删除该目录及其下所有文件
	err := os.RemoveAll(certificateDir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %v", certificateDir, err)
	}

	return nil
}

func (s *CaService) RemoveCertificate(req model.DeleteCertificateRequest) error {
	// 检查 .crt 文件是否存在
	if _, err := os.Stat(req.Source); os.IsNotExist(err) {
		return fmt.Errorf("certificate file not found: %s", req.Source)
	}

	// 删除该文件
	err := os.Remove(req.Source)
	if err != nil {
		return fmt.Errorf("failed to remove certificate file %s: %v", req.Source, err)
	}

	return nil
}

func (s *CaService) ImportCertificate(req model.ImportCertificateRequest) error {
	// 检查目录是否已存在
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)
	if _, err := os.Stat(certificateDir); os.IsExist(err) {
		return fmt.Errorf("Alias already exist: %v", err)
	}

	// 生成存储目录路径
	if err := utils.EnsurePaths([]string{certificateDir}); err != nil {
		global.LOG.Error("Failed to create dir %s, %v", certificateDir, err)
		return err
	}

	// 根据keyType，将内容保存至alias.key
	privateKeyPath := certificateDir + "/" + req.Alias + ".key"
	switch req.KeyType {
	case 0:
		// keyContent 写入文件 privateKeyPath
		if err := os.WriteFile(privateKeyPath, []byte(req.KeyContent), 0600); err != nil {
			global.LOG.Error("Failed to write key content to file %s, %v", privateKeyPath, err)
			return err
		}

	case 1:
		// keyContent 写入文件 privateKeyPath
		if err := os.WriteFile(privateKeyPath, []byte(req.KeyContent), 0600); err != nil {
			global.LOG.Error("Failed to write key content to file %s, %v", privateKeyPath, err)
			return err
		}

	case 2:
		// 从 keyPath 文件读取内容后，写入文件 privateKeyPath
		keyData, err := os.ReadFile(req.KeyPath)
		if err != nil {
			global.LOG.Error("Failed to read key file from path %s, %v", req.KeyPath, err)
			return err
		}
		if err := os.WriteFile(privateKeyPath, keyData, 0600); err != nil {
			global.LOG.Error("Failed to write key data to file %s, %v", privateKeyPath, err)
			return err
		}

	default:
		global.LOG.Error("Invalid key type")
		return fmt.Errorf("Invalid KeyType: %d", req.KeyType)
	}

	// 根据caType，保存证书内容
	timestamp := time.Now().Unix()
	caPath := fmt.Sprintf("%s/%d.crt", certificateDir, timestamp)
	switch req.CaType {
	case 0:
		// caContent 写入文件 caPath
		if err := os.WriteFile(caPath, []byte(req.CaContent), 0600); err != nil {
			global.LOG.Error("Failed to write ca content to file %s, %v", caPath, err)
			return err
		}

	case 1:
		// caContent 写入文件 caPath
		if err := os.WriteFile(caPath, []byte(req.CaContent), 0600); err != nil {
			global.LOG.Error("Failed to write ca content to file %s, %v", caPath, err)
			return err
		}

	case 2:
		// 从 CaPath 文件读取内容后，写入文件 caPath
		caData, err := os.ReadFile(req.CaPath)
		if err != nil {
			global.LOG.Error("Failed to read ca file from path %s, %v", req.CaPath, err)
			return err
		}
		if err := os.WriteFile(caPath, caData, 0600); err != nil {
			global.LOG.Error("Failed to write ca data to file %s, %v", caPath, err)
			return err
		}

	default:
		global.LOG.Error("Invalid ca type")
		return fmt.Errorf("Invalid CaType: %d", req.CaType)
	}

	// 补齐证书链
	if req.CompleteChain {
		fullChain, err := s.completeCertificateChain(caPath)
		if err != nil {
			return fmt.Errorf("error to complete chain: %v", err)
		}
		if fullChain == "" {
			return fmt.Errorf("failed to complete chain: %v", err)
		}

		// 将补齐的证书链写回覆盖原文件
		if err := os.WriteFile(caPath, []byte(fullChain), 0600); err != nil {
			return fmt.Errorf("failed to write full chain to file: %v", err)
		}
	}

	// 根据csrType，将内容保存至alias.csr
	csrPath := certificateDir + "/" + req.Alias + ".csr"
	switch req.CsrType {
	case 0:
		// csrContent 写入文件 csrPath
		if req.CsrContent != "" {
			if err := os.WriteFile(csrPath, []byte(req.CsrContent), 0600); err != nil {
				global.LOG.Error("Failed to write csr content to file %s, %v", csrPath, err)
				return err
			}
		}

	case 1:
		// csrContent 写入文件 csrPath
		if req.CsrContent != "" {
			if err := os.WriteFile(csrPath, []byte(req.CsrContent), 0600); err != nil {
				global.LOG.Error("Failed to write csr content to file %s, %v", csrPath, err)
				return err
			}
		}

	case 2:
		// 从 CsrPath 文件读取内容后，写入文件 csrPath
		if req.CsrPath != "" {
			csrData, err := os.ReadFile(req.CsrPath)
			if err != nil {
				global.LOG.Error("Failed to read csr file from path %s, %v", req.CsrPath, err)
				return err
			}
			if err := os.WriteFile(csrPath, csrData, 0600); err != nil {
				global.LOG.Error("Failed to write csr data to file %s, %v", privateKeyPath, err)
				return err
			}
		}

	default:
		global.LOG.Error("Invalid csr type")
	}

	return nil
}

func (s *CaService) UpdateCertificate(req model.UpdateCertificateRequest) error {
	// 目录必须已经存在（与 ImportCertificate 相反）
	certificateDir := filepath.Join(constant.CenterDataDir, "certificates", req.Alias)
	if _, err := os.Stat(certificateDir); os.IsNotExist(err) {
		return fmt.Errorf("Alias not found: %v", err)
	}

	// 保存新的证书（以当前时间戳命名）
	timestamp := time.Now().Unix()
	newCertPath := filepath.Join(certificateDir, fmt.Sprintf("%d.crt", timestamp))
	switch req.CaType {
	case 0, 1:
		if err := os.WriteFile(newCertPath, []byte(req.CaContent), 0600); err != nil {
			return fmt.Errorf("Failed to write new cert to %s: %w", newCertPath, err)
		}
	case 2:
		caData, err := os.ReadFile(req.CaPath)
		if err != nil {
			return fmt.Errorf("Failed to read ca file: %w", err)
		}
		if err := os.WriteFile(newCertPath, caData, 0600); err != nil {
			return fmt.Errorf("Failed to write ca file: %w", err)
		}
	default:
		return fmt.Errorf("Invalid CaType: %d", req.CaType)
	}

	// 补齐证书链
	if req.CompleteChain {
		fullChain, err := s.completeCertificateChain(newCertPath)
		if err != nil {
			return fmt.Errorf("error to complete chain: %v", err)
		}
		if fullChain == "" {
			return fmt.Errorf("failed to complete chain: %v", err)
		}

		// 将补齐的证书链写回覆盖原文件
		if err := os.WriteFile(newCertPath, []byte(fullChain), 0600); err != nil {
			return fmt.Errorf("failed to write full chain to file: %v", err)
		}
	}

	return nil
}

func extractCommonNameFromCSR(csrBytes []byte) string {
	block, _ := pem.Decode(csrBytes)
	if block == nil {
		return ""
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return ""
	}
	return csr.Subject.CommonName
}

func extractCertificateInfo(crtFile string) (*model.CertificateSimpleInfo, error) {
	var certInfo model.CertificateSimpleInfo

	certBytes, err := os.ReadFile(crtFile)
	if err != nil {
		return &certInfo, err
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return &certInfo, fmt.Errorf("failed to decode certificate file: %s", crtFile)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return &certInfo, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// 判断状态
	now := time.Now()
	status := "valid"
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		status = "expired"
	}

	// 获取证书信息
	certInfo.Domain = cert.Subject.CommonName
	certInfo.AltDomains = append(cert.DNSNames, convertIPsToStrings(cert.IPAddresses)...)
	certInfo.NotBefore = cert.NotBefore
	certInfo.NotAfter = cert.NotAfter
	certInfo.IssuerOrganization = strings.Join(cert.Issuer.Organization, ", ")
	certInfo.Status = status
	certInfo.Source = crtFile

	// 提取证书信息
	return &certInfo, nil
}

// 将 IP 地址切换为字符串格式
func convertIPsToStrings(ips []net.IP) []string {
	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	return ipStrings
}

// 判断私钥是否是 RSA 格式
// func isRSAKey(privKey []byte) bool {
// 	_, err := x509.ParsePKCS1PrivateKey(privKey)
// 	return err == nil
// }

// 判断私钥是否是 ECDSA 格式
// func isECDSAKey(privKey []byte) bool {
// 	_, err := x509.ParseECPrivateKey(privKey)
// 	return err == nil
// }

// 生成RSA私钥
func generateRSAKey(bits int) ([]byte, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate RSA key with %d bits", bits)
	}

	// Convert private key to PEM format
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	return privKeyBytes, nil
}

// 生成ECDSA私钥
func generateECDSAKey(curve string) ([]byte, error) {
	var ellipticCurve elliptic.Curve
	switch curve {
	case "P-256":
		ellipticCurve = elliptic.P256()
	case "P-384":
		ellipticCurve = elliptic.P384()
	default:
		return nil, fmt.Errorf("unsupported ECDSA curve: %s", curve)
	}

	privateKey, err := ecdsa.GenerateKey(ellipticCurve, rand.Reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate ECDSA key with curve %s", curve)
	}

	privKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ECDSA private key")
	}

	return privKeyBytes, nil
}

// 保存私钥到文件
func savePrivateKey(path string, key []byte) error {
	// Create or open the file
	keyFile, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create key file %s", path)
	}
	defer keyFile.Close()

	// Write the private key in PEM format
	err = pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: key})
	if err != nil {
		return errors.Wrapf(err, "failed to encode private key to %s", path)
	}

	return nil
}

// 生成CSR
func generateCSR(req model.CreateGroupRequest, privateKey []byte) ([]byte, error) {
	// Generate a template for the certificate request
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         req.DomainName,
			Organization:       []string{req.Organization},
			OrganizationalUnit: []string{req.OrganizationalUnit},
			Locality:           []string{req.City},
			Province:           []string{req.Province},
			Country:            []string{req.Country},
		},
		EmailAddresses: []string{req.Email},
	}

	// Parse the private key (RSA or ECDSA)
	privKey, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private key")
	}

	// Generate the CSR using the template and private key
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create CSR")
	}

	return csrBytes, nil
}

// 保存CSR到文件
func saveCSR(path string, csr []byte) error {
	// Create or open the file
	csrFile, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create CSR file %s", path)
	}
	defer csrFile.Close()

	// Write the CSR in PEM format
	err = pem.Encode(csrFile, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})
	if err != nil {
		return errors.Wrapf(err, "failed to encode CSR to %s", path)
	}

	return nil
}

// 自动补齐证书链
func (s *CaService) completeCertificateChain(source string) (string, error) {
	// 检查 .crt 文件是否存在
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return "", fmt.Errorf("Certificate file not found: %s", source)
	}

	// 读取证书文件
	certBytes, err := os.ReadFile(source)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate file: %v", err)
	}

	// 解析证书链
	var certPEMs []string
	for {
		// 解码 PEM 格式证书
		block, rest := pem.Decode(certBytes)
		if block == nil {
			break // 没有更多证书
		}

		// 将证书添加到证书链中
		certPEMs = append(certPEMs, string(pem.EncodeToMemory(block)))
		certBytes = rest
	}
	if len(certPEMs) == 0 {
		return "", fmt.Errorf("failed to parse pems: %v", err)
	}

	// 解析现有的证书链
	var certs []*x509.Certificate
	for _, certPEM := range certPEMs {
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return "", fmt.Errorf("failed to decode certificate: %s", certPEM)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse certificate: %v", err)
		}
		certs = append(certs, cert)
	}

	// 验证当前链是否完整，如果不完整，补齐缺失部分
	fullCerts, err := s.fillCertificateChain(certs, s.rootCertMap)
	if err != nil {
		return "", fmt.Errorf("failed to complete certificate chain: %v", err)
	}

	// 构建完整的 PEM 格式证书链
	var fullCertPEMs []string
	for _, cert := range fullCerts {
		fullCertPEMs = append(fullCertPEMs, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})))
	}

	if len(fullCertPEMs) == 0 {
		return "", fmt.Errorf("failed to complete certificate chain pems: %v", err)
	}

	return strings.Join(fullCertPEMs, "\n"), nil
}

// 补齐证书链
func (s *CaService) fillCertificateChain(
	certs []*x509.Certificate,
	rootCertMap map[string]*x509.Certificate,
) ([]*x509.Certificate, error) {
	chain := append([]*x509.Certificate{}, certs...)
	seen := map[string]struct{}{}

	for _, c := range chain {
		seen[string(c.Raw)] = struct{}{}
	}

	// 从最后一个证书开始补齐
	current := certs[len(certs)-1]
	for {
		// 是根证书就停止
		if isRootCertificate(current, rootCertMap) {
			break
		}

		// 下载并缓存中间证书
		s.ensureIntermediate(current)

		// 查找颁发者证书
		s.mu.Lock()
		issuer := findIssuer(current, s.intermediateCertMap, rootCertMap)
		s.mu.Unlock()
		if issuer == nil {
			return nil, fmt.Errorf("issuer not found for %s", current.Subject)
		}

		if _, ok := seen[string(issuer.Raw)]; ok {
			break
		}

		chain = append(chain, issuer)
		seen[string(issuer.Raw)] = struct{}{}
		current = issuer
	}

	return chain, nil
}

// 检查是否为根证书
func isRootCertificate(cert *x509.Certificate, rootCertMap map[string]*x509.Certificate) bool {
	// 检查是否自签名
	if cert.Subject.String() != cert.Issuer.String() || cert.CheckSignatureFrom(cert) != nil {
		return false
	}

	// 检查是否在 rootCAs 映射中
	_, exists := rootCertMap[string(cert.Raw)]
	return exists
}

func findIssuer(
	cert *x509.Certificate,
	intermediate map[string]*x509.Certificate,
	root map[string]*x509.Certificate,
) *x509.Certificate {

	for _, c := range intermediate {
		if cert.CheckSignatureFrom(c) == nil {
			return c
		}
	}

	for _, c := range root {
		if cert.CheckSignatureFrom(c) == nil {
			return c
		}
	}

	return nil
}

// 在 CA 映射中查找颁发者证书
// func findIssuerCertificate(cert *x509.Certificate, rootCertMap map[string]*x509.Certificate) (*x509.Certificate, error) {
// 	for _, rootCert := range rootCertMap {
// 		if cert.CheckSignatureFrom(rootCert) == nil {
// 			return rootCert, nil
// 		}
// 	}
// 	return nil, fmt.Errorf("issuer not found for %s", cert.Subject)
// }

// 检查证书是否已经存在于链中
// func containsCertificate(seen map[string]struct{}, cert *x509.Certificate) bool {
// 	_, exists := seen[string(cert.Raw)]
// 	return exists
// }

// 加载 Mozilla CA 存储
func (s *CaService) loadMozillaCAStore() (bool, error) {
	// 放在data目录下
	mozillaCaPath := filepath.Join(constant.CenterDataDir, "cacert.pem")

	// 如果文件不存在，下载并缓存
	if _, err := os.Stat(mozillaCaPath); os.IsNotExist(err) {
		if err := downloadMozillaCAStore(mozillaCaPath); err != nil {
			return false, fmt.Errorf("failed to download Mozilla CA store: %v", err)
		}
	}

	// 加载 PEM 文件
	data, err := os.ReadFile(mozillaCaPath)
	if err != nil {
		return false, fmt.Errorf("failed to read CA store file: %v", err)
	}

	rootCertMap, err := parseRootCAsFromPEM(data)
	if err != nil {
		return false, err
	}

	s.rootCertMap = rootCertMap
	return len(rootCertMap) > 0, nil
}

func parseRootCAsFromPEM(pemData []byte) (map[string]*x509.Certificate, error) {
	rootCertMap := make(map[string]*x509.Certificate)

	for {
		block, rest := pem.Decode(pemData)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			pemData = rest
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			pemData = rest
			continue
		}

		// 用 Raw 作为唯一键（你后面逻辑正是这么用的）
		rootCertMap[string(cert.Raw)] = cert
		pemData = rest
	}

	if len(rootCertMap) == 0 {
		return nil, errors.New("no root certificates parsed")
	}

	return rootCertMap, nil
}

// 下载 Mozilla CA 存储
func downloadMozillaCAStore(filePath string) error {
	url := "https://curl.se/ca/cacert.pem"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download CA store: %v", err)
	}
	defer resp.Body.Close()

	// 保存到文件
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CA store file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// 检查证书是否由 Mozilla CA 存储信任
// func isTrustedByMozilla(rootCAs *x509.CertPool, cert *x509.Certificate) bool {
// 	// 设置验证选项
// 	opts := x509.VerifyOptions{
// 		Roots: rootCAs, // 使用 Mozilla CA 的根证书
// 	}

// 	// 验证证书
// 	if _, err := cert.Verify(opts); err != nil {
// 		return false // 验证失败，非受信任证书
// 	}

// 	return true // 受信任
// }

func getAIAIssuerURLs(cert *x509.Certificate) []string {
	var urls []string
	for _, u := range cert.IssuingCertificateURL {
		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
			urls = append(urls, u)
		}
	}
	return urls
}

func downloadIntermediateCA(url string) (*x509.Certificate, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 支持 DER 或 PEM
	if block, _ := pem.Decode(data); block != nil {
		return x509.ParseCertificate(block.Bytes)
	}

	return x509.ParseCertificate(data)
}

// 获取并缓存中间证书
func (s *CaService) ensureIntermediate(cert *x509.Certificate) {
	if s.intermediateCertMap == nil {
		s.intermediateCertMap = make(map[string]*x509.Certificate)
	}

	for _, url := range getAIAIssuerURLs(cert) {
		s.mu.Lock()
		// 检查是否已经下载过
		already := false
		for _, c := range s.intermediateCertMap {
			if c.Subject.String() == cert.Subject.String() {
				already = true
				break
			}
		}
		s.mu.Unlock()
		if already {
			continue
		}

		issuer, err := downloadIntermediateCA(url)
		if err != nil {
			continue
		}

		key := string(issuer.Raw)
		s.mu.Lock()
		s.intermediateCertMap[key] = issuer
		s.mu.Unlock()
	}
}

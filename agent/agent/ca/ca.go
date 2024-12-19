package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"github.com/sensdata/idb/core/utils/common"
)

type CaService struct{}

type ICaService interface {
	GenerateCertificate(req model.GenerateCertificateRequest) error
	GenerateSelfSignedCertificate(req model.SelfSignedRequest) error
	GetPrivateKeyInfo(req model.PrivateKeyInfoRequest) (*model.PrivateKeyInfo, error)
	GetCSRInfo(req model.PrivateKeyInfoRequest) (*model.CSRInfo, error)
	GetCertificateInfo(req model.CertificateInfoRequest) (*model.CertificateInfo, error)
	CompleteCertificateChain(req model.CertificateInfoRequest) error
	GetCertificateGroups() (*model.PageResult, error)
	RemoveCertificateGroup(req model.RemoveCertificateGroupRequest) error
	RemoveCertificate(req model.RemoveCertificateRequest) error
}

func NewICaService() ICaService {
	return &CaService{}
}

func (s *CaService) GenerateCertificate(req model.GenerateCertificateRequest) error {

	// 1. 生成存储目录路径
	certificateDir := "/var/lib/idb/data/certificates/" + req.Alias
	if err := utils.EnsurePaths([]string{certificateDir}); err != nil {
		global.LOG.Error("Failed to create dir %s, %v", certificateDir, err)
		return err
	}

	// 2. 根据密钥算法生成私钥并保存
	privateKeyPath := certificateDir + "/" + req.Alias + ".key"
	var privateKey []byte
	var err error
	switch req.KeyAlgorithm {
	case "RSA 2048":
		privateKey, err = generateRSAKey(2048)
		if err != nil {
			return err
		}
	case "RSA 3072":
		privateKey, err = generateRSAKey(3072)
		if err != nil {
			return err
		}
	case "RSA 4096":
		privateKey, err = generateRSAKey(4096)
		if err != nil {
			return err
		}
	case "EC 256":
		privateKey, err = generateECDSAKey("P-256")
		if err != nil {
			return err
		}
	case "EC 384":
		privateKey, err = generateECDSAKey("P-384")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported key algorithm: %s", req.KeyAlgorithm)
	}

	// 保存私钥到文件
	err = savePrivateKey(privateKeyPath, privateKey)
	if err != nil {
		return err
	}

	// 3. 生成CSR文件
	csrPath := certificateDir + "/" + req.Alias + ".csr"
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
	// 校验ip
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
	certDirectory := fmt.Sprintf("/var/lib/idb/data/certificates/%s", req.Alias)

	csrPath := fmt.Sprintf("%s/%s.csr", certDirectory, req.Alias)
	keyPath := fmt.Sprintf("%s/%s.key", certDirectory, req.Alias)

	// 检查 .csr 和 .key 文件是否存在
	if _, err := os.Stat(csrPath); os.IsNotExist(err) {
		return fmt.Errorf("CSR file not found: %s", csrPath)
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return fmt.Errorf("Key file not found: %s", keyPath)
	}

	// 解析私钥和 CSR
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	csrBytes, err := os.ReadFile(csrPath)
	if err != nil {
		return fmt.Errorf("failed to read CSR file: %v", err)
	}

	// 解析私钥：根据私钥格式判断是 RSA 还是 ECDSA
	var privateKey interface{}
	if isRSAKey(privateKeyBytes) {
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBytes)
	} else if isECDSAKey(privateKeyBytes) {
		privateKey, err = x509.ParseECPrivateKey(privateKeyBytes)
	} else {
		return fmt.Errorf("unsupported key algorithm or invalid private key")
	}
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	// 确定 private key 的类型，并提取 public key
	var publicKey interface{}
	switch privKey := privateKey.(type) {
	case *rsa.PrivateKey:
		publicKey = &privKey.PublicKey
	case *ecdsa.PrivateKey:
		publicKey = &privKey.PublicKey
	default:
		return fmt.Errorf("unsupported private key type")
	}

	// 解析 CSR
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return fmt.Errorf("failed to parse CSR: %v", err)
	}

	// Step 5: 创建证书模板
	certTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               csr.Subject,
		EmailAddresses:        csr.EmailAddresses,
		Issuer:                csr.Subject, // 使用 CSR 的 Subject 作为签发者
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(expireDuration),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:                  req.IsCA, // 如果是CA证书，可以用于签发下级证书
		BasicConstraintsValid: true,
		DNSNames:              domains,
		IPAddresses:           ips,
	}

	// 生成自签名证书
	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, publicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// 将生成的证书保存到文件
	timestamp := time.Now().Unix()
	certPath := fmt.Sprintf("%s/%d.crt", certDirectory, timestamp)

	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create certificate file: %v", err)
	}
	defer certFile.Close()

	// 将证书编码为 PEM 格式并写入文件
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return fmt.Errorf("failed to encode certificate to PEM: %v", err)
	}

	return nil
}

func (s *CaService) GetPrivateKeyInfo(req model.PrivateKeyInfoRequest) (*model.PrivateKeyInfo, error) {
	var privateKeyInfo model.PrivateKeyInfo

	// 根据 alias 构建文件路径
	certDirectory := fmt.Sprintf("/var/lib/idb/data/certificates/%s", req.Alias)
	keyPath := fmt.Sprintf("%s/%s.key", certDirectory, req.Alias)

	// 检查 .key 文件是否存在
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return &privateKeyInfo, fmt.Errorf("Key file not found: %s", keyPath)
	}

	// 读取私钥文件
	privKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return &privateKeyInfo, fmt.Errorf("failed to read private key: %v", err)
	}

	// 解析私钥：根据私钥格式判断是 RSA 还是 ECDSA
	var parsedPrivKey interface{}
	var keyAlgorithm string
	var keySize int

	if isRSAKey(privKeyBytes) {
		// 解析 RSA 私钥
		parsedPrivKey, err = x509.ParsePKCS1PrivateKey(privKeyBytes)
		if err != nil {
			return &privateKeyInfo, fmt.Errorf("failed to parse RSA private key: %v", err)
		}
		keyAlgorithm = "RSA"
		// 获取 RSA 私钥的位数
		keySize = parsedPrivKey.(*rsa.PrivateKey).N.BitLen() // RSA 使用 N.BitLen() 来获取密钥长度
	} else if isECDSAKey(privKeyBytes) {
		// 解析 ECDSA 私钥
		parsedPrivKey, err = x509.ParseECPrivateKey(privKeyBytes)
		if err != nil {
			return &privateKeyInfo, fmt.Errorf("failed to parse ECDSA private key: %v", err)
		}
		keyAlgorithm = "ECDSA"
		// 获取 ECDSA 私钥的位数
		keySize = parsedPrivKey.(*ecdsa.PrivateKey).Params().BitSize
	} else {
		return &privateKeyInfo, fmt.Errorf("unsupported key algorithm or invalid private key")
	}

	// 将私钥编码为 PEM 格式
	privKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privKeyBytes,
	}

	// 生成 PEM 格式私钥
	pemKey := pem.EncodeToMemory(privKeyBlock)

	// 填充返回结构体
	privateKeyInfo = model.PrivateKeyInfo{
		Alias:        req.Alias,
		KeyAlgorithm: keyAlgorithm,
		KeySize:      keySize,
		Pem:          string(pemKey),
	}

	return &privateKeyInfo, nil
}

func (s *CaService) GetCSRInfo(req model.PrivateKeyInfoRequest) (*model.CSRInfo, error) {
	var csrInfo model.CSRInfo

	// 根据 alias 构建文件路径
	certDirectory := fmt.Sprintf("/var/lib/idb/data/certificates/%s", req.Alias)
	csrPath := fmt.Sprintf("%s/%s.csr", certDirectory, req.Alias)

	// 检查 .csr 文件是否存在
	if _, err := os.Stat(csrPath); os.IsNotExist(err) {
		return &csrInfo, fmt.Errorf("CSR file not found: %s", csrPath)
	}

	// 读取 CSR 文件
	csrBytes, err := os.ReadFile(csrPath)
	if err != nil {
		return &csrInfo, fmt.Errorf("failed to read CSR file: %v", err)
	}

	// 解析 CSR
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return &csrInfo, fmt.Errorf("failed to parse CSR: %v", err)
	}

	// 填充返回结构体
	csrInfo = model.CSRInfo{
		CommonName:     csr.Subject.CommonName,
		Country:        strings.Join(csr.Subject.Country, ", "),
		Organization:   strings.Join(csr.Subject.Organization, ", "),
		EmailAddresses: csr.EmailAddresses,
		Pem:            string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})),
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
	certInfo.Pem = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes}))

	return &certInfo, nil
}

func (s *CaService) CompleteCertificateChain(req model.CertificateInfoRequest) error {
	// 检查 .crt 文件是否存在
	if _, err := os.Stat(req.Source); os.IsNotExist(err) {
		return fmt.Errorf("Certificate file not found: %s", req.Source)
	}

	// 读取证书文件
	certBytes, err := os.ReadFile(req.Source)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %v", err)
	}

	// TODO: MozillaCA方案，可能需要增加内置和存储官方数据，并定时更新，在此处再直接使用
	fullChain, err := completeCertificateChain(string(certBytes), "")
	if err != nil {
		return fmt.Errorf("failed to complete chain: %v", err)
	}

	// 将补齐的证书链写回覆盖原文件
	if err := os.WriteFile(req.Source, []byte(fullChain), 0644); err != nil {
		return fmt.Errorf("failed to write full chain to file: %v", err)
	}

	return nil
}

func (s *CaService) GetCertificateGroups() (*model.PageResult, error) {
	var result model.PageResult

	// 扫描根目录下所有子目录
	baseDir := "/var/lib/idb/data/certificates"
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

func (s *CaService) RemoveCertificateGroup(req model.RemoveCertificateGroupRequest) error {
	certDirectory := fmt.Sprintf("/var/lib/idb/data/certificates/%s", req.Alias)

	// 检查目录是否存在
	if _, err := os.Stat(certDirectory); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", certDirectory)
	}

	// 删除该目录及其下所有文件
	err := os.RemoveAll(certDirectory)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %v", certDirectory, err)
	}

	return nil
}

func (s *CaService) RemoveCertificate(req model.RemoveCertificateRequest) error {
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
func isRSAKey(privKey []byte) bool {
	_, err := x509.ParsePKCS1PrivateKey(privKey)
	return err == nil
}

// 判断私钥是否是 ECDSA 格式
func isECDSAKey(privKey []byte) bool {
	_, err := x509.ParseECPrivateKey(privKey)
	return err == nil
}

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
func generateCSR(req model.GenerateCertificateRequest, privateKey []byte) ([]byte, error) {
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
func completeCertificateChain(serverCertPEM string, caRepositoryURL string) (string, error) {
	// 解析终端证书
	block, _ := pem.Decode([]byte(serverCertPEM))
	if block == nil {
		return "", errors.New("failed to decode server certificate")
	}
	serverCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse server certificate: %v", err)
	}

	// 初始化证书链
	certChain := []string{serverCertPEM}
	certCache := make(map[string]string) // 缓存已获取的证书

	// 循环补齐中间 CA
	currentIssuer := serverCert.Issuer
	for {
		if _, ok := certCache[currentIssuer.String()]; !ok {
			intermediateCertPEM, err := fetchCACertificate(currentIssuer, caRepositoryURL)
			if err != nil {
				fmt.Printf("warning: failed to fetch intermediate CA for issuer %s: %v\n", currentIssuer.String(), err)
				break
			}
			certCache[currentIssuer.String()] = intermediateCertPEM
		}

		intermediateCertPEM := certCache[currentIssuer.String()]
		block, _ := pem.Decode([]byte(intermediateCertPEM))
		if block == nil {
			return "", errors.New("failed to decode intermediate CA certificate")
		}
		intermediateCert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse intermediate CA: %v", err)
		}

		// 添加到证书链
		certChain = append(certChain, intermediateCertPEM)

		// 检查是否已经到根 CA
		if strings.EqualFold(intermediateCert.Subject.String(), intermediateCert.Issuer.String()) &&
			(intermediateCert.IsCA && intermediateCert.KeyUsage&x509.KeyUsageCertSign != 0) {
			break
		}

		// 更新 Issuer 为下一个
		currentIssuer = intermediateCert.Issuer
	}

	// 构建完整的 PEM 格式证书链
	fullCertChain := strings.Join(certChain, "\n\n")
	return fullCertChain, nil
}

// 从可信源获取 CA 证书
func fetchCACertificate(issuer pkix.Name, caRepositoryURL string) (string, error) {
	return "", nil
	// // 根据 Issuer 的信息构造查询 URL
	// query := fmt.Sprintf("%s?issuer=%s", caRepositoryURL, issuer.String())

	// // HTTP 请求获取证书
	// resp, err := http.Get(query)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to fetch CA certificate: %v", err)
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	// }

	// // 返回 PEM 格式证书
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to read response body: %v", err)
	// }
	// return string(body), nil
}

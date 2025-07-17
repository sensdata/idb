package model

import "time"

type PrivateKeyInfo struct {
	Alias        string `json:"alias"`
	KeyAlgorithm string `json:"key_algorithm"`
	KeySize      int    `json:"key_size"`
	Pem          string `json:"pem"`
}

type CSRInfo struct {
	CommonName     string   `json:"common_name"`
	Country        string   `json:"country"`
	Organization   string   `json:"organization"`
	EmailAddresses []string `json:"email_addresses"`
	Pem            string   `json:"pem"`
}

type CertificateGroup struct {
	Alias        string                  `json:"alias"`
	Requester    string                  `json:"requester"`
	Certificates []CertificateSimpleInfo `json:"certificates"`
}

type CertificateSimpleInfo struct {
	Domain             string    `json:"domain"`
	AltDomains         []string  `json:"alt_names"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after"`
	IssuerOrganization string    `json:"issuer_organization"`
	Status             string    `json:"status"`
	Source             string    `json:"source"`
}

type CertificateInfo struct {
	// 证书信息
	Domain       string    `json:"domain"`
	AltDomains   []string  `json:"alt_names"`
	NotBefore    time.Time `json:"not_before"`
	NotAfter     time.Time `json:"not_after"`
	Country      string    `json:"country"`
	Organization string    `json:"organization"`
	KeyAlgorithm string    `json:"key_algorithm"`
	KeySize      int       `json:"key_size"`
	IsCA         bool      `json:"is_ca"`

	// 签发机构信息
	IssuerCN           string `json:"issuer_cn"`
	IssuerCountry      string `json:"issuer_country"`
	IssuerOrganization string `json:"issuer_organization"`

	// 证书代码（PEM格式）
	Pem string `json:"pem"`

	Source string `json:"source"`
}

type CreateGroupRequest struct {
	Alias              string `json:"alias"`             // 主体别名
	DomainName         string `json:"domain_name"`       // 域名/名称
	Email              string `json:"email"`             // 邮箱
	Organization       string `json:"organization"`      // 公司/组织
	OrganizationalUnit string `json:"organization_unit"` // 部门
	Country            string `json:"country"`           // 国家
	Province           string `json:"province"`          // 省份
	City               string `json:"city"`              // 城市
	KeyAlgorithm       string `json:"key_algorithm"`     // 秘钥算法
}

type SelfSignedRequest struct {
	Alias       string `json:"alias"`        // 用于查找目录和证书项目的主体别名
	ExpireUnit  string `json:"expire_unit"`  // 有效期单位：day 或 year
	ExpireValue int    `json:"expire_value"` // 有效期值
	AltDomains  string `json:"alt_domains"`  // 备用域名，以换行符分隔
	AltIPs      string `json:"alt_ips"`      // 备用IP地址，以换行符分隔
	IsCA        bool   `json:"is_ca"`        // 是否允许使用该证书签发下级证书
}

type GroupPkRequest struct {
	Alias string `json:"alias"`
}

type CertificateInfoRequest struct {
	Source string `json:"source"`
}

type DeleteGroupRequest struct {
	Alias string `json:"alias"`
}

type DeleteCertificateRequest struct {
	Source string `json:"source"`
}

type ImportCertificateRequest struct {
	Alias         string `json:"alias"`
	KeyType       int    `json:"key_type"`
	KeyContent    string `json:"key_content"`
	KeyPath       string `json:"key_path"`
	CaType        int    `json:"ca_type"`
	CaContent     string `json:"ca_content"`
	CaPath        string `json:"ca_path"`
	CsrType       int    `json:"csr_type"`
	CsrContent    string `json:"csr_content"`
	CsrPath       string `json:"csr_path"`
	CompleteChain bool   `json:"complete_chain"`
}

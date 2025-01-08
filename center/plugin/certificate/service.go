package certificate

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/api/service"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/plugin"
	"gopkg.in/yaml.v2"
)

type CertificateMan struct {
	plugin      plugin.Plugin
	pluginConf  plugin.PluginConf
	restyClient *resty.Client
}

var LOG *log.Log

//go:embed plug.yaml
var plugYAML []byte

//go:embed conf.yaml
var confYAML []byte

func (s *CertificateMan) Initialize() {
	global.LOG.Info("certificateman init begin \n")

	if err := yaml.Unmarshal(plugYAML, &s.plugin); err != nil {
		global.LOG.Error("Failed to load info: %v", err)
		return
	}

	confPath := filepath.Join(constant.CenterConfDir, "certificate", "conf.yaml")
	// 检查配置文件的目录是否存在
	if err := os.MkdirAll(filepath.Dir(confPath), os.ModePerm); err != nil {
		global.LOG.Error("Failed to create conf directory: %v \n", err)
		return
	}
	// 检查配置文件是否存在
	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		// 创建配置文件并写入默认内容
		if err := os.WriteFile(confPath, confYAML, 0644); err != nil {
			global.LOG.Error("Failed to create conf: %v \n", err)
			return
		}
	}
	// 读取文件内容
	data, err := os.ReadFile(confPath)
	if err != nil {
		global.LOG.Error("Failed to read conf: %v \n", err)
		return
	}
	// 解析 YAML 内容
	if err := yaml.Unmarshal(data, &s.pluginConf); err != nil {
		global.LOG.Error("Failed to load conf: %v", err)
		return
	}

	//初始化日志模块
	if LOG == nil {
		logger, err := log.InitLogger(s.pluginConf.Items.LogDir, "certificate.log")
		if err != nil {
			global.LOG.Error("Failed to initialize logger: %v \n", err)
			return
		}
		LOG = logger
	}

	settingService := service.NewISettingsService()
	settingInfo, _ := settingService.Settings()
	scheme := "http"
	if settingInfo.Https == "yes" {
		scheme = "https"
	}
	host := global.Host
	if settingInfo.BindDomain != "" && settingInfo.BindDomain != host {
		host = settingInfo.BindDomain
	}
	baseUrl := fmt.Sprintf("%s://%s:%d/api/v1", scheme, host, settingInfo.BindPort)

	s.restyClient = resty.New().
		SetBaseURL(baseUrl).
		SetHeader("Content-Type", "application/json")

	api.API.SetUpPluginRouters(
		"certificates",
		[]plugin.PluginRoute{
			{Method: "GET", Path: "/info", Handler: s.GetPluginInfo},
			{Method: "GET", Path: "/menu", Handler: s.GetMenu},
			{Method: "GET", Path: "/:host/group", Handler: s.Groups},
			{Method: "POST", Path: "/:host/group", Handler: s.CreateGroup},
			{Method: "DELETE", Path: "/:host/group", Handler: s.DeleteGroup},
			{Method: "GET", Path: "/:host/group/key", Handler: s.GroupPk},
			{Method: "GET", Path: "/:host/group/csr", Handler: s.GroupCsr},

			{Method: "GET", Path: "/:host", Handler: s.GetCertificate},
			{Method: "DELETE", Path: "/:host", Handler: s.DeleteCertificate},
			{Method: "POST", Path: "/:host/sign/self", Handler: s.SelfSignCertificate},
			{Method: "POST", Path: "/:host/complete", Handler: s.CompleteCertificate},

			{Method: "POST", Path: "/:host/import", Handler: s.Import},
		},
	)

	global.LOG.Info("certificateman init end")
}

func (s *CertificateMan) Release() {

}

// @Tags Certificates
// @Summary Get plugin info
// @Description Get plugin information
// @Accept json
// @Produce json
// @Success 200 {object} plugin.PluginInfo
// @Router /certificates/info [get]
func (s *CertificateMan) GetPluginInfo(c *gin.Context) {
	pluginInfo, err := s.getPluginInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "info", "payload": pluginInfo})
}

// @Tags Certificates
// @Summary Get plugin menu
// @Description Get plugin menu items
// @Accept json
// @Produce json
// @Success 200 {array} plugin.MenuItem
// @Router /certificates/menu [get]
func (s *CertificateMan) GetMenu(c *gin.Context) {
	menuItems, err := s.getMenus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"type": "menu", "payload": menuItems})
}

func (s *CertificateMan) getPluginInfo() (plugin.PluginInfo, error) {
	return s.plugin.Info, nil
}

func (s *CertificateMan) getMenus() ([]plugin.MenuItem, error) {
	return s.plugin.Menu, nil
}

// @Tags Certificates
// @Summary Get certificate groups
// @Description Get list of certificate groups
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Success 200 {object} model.PageResult
// @Router /certificates/{host}/group [get]
func (s *CertificateMan) Groups(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	groups, err := s.groups(hostID)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, groups)
}

// @Tags Certificates
// @Summary Create certificate group
// @Description Create certificate group
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CreateGroupRequest true "Certificate group creation details"
// @Success 200
// @Router /certificates/{host}/group [post]
func (s *CertificateMan) CreateGroup(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CreateGroupRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.createGroup(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Certificates
// @Summary Delete certificate group
// @Description Delete certificate group
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param alias query string true "Group Alias"
// @Success 200
// @Router /certificates/{host}/group [delete]
func (s *CertificateMan) DeleteGroup(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	alias := c.Query("alias")
	if alias == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid alias", err)
		return
	}

	req := model.DeleteGroupRequest{Alias: alias}
	err = s.deleteGroup(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Certificates
// @Summary Get certificate group private key
// @Description Get certificate group private key
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param alias query string true "Group Alias"
// @Success 200 {object} model.PrivateKeyInfo
// @Router /certificates/{host}/group/key [get]
func (s *CertificateMan) GroupPk(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	alias := c.Query("alias")
	if alias == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid alias", err)
		return
	}

	req := model.GroupPkRequest{Alias: alias}
	key, err := s.groupPk(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, key)
}

// @Tags Certificates
// @Summary Get certificate group csr
// @Description Get certificate group csr
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param alias query string true "Group Alias"
// @Success 200 {object} model.CSRInfo
// @Router /certificates/{host}/group/csr [get]
func (s *CertificateMan) GroupCsr(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	alias := c.Query("alias")
	if alias == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid alias", err)
		return
	}

	req := model.GroupPkRequest{Alias: alias}
	csr, err := s.groupCsr(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, csr)
}

// @Tags Certificates
// @Summary Get certificate detail
// @Description Get certificate detail
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param source query string true "source"
// @Success 200 {object} model.CertificateInfo
// @Router /certificates/{host} [get]
func (s *CertificateMan) GetCertificate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	source := c.Query("source")
	if source == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid source", err)
		return
	}

	req := model.CertificateInfoRequest{Source: source}
	detail, err := s.getCertificate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, detail)
}

// @Tags Certificates
// @Summary Delete certificate
// @Description Delete certificate
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param source query string true "source"
// @Success 200
// @Router /certificates/{host} [delete]
func (s *CertificateMan) DeleteCertificate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	source := c.Query("source")
	if source == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid source", err)
		return
	}

	req := model.DeleteCertificateRequest{Source: source}
	err = s.deleteCertificate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Certificates
// @Summary Self sign certificate
// @Description Self sign certificate
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.SelfSignedRequest true "Certificate self sign details"
// @Success 200
// @Router /certificates/{host}/sign/self [post]
func (s *CertificateMan) SelfSignCertificate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.SelfSignedRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.selfSignCertificate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Certificates
// @Summary Complete certificate
// @Description Complete certificate
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param request body model.CertificateInfoRequest true "Certificate complete details"
// @Success 200
// @Router /certificates/{host}/complete [post]
func (s *CertificateMan) CompleteCertificate(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	var req model.CertificateInfoRequest
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}

	err = s.completeCertificate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

// @Tags Certificates
// @Summary Import certificate
// @Description Import certificate
// @Accept json
// @Produce json
// @Param host path uint true "Host ID"
// @Param alias formData string true "Alias"
// @Param key_type formData int true "Key import type"
// @Param key_file formData file false "Key file to import"
// @Param key_content formData string false "Key file content to import"
// @Param key_path formData string false "Local key file path"
// @Param ca_type formData int true "Certificate import type"
// @Param ca_file formData file false "Certificate file to import"
// @Param ca_content formData string false "Certificate file content to import"
// @Param ca_path formData string false "Local ca file path"
// @Param csr_type formData int true "Csr import type"
// @Param csr_file formData file false "Csr file to import"
// @Param csr_content formData string false "Csr file content to import"
// @Param csr_path formData string false "Local csr file path"
// @Success 200
// @Router /certificates/{host}/import [post]
func (s *CertificateMan) Import(c *gin.Context) {
	// 获取路径参数中的 Host ID
	hostID, err := strconv.ParseUint(c.Param("host"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid host", err)
		return
	}

	// 获取表单字段
	alias := c.PostForm("alias") // 获取 alias 字段
	if alias == "" {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Alias is required", nil)
		return
	}

	keyType, err := strconv.ParseUint(c.PostForm("key_type"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid key_type value", err)
		return
	}
	caType, err := strconv.ParseUint(c.PostForm("ca_type"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid ca_type value", err)
		return
	}
	csrType, err := strconv.ParseUint(c.PostForm("csr_type"), 10, 32)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid csr_type value", err)
		return
	}

	// 获取表单中的文件内容
	var (
		keyContent string
		keyPath    string
		caContent  string
		caPath     string
		csrContent string
		csrPath    string
	)

	// 秘钥
	switch keyType {
	// 上传文件
	case 0:
		file, err := c.FormFile("key_file")
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Failed to read key_file", err)
			return
		}
		// 打开文件
		srcFile, err := file.Open()
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, "Failed to open key_file", err)
			return
		}
		defer srcFile.Close()
		// 读取文件内容
		buf, err := io.ReadAll(srcFile)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, "Failed to read key_file content", err)
			return
		}
		keyContent = string(buf)

	// 粘贴文件内容
	case 1:
		keyContent = c.PostForm("key_content")
		if keyContent == "" {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "key_content required", nil)
			return
		}

	// 从本地文件导入
	case 2:
		keyPath = c.PostForm("key_path")
		if keyPath == "" {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "key_path required", nil)
			return
		}

	default:
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid key_type value", nil)
		return
	}

	// 证书
	switch caType {
	// 上传文件
	case 0:
		file, err := c.FormFile("ca_file")
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Failed to read ca_file", err)
			return
		}
		// 打开文件
		srcFile, err := file.Open()
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, "Failed to open ca_file", err)
			return
		}
		defer srcFile.Close()
		// 读取文件内容
		buf, err := io.ReadAll(srcFile)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, "Failed to read ca_file content", err)
			return
		}
		caContent = string(buf)

	// 粘贴文件内容
	case 1:
		caContent = c.PostForm("ca_content")
		if caContent == "" {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "ca_content required", nil)
			return
		}

	// 从本地文件导入
	case 2:
		caPath = c.PostForm("ca_path")
		if caPath == "" {
			helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "ca_path required", nil)
			return
		}

	default:
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid ca_type value", nil)
		return
	}

	// csr，可以不传
	switch csrType {
	// 上传文件
	case 0:
		file, err := c.FormFile("csr_file")
		if err != nil {
			LOG.Error("Failed to read csr_file")
		} else {
			// 打开文件
			srcFile, err := file.Open()
			if err != nil {
				LOG.Error("Failed to open csr_file")
			} else {
				defer srcFile.Close()
				// 读取文件内容
				buf, err := io.ReadAll(srcFile)
				if err != nil {
					LOG.Error("Failed to read csr_file content")
				} else {
					csrContent = string(buf)
				}
			}
		}

	// 粘贴文件内容
	case 1:
		csrContent = c.PostForm("csr_content")

	// 从本地文件导入
	case 2:
		csrPath = c.PostForm("csr_path")

	default:
		helper.ErrorWithDetail(c, constant.CodeErrBadRequest, "Invalid csr_type value", nil)
		return
	}

	req := model.ImportCertificateRequest{
		Alias:      alias,
		KeyType:    int(keyType),
		KeyContent: keyContent,
		KeyPath:    keyPath,
		CaType:     int(caType),
		CaContent:  caContent,
		CaPath:     caPath,
		CsrType:    int(csrType),
		CsrContent: csrContent,
		CsrPath:    csrPath,
	}
	err = s.importCertificate(hostID, req)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeErrInternalServer, err.Error(), err)
		return
	}

	helper.SuccessWithData(c, nil)
}

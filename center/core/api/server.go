package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
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
	router: gin.Default(),
}

type ApiServer struct {
	router *gin.Engine
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
		Handler: s.router,
	}
	tcpItem := "tcp4"
	ln, err := net.Listen(tcpItem, server.Addr)
	if err != nil {
		global.LOG.Error("Failed to listen to %s", server.Addr)
		return err
	}

	if settings.Https == "yes" {
		var cert tls.Certificate
		var certPath string
		var keyPath string
		var err error
		if settings.HttpsCertType == "default" {
			certPath = filepath.Join(constant.CenterBinDir, "cert.pem")
			keyPath = filepath.Join(constant.CenterBinDir, "key.pem")

		} else {
			certPath = settings.HttpsCertPath
			keyPath = settings.HttpsKeyPath
		}
		certificate, err := os.ReadFile(certPath)
		if err != nil {
			global.LOG.Error("Failed to read cert file %s : %v", settings.HttpsCertPath, err)
			return err
		}
		key, err := os.ReadFile(keyPath)
		if err != nil {
			global.LOG.Error("Failed to read key file %s : %v", settings.HttpsKeyPath, err)
			return err
		}
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
	apiGroup := s.router.Group("api/v1")
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
	s.router.NoRoute(func(c *gin.Context) {
		// 这里可以使用 c.FileServer 来处理目录下的所有请求
		c.File("/var/lib/idb/home/index.html") // 默认返回 index.html
	})

	// 设置静态文件路由，确保可以访问 assets 目录
	s.router.Static("/assets", "/var/lib/idb/home/assets") // 处理 assets 目录
}

// SetUpPluginRouters sets up routers from plugins
func (s *ApiServer) SetUpPluginRouters(group string, routes []plugin.PluginRoute) {
	global.LOG.Info("register router - %s", group)
	pluginGroup := s.router.Group("api/v1/" + group)
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
}

package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

	// 设置 git 路由
	gitGroup := s.router.Group("api/v1/git/" + group)
	repoPath := fmt.Sprintf("/var/lib/idb/data/%s/global", group)
	// 处理 git clone/pull 请求
	gitGroup.GET("/*path", s.handleGitRoute(repoPath))
	gitGroup.POST("/*path", s.handleGitRoute(repoPath))
}

func (s *ApiServer) handleGitRoute(repoPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, err := git.PlainOpen(repoPath)
		if err != nil {
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

	// 发送服务能力声明
	c.Writer.Write([]byte("# service=git-upload-pack\n"))
	c.Writer.Write([]byte("0000")) // 分隔符

	// 获取并发送引用信息
	refs, err := repo.References()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get references")
		return
	}

	// 首先发送 HEAD 引用
	head, err := repo.Head()
	if err == nil {
		c.Writer.Write([]byte(fmt.Sprintf("%s HEAD\x00multi_ack thin-pack\n", head.Hash())))
	}

	// 发送其他引用
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() || ref.Name().IsTag() {
			c.Writer.Write([]byte(fmt.Sprintf("%s %s\n", ref.Hash(), ref.Name())))
		}
		return nil
	})

	c.Writer.Write([]byte("0000")) // 结束标记
}

func (s *ApiServer) handleGitUploadPack(c *gin.Context, repo *git.Repository) {
	// 读取客户端想要的对象列表
	wants, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read request")
		return
	}

	// 解析客户端请求的对象
	wantList := strings.Split(string(wants), "\n")
	if len(wantList) == 0 {
		c.String(http.StatusBadRequest, "No wants specified")
		return
	}

	c.Header("Content-Type", "application/x-git-upload-pack-result")

	// 遍历客户端请求的对象
	for _, want := range wantList {
		if want == "" {
			continue
		}

		// 解析对象哈希
		hash := plumbing.NewHash(strings.TrimSpace(want))

		// 获取对象
		obj, err := repo.Object(plumbing.AnyObject, hash)
		if err != nil {
			global.LOG.Error("Failed to get object %s: %v", hash, err)
			continue
		}

		// 如果是提交对象，获取其树
		if commit, ok := obj.(*object.Commit); ok {
			tree, err := commit.Tree()
			if err != nil {
				global.LOG.Error("Failed to get tree for commit %s: %v", hash, err)
				continue
			}

			// 遍历并发送所有文件
			tree.Files().ForEach(func(f *object.File) error {
				contents, err := f.Contents()
				if err != nil {
					return err
				}
				c.Writer.Write([]byte(contents))
				return nil
			})
		}
	}
}

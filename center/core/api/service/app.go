package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"gorm.io/gorm"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
	"gopkg.in/yaml.v2"
)

type AppService struct {
	AppDir string // App目录
}

type IAppService interface {
	SyncApp() error
	RemoveApp(req core.RemoveApp) error
	AppPage(hostID uint64, req core.QueryApp) (*core.PageResult, error)
	InstalledAppPage(hostID uint64, req core.QueryInstalledApp) (*core.PageResult, error)
	AppDetail(hostID uint64, req core.QueryAppDetail) (*core.App, error)
	AppInstall(hostID uint64, req core.InstallApp) (*core.ComposeCreateResult, error)
	AppUninstall(hostID uint64, req core.UninstallApp) (*core.ComposeCreateResult, error)
	AppUpgrade(hostID uint64, req core.UpgradeApp) (*core.ComposeCreateResult, error)
}

func NewIAppService() IAppService {
	return &AppService{AppDir: constant.AgentDockerDir}
}

func (s *AppService) SyncApp() error {
	global.LOG.Info("SyncApp begin")

	// 仓库路径
	repoURL := "https://github.com/sensdata/idb-store.git"
	// 本地路径
	repoPath := filepath.Join(constant.CenterDataDir, constant.StoreDir)

	// 检查目录是否已存在
	var repo *git.Repository
	var err error
	if _, err = os.Stat(repoPath); os.IsNotExist(err) {
		// 如果目录不存在，克隆仓库
		global.LOG.Info("Cloning repository...")
		repo, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:      repoURL,
			Progress: os.Stdout,
			// Auth: &http.BasicAuth{
			// 	Username: "idb",              // GitHub用户名
			// 	Password: "idb-access-token", // GitHub访问令牌
			// },
			ReferenceName: plumbing.ReferenceName("refs/heads/main"), // 指定分支
		})
		if err != nil {
			global.LOG.Error("Error cloning repository: %s\n", err)
			return err
		}
	} else {
		// 如果目录已存在，打开仓库
		global.LOG.Info("Opening repository...")
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			global.LOG.Error("Error opening repository: %s\n", err)
			return err
		}
	}

	// 拉取最新代码
	global.LOG.Info("Pulling latest changes...")
	worktree, err := repo.Worktree()
	if err != nil {
		global.LOG.Error("Error getting worktree: %s\n", err)
		return err
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			global.LOG.Info("Repository is already up-to-date.")
		} else {
			global.LOG.Error("Error pulling changes: %s\n", err)
		}
	} else {
		global.LOG.Info("Successfully pulled latest changes.")
	}

	// 打印当前HEAD
	ref, err := repo.Head()
	if err != nil {
		global.LOG.Error("Error getting HEAD: %s\n", err)
		return err
	}
	global.LOG.Info("Current HEAD: %s\n", ref.Hash())

	// 扫描应用目录
	appsDir := filepath.Join(repoPath, "apps")
	dirEntries, err := os.ReadDir(appsDir)
	if err != nil {
		global.LOG.Error("Error reading apps dir")
		return err
	}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			// 应用目录
			appDir := filepath.Join(appsDir, dirEntry.Name())

			// manifest.yaml
			app, err := loadManifest(appDir)
			if err != nil {
				continue
			}
			// form.yaml
			form, err := loadForm(appDir)
			if err != nil {
				continue
			}
			app.FormContent = form

			// create or update app
			var appId uint
			appRecord, err := AppRepo.Get(AppRepo.WithByName(app.Name))
			if err != nil {
				// 如果是未找到记录，创建新应用
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if err := AppRepo.Create(app); err != nil {
						global.LOG.Error("Failed to create app, %v", err)
						continue
					}
					appId = app.ID
				} else {
					// 其他数据库错误
					global.LOG.Error("Error when checking app record, %v", err)
					continue
				}
			} else {
				upMap := make(map[string]interface{})
				upMap["name"] = app.Name
				upMap["display_name"] = app.DisplayName
				upMap["category"] = app.Category
				upMap["tags"] = app.Tags
				upMap["title"] = app.Title
				upMap["description"] = app.Description
				upMap["vendor"] = app.Vendor
				upMap["vendor_url"] = app.VendorUrl
				upMap["packager"] = app.Packager
				upMap["packager_url"] = app.PackagerUrl
				upMap["form_content"] = app.FormContent

				if err := AppRepo.Update(appRecord.ID, upMap); err != nil {
					global.LOG.Error("Failed to update app, %v", err)
				}
				appId = appRecord.ID
			}

			// versions
			appVersions, err := loadVersions(appId, appDir)
			if err != nil {
				continue
			}
			// delete app versions
			if err := AppVersionRepo.Delete(AppVersionRepo.WithByAppID(appId)); err != nil {
				global.LOG.Error("Failed to delete versions for app %s with id %d", app.Name, appId)
			}
			// create app versions
			for _, appVersion := range appVersions {
				if err := AppVersionRepo.Create(&appVersion); err != nil {
					global.LOG.Error("Failed to create version for app %s", app.Name)
				}
			}
		}
	}

	return nil
}

func loadManifest(appDir string) (*model.App, error) {
	var app model.App

	var appData core.App
	manifestYamlPath := filepath.Join(appDir, "manifest.yaml")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(manifestYamlPath) {
		global.LOG.Error("manifest.yaml missed in %s", appDir)
		return &app, errors.New(constant.ErrFileNotFound)
	}
	manifestYamlByte, err := fileOp.GetContent(manifestYamlPath)
	if err != nil {
		global.LOG.Error("faile to get manifest.yaml content %s", manifestYamlPath)
		return &app, err
	}
	if err = yaml.Unmarshal(manifestYamlByte, &appData); err != nil {
		global.LOG.Error("faile to unmarshal manifest.yaml %s", manifestYamlPath)
		return &app, err
	}
	return &model.App{
		Name:        appData.Name,
		DisplayName: appData.DisplayName,
		Category:    appData.Category,
		Tags:        strings.Join(appData.Tags, ","),
		Title:       appData.Title,
		Description: appData.Description,
		Vendor:      appData.Vendor.Name,
		VendorUrl:   appData.Vendor.Url,
		Packager:    appData.Packager.Name,
		PackagerUrl: appData.Packager.Url,
	}, nil
}

func loadForm(appDir string) (string, error) {
	formYamlPath := filepath.Join(appDir, "form.yaml")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(formYamlPath) {
		global.LOG.Error("form.yaml missed in %s", appDir)
		return "", errors.New(constant.ErrFileNotFound)
	}
	formYamlByte, err := fileOp.GetContent(formYamlPath)
	if err != nil {
		global.LOG.Error("faile to get form.yaml content %s", formYamlPath)
		return "", err
	}
	return string(formYamlByte), nil
}

func loadVersions(appId uint, appDir string) ([]model.AppVersion, error) {
	var appVersions []model.AppVersion
	appDirEntries, err := os.ReadDir(appDir)
	if err != nil {
		global.LOG.Error("Error reading apps dir")
		return appVersions, err
	}
	for _, appDirEntry := range appDirEntries {
		if appDirEntry.IsDir() {
			var appVersion model.AppVersion
			versionDir := filepath.Join(appDir, appDirEntry.Name())

			// app id
			appVersion.AppId = appId

			// docker-compose.yaml
			fileOp := files.NewFileOp()
			dockerComposePath := filepath.Join(versionDir, "docker-compose.yaml")
			if !fileOp.Stat(dockerComposePath) {
				global.LOG.Error("docker-compose.yaml missed in %s", versionDir)
				continue
			}
			dockerComposeByte, err := fileOp.GetContent(dockerComposePath)
			if err != nil {
				global.LOG.Error("faile to get manifest.yaml content %s", dockerComposePath)
				continue
			}
			appVersion.ComposeContent = string(dockerComposeByte)

			// .env
			envPath := filepath.Join(versionDir, ".env")
			if !fileOp.Stat(envPath) {
				global.LOG.Error(".env missed in %s", versionDir)
				continue
			}
			envByte, err := fileOp.GetContent(envPath)
			if err != nil {
				global.LOG.Error("faile to get .env content %s", envByte)
				continue
			}
			appVersion.EnvContent = string(envByte)

			// 获取project信息
			envMap, _ := godotenv.Read(envPath)

			project, err := loader.Load(types.ConfigDetails{
				ConfigFiles: []types.ConfigFile{
					{
						Filename: dockerComposePath,
					},
				},
				Environment: envMap,
			})
			if err != nil {
				global.LOG.Error("Failed to load Compose YAML: %v", err)
				continue
			}

			// 遍历 services 部分，拿到版本和升级版本
			for _, service := range project.Services {
				if service.Labels != nil {
					for key, value := range service.Labels {
						switch key {
						case constant.IDBVersion:
							if appVersion.Version == "" {
								appVersion.Version = value
							}
						case constant.IDBUpdateVersion:
							if appVersion.UpdateVersion == "" {
								appVersion.UpdateVersion = value
							}
						default:
						}
					}
				}
			}

			// config (might not exist)
			configDir := filepath.Join(versionDir, "config")
			if fileOp.Stat(configDir) {
				// find conf file
				confFiles, err := os.ReadDir(configDir)
				if err == nil && len(confFiles) > 0 {
					confFile := confFiles[0]
					confFilePath := filepath.Join(configDir, confFile.Name())
					confByte, err := fileOp.GetContent(confFilePath)
					if err == nil {
						appVersion.ConfigName = confFile.Name()
						appVersion.ConfigContent = string(confByte)
					}
				}
			}

			appVersions = append(appVersions, appVersion)
		}
	}

	return appVersions, nil
}

func (s *AppService) RemoveApp(req core.RemoveApp) error {
	// 从数据库中删除应用版本
	if err := AppVersionRepo.Delete(AppVersionRepo.WithByAppID(req.ID)); err != nil {
		global.LOG.Error("Failed to delete versions for app %s with id %d", req.ID, err)
		return err
	}

	// 从数据库中删除应用
	if err := AppRepo.Delete(AppRepo.WithByID(req.ID)); err != nil {
		global.LOG.Error("Failed to delete app %s, %v", req.ID, err)
		return err
	}

	return nil
}

func (s *AppService) AppPage(hostID uint64, req core.QueryApp) (*core.PageResult, error) {
	var result core.PageResult

	// 查询host上已安装的应用
	composeInfos, err := s.composePageInHost(uint(hostID), req.Name, s.AppDir)
	if err != nil {
		global.LOG.Error("Error query compose in host %d, %v", hostID, err)
		return &result, err
	}
	// 转成map
	composeMap := make(map[string]core.ComposeInfo)
	for _, composeInfo := range composeInfos {
		composeMap[composeInfo.IdbName] = composeInfo
	}

	var opts []repo.DBOption
	if req.Name != "" {
		opts = append(opts, CommonRepo.WithLikeName(req.Name))
	}
	if req.Category != "" {
		opts = append(opts, AppRepo.WithByName(req.Name))
	}
	total, apps, err := AppRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return &result, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}
	// db -> dto
	var items []core.App
	for _, appData := range apps {
		status := "uninstalled"
		if _, ok := composeMap[appData.Name]; ok {
			status = "installed"
		}

		items = append(items, core.App{
			ID:          appData.ID,
			Type:        constant.TYPE_APP,
			Name:        appData.Name,
			DisplayName: appData.DisplayName,
			Category:    appData.Category,
			Tags:        strings.Split(appData.Tags, ","),
			Title:       appData.Title,
			Description: appData.Description,
			Vendor:      core.NameUrl{Name: appData.Vendor, Url: appData.VendorUrl},
			Packager:    core.NameUrl{Name: appData.Packager, Url: appData.PackagerUrl},
			Status:      status,
		})
	}
	return &core.PageResult{Total: total, Items: items}, nil
}

func (s *AppService) InstalledAppPage(hostID uint64, req core.QueryInstalledApp) (*core.PageResult, error) {
	var result core.PageResult

	composeInfos, err := s.composePageInHost(uint(hostID), req.Name, s.AppDir)
	if err != nil {
		global.LOG.Error("Error query compose in host %d, %v", hostID, err)
		return &result, err
	}

	// 遍历 ComposeInfo 查询对应的 App
	var (
		apps      []core.App
		BackDatas []core.App
	)
	for _, compose := range composeInfos {
		// 查询App
		appData, err := AppRepo.Get(AppRepo.WithByName(compose.IdbName))
		if err != nil {
			global.LOG.Error("Error query app %s, %v", compose.IdbName, err)
			continue
		}
		// 查询版本信息
		appVersions, err := AppVersionRepo.GetList(AppVersionRepo.WithByAppID(appData.ID))
		if err != nil {
			global.LOG.Error("Error query app %s version, %v", compose.IdbName, err)
			continue
		}

		composeUpdVersion, err := strconv.Atoi(compose.IdbUpdateVersion)
		if err != nil {
			global.LOG.Error("Failed to convert Compose update version: %v", err)
			composeUpdVersion = 0
		}
		hasUpgrade := false
		status := "uninstalled"
		var versions []core.AppVersion
		for _, appVersion := range appVersions {
			v := core.AppVersion{
				ID:             appVersion.ID,
				Version:        appVersion.Version,
				UpdateVersion:  appVersion.UpdateVersion,
				ComposeContent: appVersion.ComposeContent,
				Status:         "uninstalled",
				CreatedAt:      "",
				CanUpgrade:     false,
			}

			if appVersion.Version == compose.IdbVersion {
				status = "installed"
				v.Status = "installed"
				v.CreatedAt = compose.CreatedAt
			}

			dbUpdVersion, err := strconv.Atoi(appVersion.UpdateVersion)
			if err != nil {
				global.LOG.Error("Failed to convert DB update version: %v", err)
				dbUpdVersion = 0
			}
			v.CanUpgrade = dbUpdVersion > composeUpdVersion
			if v.CanUpgrade {
				hasUpgrade = true
			}

			versions = append(versions, v)
		}

		apps = append(apps, core.App{
			ID:             appData.ID,
			Type:           constant.TYPE_APP,
			Name:           compose.Name,
			DisplayName:    appData.DisplayName,
			Category:       appData.Category,
			Tags:           strings.Split(appData.Tags, ","),
			Title:          appData.Title,
			Description:    appData.Description,
			Vendor:         core.NameUrl{Name: appData.Vendor, Url: appData.VendorUrl},
			Packager:       core.NameUrl{Name: appData.Packager, Url: appData.PackagerUrl},
			HasUpgrade:     hasUpgrade,
			Versions:       versions,
			CurrentVersion: fmt.Sprintf("%s.%s", compose.IdbVersion, compose.IdbUpdateVersion),
			Status:         status,
		})
	}

	// 分页
	total, start, end := len(apps), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		BackDatas = make([]core.App, 0)
	} else {
		if end >= total {
			end = total
		}
		BackDatas = apps[start:end]
	}
	result.Total = int64(total)
	result.Items = BackDatas

	return &result, nil
}

func (s *AppService) composePageInHost(hostID uint, info string, workDir string) ([]core.ComposeInfo, error) {
	var result []core.ComposeInfo
	queryCompose := core.QueryCompose{
		PageInfo: core.PageInfo{Page: 1, PageSize: 10000}, // get all
		Info:     info,
		WorkDir:  workDir,
		IdbType:  constant.TYPE_APP,
	}
	data, err := utils.ToJSONString(queryCompose)
	if err != nil {
		return result, err
	}
	actionRequest := core.HostAction{
		HostID: hostID,
		Action: core.Action{
			Action: core.Docker_Compose_Page,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Docker_Compose_Page %v", err)
		return result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Docker_Compose_Page failed")
		return result, fmt.Errorf("failed to query compose")
	}
	var composeResult core.PageResult
	err = utils.FromJSONString(actionResponse.Data, &composeResult)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose query result: %v", err)
		return result, fmt.Errorf("unmarshal err: %v", err)
	}
	// 将 Items 转换为 []ComposeInfo 类型
	itemsJSON, err := utils.ToJSONString(composeResult.Items)
	if err != nil {
		global.LOG.Error("Error marshaling Items: %v", err)
		return result, fmt.Errorf("marshal err: %v", err)
	}
	if err := utils.FromJSONString(itemsJSON, &result); err != nil {
		global.LOG.Error("Error unmarshaling Items to []ComposeInfo: %v", err)
		return result, fmt.Errorf("unmarshal err: %v", err)
	}
	return result, nil
}

func (s *AppService) AppDetail(hostID uint64, req core.QueryAppDetail) (*core.App, error) {
	app, err := AppRepo.Get(AppRepo.WithByID(req.ID))
	if err != nil {
		global.LOG.Error("App %d data not found", req.ID)
		return nil, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	appInfo := core.App{
		ID:          app.ID,
		Name:        app.Name,
		DisplayName: app.DisplayName,
		Category:    app.Category,
		Tags:        strings.Split(app.Tags, ","),
		Title:       app.Title,
		Description: app.Description,
		Vendor:      core.NameUrl{Name: app.Vendor, Url: app.VendorUrl},
		Packager:    core.NameUrl{Name: app.Packager, Url: app.PackagerUrl},
	}

	versions, _ := AppVersionRepo.GetList(AppVersionRepo.WithByAppID(app.ID))
	for _, version := range versions {
		appInfo.Versions = append(appInfo.Versions, core.AppVersion{
			ID:             version.ID,
			Version:        version.Version,
			UpdateVersion:  version.UpdateVersion,
			ComposeContent: version.ComposeContent,
		})
	}

	var form core.Form
	if err := yaml.Unmarshal([]byte(app.FormContent), &form); err != nil {
		global.LOG.Error("Failed to unmarshal app form data: %v", err)
		return nil, fmt.Errorf("unmarshal form err: %v", err)
	}
	appInfo.Form = form

	return &appInfo, nil
}

func (s *AppService) AppInstall(hostID uint64, req core.InstallApp) (*core.ComposeCreateResult, error) {
	var result core.ComposeCreateResult

	// 查找应用
	app, err := AppRepo.Get(AppRepo.WithByID(req.ID))
	if err != nil {
		global.LOG.Error("App %d not found", req.ID)
		return &result, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	var form core.Form
	if err := yaml.Unmarshal([]byte(app.FormContent), &form); err != nil {
		global.LOG.Error("Failed to unmarshal app form data: %v", err)
		return &result, fmt.Errorf("unmarshal form err: %v", err)
	}

	// 使用传入的compose名称
	appName := req.ComposeName
	if appName == "" {
		appName = app.Name
	}

	// 找版本
	version, err := AppVersionRepo.Get(AppVersionRepo.WithByID(req.VersionID))
	if err != nil {
		global.LOG.Error("App version %d not found", req.VersionID)
		return &result, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 处理env
	envMap, err := godotenv.Unmarshal(version.EnvContent)
	if err != nil {
		return &result, fmt.Errorf("unmarshal env err : %v", err)
	}
	// 校验form params
	if len(req.FormParams) > 0 {
		// 字段规则
		validKeys := make(map[string]core.FormField)
		for _, field := range form.Fields {
			validKeys[field.Name] = field
		}
		// 校验env params
		for _, param := range req.FormParams {
			// 检查 key 是否在 validKeys 中
			formField, exists := validKeys[param.Key]
			if !exists {
				// 不存在，返回错误
				global.LOG.Error("Invalid form key: %s", param.Key)
				return &result, fmt.Errorf("invalid key: %s", param.Key)
			}

			// 设置了校验规则
			if formField.Validation != nil {
				// 设置了正则匹配，优先正则匹配
				if formField.Validation.Pattern != "" {
					// 使用正则表达式校验
					matched, err := regexp.MatchString(formField.Validation.Pattern, param.Value)
					if err != nil {
						global.LOG.Error("Invalid regex pattern: %v", err)
						return &result, fmt.Errorf("invalid regex pattern for key %s: %v", param.Key, err)
					}
					if !matched {
						global.LOG.Error("Value %s does not match the required pattern for key %s", param.Value, param.Key)
						return &result, fmt.Errorf("invalid value for key %s", param.Key)
					}
				}
				// 设置了长度限制
				if formField.Validation.MinLength >= 0 && formField.Validation.MaxLength != 0 && formField.Validation.MaxLength >= formField.Validation.MinLength {
					if len(param.Value) < formField.Validation.MinLength || len(param.Value) > formField.Validation.MaxLength {
						global.LOG.Error("Value %s does not has valid length for key %s", param.Value, param.Key)
						return &result, fmt.Errorf("invalid value for key %s", param.Key)
					}
				}
				// 是数值类型，且设置了值大小
				if formField.Type == "number" && formField.Validation.MaxValue >= formField.Validation.MinValue {
					paramValue, err := strconv.Atoi(param.Value)
					if err != nil || (paramValue < formField.Validation.MinValue || paramValue > formField.Validation.MaxValue) {
						global.LOG.Error("Value %s is not valid number for key %s", param.Value, param.Key)
						return &result, fmt.Errorf("invalid number value for key %s", param.Key)
					}
				}
			}

			// 校验通过，传递值到envMap中，formField.Name -> envMap key
			if _, exist := envMap[formField.Name]; exist {
				// 密码类型，可能包含特殊字符，以单引号包含，避免转义错误
				if formField.Type == "password" {
					envMap[formField.Name] = fmt.Sprintf("'%s'", param.Value)
				} else {
					envMap[formField.Name] = param.Value
				}
			}
		}
	}
	// 额外的params
	for _, param := range req.ExtraParams {
		envMap[param.Key] = param.Value
	}
	// 转换成env内容
	var envArray []string
	for key, value := range envMap {
		// 应用名，也支持从form中传入
		if key == constant.IDB_compose_name {
			appName = value
		}
		envArray = append(envArray, fmt.Sprintf("%s=%s", key, value))
	}
	envContent := strings.Join(envArray, "\n")

	// 处理compose内容
	var composeContent string
	if req.ComposeContent != "" {
		// 使用传入的内容
		composeContent = req.ComposeContent
	} else {
		// 使用版本内容
		composeContent = version.ComposeContent
	}

	// 处理conf
	var confPath, confContent string
	if confDir, exist := envMap[constant.IDB_service_config_path]; exist {
		confPath = filepath.Join(confDir, version.ConfigName)
		confContent = version.ConfigContent
	}

	// 发送compose create请求
	composeCreate := core.ComposeCreate{
		Name:           appName,
		ComposeContent: composeContent,
		EnvContent:     envContent,
		ConfContent:    confContent,
		ConfPath:       confPath,
		WorkDir:        s.AppDir,
	}
	data, err := utils.ToJSONString(composeCreate)
	if err != nil {
		return &result, err
	}

	actionRequest := core.HostAction{
		HostID: uint(hostID),
		Action: core.Action{
			Action: core.Docker_Compose_Create,
			Data:   data,
		},
	}

	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Docker_Compose_Create %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Docker_Compose_Create failed")
		return &result, fmt.Errorf("failed to create compose")
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose create result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *AppService) AppUninstall(hostID uint64, req core.UninstallApp) (*core.ComposeCreateResult, error) {
	var result core.ComposeCreateResult
	composeRemove := core.ComposeRemove{
		Name:    req.ComposeName,
		WorkDir: s.AppDir,
	}
	data, err := utils.ToJSONString(composeRemove)
	if err != nil {
		return &result, err
	}
	actionRequest := core.HostAction{
		HostID: uint(hostID),
		Action: core.Action{
			Action: core.Docker_Compose_Remove,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Docker_Compose_Remove %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Docker_Compose_Remove failed")
		return &result, fmt.Errorf("failed to remove compose: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose remove result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *AppService) AppUpgrade(hostID uint64, req core.UpgradeApp) (*core.ComposeCreateResult, error) {
	var result core.ComposeCreateResult

	// 查找应用
	app, err := AppRepo.Get(AppRepo.WithByID(req.ID))
	if err != nil {
		global.LOG.Error("App %d not found", req.ID)
		return &result, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}
	var form core.Form
	if err := yaml.Unmarshal([]byte(app.FormContent), &form); err != nil {
		global.LOG.Error("Failed to unmarshal app form data: %v", err)
		return &result, fmt.Errorf("unmarshal form err: %v", err)
	}

	// 找版本
	version, err := AppVersionRepo.Get(AppVersionRepo.WithByID(req.UpgradeVersionID))
	if err != nil {
		global.LOG.Error("App version %d not found", req.UpgradeVersionID)
		return &result, errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	// 新版env
	envMap, err := godotenv.Unmarshal(version.EnvContent)
	if err != nil {
		return &result, fmt.Errorf("unmarshal env err : %v", err)
	}

	// 旧版env
	var composeDetailRsp core.ComposeDetailRsp
	composeDetailReq := core.ComposeDetailReq{
		Name:    req.ComposeName,
		WorkDir: s.AppDir,
	}
	data, err := utils.ToJSONString(composeDetailReq)
	if err != nil {
		return &result, err
	}
	actionRequest := core.HostAction{
		HostID: uint(hostID),
		Action: core.Action{
			Action: core.Docker_Compose_Detail,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Docker_Compose_Detail %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Docker_Compose_Detail failed")
		return &result, fmt.Errorf("failed to query compose detail")
	}
	err = utils.FromJSONString(actionResponse.Data, &composeDetailRsp)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose detail result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}
	oldEnvMap, err := godotenv.Unmarshal(composeDetailRsp.EnvContent)
	if err != nil {
		return &result, fmt.Errorf("unmarshal env err : %v", err)
	}

	// 合并env，旧值覆盖
	mergedEnvMap := map[string]string{}
	for k, v := range envMap {
		mergedEnvMap[k] = v
	}
	for k, v := range oldEnvMap {
		mergedEnvMap[k] = v
	}
	// 转换成env内容
	var envArray []string
	for key, value := range mergedEnvMap {
		envArray = append(envArray, fmt.Sprintf("%s=%s", key, value))
	}
	envContent := strings.Join(envArray, "\n")

	// 处理compose内容
	composeContent := version.ComposeContent

	// 处理conf
	var confPath, confContent string
	if confDir, exist := envMap[constant.IDB_service_config_path]; exist {
		confPath = filepath.Join(confDir, version.ConfigName)
		confContent = version.ConfigContent
	}

	composeUpgrade := core.ComposeUpgrade{
		Name:           req.ComposeName,
		ComposeContent: composeContent,
		EnvContent:     envContent,
		ConfContent:    confContent,
		ConfPath:       confPath,
		WorkDir:        s.AppDir,
	}
	data, err = utils.ToJSONString(composeUpgrade)
	if err != nil {
		return &result, err
	}

	actionRequest = core.HostAction{
		HostID: uint(hostID),
		Action: core.Action{
			Action: core.Docker_Compose_Upgrade,
			Data:   data,
		},
	}
	actionResponse, err = conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action Docker_Compose_Upgrade %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action Docker_Compose_Upgrade failed")
		return &result, fmt.Errorf("failed to upgrade compose: %s", actionResponse.Data)
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose upgrade result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

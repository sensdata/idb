package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/types"

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
	AppPage(req core.QueryApp) (*core.PageResult, error)
	InstalledAppPage(hostID uint64, req core.QueryInstalledApp) (*core.PageResult, error)
	AppDetail(req core.QueryAppDetail) (*core.App, error)
	AppInstall(hostID uint64, req core.InstallApp) (*core.ComposeCreateResult, error)
	AppUninstall(hostID uint64, req core.UninstallApp) error
}

func NewIAppService() IAppService {
	return &AppService{AppDir: constant.AgentDockerDir}
}

func (s *AppService) SyncApp() error {
	global.LOG.Info("SyncApp begin")

	// 先检查应用目录是否存在
	repoPath := filepath.Join(constant.CenterDataDir, constant.StoreDir)
	if err := utils.EnsurePaths([]string{repoPath}); err != nil {
		global.LOG.Error("Failed to create repo dir %s, %v", repoPath, err)
		return err
	}

	// 定义仓库路径
	repoURL := "https://github.com/sensdata/idb-store.git"

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
	if err != nil && err != git.NoErrAlreadyUpToDate {
		global.LOG.Error("Error pulling changes: %s\n", err)
		return err
	}

	if err == git.NoErrAlreadyUpToDate {
		global.LOG.Error("Repository is already up-to-date.")
	} else {
		global.LOG.Error("Successfully pulled latest changes.")
	}

	// 打印当前HEAD
	ref, err := repo.Head()
	if err != nil {
		global.LOG.Error("Error getting HEAD: %s\n", err)
		return err
	}
	global.LOG.Error("Current HEAD: %s\n", ref.Hash())

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
			appRecord, err := AppRepo.Get(AppRepo.WithByName(app.Name))
			if err != nil {
				global.LOG.Error("Error when checking app record, %v", err)
				continue
			}
			var appId uint
			if appRecord.Name == "" {
				if err := AppRepo.Create(app); err != nil {
					global.LOG.Error("Failed to create app, %v", err)
				}
				appId = app.ID
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
			if err := AppVersionRepo.Delete(AppVersionRepo.WithByID(appId)); err != nil {
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

			// compose.yml
			fileOp := files.NewFileOp()
			dockerComposePath := filepath.Join(versionDir, "docker-compose.yml")
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

			var composeConfig types.Config
			err = yaml.Unmarshal(dockerComposeByte, &composeConfig)
			if err != nil {
				global.LOG.Error("Failed to unmarshal Compose YAML: %v", err)
				continue
			}
			// 遍历 services 部分，拿到版本和升级版本
			for _, service := range composeConfig.Services {
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

func (s *AppService) AppPage(req core.QueryApp) (*core.PageResult, error) {
	var opts []repo.DBOption
	if req.Name != "" {
		opts = append(opts, CommonRepo.WithLikeName(req.Name))
	}
	if req.Category != "" {
		opts = append(opts, AppRepo.WithByName(req.Name))
	}
	total, apps, err := AppRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}
	// db -> dto
	var items []core.App
	for _, appData := range apps {
		items = append(items, core.App{
			Name:        appData.Name,
			DisplayName: appData.DisplayName,
			Category:    appData.Category,
			Tags:        strings.Split(appData.Tags, ","),
			Title:       appData.Title,
			Description: appData.Description,
			Vendor:      core.NameUrl{Name: appData.Vendor, Url: appData.VendorUrl},
			Packager:    core.NameUrl{Name: appData.Packager, Url: appData.PackagerUrl},
		})
	}
	return &core.PageResult{Total: total, Items: items}, nil
}

func (s *AppService) InstalledAppPage(hostID uint64, req core.QueryInstalledApp) (*core.PageResult, error) {
	var result core.PageResult

	queryCompose := core.QueryCompose{
		PageInfo: core.PageInfo{Page: 1, PageSize: 10000}, // get all
		Info:     req.Name,
		WorkDir:  s.AppDir,
		IdbType:  constant.TYPE_APP,
	}
	data, err := utils.ToJSONString(queryCompose)
	if err != nil {
		return &result, err
	}

	actionRequest := core.HostAction{
		HostID: uint(hostID),
		Action: core.Action{
			Action: core.Docker_Compose_Page,
			Data:   data,
		},
	}
	actionResponse, err := conn.CENTER.ExecuteAction(actionRequest)
	if err != nil {
		global.LOG.Error("Failed to send action %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query compose")
	}
	var composeResult core.PageResult
	err = utils.FromJSONString(actionResponse.Data, &composeResult)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose query result: %v", err)
		return &result, fmt.Errorf("unmarshal err: %v", err)
	}

	// 将 Items 转换为 []ComposeInfo 类型
	itemsJSON, err := utils.ToJSONString(composeResult.Items)
	if err != nil {
		global.LOG.Error("Error marshaling Items: %v", err)
		return &result, fmt.Errorf("marshal err: %v", err)
	}
	var composeInfos []core.ComposeInfo
	if err := utils.FromJSONString(itemsJSON, &composeInfos); err != nil {
		global.LOG.Error("Error unmarshaling Items to []ComposeInfo: %v", err)
		return &result, fmt.Errorf("unmarshal err: %v", err)
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
		appVersions, err := AppVersionRepo.GetList(AppVersionRepo.WithByID(appData.ID))
		if err != nil {
			global.LOG.Error("Error query app %s version, %v", compose.IdbName, err)
			continue
		}
		hasUpdate := false
		var versions []core.AppVersion
		for _, version := range appVersions {
			versions = append(versions, core.AppVersion{
				ID:             version.ID,
				Version:        version.Version,
				UpdateVersion:  version.UpdateVersion,
				ComposeContent: version.ComposeContent,
			})
			if version.Version == compose.IdbVersion {
				composeUpdtVersion, err := strconv.Atoi(compose.IdbUpdateVersion)
				if err != nil {
					global.LOG.Error("Failed to convert Compose update version: %v", err)
					composeUpdtVersion = 0
				}

				dbUpdtVersion, err := strconv.Atoi(version.UpdateVersion)
				if err != nil {
					global.LOG.Error("Failed to convert DB update version: %v", err)
					dbUpdtVersion = 0
				}

				// 判断是否可以更新
				hasUpdate = dbUpdtVersion > composeUpdtVersion
			}
		}
		apps = append(apps, core.App{
			Name:        appData.Name,
			DisplayName: appData.DisplayName,
			Category:    appData.Category,
			Tags:        strings.Split(appData.Tags, ","),
			Title:       appData.Title,
			Description: appData.Description,
			Vendor:      core.NameUrl{Name: appData.Vendor, Url: appData.VendorUrl},
			Packager:    core.NameUrl{Name: appData.Packager, Url: appData.PackagerUrl},
			HasUpdate:   hasUpdate,
			Versions:    versions,
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

func (s *AppService) AppDetail(req core.QueryAppDetail) (*core.App, error) {
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

	versions, _ := AppVersionRepo.GetList(AppVersionRepo.WithByID(app.ID))
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
	appName := app.Name

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
			validKeys[field.Key] = field
		}
		// 校验env params
		for _, param := range req.FormParams {
			// 检查 key 是否在 validKeys 中
			formField, exists := validKeys[param.Key]
			if exists {
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
						// 校验通过
						continue
					}
					// 设置了长度限制
					if formField.Validation.MinLength >= 0 && formField.Validation.MaxLength >= formField.Validation.MinLength {
						if len(param.Value) < formField.Validation.MinLength || len(param.Value) > formField.Validation.MaxLength {
							global.LOG.Error("Value %s does not has valid length for key %s", param.Value, param.Key)
							return &result, fmt.Errorf("invalid value for key %s", param.Key)
						}
						// 校验通过
						continue
					}
					// 是数值类型，且设置了值大小
					if formField.Type == "number" && formField.Validation.MinValue >= 0 && formField.Validation.MaxValue >= formField.Validation.MinValue {
						paramValue, err := strconv.Atoi(param.Value)
						if err != nil || (paramValue < formField.Validation.MinValue || paramValue > formField.Validation.MaxValue) {
							global.LOG.Error("Value %s is not valid number for key %s", param.Value, param.Key)
							return &result, fmt.Errorf("invalid number value for key %s", param.Key)
						}
					}
				}
				// 校验通过，根据传入的form params，替换env中的对应值
				if _, exist := envMap[param.Key]; exist {
					// 密码类型，可能包含特殊字符，以单引号包含，避免转义错误
					if formField.Type == "password" {
						envMap[param.Key] = fmt.Sprintf("'%s'", param.Value)
					} else {
						envMap[param.Key] = param.Value
					}
				}
			} else {
				// 不存在，返回错误
				global.LOG.Error("Invalid form key: %s", param.Key)
				return &result, fmt.Errorf("invalid key: %s", param.Key)
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
		// 应用名
		if key == "iDB_name" {
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
	if confDir, exist := envMap["iDB_service_conf_path"]; exist {
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
		global.LOG.Error("Failed to send action %v", err)
		return &result, err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to create compose")
	}

	err = utils.FromJSONString(actionResponse.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to compose create result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *AppService) AppUninstall(hostID uint64, req core.UninstallApp) error {

	// 查找应用
	app, err := AppRepo.Get(AppRepo.WithByID(req.ID))
	if err != nil {
		global.LOG.Error("App %d not found", req.ID)
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	appName := app.Name

	composeRemove := core.ComposeRemove{
		Name:    appName,
		WorkDir: s.AppDir,
	}
	data, err := utils.ToJSONString(composeRemove)
	if err != nil {
		return err
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
		global.LOG.Error("Failed to send action %v", err)
		return err
	}
	if !actionResponse.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to remove compose")
	}
	return nil
}

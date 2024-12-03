package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/types"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	AppInstall(hostID uint64, req core.ComposeCreate) (*core.ComposeCreateResult, error)
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
			// TODO: load form.yaml and save in app db

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
		return &result, fmt.Errorf("json err: %v", err)
	}

	// 将 Items 转换为 []ComposeInfo 类型
	itemsJSON, err := utils.ToJSONString(composeResult.Items)
	if err != nil {
		global.LOG.Error("Error marshaling Items: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}
	var composeInfos []core.ComposeInfo
	if err := utils.FromJSONString(itemsJSON, &composeInfos); err != nil {
		global.LOG.Error("Error unmarshaling Items to []ComposeInfo: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
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
				EnvContent:     version.EnvContent,
				// ConfigName:     version.ConfigName,    // TODO：config文件的处理
				// ConfigContent:  version.ConfigContent,
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
			EnvContent:     version.EnvContent,
			// ConfigName:     version.ConfigName,    // TODO：config文件的处理
			// ConfigContent:  version.ConfigContent,
		})
	}

	return &appInfo, nil
}

func (s *AppService) AppInstall(hostID uint64, req core.ComposeCreate) (*core.ComposeCreateResult, error) {
	var result core.ComposeCreateResult

	data, err := utils.ToJSONString(req)
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

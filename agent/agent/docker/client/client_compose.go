package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/compose-spec/compose-go/loader"
	composeTypes "github.com/compose-spec/compose-go/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/files"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func isInManagedRoot(workingDir, managedRoot string) bool {
	return strings.HasPrefix(workingDir, managedRoot)
}

func isProjectManaged(projectName, managedRoot string) bool {
	projectPath := filepath.Join(managedRoot, projectName)
	_, err := os.Stat(projectPath)
	return err == nil
}

func isContainerFromManagedCompose(labels map[string]string, managedRoot string) (string, string, string, bool) {
	projectName, hasProjectName := labels[constant.ComposeProjectLabel]
	workingDir, hasWorkingDir := labels[constant.ComposeWorkDirLabel]
	configFiles, _ := labels[constant.ComposeConfFilesLabel]

	if !hasWorkingDir || !hasProjectName {
		return "", "", "", false
	}

	isContainerFromManagedCompose := isInManagedRoot(workingDir, managedRoot) && isProjectManaged(projectName, managedRoot)

	return projectName, workingDir, configFiles, isContainerFromManagedCompose
}

// resolveComposeConfigPaths 将 config_files label 中的路径转换为绝对路径列表
func resolveComposeConfigPaths(workdir, configFiles string) []string {
	var result []string
	for _, f := range strings.Split(configFiles, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if filepath.IsAbs(f) {
			result = append(result, f)
		} else {
			result = append(result, filepath.Join(workdir, f))
		}
	}
	return result
}

// mergeComposeConfigPaths 合并已有和新增的配置路径，并生成去重后的逗号分隔字符串
func mergeComposeConfigPaths(existing string, incoming []string) string {
	existingSet := make(map[string]struct{})

	// 处理 existing 字符串
	for _, f := range strings.Split(existing, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			existingSet[f] = struct{}{}
		}
	}

	// 处理 incoming 切片
	for _, f := range incoming {
		f = strings.TrimSpace(f)
		if f != "" {
			existingSet[f] = struct{}{}
		}
	}

	var merged []string
	for f := range existingSet {
		merged = append(merged, f)
	}
	sort.Strings(merged) // 保持一致性
	return strings.Join(merged, ",")
}

// 列举compose项目: /{workdir}/docker/{project}
func (c DockerClient) listComposeProjects(workDir string) ([]string, error) {
	var projects []string

	// 检查workDir是否存在
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return projects, fmt.Errorf("workDir %s does not exist", workDir)
	}

	// 枚举workDir下的所有文件夹
	files, err := os.ReadDir(workDir)
	if err != nil {
		return projects, err
	}

	for _, file := range files {
		if file.IsDir() {
			projects = append(projects, file.Name())
		}
	}

	return projects, nil
}

func (c DockerClient) initComposeAndEnv(req *model.ComposeCreate) (string, error) {
	dir := filepath.Join(req.WorkDir, req.Name)
	composePath := filepath.Join(dir, "docker-compose.yaml")
	file, err := os.OpenFile(composePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.ComposeContent)
	write.Flush()

	envPath := filepath.Join(dir, ".env")
	file, err = os.OpenFile(envPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	write = bufio.NewWriter(file)
	_, _ = write.WriteString(req.EnvContent)
	write.Flush()

	return composePath, nil
}

func (c DockerClient) initConf(path string, content string, upgrade bool) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	// 如果是升级，并且文件已存在，则跳过写入
	if upgrade {
		if _, err := os.Stat(path); err == nil {
			global.LOG.Info("Conf file %s already exists during upgrade, skipping overwrite", path)
			return nil
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(content)
	write.Flush()

	return nil
}

func pull(filePath string) (string, error) {
	stdout, err := utils.Execf("docker compose -f %s pull", filePath)
	return stdout, err
}

func up(filePath string) (string, error) {
	stdout, err := utils.Execf("docker compose -f %s up -d", filePath)
	return stdout, err
}

func down(filePath string, removeVolumes bool) (string, error) {
	var removeVolumesFlag string
	if removeVolumes {
		removeVolumesFlag = "--volumes"
	}
	stdout, err := utils.Execf("docker compose -f %s down --remove-orphans %s", filePath, removeVolumesFlag)
	return stdout, err
}

func start(filePath string) (string, error) {
	stdout, err := utils.Execf("docker compose -f %s start", filePath)
	return stdout, err
}

func stop(filePath string) (string, error) {
	stdout, err := utils.Execf("docker compose -f %s stop", filePath)
	return stdout, err
}

func restart(filePath string) (string, error) {
	stdout, err := utils.Execf("docker compose -f %s restart", filePath)
	return stdout, err
}

func operate(filePath, operation string) (string, error) {
	stdout, err := utils.Execf("docker compose -f %s %s", filePath, operation)
	return stdout, err
}

func (c DockerClient) ComposePage(req model.QueryCompose) (*model.PageResult, error) {
	var result model.PageResult

	if utils.CheckIllegal(req.Info, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 获取所有容器
	allContainers, err := c.cli.ContainerList(
		context.Background(),
		container.ListOptions{All: true},
	)
	if err != nil {
		return &result, err
	}

	// 构建 容器workdir -> 容器列表 映射
	containerMap := make(map[string][]types.Container)
	for _, container := range allContainers {
		if container.Labels != nil {
			_, workDir, _, ok := isContainerFromManagedCompose(container.Labels, req.WorkDir)
			if ok {
				containerMap[workDir] = append(containerMap[workDir], container)
			}
		}
	}

	// 列出 req.WorkDir 下所有项目
	projects, err := c.listComposeProjects(req.WorkDir)
	if err != nil {
		return &result, err
	}

	var records []model.ComposeInfo
	for _, project := range projects {
		// 项目路径
		workDir := filepath.Join(req.WorkDir, project)
		containers := containerMap[workDir]

		info := model.ComposeInfo{
			Name:        project,
			Workdir:     req.WorkDir,
			Path:        workDir,
			Containers:  make([]model.ComposeContainer, 0),
			CreatedAt:   "",
			ConfigFiles: "",
		}

		confSet := make(map[string]struct{})
		statusCount := make(map[string]int)

		if len(containers) == 0 {
			// 当未启动任何容器时，尝试从 docker-compose.yaml 中提取元信息
			composePath := filepath.Join(workDir, "docker-compose.yaml")
			envPath := filepath.Join(workDir, ".env")
			envMap, _ := godotenv.Read(envPath)
			project, err := loader.Load(composeTypes.ConfigDetails{
				ConfigFiles: []composeTypes.ConfigFile{
					{
						Filename: composePath,
					},
				},
				Environment: envMap,
				WorkingDir:  workDir,
			})
			if err == nil {
				for _, svc := range project.Services {
					labels := svc.Labels
					if info.IdbType == "" {
						info.IdbType = labels["net.idb.type"]
					}
					if info.IdbName == "" {
						info.IdbName = labels["net.idb.name"]
					}
					if info.IdbVersion == "" {
						info.IdbVersion = labels["net.idb.version"]
					}
					if info.IdbUpdateVersion == "" {
						info.IdbUpdateVersion = labels["net.idb.update_version"]
					}
					if info.IdbPanel == "" {
						info.IdbPanel = labels["net.idb.panel"]
					}
					// 如果都已获取，提前退出
					if info.IdbType != "" && info.IdbName != "" && info.IdbVersion != "" &&
						info.IdbUpdateVersion != "" && info.IdbPanel != "" {
						break
					}
				}
			} else {
				global.LOG.Error("Compose file load failed in %s: %v", workDir, err)
			}

			info.Status = model.ComposeStatusNotDeployed

			if _, err := os.Stat(composePath); err == nil {
				confSet[composePath] = struct{}{}
			}
		} else {
			for _, container := range containers {
				containerItem := model.ComposeContainer{
					ContainerID: container.ID,
					Name:        strings.TrimPrefix(container.Names[0], "/"),
					State:       container.State,
					CreateTime:  time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
				}
				info.Containers = append(info.Containers, containerItem)
				info.ContainerNumber++

				if info.CreatedAt == "" {
					info.CreatedAt = containerItem.CreateTime
				}

				statusCount[container.State]++

				// 合并配置路径
				if container.Labels != nil {
					_, _, configFiles, ok := isContainerFromManagedCompose(container.Labels, req.WorkDir)
					if ok {
						for _, f := range resolveComposeConfigPaths(workDir, configFiles) {
							confSet[f] = struct{}{}
						}
					}
				}

				// 提取标签（首次有效即可）
				if info.IdbType == "" {
					info.IdbType = container.Labels[constant.IDBType]
				}
				if info.IdbName == "" {
					info.IdbName = container.Labels[constant.IDBName]
				}
				if info.IdbVersion == "" {
					info.IdbVersion = container.Labels[constant.IDBVersion]
				}
				if info.IdbUpdateVersion == "" {
					info.IdbUpdateVersion = container.Labels[constant.IDBUpdateVersion]
				}
				if info.IdbPanel == "" {
					info.IdbPanel = container.Labels[constant.IDBPanel]
				}
			}

			// 组装状态
			switch len(statusCount) {
			case 0:
				info.Status = model.ComposeStatusNotDeployed
			case 1:
				for state := range statusCount {
					switch state {
					case "running":
						info.Status = model.ComposeStatusRunning
					case "exited":
						info.Status = model.ComposeStatusExited
					case "paused":
						info.Status = model.ComposeStatusPaused
					case "restarting":
						info.Status = model.ComposeStatusRestarting
					case "removing":
						info.Status = model.ComposeStatusRemoving
					case "dead":
						info.Status = model.ComposeStatusDead
					default:
						info.Status = model.ComposeStatusUnknown
					}
				}
			default:
				info.Status = model.ComposeStatusMixed
			}
		}

		// 配置路径合并
		if len(confSet) > 0 {
			confList := make([]string, 0, len(confSet))
			for f := range confSet {
				confList = append(confList, f)
			}
			info.ConfigFiles = strings.Join(confList, ",")
		}

		// 筛选 idbType
		if req.IdbType != "" && info.IdbType != req.IdbType {
			continue
		}
		// 筛选关键字
		if req.Info != "" && !strings.Contains(info.Name, req.Info) {
			continue
		}

		records = append(records, info)
	}

	// 排序
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})

	// 分页
	total := len(records)
	start := (req.Page - 1) * req.PageSize
	end := req.Page * req.PageSize
	if start > total {
		result.Items = []model.ComposeInfo{}
	} else {
		if end > total {
			end = total
		}
		result.Items = records[start:end]
	}
	result.Total = int64(total)
	return &result, nil
}

func (c DockerClient) ComposeTest(req model.ComposeCreate) (*model.ComposeTestResult, error) {
	result := model.ComposeTestResult{Success: false}
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		result.Error = constant.ErrCmdIllegal
		return &result, errors.New(constant.ErrCmdIllegal)
	}
	composeProjects, _ := c.listComposeProjects(req.WorkDir)
	// 检查project是否已经存在
	for _, project := range composeProjects {
		if project == req.Name {
			result.Error = "compose project already exists"
			return &result, errors.New("compose project already exists")
		}
	}
	// 写入docker-compose.yaml和.env
	composePath, err := c.initComposeAndEnv(&req)
	if err != nil {
		result.Error = "failed to init compose and env"
		return &result, err
	}
	cmd := exec.Command("docker", "compose", "-f", composePath, "config")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = string(stdout)
		return &result, errors.New(string(stdout))
	}
	result.Success = true
	return &result, nil
}

func (c DockerClient) ComposeCreate(req model.ComposeCreate) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 初始化日志
	dockerLogDir := filepath.Join(req.WorkDir, req.Name, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, errors.New("failed to mkdir docker_logs")
		}
	}
	logPath := filepath.Join(dockerLogDir, fmt.Sprintf("compose_create_%s.log", time.Now().Format("20060102150405")))
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, errors.New("failed to create compose_create log")
	}
	result.Log = logPath

	// logger
	logger := utils.NewStepLogger(file)

	// 异步执行，写入日志
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				global.LOG.Error("ComposeCreate panic recovered: %v", r)
			}
			file.Close()
		}()

		// 写入docker-compose.yaml和.env
		logger.Info("try init compose and env")
		composePath, err := c.initComposeAndEnv(&req)
		if err != nil {
			logger.Error("init compose and env failed: %v", err)
			global.LOG.Error("Failed to init compose and env, %v", err)
			return
		}
		logger.Info("init compose and env successful")
		global.LOG.Info("init compose and env successful")

		// 写入conf
		if req.ConfPath != "" && req.ConfContent != "" {
			if err := c.initConf(req.ConfPath, req.ConfContent, false); err != nil {
				logger.Error("init conf failed: %v", err)
				global.LOG.Error("Failed to init conf %s, %v", req.ConfPath, err)
				return
			}
			logger.Info("init conf successful")
			global.LOG.Info("init conf successful")
		}

		logger.Info("try docker compose up %s", req.Name)
		global.LOG.Info("try docker compose up %s", req.Name)

		if stdout, err := up(composePath); err != nil {
			logger.Error("docker compose up %s failed, err: %v", req.Name, strings.TrimSpace(stdout), err)
			global.LOG.Error("docker compose up %s failed, stdout: %s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			_, _ = down(composePath, true)
			return
		}

		logger.Info("docker compose up %s successful!", req.Name)
		global.LOG.Info("docker compose up %s successful!", req.Name)
	}()

	return &result, nil
}

func (c DockerClient) ComposeRemove(req model.ComposeRemove) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 初始化日志 (删除时，使用临时目录，因为compose目录会被清理)
	dockerLogDir := filepath.Join(os.TempDir(), "idb_docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, errors.New("failed to mkdir docker_logs")
		}
	}
	logPath := filepath.Join(dockerLogDir, fmt.Sprintf("compose_remove_%s.log", time.Now().Format("20060102150405")))
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, errors.New("failed to create compose_remove log")
	}
	result.Log = logPath

	// logger
	logger := utils.NewStepLogger(file)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				global.LOG.Error("ComposeRemove panic recovered: %v", r)
			}
			file.Close()
		}()

		logger.Info("try docker compose down %s", req.Name)
		global.LOG.Info("try docker compose down %s", req.Name)

		composePath := filepath.Join(req.WorkDir, req.Name, "docker-compose.yaml")
		if _, err := os.Stat(composePath); err != nil {
			logger.Error("Compose file %s not found", composePath)
			global.LOG.Error("Compose file %s not found", composePath)
			return
		}
		if stdout, err := down(composePath, true); err != nil {
			logger.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			global.LOG.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			return
		}
		logger.Info("docker compose down %s successful", req.Name)
		global.LOG.Info("docker compose down %s successful", req.Name)

		// 删除工作目录
		logger.Info("try remove work dir %s", req.WorkDir)
		global.LOG.Info("try remove work dir %s", req.WorkDir)
		dir := filepath.Join(req.WorkDir, req.Name)
		err := os.RemoveAll(dir)
		if err != nil {
			logger.Error("failed to remove directory %s: %v", dir, err)
			global.LOG.Error("failed to remove directory %s: %v", dir, err)
			return
		}
	}()

	return &result, nil
}

func (c DockerClient) ComposeOperation(req model.ComposeOperation) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	if utils.CheckIllegal(req.Name, req.Operation, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 初始化日志
	dockerLogDir := filepath.Join(req.WorkDir, req.Name, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, errors.New("failed to mkdir docker_logs")
		}
	}
	logPath := filepath.Join(dockerLogDir, fmt.Sprintf("compose_operation_%s.log", time.Now().Format("20060102150405")))
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, errors.New("failed to create compose_operation log")
	}
	result.Log = logPath

	// logger
	logger := utils.NewStepLogger(file)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				global.LOG.Error("ComposeOperation panic recovered: %v", r)
			}
			file.Close()
		}()

		logger.Info("try docker compose %s %s", req.Operation, req.Name)
		global.LOG.Info("try docker compose %s %s", req.Operation, req.Name)

		composePath := filepath.Join(req.WorkDir, req.Name, "docker-compose.yaml")
		if _, err := os.Stat(composePath); err != nil {
			logger.Error("Failed to load compose file %s", composePath)
			global.LOG.Error("Failed to load compose file %s", composePath)
			return
		}

		switch req.Operation {
		case "start":
			if stdout, err := start(composePath); err != nil {
				logger.Error("docker compose start %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				global.LOG.Error("docker compose start %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				return
			}
		case "stop":
			if stdout, err := stop(composePath); err != nil {
				logger.Error("docker compose stop %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				global.LOG.Error("docker compose stop %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				return
			}
		case "restart":
			if stdout, err := restart(composePath); err != nil {
				logger.Error("docker compose restart %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				global.LOG.Error("docker compose restart %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				return
			}
		case "up":
			if stdout, err := up(composePath); err != nil {
				logger.Error("docker compose up %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				global.LOG.Error("docker compose up %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				return
			}
		case "down":
			if stdout, err := down(composePath, req.RemoveVolumes); err != nil {
				logger.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				global.LOG.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
				return
			}
		default:
			logger.Error("invalid operation %s", req.Operation)
			global.LOG.Error("invalid operation %s", req.Operation)
			return
		}
		logger.Info("docker compose %s %s successful", req.Operation, req.Name)
		global.LOG.Info("docker compose %s %s successful", req.Operation, req.Name)
	}()

	return &result, nil
}

func (c DockerClient) ComposeDetail(req model.ComposeDetailReq) (*model.ComposeDetailRsp, error) {
	var rsp model.ComposeDetailRsp
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &rsp, errors.New(constant.ErrCmdIllegal)
	}

	composePath := filepath.Join(req.WorkDir, req.Name, "docker-compose.yaml")
	if _, err := os.Stat(composePath); err != nil {
		global.LOG.Error("Failed to load compose file %s", composePath)
		return &rsp, fmt.Errorf("load compose file failed, %v", err)
	}
	// 读取docker-compose.yaml
	composeFile, err := os.Open(composePath)
	if err != nil {
		return &rsp, err
	}
	defer composeFile.Close()
	composeBytes, err := io.ReadAll(composeFile)
	if err != nil {
		return &rsp, err
	}
	rsp.ComposeContent = string(composeBytes)
	// 读取.env
	envPath := filepath.Join(req.WorkDir, req.Name, ".env")
	envContent := ""
	if _, err := os.Stat(envPath); err == nil {
		envFile, err := os.Open(envPath)
		if err != nil {
			global.LOG.Error("Failed to open env file %s: %v", envPath, err)
		} else {
			defer envFile.Close()
			envBytes, err := io.ReadAll(envFile)
			if err != nil {
				global.LOG.Error("Failed to read env file %s: %v", envPath, err)
			} else {
				envContent = string(envBytes)
			}
		}
	} else if os.IsNotExist(err) {
		global.LOG.Info("Env file %s not found", envPath)
	} else {
		global.LOG.Error("Failed to stat env file %s: %v", envPath, err)
	}
	rsp.EnvContent = envContent

	return &rsp, nil
}

func (c DockerClient) ComposeUpdate(req model.ComposeUpdate) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 初始化日志
	dockerLogDir := filepath.Join(req.WorkDir, req.Name, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, errors.New("failed to mkdir docker_logs")
		}
	}
	logPath := filepath.Join(dockerLogDir, fmt.Sprintf("compose_update_%s.log", time.Now().Format("20060102150405")))
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, errors.New("failed to create compose_update log")
	}
	result.Log = logPath

	// logger
	logger := utils.NewStepLogger(file)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				global.LOG.Error("ComposeUpdate panic recovered: %v", r)
			}
			file.Close()
		}()

		logger.Info("try docker compose down %s", req.Name)
		global.LOG.Info("try docker compose down %s", req.Name)

		composePath := filepath.Join(req.WorkDir, req.Name, "docker-compose.yaml")
		if _, err := os.Stat(composePath); err != nil {
			logger.Error("Compose file %s not found", composePath)
			global.LOG.Error("Compose file %s not found", composePath)
			return
		}
		if stdout, err := down(composePath, true); err != nil {
			logger.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			global.LOG.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			return
		}
		logger.Info("docker compose down %s successful", req.Name)
		global.LOG.Info("docker compose down %s successful", req.Name)

		// 覆盖docker-compose.yaml
		composeFile, err := os.OpenFile(composePath, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			logger.Error("Failed to open compose file %s: %v", composePath, err)
			global.LOG.Error("Failed to open compose file %s: %v", composePath, err)
			return
		}
		defer composeFile.Close()
		write := bufio.NewWriter(composeFile)
		_, _ = write.WriteString(req.ComposeContent)
		write.Flush()

		// 覆盖.env
		envPath := filepath.Join(req.WorkDir, req.Name, ".env")
		envFile, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			logger.Error("Failed to open env file %s: %v", envPath, err)
			global.LOG.Error("Failed to open env file %s: %v", envPath, err)
			return
		}
		defer envFile.Close()
		write = bufio.NewWriter(envFile)
		_, _ = write.WriteString(req.EnvContent)
		write.Flush()

		logger.Info("config files has been replaced, try docker compose up")
		global.LOG.Info("config files has been replaced, try docker compose up")

		if stdout, err := up(composePath); err != nil {
			logger.Error("docker compose up %s failed, err: %v", req.Name, strings.TrimSpace(stdout), err)
			global.LOG.Error("docker compose up %s failed, stdout: %s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			return
		}

		logger.Info("docker compose up %s successful!", req.Name)
		global.LOG.Info("docker compose up %s successful!", req.Name)
	}()

	return &result, nil
}

func (c DockerClient) ComposeUpgrade(req model.ComposeUpgrade) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 初始化日志
	dockerLogDir := filepath.Join(req.WorkDir, req.Name, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, errors.New("failed to mkdir docker_logs")
		}
	}
	logPath := filepath.Join(dockerLogDir, fmt.Sprintf("compose_upgrade_%s.log", time.Now().Format("20060102150405")))
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, errors.New("failed to create compose_upgrade log")
	}
	result.Log = logPath

	// logger
	logger := utils.NewStepLogger(file)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				global.LOG.Error("ComposeUpgrade panic recovered: %v", r)
			}
			file.Close()
		}()

		logger.Info("try docker compose down %s", req.Name)
		global.LOG.Info("try docker compose down %s", req.Name)

		composePath := filepath.Join(req.WorkDir, req.Name, "docker-compose.yaml")
		envPath := filepath.Join(req.WorkDir, req.Name, ".env")

		// 先停止原compose
		if _, err := os.Stat(composePath); err != nil {
			logger.Error("Compose file %s not found", composePath)
			global.LOG.Error("Compose file %s not found", composePath)
			return
		}
		if stdout, err := down(composePath, true); err != nil {
			logger.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			global.LOG.Error("docker compose down %s failed, out:%s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			return
		}

		// 备份至 /WorkDir/Name/backup/serial/*
		fo := files.NewFileOp()
		backupID := time.Now().Format("20060102T150405")
		backupDir := filepath.Join(req.WorkDir, req.Name, "backup", backupID)
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			logger.Error("Failed to mkdir backup dir %s, %v", backupDir, err)
			global.LOG.Error("Failed to mkdir backup dir %s, %v", backupDir, err)
			return
		}
		// 备份 .env
		backupEnv := filepath.Join(backupDir, ".env")
		fo.Copy(envPath, backupEnv)

		// 备份 docker-compose.yaml
		backupCompose := filepath.Join(backupDir, "docker-compose.yaml")
		fo.Copy(composePath, backupCompose)

		logger.Info("backup compose and env to %s", backupDir)
		global.LOG.Info("backup compose and env to %s", backupDir)

		// 写入conf
		if req.ConfPath != "" && req.ConfContent != "" {
			if err := c.initConf(req.ConfPath, req.ConfContent, true); err != nil {
				logger.Error("Failed to init conf %s, %v", req.ConfPath, err)
				global.LOG.Error("Failed to init conf %s, %v", req.ConfPath, err)
				return
			}
			logger.Info("init conf successful")
			global.LOG.Info("init conf successful")
		}

		// 覆盖docker-compose.yaml
		composeFile, err := os.OpenFile(composePath, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			logger.Error("Failed to open compose file %s: %v", composePath, err)
			global.LOG.Error("Failed to open compose file %s: %v", composePath, err)
			return
		}
		defer composeFile.Close()
		write := bufio.NewWriter(composeFile)
		_, _ = write.WriteString(req.ComposeContent)
		write.Flush()

		// 覆盖.env
		envFile, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0640)
		if err != nil {
			logger.Error("Failed to open env file %s: %v", envPath, err)
			global.LOG.Error("Failed to open env file %s: %v", envPath, err)
			return
		}
		defer envFile.Close()
		write = bufio.NewWriter(envFile)
		_, _ = write.WriteString(req.EnvContent)
		write.Flush()

		logger.Info("config files has been replaced, try docker compose up %s", req.Name)
		global.LOG.Info("config files has been replaced, try docker compose up %s", req.Name)

		if stdout, err := up(composePath); err != nil {
			logger.Error("docker compose up %s failed, err: %v", req.Name, strings.TrimSpace(stdout), err)
			global.LOG.Error("docker compose up %s failed, stdout: %s, err: %v", req.Name, strings.TrimSpace(stdout), err)
			return
		}

		logger.Info("docker compose up %s successful!", req.Name)
		global.LOG.Info("docker compose up %s successful!", req.Name)
	}()

	return &result, nil
}

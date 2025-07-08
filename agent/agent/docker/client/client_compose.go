package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
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
	dir := fmt.Sprintf("%s/%s", req.WorkDir, req.Name)

	// 第一步：判断目录是否已存在，存在就报错
	if _, err := os.Stat(dir); err == nil {
		return "", errors.New("compose already exist")
	} else if !os.IsNotExist(err) {
		// 其他类型的错误（权限等）
		return "", err
	}

	// 第二步：目录不存在时创建
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	composePath := fmt.Sprintf("%s/docker-compose.yaml", dir)
	file, err := os.OpenFile(composePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.ComposeContent)
	write.Flush()

	envPath := fmt.Sprintf("%s/.env", dir)
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
	var (
		result    model.PageResult
		records   []model.ComposeInfo
		BackDatas []model.ComposeInfo
	)

	if utils.CheckIllegal(req.Info, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	// 枚举所有的 container
	list, err := c.cli.ContainerList(
		context.Background(),
		container.ListOptions{All: true},
	)
	if err != nil {
		return &result, err
	}

	// composeProjects, _ := c.listComposeProjects(req.WorkDir) // workDir下实际的项目，暂时不做限制
	composeMap := make(map[string]*model.ComposeInfo)
	for _, container := range list {
		if projectName, workingDir, configFiles, ok := isContainerFromManagedCompose(container.Labels, req.WorkDir); ok {
			containerItem := model.ComposeContainer{
				ContainerID: container.ID,
				Name:        container.Names[0][1:],
				State:       container.State,
				CreateTime:  time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
			}
			confFiles := resolveComposeConfigPaths(workingDir, configFiles)
			if compose, has := composeMap[projectName]; has {
				compose.ContainerNumber++
				compose.Containers = append(compose.Containers, containerItem)
				composeMap[projectName] = compose
				compose.ConfigFiles = mergeComposeConfigPaths(compose.ConfigFiles, confFiles)
			} else {
				composeItem := &model.ComposeInfo{
					ContainerNumber: 1,
					CreatedAt:       time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
					ConfigFiles:     strings.Join(confFiles, ","),
					Workdir:         workingDir,
					Path:            workingDir,
					Containers:      []model.ComposeContainer{containerItem},
				}
				idbType, ok := container.Labels[constant.IDBType]
				if ok {
					composeItem.IdbType = idbType
				}
				// 如果限制了类型
				if req.IdbType != "" && idbType != req.IdbType {
					global.LOG.Info("Container %s type %s, is not %s, ignoring", containerItem.Name, idbType, req.IdbType)
					continue
				}

				idbName, ok := container.Labels[constant.IDBName]
				if ok {
					composeItem.IdbName = idbName
				}
				idbVersion, ok := container.Labels[constant.IDBVersion]
				if ok {
					composeItem.IdbVersion = idbVersion
				}
				idbUpdateVersion, ok := container.Labels[constant.IDBUpdateVersion]
				if ok {
					composeItem.IdbUpdateVersion = idbUpdateVersion
				}
				idbPanel, ok := container.Labels[constant.IDBPanel]
				if ok {
					composeItem.IdbPanel = idbPanel
				}

				composeMap[projectName] = composeItem
			}
		}
	}

	// 遍历计算状态
	for name, compose := range composeMap {
		statusCount := make(map[string]int)
		for _, c := range compose.Containers {
			state := c.State // e.g. "running", "exited"
			statusCount[state]++
		}

		switch len(statusCount) {
		case 0:
			compose.Status = model.ComposeStatusUnknown
		case 1:
			for state := range statusCount {
				switch state {
				case "running":
					compose.Status = model.ComposeStatusRunning
				case "exited":
					compose.Status = model.ComposeStatusExited
				case "paused":
					compose.Status = model.ComposeStatusPaused
				case "restarting":
					compose.Status = model.ComposeStatusRestarting
				case "removing":
					compose.Status = model.ComposeStatusRemoving
				case "dead":
					compose.Status = model.ComposeStatusDead
				default:
					compose.Status = model.ComposeStatusUnknown
				}
			}
		default:
			compose.Status = model.ComposeStatusMixed
		}
		compose.Name = name
		records = append(records, *compose)
	}

	if len(req.Info) != 0 {
		length, count := len(records), 0
		for count < length {
			if !strings.Contains(records[count].Name, req.Info) {
				records = append(records[:count], records[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt > records[j].CreatedAt
	})
	total, start, end := len(records), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		BackDatas = make([]model.ComposeInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		BackDatas = records[start:end]
	}
	result.Total = int64(total)
	result.Items = BackDatas
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

	// 写入docker-compose.yaml和.env
	composePath, err := c.initComposeAndEnv(&req)
	if err != nil {
		return &result, err
	}
	global.LOG.Info("init compose and env successful")

	// 写入conf
	if req.ConfPath != "" && req.ConfContent != "" {
		if err := c.initConf(req.ConfPath, req.ConfContent, false); err != nil {
			global.LOG.Error("Failed to init conf %s, %v", req.ConfPath, err)
			return &result, err
		}
		global.LOG.Info("init conf successful")
	}

	global.LOG.Info("try docker compose up %s", req.Name)

	// 初始化日志
	projectDir := filepath.Dir(composePath)
	dockerLogDir := path.Join(projectDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, err
		}
	}
	logPath := fmt.Sprintf("%s/compose_create_%s_%s.log", dockerLogDir, req.Name, time.Now().Format("20060102150405"))
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, err
	}
	result.Log = logPath

	defer file.Close()
	if stdout, err := up(composePath); err != nil {
		global.LOG.Error("docker compose up %s failed, stdout: %s, err: %v", req.Name, string(stdout), err)
		_, _ = down(composePath, true)
		_, _ = file.WriteString("docker compose up failed!")
		return &result, errors.New(string(stdout))
	}
	global.LOG.Info("docker compose up %s successful!", req.Name)
	_, _ = file.WriteString("docker compose up successful!")

	return &result, nil
}

func (c DockerClient) ComposeRemove(req model.ComposeRemove) error {
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return errors.New(constant.ErrCmdIllegal)
	}
	composePath := fmt.Sprintf("%s/%s/docker-compose.yaml", req.WorkDir, req.Name)
	if _, err := os.Stat(composePath); err != nil {
		global.LOG.Error("Compose file %s not found", composePath)
		return fmt.Errorf("%s not found, %v", composePath, err)
	}
	if stdout, err := down(composePath, true); err != nil {
		return errors.New(string(stdout))
	}
	global.LOG.Info("docker compose down %s successful", req.Name)

	// 删除工作目录
	dir := fmt.Sprintf("%s/%s", req.WorkDir, req.Name)
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %v", dir, err)
	}

	return nil
}

func (c DockerClient) ComposeOperation(req model.ComposeOperation) error {
	if utils.CheckIllegal(req.Name, req.Operation, req.WorkDir) {
		return errors.New(constant.ErrCmdIllegal)
	}
	composePath := fmt.Sprintf("%s/%s/docker-compose.yaml", req.WorkDir, req.Name)
	if _, err := os.Stat(composePath); err != nil {
		global.LOG.Error("Failed to load compose file %s", composePath)
		return fmt.Errorf("load compose file failed, %v", err)
	}
	switch req.Operation {
	case "start":
		if stdout, err := start(composePath); err != nil {
			return errors.New(string(stdout))
		}
	case "stop":
		if stdout, err := stop(composePath); err != nil {
			return errors.New(string(stdout))
		}
	case "restart":
		if stdout, err := restart(composePath); err != nil {
			return errors.New(string(stdout))
		}
	case "up":
		if stdout, err := up(composePath); err != nil {
			return errors.New(string(stdout))
		}
	case "down":
		if stdout, err := down(composePath, req.RemoveVolumes); err != nil {
			return errors.New(string(stdout))
		}
	default:
		return errors.New("invalid operation")
	}
	global.LOG.Info("docker compose %s %s successful", req.Operation, req.Name)
	return nil
}

func (c DockerClient) ComposeDetail(req model.ComposeDetailReq) (*model.ComposeDetailRsp, error) {
	var rsp model.ComposeDetailRsp
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &rsp, errors.New(constant.ErrCmdIllegal)
	}

	composePath := fmt.Sprintf("%s/%s/docker-compose.yaml", req.WorkDir, req.Name)
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
	envPath := fmt.Sprintf("%s/%s/.env", req.WorkDir, req.Name)
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

func (c DockerClient) ComposeUpdate(req model.ComposeUpdate) error {
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return errors.New(constant.ErrCmdIllegal)
	}

	composePath := fmt.Sprintf("%s/%s/docker-compose.yaml", req.WorkDir, req.Name)

	// 先停止原compose
	if stdout, err := down(composePath, true); err != nil {
		return fmt.Errorf("failed to docker compose down %s, err: %s", req.Name, string(stdout))
	}

	// 覆盖docker-compose.yaml
	composeFile, err := os.OpenFile(composePath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer composeFile.Close()
	write := bufio.NewWriter(composeFile)
	_, _ = write.WriteString(req.ComposeContent)
	write.Flush()

	// 覆盖.env
	envPath := fmt.Sprintf("%s/%s/.env", req.WorkDir, req.Name)
	envFile, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer envFile.Close()
	write = bufio.NewWriter(envFile)
	_, _ = write.WriteString(req.EnvContent)
	write.Flush()

	global.LOG.Info("config files has been replaced, try docker compose up")

	if stdout, err := up(composePath); err != nil {
		return fmt.Errorf("docker compose up %s failed: %s", req.Name, string(stdout))
	}

	global.LOG.Info("docker compose up %s successful!", req.Name)
	return nil
}

func (c DockerClient) ComposeUpgrade(req model.ComposeUpgrade) (*model.ComposeCreateResult, error) {
	var result model.ComposeCreateResult
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return &result, errors.New(constant.ErrCmdIllegal)
	}

	composePath := filepath.Join(req.WorkDir, req.Name, "docker-compose.yaml")
	envPath := filepath.Join(req.WorkDir, req.Name, ".env")

	// 先停止原compose
	if stdout, err := down(composePath, true); err != nil {
		return &result, fmt.Errorf("failed to docker compose down %s, err: %s", req.Name, string(stdout))
	}

	// 备份至 /WorkDir/Name/backup/serial/*
	fo := files.NewFileOp()
	backupID := time.Now().Format("20060102T150405")
	backupDir := filepath.Join(req.WorkDir, req.Name, "backup", backupID)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, err
	}
	// 备份 .env
	backupEnv := filepath.Join(backupDir, ".env")
	fo.CopyFile(envPath, backupEnv)

	// 备份 docker-compose.yaml
	backupCompose := filepath.Join(backupDir, "docker-compose.yaml")
	fo.CopyFile(composePath, backupCompose)
	global.LOG.Info("backup compose and env to %s", backupDir)

	// 写入conf
	if req.ConfPath != "" && req.ConfContent != "" {
		if err := c.initConf(req.ConfPath, req.ConfContent, true); err != nil {
			global.LOG.Error("Failed to init conf %s, %v", req.ConfPath, err)
			return &result, err
		}
		global.LOG.Info("init conf successful")
	}

	// 覆盖docker-compose.yaml
	composeFile, err := os.OpenFile(composePath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return &result, err
	}
	defer composeFile.Close()
	write := bufio.NewWriter(composeFile)
	_, _ = write.WriteString(req.ComposeContent)
	write.Flush()

	// 覆盖.env
	envFile, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return &result, err
	}
	defer envFile.Close()
	write = bufio.NewWriter(envFile)
	_, _ = write.WriteString(req.EnvContent)
	write.Flush()

	global.LOG.Info("config files has been replaced, try docker compose up %s", req.Name)

	// 初始化日志
	projectDir := filepath.Dir(composePath)
	dockerLogDir := path.Join(projectDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return &result, err
		}
	}
	logPath := fmt.Sprintf("%s/compose_upgrade_%s_%s.log", dockerLogDir, req.Name, backupID)
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return &result, err
	}
	result.Log = logPath

	defer file.Close()
	if stdout, err := up(composePath); err != nil {
		global.LOG.Error("docker compose up %s failed, stdout: %s, err: %v", req.Name, string(stdout), err)
		_, _ = down(composePath, true)
		_, _ = file.WriteString("docker compose up failed!")
		return &result, errors.New(string(stdout))
	}
	global.LOG.Info("docker compose up %s successful!", req.Name)
	_, _ = file.WriteString("docker compose up successful!")

	return &result, nil
}

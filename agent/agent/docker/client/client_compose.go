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
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return "", err
		}
	}

	composePath := fmt.Sprintf("%s/compose.yml", dir)
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

func (c DockerClient) initConf(req *model.ComposeCreate) error {
	dir := filepath.Dir(req.ConfPath)
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(req.ConfPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.ConfContent)
	write.Flush()

	return nil
}

func pull(filePath string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s pull", filePath)
	return stdout, err
}

func up(filePath string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s up -d", filePath)
	return stdout, err
}

func down(filePath string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s down --remove-orphans", filePath)
	return stdout, err
}

func start(filePath string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s start", filePath)
	return stdout, err
}

func stop(filePath string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s stop", filePath)
	return stdout, err
}

func restart(filePath string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s restart", filePath)
	return stdout, err
}

func operate(filePath, operation string) (string, error) {
	stdout, err := utils.Execf("docker-compose -f %s %s", filePath, operation)
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
	composeMap := make(map[string]model.ComposeInfo)
	for _, container := range list {
		if projectName, workingDir, configFiles, ok := isContainerFromManagedCompose(container.Labels, req.WorkDir); ok {
			containerItem := model.ComposeContainer{
				ContainerID: container.ID,
				Name:        container.Names[0][1:],
				State:       container.State,
				CreateTime:  time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
			}
			if compose, has := composeMap[projectName]; has {
				compose.ContainerNumber++
				compose.Containers = append(compose.Containers, containerItem)
				composeMap[projectName] = compose
			} else {
				composeItem := model.ComposeInfo{
					ContainerNumber: 1,
					CreatedAt:       time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
					ConfigFile:      configFiles,
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
	for key, value := range composeMap {
		value.Name = key
		records = append(records, value)
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
	// 写入compose.yml和.env
	composePath, err := c.initComposeAndEnv(&req)
	if err != nil {
		result.Error = "failed to init compose and env"
		return &result, err
	}
	cmd := exec.Command("docker-compose", "-f", composePath, "config")
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

	// 写入compose.yml和.env
	composePath, err := c.initComposeAndEnv(&req)
	if err != nil {
		return &result, err
	}
	global.LOG.Info("init compose and env successful")

	// 写入conf
	if req.ConfPath != "" && req.ConfContent != "" {
		if err := c.initConf(&req); err != nil {
			global.LOG.Error("Failed to init conf %s, %v", req.ConfPath, err)
			return &result, err
		}
		global.LOG.Info("init conf successful")
	}

	global.LOG.Info("try docker-compose up %s", req.Name)

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
	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	multiWriter := io.MultiWriter(os.Stdout, file)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter
	if err := cmd.Run(); err != nil {
		global.LOG.Error("docker-compose up %s failed, err: %v", req.Name, err)
		_, _ = down(composePath)
		_, _ = file.WriteString("docker-compose up failed!")
		return &result, err
	}
	global.LOG.Info("docker-compose up %s successful!", req.Name)
	_, _ = file.WriteString("docker-compose up successful!")

	return &result, nil
}

func (c DockerClient) ComposeRemove(req model.ComposeRemove) error {
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return errors.New(constant.ErrCmdIllegal)
	}
	composePath := fmt.Sprintf("%s/%s/compose.yml", req.WorkDir, req.Name)
	if _, err := os.Stat(composePath); err != nil {
		global.LOG.Error("Failed to load compose file %s", composePath)
		return fmt.Errorf("load compose file failed, %v", err)
	}
	if stdout, err := down(composePath); err != nil {
		return errors.New(string(stdout))
	}
	global.LOG.Info("docker-compose down %s successful", req.Name)

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
	composePath := fmt.Sprintf("%s/%s/compose.yml", req.WorkDir, req.Name)
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
		if stdout, err := down(composePath); err != nil {
			return errors.New(string(stdout))
		}
	default:
		return errors.New("invalid operation")
	}
	global.LOG.Info("docker-compose %s %s successful", req.Operation, req.Name)
	return nil
}

func (c DockerClient) ComposeUpdate(req model.ComposeUpdate) error {
	if utils.CheckIllegal(req.Name, req.WorkDir) {
		return errors.New(constant.ErrCmdIllegal)
	}

	composePath := fmt.Sprintf("%s/%s/compose.yml", req.WorkDir, req.Name)

	// 先停止原compose
	if stdout, err := down(composePath); err != nil {
		return fmt.Errorf("failed to docker-compose down %s, err: %s", req.Name, string(stdout))
	}

	// 覆盖compose.yml
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

	global.LOG.Info("config files has been replaced, try docker-compose up")

	if stdout, err := up(composePath); err != nil {
		return fmt.Errorf("docker-compose up %s failed: %s", req.Name, string(stdout))
	}

	global.LOG.Info("docker-compose up %s successful!", req.Name)
	return nil
}

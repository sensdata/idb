package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"syscall"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type ContainerService struct{}

type IContainerService interface {
	Inspect(req model.Inspect) (string, error)
	Prune(req model.Prune) (*model.PruneResult, error)

	ContainerQuery(req model.QueryContainer) (*model.PageResult, error)
	ContainerNames() ([]string, error)
	ContainerCreate(req model.ContainerOperate) error
	ContainerUpdate(req model.ContainerOperate) error
	ContainerUpgrade(req model.ContainerUpgrade) error
	ContainerInfo(containerID string) (*model.ContainerOperate, error)
	ContainerResourceUsage() ([]model.ContainerResourceUsage, error)
	ContainerResourceLimit() (*model.ContainerResourceLimit, error)
	ContainerStats(id string) (*model.ContainerStats, error)
	ContainerRename(req model.Rename) error
	ContainerLogClean(containerID string) error
	ContainerOperation(req model.ContainerOperation) error
	ContainerLogs(wsConn *websocket.Conn, containerType, container, since, tail string, follow bool) error

	// ComposePage(req model.SearchPageInfo) (int64, interface{}, error)
	// ComposeCreate(req model.ComposeCreate) (string, error)
	// ComposeOperation(req model.ComposeOperation) error
	// ComposeTest(req model.ComposeCreate) (bool, error)
	// ComposeUpdate(req model.ComposeUpdate) error

	ImagePage(req model.SearchPageInfo) (*model.PageResult, error)
	ImageList() ([]model.Options, error)
	ImageBuild(req model.ImageBuild) (string, error)
	ImagePull(req model.ImagePull) (string, error)
	ImageLoad(req model.ImageLoad) error
	ImageSave(req model.ImageSave) error
	ImagePush(req model.ImagePush) (string, error)
	ImageRemove(req model.BatchDelete) error
	ImageTag(req model.ImageTag) error

	VolumePage(req model.SearchPageInfo) (*model.PageResult, error)
	VolumeList() ([]model.Options, error)
	VolumeDelete(req model.BatchDelete) error
	VolumeCreate(req model.VolumeCreate) error

	NetworkPage(req model.SearchPageInfo) (*model.PageResult, error)
	NetworkList() ([]model.Options, error)
	NetworkDelete(req model.BatchDelete) error
	NetworkCreate(req model.NetworkCreate) error
}

func NewIContainerService() IContainerService {
	return &ContainerService{}
}

func (u *ContainerService) Inspect(req model.Inspect) (string, error) {
	client, err := client.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()

	inspectInfo, err := client.Inspect(req)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(inspectInfo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *ContainerService) Prune(req model.Prune) (*model.PruneResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PruneResult{}, err
	}
	defer client.Close()
	return client.Prune(req)
}

func (u *ContainerService) ContainerQuery(req model.QueryContainer) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ContainerQuery(req)
}

func (u *ContainerService) ContainerNames() ([]string, error) {
	var names []string
	client, err := client.NewClient()
	if err != nil {
		return names, err
	}
	defer client.Close()
	return client.ContainerNames()
}

func (u *ContainerService) ContainerCreate(req model.ContainerOperate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerCreate(req, global.LOG)
}

func (u *ContainerService) ContainerUpdate(req model.ContainerOperate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerUpdate(req, global.LOG)
}

func (u *ContainerService) ContainerUpgrade(req model.ContainerUpgrade) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerUpgrade(req, global.LOG)
}

func (u *ContainerService) ContainerInfo(containerID string) (*model.ContainerOperate, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ContainerInfo(containerID)
}

func (u *ContainerService) ContainerResourceUsage() ([]model.ContainerResourceUsage, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ContainerResourceUsage()
}

func (u *ContainerService) ContainerResourceLimit() (*model.ContainerResourceLimit, error) {
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("load cpu limit failed, err: %v", err)
	}
	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("load memory limit failed, err: %v", err)
	}

	data := model.ContainerResourceLimit{
		CPU:    cpuCounts,
		Memory: memoryInfo.Total,
	}
	return &data, nil
}

func (u *ContainerService) ContainerStats(id string) (*model.ContainerStats, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ContainerStats(id)
}

func (u *ContainerService) ContainerRename(req model.Rename) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerRename(req)
}

func (u *ContainerService) ContainerLogClean(containerID string) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerLogClean(containerID)
}

func (u *ContainerService) ContainerOperation(req model.ContainerOperation) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerOperation(req)
}

func (u *ContainerService) ContainerLogs(wsConn *websocket.Conn, containerType, container, since, tail string, follow bool) error {
	defer func() { wsConn.Close() }()
	if utils.CheckIllegal(container, since, tail) {
		return errors.New(constant.ErrCmdIllegal)
	}
	commandName := "docker"
	commandArg := []string{"logs", container}
	if containerType == "compose" {
		commandName = "docker-compose"
		commandArg = []string{"-f", container, "logs"}
	}
	if tail != "0" {
		commandArg = append(commandArg, "--tail")
		commandArg = append(commandArg, tail)
	}
	if since != "all" {
		commandArg = append(commandArg, "--since")
		commandArg = append(commandArg, since)
	}
	if follow {
		commandArg = append(commandArg, "-f")
	}
	if !follow {
		cmd := exec.Command(commandName, commandArg...)
		cmd.Stderr = cmd.Stdout
		stdout, _ := cmd.CombinedOutput()
		if !utf8.Valid(stdout) {
			return errors.New("invalid utf8")
		}
		if err := wsConn.WriteMessage(websocket.TextMessage, stdout); err != nil {
			global.LOG.Error("send message with log to ws failed, err: %v", err)
		}
		return nil
	}

	cmd := exec.Command(commandName, commandArg...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return err
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		return err
	}
	exitCh := make(chan struct{})
	go func() {
		_, wsData, _ := wsConn.ReadMessage()
		if string(wsData) == "close conn" {
			_ = cmd.Process.Signal(syscall.SIGTERM)
			exitCh <- struct{}{}
		}
	}()

	go func() {
		buffer := make([]byte, 1024)
		for {
			select {
			case <-exitCh:
				return
			default:
				n, err := stdout.Read(buffer)
				if err != nil {
					if err == io.EOF {
						return
					}
					global.LOG.Error("read bytes from log failed, err: %v", err)
					return
				}
				if !utf8.Valid(buffer[:n]) {
					continue
				}
				if err = wsConn.WriteMessage(websocket.TextMessage, buffer[:n]); err != nil {
					global.LOG.Error("send message with log to ws failed, err: %v", err)
					return
				}
			}
		}
	}()
	_ = cmd.Wait()
	return nil
}

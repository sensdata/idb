package docker

import (
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

func (s *DockerService) ContainerQuery(req model.QueryContainer) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ContainerQuery(req)
}

func (s *DockerService) ContainerNames() (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ContainerNames()
}

func (s *DockerService) ContainerCreate(req model.ContainerOperate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerCreate(req, global.LOG)
}

func (s *DockerService) ContainerUpdate(req model.ContainerOperate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerUpdate(req, global.LOG)
}

func (s *DockerService) ContainerUpgrade(req model.ContainerUpgrade) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerUpgrade(req, global.LOG)
}

func (s *DockerService) ContainerInfo(containerID string) (*model.ContainerOperate, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ContainerInfo(containerID)
}

func (s *DockerService) ContainerResourceUsage() (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ContainerResourceUsage()
}

func (s *DockerService) ContainerResourceLimit() (*model.ContainerResourceLimit, error) {
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

func (s *DockerService) ContainerStats(id string) (*model.ContainerStats, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ContainerStats(id)
}

func (s *DockerService) ContainerRename(req model.Rename) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerRename(req)
}

func (s *DockerService) ContainerLogClean(containerID string) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerLogClean(containerID)
}

func (s *DockerService) ContainerOperation(req model.ContainerOperation) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ContainerOperation(req)
}

func (s *DockerService) ContainerLogs(wsConn *websocket.Conn, containerType, container, since, tail string, follow bool) error {
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

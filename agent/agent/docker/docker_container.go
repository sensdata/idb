package docker

import (
	"fmt"

	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/model"

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

func (s *DockerService) ContainerLogs(req model.FileContentPartReq) (*model.FileContentPartRsp, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ContainerLogs(req)
}

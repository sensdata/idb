package docker

import (
	"github.com/sensdata/idb/core/model"
)

/**
 * {"Command":"\"docker-entrypoint.s…\"","CreatedAt":"2024-11-19 11:47:04 +0800 CST","ID":"6d198be1a9e7","Image":"redis:latest","Labels":"","LocalVolumes":"1","Mounts":"028f6808b833c4…","Names":"redis","Networks":"bridge","Ports":"0.0.0.0:3309-\u003e3309/tcp, :::3309-\u003e3309/tcp, 6379/tcp","RunningFor":"8 hours ago","Size":"0B","State":"running","Status":"Up 8 hours"}
 */
func (s *DockerMan) getContainers(hostID uint64) (*model.PageResult, error) {
	var result model.PageResult
	var items []model.ContainerInfo = []model.ContainerInfo{}

	result.Items = items
	result.Total = (int64(len(items)))

	return &result, nil
}

func (s *DockerMan) createContainer(hostID uint64, req model.ContainerOperate) error {
	return nil
}

func (s *DockerMan) pruneContainer(hostID uint64) error {
	return nil
}

func (s *DockerMan) getContainer(hostID uint64, containerID string) (string, error) {
	return "", nil
}

func (s *DockerMan) updateContainer(hostID uint64, containerID string, req model.ContainerOperate) error {
	return nil
}

func (s *DockerMan) deleteContainer(hostID uint64, containerID string) error {
	return nil
}

// func (s *DockerMan) getContainerLog(hostID uint64, containerID string) (*model.ContainerLog, error) {
// 	return nil, nil
// }

func (s *DockerMan) getContainerStatus(hostID uint64, containerID string) (*model.ContainerStats, error) {
	return nil, nil
}

func (s *DockerMan) upgradeContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) startContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) stopContainer(hostID uint64, containerID string, force bool) error {
	return nil
}

func (s *DockerMan) rebootContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) pauseContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) resumeContainer(hostID uint64, containerID string) error {
	return nil
}

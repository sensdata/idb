package docker

import "github.com/sensdata/idb/core/model"

func (s *DockerMan) getContainers(hostID uint64) (*model.PageResult, error) {
	return nil, nil
}

func (s *DockerMan) createContainer(hostID uint64, req model.CreateContainer) error {
	return nil
}

func (s *DockerMan) pruneContainer(hostID uint64) error {
	return nil
}

func (s *DockerMan) getContainer(hostID uint64, containerID string) (*model.Container, error) {
	return nil, nil
}

func (s *DockerMan) updateContainer(hostID uint64, containerID string, req model.UpdateContainer) error {
	return nil
}

func (s *DockerMan) deleteContainer(hostID uint64, containerID string) error {
	return nil
}

func (s *DockerMan) getContainerLog(hostID uint64, containerID string) (*model.ContainerLog, error) {
	return nil, nil
}

func (s *DockerMan) getContainerStatus(hostID uint64, containerID string) (*model.ContainerStatus, error) {
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

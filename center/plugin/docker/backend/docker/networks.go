package docker

import "github.com/sensdata/idb/core/model"

func (s *DockerMan) getNetworks(hostID uint64) (*model.PageResult, error) {
	return nil, nil
}

func (s *DockerMan) createNetwork(hostID uint64, req model.NetworkCreate) error {
	return nil
}

func (s *DockerMan) pruneNetwork(hostID uint64) error {
	return nil
}

func (s *DockerMan) getNetwork(hostID uint64, networkID string) (*model.Network, error) {
	return nil, nil
}

func (s *DockerMan) deleteNetwork(hostID uint64, networkID string) error {
	return nil
}

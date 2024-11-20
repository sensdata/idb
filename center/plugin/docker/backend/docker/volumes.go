package docker

import "github.com/sensdata/idb/core/model"

func (s *DockerMan) getVolumes(hostID uint64) (*model.PageResult, error) {
	return nil, nil
}

func (s *DockerMan) createVolume(hostID uint64, req model.VolumeCreate) error {
	return nil
}

func (s *DockerMan) pruneVolume(hostID uint64) error {
	return nil
}

func (s *DockerMan) getVolume(hostID uint64, volumeID string) (*model.Volume, error) {
	return nil, nil
}

func (s *DockerMan) deleteVolume(hostID uint64, volumeID string) error {
	return nil
}

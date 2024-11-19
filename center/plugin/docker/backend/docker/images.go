package docker

import "github.com/sensdata/idb/core/model"

func (s *DockerMan) getImages(hostID uint64) (*model.PageResult, error) {
	return nil, nil
}

func (s *DockerMan) pruneImage(hostID uint64, req model.PruneImage) error {
	return nil
}

func (s *DockerMan) pullImage(hostID uint64, req model.PullImage) error {
	return nil
}

func (s *DockerMan) importImage(hostID uint64, req model.ImportImage) error {
	return nil
}

func (s *DockerMan) buildImage(hostID uint64, req model.BuildImage) error {
	return nil
}

func (s *DockerMan) cleanBuild(hostID uint64) error {
	return nil
}

func (s *DockerMan) getImage(hostID uint64, imageID string) (*model.Image, error) {
	return nil, nil
}

func (s *DockerMan) pushImage(hostID uint64, imageID string, req model.PushImage) error {
	return nil
}

func (s *DockerMan) exportImage(hostID uint64, imageID string, req model.ExportImage) error {
	return nil
}

func (s *DockerMan) setImageTag(hostID uint64, imageID string, req model.SetImageTag) error {
	return nil
}

func (s *DockerMan) deleteImage(hostID uint64, imageID string) error {
	return nil
}

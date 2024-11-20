package docker

import "github.com/sensdata/idb/core/model"

func (s *DockerMan) getImages(hostID uint64) (*model.PageResult, error) {
	return nil, nil
}

func (s *DockerMan) pruneImage(hostID uint64, req model.Prune) error {
	return nil
}

func (s *DockerMan) pullImage(hostID uint64, req model.ImagePull) error {
	return nil
}

func (s *DockerMan) importImage(hostID uint64, req model.ImageLoad) error {
	return nil
}

func (s *DockerMan) buildImage(hostID uint64, req model.ImageBuild) error {
	return nil
}

func (s *DockerMan) cleanBuild(hostID uint64) error {
	return nil
}

func (s *DockerMan) getImage(hostID uint64, imageID string) (*model.Image, error) {
	return nil, nil
}

func (s *DockerMan) pushImage(hostID uint64, imageID string, req model.ImagePush) error {
	return nil
}

func (s *DockerMan) exportImage(hostID uint64, imageID string, req model.ImageSave) error {
	return nil
}

func (s *DockerMan) setImageTag(hostID uint64, imageID string, req model.ImageTag) error {
	return nil
}

func (s *DockerMan) deleteImage(hostID uint64, imageID string) error {
	return nil
}

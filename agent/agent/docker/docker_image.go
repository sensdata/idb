package docker

import (
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func (s *DockerService) ImagePage(req model.SearchPageInfo) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ImagePage(req)
}

func (s *DockerService) ImageList() (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ImageList()
}

func (s *DockerService) ImageBuild(req model.ImageBuild) (*model.ImageOperationResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ImageOperationResult{}, err
	}
	defer client.Close()
	return client.ImageBuild(req, constant.AgentDataDir, constant.AgentLogDir, global.LOG)
}

func (s *DockerService) ImagePull(req model.ImagePull) (*model.ImageOperationResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ImageOperationResult{}, err
	}
	defer client.Close()
	return client.ImagePull(req, constant.AgentLogDir, global.LOG)
}

func (s *DockerService) ImageLoad(req model.ImageLoad) (*model.ImageOperationResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ImageOperationResult{}, err
	}
	defer client.Close()
	return client.ImageLoad(req)
}

func (s *DockerService) ImageSave(req model.ImageSave) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageSave(req)
}

func (s *DockerService) ImageTag(req model.ImageTag) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageTag(req)
}

func (s *DockerService) ImagePush(req model.ImagePush) (*model.ImageOperationResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ImageOperationResult{}, err
	}
	defer client.Close()
	return client.ImagePush(req, constant.AgentLogDir, global.LOG)
}

func (s *DockerService) ImageRemove(req model.BatchDelete) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageRemove(req)
}

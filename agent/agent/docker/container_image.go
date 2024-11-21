package docker

import (
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
)

func (u *ContainerService) ImagePage(req model.SearchPageInfo) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ImagePage(req)
}

func (u *ContainerService) ImageList() ([]model.Options, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.ImageList()
}

func (u *ContainerService) ImageBuild(req model.ImageBuild) (string, error) {
	client, err := client.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	return client.ImageBuild(req, constant.AgentDataDir, constant.AgentLogDir, global.LOG)
}

func (u *ContainerService) ImagePull(req model.ImagePull) (string, error) {
	client, err := client.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	return client.ImagePull(req, constant.AgentLogDir, global.LOG)
}

func (u *ContainerService) ImageLoad(req model.ImageLoad) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageLoad(req)
}

func (u *ContainerService) ImageSave(req model.ImageSave) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageSave(req)
}

func (u *ContainerService) ImageTag(req model.ImageTag) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageTag(req)
}

func (u *ContainerService) ImagePush(req model.ImagePush) (string, error) {
	client, err := client.NewClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	return client.ImagePush(req, constant.AgentLogDir, global.LOG)
}

func (u *ContainerService) ImageRemove(req model.BatchDelete) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ImageRemove(req)
}

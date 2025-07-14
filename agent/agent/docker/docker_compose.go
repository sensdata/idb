package docker

import (
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/core/model"
)

func (s *DockerService) ComposePage(req model.QueryCompose) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.ComposePage(req)
}

func (s *DockerService) ComposeTest(req model.ComposeCreate) (*model.ComposeTestResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeTestResult{}, err
	}
	defer client.Close()
	return client.ComposeTest(req)
}

func (s *DockerService) ComposeCreate(req model.ComposeCreate) (*model.ComposeCreateResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeCreateResult{}, err
	}
	defer client.Close()
	return client.ComposeCreate(req)
}

func (s *DockerService) ComposeRemove(req model.ComposeRemove) (*model.ComposeCreateResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeCreateResult{}, err
	}
	defer client.Close()
	return client.ComposeRemove(req)
}

func (s *DockerService) ComposeOperation(req model.ComposeOperation) (*model.ComposeCreateResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeCreateResult{}, err
	}
	defer client.Close()
	return client.ComposeOperation(req)
}

func (s *DockerService) ComposeDetail(req model.ComposeDetailReq) (*model.ComposeDetailRsp, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeDetailRsp{}, err
	}
	defer client.Close()
	return client.ComposeDetail(req)
}

func (s *DockerService) ComposeUpdate(req model.ComposeUpdate) (*model.ComposeCreateResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeCreateResult{}, err
	}
	defer client.Close()
	return client.ComposeUpdate(req)
}

func (s *DockerService) ComposeUpgrade(req model.ComposeUpgrade) (*model.ComposeCreateResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.ComposeCreateResult{}, err
	}
	defer client.Close()
	return client.ComposeUpgrade(req)
}

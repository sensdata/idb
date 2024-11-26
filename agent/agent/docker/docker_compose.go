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

func (s *DockerService) ComposeOperation(req model.ComposeOperation) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ComposeOperation(req)
}

func (s *DockerService) ComposeUpdate(req model.ComposeUpdate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.ComposeUpdate(req)
}

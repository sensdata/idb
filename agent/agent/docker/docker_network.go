package docker

import (
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/core/model"
)

func (s *DockerService) NetworkPage(req model.SearchPageInfo) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.NetworkPage(req)
}

func (s *DockerService) NetworkList() (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.NetworkList()
}

func (s *DockerService) NetworkDelete(req model.BatchDelete) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.NetworkDelete(req)
}

func (s *DockerService) NetworkCreate(req model.NetworkCreate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.NetworkCreate(req)
}

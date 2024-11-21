package docker

import (
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/core/model"
)

func (u *ContainerService) NetworkPage(req model.SearchPageInfo) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.NetworkPage(req)
}

func (u *ContainerService) NetworkList() ([]model.Options, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.NetworkList()
}

func (u *ContainerService) NetworkDelete(req model.BatchDelete) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.NetworkDelete(req)
}

func (u *ContainerService) NetworkCreate(req model.NetworkCreate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.NetworkCreate(req)
}

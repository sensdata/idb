package docker

import (
	"github.com/sensdata/idb/agent/agent/docker/client"
	"github.com/sensdata/idb/core/model"
)

func (u *ContainerService) VolumePage(req model.SearchPageInfo) (*model.PageResult, error) {
	client, err := client.NewClient()
	if err != nil {
		return &model.PageResult{}, err
	}
	defer client.Close()
	return client.VolumePage(req)
}

func (u *ContainerService) VolumeList() ([]model.Options, error) {
	client, err := client.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return client.VolumeList()
}

func (u *ContainerService) VolumeDelete(req model.BatchDelete) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.VolumeDelete(req)
}

func (u *ContainerService) VolumeCreate(req model.VolumeCreate) error {
	client, err := client.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return client.VolumeCreate(req)
}

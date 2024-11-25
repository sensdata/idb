package docker

import (
	"fmt"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

func (s *DockerMan) getImages(hostID uint64, req model.SearchPageInfo) (*model.PageResult, error) {
	var result model.PageResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Page,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query images")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to images result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) imageNames(hostID uint64) (*model.PageResult, error) {
	var result model.PageResult

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_List,
			Data:   "",
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to query image names")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to image names result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) buildImage(hostID uint64, req model.ImageBuild) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Build,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to build image")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to build image result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) pullImage(hostID uint64, req model.ImagePull) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Pull,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to pull image")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to pull image result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) loadImage(hostID uint64, req model.ImageLoad) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Load,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to load image")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to load image result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) exportImage(hostID uint64, req model.ImageSave) error {

	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Save,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to save image")
	}

	return nil
}

func (s *DockerMan) pushImage(hostID uint64, req model.ImagePush) (*model.ImageOperationResult, error) {
	var result model.ImageOperationResult

	data, err := utils.ToJSONString(req)
	if err != nil {
		return &result, err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Push,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return &result, err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return &result, fmt.Errorf("failed to push image")
	}

	err = utils.FromJSONString(actionResponse.Data.Action.Data, &result)
	if err != nil {
		global.LOG.Error("Error unmarshaling data to push image result: %v", err)
		return &result, fmt.Errorf("json err: %v", err)
	}

	return &result, nil
}

func (s *DockerMan) deleteImage(hostID uint64, req model.BatchDelete) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Remove,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to remove images")
	}

	return nil
}

func (s *DockerMan) setImageTag(hostID uint64, req model.ImageTag) error {
	data, err := utils.ToJSONString(req)
	if err != nil {
		return err
	}

	actionRequest := model.HostAction{
		HostID: uint(hostID),
		Action: model.Action{
			Action: model.Docker_Image_Tag,
			Data:   data,
		},
	}

	actionResponse, err := s.sendAction(actionRequest)
	if err != nil {
		return err
	}

	if !actionResponse.Data.Action.Result {
		global.LOG.Error("action failed")
		return fmt.Errorf("failed to set image tag")
	}

	return nil
}

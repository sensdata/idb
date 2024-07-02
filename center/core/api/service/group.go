package service

import (
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/constant"
	"github.com/sensdata/idb/center/core/api/dto"
	"github.com/sensdata/idb/center/db/model"
)

type GroupService struct{}

type IGroupService interface {
	List(req dto.PageInfo) (*dto.PageResult, error)
	Create(req dto.CreateGroup) (*dto.GroupInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
}

func NewIGroupService() IGroupService {
	return &GroupService{}
}

// List group
func (s *GroupService) List(req dto.PageInfo) (*dto.PageResult, error) {
	total, groups, err := GroupRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &dto.PageResult{Total: total, Items: groups}, nil
}

// Create group
func (s *GroupService) Create(req dto.CreateGroup) (*dto.GroupInfo, error) {
	var group model.Group
	if err := copier.Copy(&group, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	if err := GroupRepo.Create(&group); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}
	var groupInfo dto.GroupInfo
	if err := copier.Copy(&groupInfo, &group); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}
	return &groupInfo, nil
}

func (s *GroupService) Update(id uint, upMap map[string]interface{}) error {
	return GroupRepo.Update(id, upMap)
}

func (s *GroupService) Delete(ids []uint) error {
	return GroupRepo.Delete(CommonRepo.WithIdsIn(ids))
}

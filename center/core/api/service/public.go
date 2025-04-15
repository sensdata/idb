package service

import (
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/model"
)

type PublicService struct{}

type IPublicService interface {
	Version() (*model.About, error)
}

func NewIPublicService() IPublicService {
	return &PublicService{}
}

func (s *PublicService) Version() (*model.About, error) {
	var about model.About

	about.Version = global.Version

	return &about, nil
}

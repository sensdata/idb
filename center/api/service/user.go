package service

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

type UserService struct{}

type IUserService interface {
	List(c *gin.Context, info dto.PageInfo) (*dto.PageResult, error)
}

func NewIUserService() IUserService {
	return &UserService{}
}

// List user
func (s *UserService) List(c *gin.Context, info dto.PageInfo) (*dto.PageResult, error) {
	users, err := UserRepo.GetList()
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &dto.PageResult{Items: users}, nil
}

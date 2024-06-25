package service

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/center/api/dto"
	"github.com/sensdata/idb/center/constant"
)

type UserService struct{}

type IUserService interface {
	List(c *gin.Context, info dto.ListUser) (*dto.ListUserResult, error)
}

func NewIUserService() IUserService {
	return &UserService{}
}

// List user
func (s *UserService) List(c *gin.Context, info dto.ListUser) (*dto.ListUserResult, error) {
	users, err := UserRepo.GetList()
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &dto.ListUserResult{Users: users}, nil
}

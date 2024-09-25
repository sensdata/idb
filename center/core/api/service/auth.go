package service

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type AuthService struct{}

type IAuthService interface {
	Login(c *gin.Context, info model.Login) (*model.LoginResult, error)
	LogOut(c *gin.Context) error
}

func NewIAuthService() IAuthService {
	return &AuthService{}
}

// LogOut implements IAuthService.
func (s *AuthService) LogOut(c *gin.Context) error {
	return nil
}

// Login implements IAuthService.
func (s *AuthService) Login(c *gin.Context, info model.Login) (*model.LoginResult, error) {
	user, err := UserRepo.Get(UserRepo.WithByName(info.Name))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrInvalidAccountOrPassword, err.Error())
	}

	if !utils.ValidatePassword(user.Password, info.Password, user.Salt) {
		return nil, errors.WithMessage(constant.ErrInvalidAccountOrPassword, constant.ErrInvalidAccountOrPassword.Error())
	}

	// TODO: jwt key需要初始化
	key := "abcd:2024:qwer"
	token, err := utils.GenerateJWT(user.ID, user.Username, 3600, key)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	return &model.LoginResult{Name: user.Username, Token: token}, nil
}

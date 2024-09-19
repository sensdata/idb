package service

import (
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/sensdata/idb/core/constant"

	"github.com/sensdata/idb/center/db/model"
	core "github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

type UserService struct{}

type IUserService interface {
	List(req core.PageInfo) (*core.PageResult, error)
	Create(req core.CreateUser) (*core.UserInfo, error)
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
	ChangePassword(req core.ChangePassword) error
}

func NewIUserService() IUserService {
	return &UserService{}
}

// List user
func (s *UserService) List(req core.PageInfo) (*core.PageResult, error) {
	total, users, err := UserRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return nil, errors.WithMessage(constant.ErrNoRecords, err.Error())
	}

	return &core.PageResult{Total: total, Items: users}, nil
}

// Create user
func (s *UserService) Create(req core.CreateUser) (*core.UserInfo, error) {
	var user model.User
	if err := copier.Copy(&user, &req); err != nil {
		return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	//找角色
	role, err := RoleRepo.Get(RoleRepo.WithByName("user"))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	//找组
	group, err := GroupRepo.Get(GroupRepo.WithByID(req.GroupID))
	if err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	//密码加盐
	salt := utils.GenerateNonce(8)
	passwordHash := utils.HashPassword(user.Password, salt)

	user.Salt = salt
	user.Password = passwordHash
	user.RoleID = role.ID
	user.GroupID = group.ID
	user.Valid = 1

	if err := UserRepo.Create(&user); err != nil {
		return nil, errors.WithMessage(constant.ErrInternalServer, err.Error())
	}

	return &core.UserInfo{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UserName:  user.Username,
		RoleInfo:  core.RoleInfo{ID: role.ID, RoleName: role.Name, CreatedAt: role.CreatedAt},
		GroupInfo: core.GroupInfo{ID: group.ID, GroupName: group.GroupName, CreatedAt: group.CreatedAt},
		Valid:     user.Valid,
	}, nil
}

func (s *UserService) Update(id uint, upMap map[string]interface{}) error {
	return UserRepo.Update(id, upMap)
}

func (s *UserService) Delete(ids []uint) error {
	return UserRepo.Delete(CommonRepo.WithIdsIn(ids))
}

func (s *UserService) ChangePassword(req core.ChangePassword) error {
	//找用户
	user, err := UserRepo.Get(UserRepo.WithByID(req.UserID))
	if err != nil {
		return errors.WithMessage(constant.ErrRecordNotFound, err.Error())
	}

	passwordHash := utils.HashPassword(req.Password, user.Salt)
	upMap := make(map[string]interface{})
	upMap["password"] = passwordHash

	return UserRepo.Update(user.ID, upMap)
}

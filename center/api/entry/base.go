package entry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/db"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/utils"
)

// Login handles user login and returns a JWT token
func (s *BaseApi) Login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user db.User
	if err := global.DB.Where("username = ?", credentials.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	if !utils.ValidatePassword(user.Password, credentials.Password, user.Salt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// TODO: jwt key需要初始化
	key := "abcd:2024:qwer"
	token, err := utils.GenerateJWT(user.ID, user.Username, 3600, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *BaseApi) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/utils"
)

type JWT struct{}

func NewJWT() *JWT {
	return &JWT{}
}

// to check JWT tokens
func (j *JWT) JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string
		var err error
		token = c.GetHeader("Authorization")
		if token == "" {
			// not in headers, check cookies
			token, err = c.Cookie("idb-token")
			if err != nil || token == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
				c.Abort()
				return
			}
		}

		claims, err := utils.ValidateJWT(token, global.JWTKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}

// Additional middlewares such as logging, error handling can be defined here

package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/db/repo"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/helper"
)

func BindDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewSettingsRepo()
		status, err := settingRepo.Get(settingRepo.WithByKey("BindDomain"))
		if err != nil {
			global.LOG.Error("Failed to get bind domain: %v", err)
			helper.ErrorWithDetail(c, constant.CodeFailed, err.Error(), nil)
			return
		}
		if len(status.Value) == 0 {
			c.Next()
			return
		}
		domains := c.Request.Host
		parts := strings.Split(c.Request.Host, ":")
		if len(parts) > 0 {
			domains = parts[0]
		}

		if domains != status.Value {
			global.LOG.Error("domain not allowed, domain: %s, bind domain: %s", domains, status.Value)
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrInternalServer.Error(), errors.New("domain not allowed"))
			return
		}
		c.Next()
	}
}

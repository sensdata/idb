package helper

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/core/constant"
	"github.com/sensdata/idb/core/model"
	"github.com/sensdata/idb/core/utils"
)

var Validator = utils.InitValidator()

func CheckQueryAndValidate(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindQuery(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrInvalidParams.Error(), err)
		return err
	}
	if err := Validator.Struct(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrInvalidParams.Error(), err)
		return err
	}
	return nil
}

func CheckQuery(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindQuery(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrInvalidParams.Error(), err)
		return err
	}
	return nil
}

func CheckBindAndValidate(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrInvalidParams.Error(), err)
		return err
	}
	if err := Validator.Struct(req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrInvalidParams.Error(), err)
		return err
	}
	return nil
}

func CheckBind(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, constant.CodeErrBadRequest, constant.ErrInvalidParams.Error(), err)
		return err
	}
	return nil
}

func ErrorWithDetail(ctx *gin.Context, code int, msg string, err error) {
	res := model.Response{
		Code:    code,
		Message: msg,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithData(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := model.Response{
		Code: constant.CodeSuccess,
		Data: data,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithOutData(ctx *gin.Context) {
	res := model.Response{
		Code:    constant.CodeSuccess,
		Message: "success",
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func SuccessWithMsg(ctx *gin.Context, msg string) {
	res := model.Response{
		Code:    constant.CodeSuccess,
		Message: msg,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

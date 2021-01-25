package ginutil

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/library/bcode"
)

// Result represents a response for restful api.
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// JSON -
func JSON(c *gin.Context, data interface{}, err error) {
	bc := bcode.Err2Coder(err)
	if bc == bcode.ServerErr {
		logrus.Errorf("server error: %v", err)
	}
	result := &Result{
		Code: bc.Code(),
		Msg:  bc.Error(),
		Data: data,
	}
	c.AbortWithStatusJSON(bc.Status(), result)
}

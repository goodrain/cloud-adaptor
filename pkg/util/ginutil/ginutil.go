package ginutil

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"goodrain.com/cloud-adaptor/pkg/bcode"
)

// Result represents a response for restful api.
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
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
	}
	if bc.Status() >= 200 && bc.Status() < 300 {
		result.Data = data
	}

	c.AbortWithStatusJSON(bc.Status(), result)
}

// Error -
func Error(c *gin.Context, err error) {
	bc := bcode.Err2Coder(err)
	if bc == bcode.ServerErr {
		logrus.Errorf("server error: %v", err)
	}
	result := &Result{
		Code: bc.Code(),
		Msg:  bc.Error(),
	}

	c.AbortWithStatusJSON(bc.Status(), result)
}

// ShouldBindJSON -
func ShouldBindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return bcode.NewBadRequest(err.Error())
	}
	return nil
}

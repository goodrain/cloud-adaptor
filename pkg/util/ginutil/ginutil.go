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

// JSONv2 -
func JSONv2(c *gin.Context, data interface{}, errs ...error) {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	bc := bcode.Err2Coder(err)
	if bc == bcode.ServerErr {
		logrus.Errorf("server error: %v", err)
	}
	if bc.Status() >= 200 && bc.Status() < 300 {
		c.AbortWithStatusJSON(bc.Status(), data)
		return
	}

	c.AbortWithStatusJSON(bc.Status(), &Result{
		Code: bc.Code(),
		Msg:  bc.Error(),
	})
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

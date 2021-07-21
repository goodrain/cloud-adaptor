// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
func JSON(c *gin.Context, data interface{}, errs ...error) {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	}
	bc := bcode.Err2Coder(err)
	if bc == bcode.ServerErr {
		logrus.Errorf("server error: %+v", err)
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
		logrus.Errorf("server error: %+v", err)
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

// MustGet -
func MustGet(c *gin.Context, key string) interface{} {
	obj, _ := c.Get(key)
	return obj
}

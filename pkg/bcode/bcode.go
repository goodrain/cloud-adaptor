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

package bcode

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Coder has ability to get code, msg or detail from error.
type Coder interface {
	// status code
	Status() int
	// business code
	Code() int
	Error() string
}

var (
	codes = make(map[int]struct{})
)

func new(status, code int) Coder {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("bcode %d already exists", code))
	}
	codes[code] = struct{}{}
	return newCode(status, code, "")
}

func newByMessage(status, code int, message string) Coder {
	if _, ok := codes[code]; ok {
		panic(fmt.Sprintf("bcode %d already exists", code))
	}
	codes[code] = struct{}{}
	return newCode(status, code, message)
}

// Code business a bussiness code
type Code struct {
	status, code int
	message      string
}

func newCode(status, code int, message string) Coder {
	return &Code{status: status, code: code, message: message}
}

// Status returns the status code
func (c Code) Status() int {
	return c.status
}

// Code returns the business code
func (c Code) Code() int {
	return c.code
}

func (c Code) Error() string {
	if c.message != "" {
		return c.message
	}
	return strconv.FormatInt(int64(c.code), 10)
}

// Err2Coder converts the given err to Coder.
func Err2Coder(err error) Coder {
	if err == nil {
		return OK
	}
	coder, ok := errors.Cause(err).(Coder)
	if ok {
		return coder
	}
	return Str2Coder(err.Error())
}

// Str2Coder converts the given str to Coder.
func Str2Coder(str string) Coder {
	str = strings.TrimSpace(str)
	if str == "" {
		return OK
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return ServerErr
	}
	return newCode(400, i, "")
}

// NewBadRequest creates a bad request error.
func NewBadRequest(msg string) error {
	return newCode(400, 400, msg)
}

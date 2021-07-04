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
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		registerValidation(v, "appStoreName", appStoreName)
	}
}

func registerValidation(v *validator.Validate, tag string, fn validator.Func, callValidationEvenIfNull ...bool) {
	if err := v.RegisterValidation(tag, fn, callValidationEvenIfNull...); err != nil {
		panic(fmt.Sprintf("register %s validation: %v", tag, err))
	}
}

var appStoreName validator.Func = func(fl validator.FieldLevel) bool {
	return validateAppStoreName(fl.Field().String())
}

func validateAppStoreName(name string) bool {
	return regexp.MustCompile("^[a-z][a-z0-9]{3,31}$").MatchString(name)
}

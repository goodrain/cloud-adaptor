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

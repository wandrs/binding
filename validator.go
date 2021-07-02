package binding

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

const (
	usernameRegexString = `^([a-zA-Z]+([a-zA-Z0-9]*[._-]{1}[a-zA-Z0-9]+)*)$`
)

var (
	usernameRegex = regexp.MustCompile(usernameRegexString)
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("username", userNameValidator); err != nil {
		panic(err)
	}
}

func userNameValidator(fl validator.FieldLevel) bool {
	return usernameRegex.MatchString(fl.Field().String())
}

func check(val reflect.Value) error {
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		return validate.Struct(val.Interface())
	} else if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := validate.Struct(val.Index(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

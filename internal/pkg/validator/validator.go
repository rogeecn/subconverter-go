package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/subconverter/subconverter-go/internal/pkg/errors"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// 自定义验证器
	validate.RegisterValidation("url", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
	})
	
	validate.RegisterValidation("proxy_type", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		validTypes := map[string]bool{
			"ss": true, "ssr": true, "vmess": true, "vless": true,
			"trojan": true, "hysteria": true, "hysteria2": true,
			"snell": true, "http": true, "https": true, "socks5": true,
		}
		return validTypes[value]
	})
}

func Validate(i interface{}) error {
	err := validate.Struct(i)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				return errors.ValidationError(
					strings.ToLower(e.Field()),
					getErrorMessage(e),
				)
			}
		}
		return errors.BadRequest("VALIDATION_ERROR", err.Error())
	}
	return nil
}

func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "url":
		return fe.Field() + " must be a valid URL"
	case "proxy_type":
		return fe.Field() + " must be a valid proxy type"
	case "min":
		return fe.Field() + " must be greater than " + fe.Param()
	case "max":
		return fe.Field() + " must be less than " + fe.Param()
	case "oneof":
		return fe.Field() + " must be one of " + fe.Param()
	default:
		return fe.Field() + " is invalid"
	}
}

// ValidateProxy validates proxy configuration
func ValidateProxy(proxy interface{}) error {
	return Validate(proxy)
}

// ValidateSubscription validates subscription configuration
func ValidateSubscription(subscription interface{}) error {
	return Validate(subscription)
}

// ValidateConvertRequest validates convert request
func ValidateConvertRequest(request interface{}) error {
	return Validate(request)
}
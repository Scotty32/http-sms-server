package validators

import (
	"context"
	"fmt"
	"net/url"

	"github.com/NdoleStudio/httpsms/pkg/requests"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/thedevsaddam/govalidator"
)

// AuthHandlerValidator validates models used in handlers.AuthHandler
type AuthHandlerValidator struct {
	validator
	logger telemetry.Logger
	tracer telemetry.Tracer
}

// NewAuthHandlerValidator creates a new AuthHandlerValidator
func NewAuthHandlerValidator(logger telemetry.Logger, tracer telemetry.Tracer) (v *AuthHandlerValidator) {
	return &AuthHandlerValidator{
		logger: logger.WithService(fmt.Sprintf("%T", v)),
		tracer: tracer,
	}
}

// ValidateLogin validates requests.AuthLoginRequest
func (validator *AuthHandlerValidator) ValidateLogin(_ context.Context, request requests.AuthLoginRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"email":    []string{"required", "email"},
			"password": []string{"required", "min:6"},
		},
	})
	return v.ValidateStruct()
}

// ValidateRegister validates requests.AuthRegisterRequest
func (validator *AuthHandlerValidator) ValidateRegister(_ context.Context, request requests.AuthRegisterRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"name":     []string{"required", "min:1", "max:100"},
			"email":    []string{"required", "email"},
			"password": []string{"required", "min:6", "max:72"},
		},
	})
	return v.ValidateStruct()
}

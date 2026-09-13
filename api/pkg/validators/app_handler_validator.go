package validators

import (
	"context"
	"fmt"
	"net/url"

	"github.com/NdoleStudio/httpsms/pkg/requests"
	"github.com/NdoleStudio/httpsms/pkg/telemetry"
	"github.com/thedevsaddam/govalidator"
)

// AppHandlerValidator validates models used in handlers.AppHandler
type AppHandlerValidator struct {
	validator
	logger telemetry.Logger
	tracer telemetry.Tracer
}

// NewAppHandlerValidator creates a new AppHandlerValidator
func NewAppHandlerValidator(
	logger telemetry.Logger,
	tracer telemetry.Tracer,
) (v *AppHandlerValidator) {
	return &AppHandlerValidator{
		logger: logger.WithService(fmt.Sprintf("%T", v)),
		tracer: tracer,
	}
}

// ValidateStore validates requests.AppStoreRequest
func (validator *AppHandlerValidator) ValidateStore(_ context.Context, request requests.AppStoreRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"name": []string{
				"required",
				"min:1",
				"max:100",
			},
			"webhook_url": []string{
				"url",
			},
		},
	})
	return v.ValidateStruct()
}

// ValidateIndex validates requests.AppIndex
func (validator *AppHandlerValidator) ValidateIndex(_ context.Context, request requests.AppIndex) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"limit": []string{
				"required",
				"numeric",
				"min:1",
				"max:100",
			},
			"skip": []string{
				"required",
				"numeric",
				"min:0",
			},
		},
	})
	return v.ValidateStruct()
}

// ValidatePhoneNumber validates requests.AppPhoneNumberRequest
func (validator *AppHandlerValidator) ValidatePhoneNumber(_ context.Context, request requests.AppPhoneNumberRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"phone_number": []string{
				"required",
				"min:7",
				"max:20",
			},
		},
	})
	return v.ValidateStruct()
}

// ValidateSendMessage validates requests.AppSendMessageRequest
func (validator *AppHandlerValidator) ValidateSendMessage(_ context.Context, request requests.AppSendMessageRequest) url.Values {
	v := govalidator.New(govalidator.Options{
		Data: &request,
		Rules: govalidator.MapData{
			"to": []string{
				"required",
				"min:7",
				"max:20",
				phoneNumberRule,
			},
			"content": []string{
				"required",
				"min:1",
				"max:1000",
			},
		},
	})
	return v.ValidateStruct()
}

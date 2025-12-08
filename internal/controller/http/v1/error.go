package v1

import (
	"net/http"

	"github.com/evrone/go-clean-template/internal/controller/http/v1/response"
	"github.com/evrone/go-clean-template/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

func errorResponse(ctx *fiber.Ctx, err error) error {
	appErr, ok := apperror.AsAppError(err)
	if !ok {
		return ctx.Status(http.StatusInternalServerError).JSON(response.Error{
			Code:    apperror.KindInternal.String(),
			Message: "An unexpected error occurred",
		})
	}

	status := kindToHTTPStatus(appErr.Kind())

	return ctx.Status(status).JSON(response.Error{
		Code:    appErr.Code(),
		Message: appErr.Message(),
		Details: appErr.Fields(),
	})
}

func kindToHTTPStatus(kind apperror.Kind) int {
	switch kind {
	case apperror.KindUnknown, apperror.KindInternal:
		return http.StatusInternalServerError
	case apperror.KindValidation:
		return http.StatusBadRequest
	case apperror.KindNotFound:
		return http.StatusNotFound
	case apperror.KindConflict:
		return http.StatusConflict
	case apperror.KindUnauthorized:
		return http.StatusUnauthorized
	case apperror.KindForbidden:
		return http.StatusForbidden
	case apperror.KindTimeout:
		return http.StatusGatewayTimeout
	case apperror.KindExternal:
		return http.StatusBadGateway
	}

	return http.StatusInternalServerError
}

func validationError(ctx *fiber.Ctx, msg string) error {
	return errorResponse(ctx, apperror.Validation(msg))
}

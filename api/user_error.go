package api

import (
	"net/http"

	"github.com/andrebq/auth/internal/usererror"
)

func InternalError() error {
	return usererror.E{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
	}
}

func UnauthorizedError(msg string) error {
	return usererror.E{
		Status:  http.StatusUnauthorized,
		Message: msg,
	}
}

func BadRequestError(msg string) error {
	return usererror.E{
		Status:  http.StatusBadRequest,
		Message: msg,
	}
}

func NotImplementedError() error {
	return usererror.E{
		Status:  http.StatusInternalServerError,
		Message: "not implemented",
	}
}

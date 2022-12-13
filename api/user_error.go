package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	userError struct {
		Status  int
		Message string
	}
)

func InternalError() error {
	return userError{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
	}
}

func UnauthorizedError(msg string) error {
	return userError{
		Status:  http.StatusUnauthorized,
		Message: msg,
	}
}

func BadRequestError(msg string) error {
	return userError{
		Status:  http.StatusBadRequest,
		Message: msg,
	}
}

func NotImplementedError() error {
	return userError{
		Status:  http.StatusInternalServerError,
		Message: "not implemented",
	}
}

func (ue userError) HTTPStatus() int {
	return ue.Status
}

func (ue userError) Error() string {
	return fmt.Sprintf("[%v]: %v", ue.Status, ue.Message)
}

func (ue *userError) Marshal() ([]byte, error) {
	val := struct {
		Status int    `json:"status"`
		Title  string `json:"title"`
	}{
		Status: ue.Status,
		Title:  ue.Message,
	}
	if val.Status == 0 {
		val.Status = http.StatusBadRequest
	}
	return json.Marshal(val)
}

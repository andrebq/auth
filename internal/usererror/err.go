package usererror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	E struct {
		Status  int
		Message string
	}
)

func (ue E) HTTPStatus() int {
	return ue.Status
}

func (ue E) Error() string {
	return fmt.Sprintf("[%v]: %v", ue.Status, ue.Message)
}

func (ue E) MarshalJSON() ([]byte, error) {
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

func (ue *E) UnmarshalJSON(buf []byte) error {
	println(string(buf))
	val := struct {
		Status int    `json:"status"`
		Title  string `json:"title"`
	}{}
	err := json.Unmarshal(buf, &val)
	if err != nil {
		return err
	}
	ue.Message = val.Title
	ue.Status = val.Status
	return nil
}

func (ue *E) Failure() bool {
	return ue != nil &&
		*ue != E{}
}

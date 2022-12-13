package api

import (
	"encoding/json"
	"time"
)

type (
	apiDuration time.Duration
)

func (a apiDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(a).String())
}

func (a *apiDuration) UnmarshalJSON(buf []byte) error {
	var str string
	err := json.Unmarshal(buf, &str)
	if err != nil {
		return err
	}
	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	*a = apiDuration(dur)
	return nil
}

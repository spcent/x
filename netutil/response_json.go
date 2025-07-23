package netutil

import (
	"encoding/json"
	"errors"
	"log"
)

// JSONResponse
type JSONResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// JSONResponse
func NewJSONResponse(data any) *JSONResponse {
	return &JSONResponse{Code: 0, Msg: "ok", Data: data}
}

func BindJSONResponse(data []byte) *JSONResponse {
	jr := new(JSONResponse)
	if err := json.Unmarshal(data, jr); err != nil {
		log.Panicf("%s:%s", err, data)
	}

	return jr
}

func (e *JSONResponse) Error() error {
	if e.Code == 0 {
		return nil
	}

	return errors.New(e.Msg)
}

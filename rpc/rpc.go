package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Handler[Request any, Response any] func(ctx context.Context, req *Request) (*Response, error)

var (
	ErrBadRequest = func(err error) Error { return newError(http.StatusBadRequest, err) }
	ErrNotFound   = func(err error) Error { return newError(http.StatusNotFound, err) }
	ErrInternal   = func(err error) Error { return newError(http.StatusInternalServerError, err) }
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

func newError(code int, err error) Error {
	return Error{
		Code:    code,
		Message: err.Error(),
	}
}

func API[Request any, Response any](handler Handler[Request, Response]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)

		if r.Method == "OPTIONS" {
			return
		}

		var req Request
		if r.Body != http.NoBody {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		}

		resp, err := handler(r.Context(), &req)
		if err != nil {
			var rpcErr Error
			if errors.As(err, &rpcErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(rpcErr.Code)
				_ = json.NewEncoder(w).Encode(rpcErr)

				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	})
}

func allowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)
}
